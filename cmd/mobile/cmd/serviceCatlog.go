package cmd

import (
	"encoding/json"

	"github.com/feedhenry/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/feedhenry/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	v1alpha1 "k8s.io/client-go/pkg/apis/settings/v1alpha1"
)

// TODO CHANGE THIS NOW THAT WE HAVE A CLIENT
type serviceCatalogClient struct {
	k8host    string
	token     string
	namespace string
	k8client  kubernetes.Interface
	scClient  sc.Interface
}

//TODO this is fragile and should be changed to use the real types and client https://github.com/kubernetes-incubator/service-catalog/issues/1367
func createBindingObject(instance string, params map[string]interface{}, secretName string) (*v1beta1.ServiceBinding, error) {
	pdata, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	b := &v1beta1.ServiceBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			GenerateName: instance + "-",
		},
		Spec: v1beta1.ServiceBindingSpec{
			ServiceInstanceRef: v1beta1.LocalObjectReference{Name: instance},
			Parameters:         &runtime.RawExtension{Raw: pdata},
			SecretName:         secretName,
		},
	}
	return b, nil
}

func (sc *serviceCatalogClient) podPreset(objectName, secretName, svcName, targetSvcName, namespace string) error {
	podPreset := v1alpha1.PodPreset{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: objectName,
			Labels: map[string]string{
				"group":   "mobile",
				"service": svcName,
			},
		},
		Spec: v1alpha1.PodPresetSpec{
			Selector: meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"run":   targetSvcName,
					svcName: "enabled",
				},
			},
			Volumes: []v1.Volume{
				{
					Name: svcName,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      svcName,
					MountPath: "/etc/secrets/" + svcName,
				},
			},
		},
	}
	if _, err := sc.k8client.SettingsV1alpha1().PodPresets(namespace).Create(&podPreset); err != nil {
		return errors.Wrap(err, "failed to create pod preset for service ")
	}
	return nil
}

// BindToService will create a binding and pod preset
// finds the service class based on the service name
// finds the first service instances
// creates a binding via service catalog which kicks of the bind apb for the service
// finally creates a pod preset for sync pods to pick up as a volume mount
func (sc *serviceCatalogClient) BindToService(bindableService, targetSvcName string, params map[string]interface{}, bindableSvcNamespace, targetSvcNamespace string) error {
	objectName := bindableService + "-" + targetSvcName
	bindableServiceClass, err := sc.serviceClassByServiceName(bindableService, sc.token)
	if err != nil {
		return err
	}
	if nil == bindableServiceClass {
		return errors.New("failed to find service class for service " + bindableService)
	}

	svcInstList, err := sc.serviceInstancesForServiceClass(sc.token, bindableServiceClass.Spec.ExternalName, targetSvcNamespace)
	if err != nil {
		return err
	}
	if len(svcInstList.Items) == 0 {
		return errors.New("no service instance of " + bindableService + " found in ns " + targetSvcNamespace)
	}

	// only care about the first one as there only should ever be one.
	svcInst := svcInstList.Items[0]
	binding, _ := createBindingObject(svcInst.Name, params, objectName)
	bindingResp, err := sc.scClient.ServicecatalogV1beta1().ServiceBindings(targetSvcNamespace).Create(binding)
	if err := sc.podPreset(objectName, objectName, bindableService, targetSvcName, targetSvcNamespace); err != nil {
		return errors.Wrap(err, "failed to get pod preset")
	}
	//update the deployment with an annotation
	dep, err := sc.k8client.AppsV1beta1().Deployments(targetSvcNamespace).Get(targetSvcName, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get deployment for service "+targetSvcName)
	}

	dep.Spec.Template.Labels[bindableService] = "enabled"
	dep.Spec.Template.Labels[bindableService+"-binding"] = bindingResp.Name
	if _, err := sc.k8client.AppsV1beta1().Deployments(targetSvcNamespace).Update(dep); err != nil {
		return errors.Wrap(err, "failed up update deployment for "+targetSvcName)
	}
	return nil
}

//UnBindFromService will Delete the binding, the pod preset and the update the deployment
func (sc *serviceCatalogClient) UnBindFromService(bindableService, targetSvcName, targetSvcNamespace string) error {
	objectName := bindableService + "-" + targetSvcName
	dep, err := sc.k8client.AppsV1beta1().Deployments(targetSvcNamespace).Get(targetSvcName, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get deployment for service "+targetSvcName)
	}
	bindingID, ok := dep.Spec.Template.Labels[bindableService+"-binding"]
	if !ok {
		return errors.New("no binding id found for service " + targetSvcName)
	}
	delete(dep.Spec.Template.Labels, bindableService+"-binding")
	delete(dep.Spec.Template.Labels, bindableService)

	if err := sc.scClient.ServicecatalogV1beta1().ServiceBindings(targetSvcNamespace).Delete(bindingID, &meta_v1.DeleteOptions{}); err != nil {
		return err
	}
	// binding deleted we will remove the pod preset and update deployment
	if err := sc.k8client.SettingsV1alpha1().PodPresets(targetSvcNamespace).Delete(objectName, meta_v1.NewDeleteOptions(0)); err != nil {
		return errors.Wrap(err, "unbinding "+bindableService+" and "+targetSvcName+" failed to delete pod preset")
	}
	if _, err := sc.k8client.AppsV1beta1().Deployments(targetSvcNamespace).Update(dep); err != nil {
		return errors.Wrap(err, "failed to update the deployment for "+targetSvcName+" after unbinding "+bindableService)
	}
	return nil
}

// create pod preset with apikeys secret, update deployment with label
func (sc *serviceCatalogClient) AddMobileApiKeys(targetSvcName, namespace string) error {
	objectName := IntegrationAPIKeys + "-" + targetSvcName
	if err := sc.podPreset(objectName, IntegrationAPIKeys, IntegrationAPIKeys, targetSvcName, namespace); err != nil {
		return errors.Wrap(err, "")
	}
	dep, err := sc.k8client.AppsV1beta1().Deployments(namespace).Get(targetSvcName, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get deployment for service "+targetSvcName+" cannot redeploy.")
	}
	dep.Spec.Template.Labels[IntegrationAPIKeys] = "enabled"
	if _, err := sc.k8client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
		return errors.Wrap(err, "failed up update deployment for "+targetSvcName)
	}
	return nil
}

// create pod preset with apikeys secret, update deployment with label
func (sc *serviceCatalogClient) RemoveMobileApiKeys(targetSvcName, namespace string) error {
	objectName := IntegrationAPIKeys + "-" + targetSvcName
	if err := sc.k8client.SettingsV1alpha1().PodPresets(namespace).Delete(objectName, meta_v1.NewDeleteOptions(0)); err != nil {
		return errors.Wrap(err, "removing api keys failed to delete pod preset")
	}
	dep, err := sc.k8client.AppsV1beta1().Deployments(namespace).Get(targetSvcName, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get deployment for service "+targetSvcName+" cannot redeploy.")
	}

	delete(dep.Spec.Template.Labels, IntegrationAPIKeys)
	if _, err := sc.k8client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
		return errors.Wrap(err, "failed up update deployment for "+targetSvcName)
	}
	return nil
}

func (sc *serviceCatalogClient) serviceClassByServiceName(name, token string) (*v1beta1.ClusterServiceClass, error) {
	serviceClasses, err := sc.scClient.ServicecatalogV1beta1().ClusterServiceClasses().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, sc := range serviceClasses.Items {
		meta := map[string]string{}
		json.Unmarshal(sc.Spec.ExternalMetadata.Raw, &meta)
		if v, ok := meta["serviceName"]; ok && v == name {
			return &sc, nil
		}
	}
	return nil, nil
}

func (sc *serviceCatalogClient) serviceInstancesForServiceClass(token, serviceClass string, ns string) (*v1beta1.ServiceInstanceList, error) {
	si, err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	sl := &v1beta1.ServiceInstanceList{}
	for _, i := range si.Items {
		if i.Spec.ClusterServiceClassExternalName == serviceClass {
			sl.Items = append(sl.Items, i)
		}
	}
	return sl, nil
}
