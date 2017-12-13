package cmd

import (
	"encoding/json"

	"fmt"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	kalpha "k8s.io/client-go/pkg/apis/settings/v1alpha1"
)

type IntegrationCmd struct {
	scClient sc.Interface
	k8Client kubernetes.Interface
}

func NewIntegrationCmd(scClient sc.Interface, k8Client kubernetes.Interface) *IntegrationCmd {
	return &IntegrationCmd{scClient: scClient, k8Client: k8Client}
}

func createBindingObject(bindingName, instance string, params map[string]interface{}, secretName string) (*v1beta1.ServiceBinding, error) {
	pdata, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	b := &v1beta1.ServiceBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: bindingName,
		},
		Spec: v1beta1.ServiceBindingSpec{
			ServiceInstanceRef: v1beta1.LocalObjectReference{Name: instance},
			Parameters:         &runtime.RawExtension{Raw: pdata},
			SecretName:         secretName,
		},
	}
	return b, nil
}

func podPreset(objectName, secretName, producerSvcName, consumerSvcName string) *kalpha.PodPreset {
	podPreset := kalpha.PodPreset{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: objectName,
			Labels: map[string]string{
				"group":   "mobile",
				"service": producerSvcName,
			},
		},
		Spec: kalpha.PodPresetSpec{
			Selector: meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"run":           consumerSvcName,
					producerSvcName: "enabled",
				},
			},
			Volumes: []v1.Volume{
				{
					Name: producerSvcName,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      producerSvcName,
					MountPath: "/etc/secrets/" + producerSvcName,
				},
			},
		},
	}
	return &podPreset
	//if _, err := sc.k8client.SettingsV1alpha1().PodPresets(namespace).Create(&podPreset); err != nil {
	//	return errors.Wrap(err, "failed to create pod preset for service ")
	//}

}

func (bc *IntegrationCmd) CreateIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration <consuming_service_instance_id> <providing_service_instance_id>",
		Short: "integrate certain mobile services together",
		Long: `example usage: kubectl plugin mobile create integration <consuming_service_instance_id> <providing_service_instance_id>
mobile --namespace=myproject create integration <consuming_service_instance_id> <providing_service_instance_id>
oc plugin mobile create integration <consuming_service_instance_id> <providing_service_instance_id>
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing arguments: " + cmd.Use)
			}
			namespace := currentNamespace(cmd.Flags())
			consumerSvcInstName := args[0]
			providerSvcInstName := args[1]
			providerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(providerSvcInstName, meta_v1.GetOptions{})
			if err != nil {
				return err
			}
			consumerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(consumerSvcInstName, meta_v1.GetOptions{})
			if err != nil {
				return err
			}
			consumerSvc := getService(namespace, consumerSvcInst.Labels["serviceName"], bc.k8Client) // the consumer service
			providerSvc := getService(namespace, providerSvcInst.Labels["serviceName"], bc.k8Client) // the provider service
			bindParams := buildBindParams(providerSvc, consumerSvc)
			objectName := objectName(consumerSvcInstName, providerSvcInstName)
			binding, err := createBindingObject(objectName, providerSvcInst.Name, bindParams, objectName)
			if err != nil {
				return err
			}
			if _, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Create(binding); err != nil {
				return err
			}
			preset := podPreset(objectName, objectName, providerSvc.Name, consumerSvc.Name)
			if _, err := bc.k8Client.SettingsV1alpha1().PodPresets(namespace).Create(preset); err != nil {
				return errors.Wrap(err, "failed to create pod preset for service ")
			}
			redeploy, err := cmd.PersistentFlags().GetBool("auto-redeploy")
			if err != nil {
				return err
			}
			if !redeploy {
				fmt.Println("you will need to redeploy your service/pod to pick up the changes")
				return nil
			}
			//update the deployment with an annotation
			dep, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Get(consumerSvc.Name, meta_v1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get deployment for service "+consumerSvcInstName)
			}
			dep.Spec.Template.Labels[providerSvc.Name] = "enabled"
			if _, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
				return errors.Wrap(err, "failed up update deployment for "+consumerSvcInstName)
			}

			return nil
		},
	}
	cmd.PersistentFlags().Bool("auto-redeploy", false, "--auto-redeploy=true will cause a backing deployment to be rolled out")
	return cmd
}

func objectName(consumer, provider string) string {
	return consumer + "-" + provider
}

func (bc *IntegrationCmd) DeleteIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration <consuming_service_instance_id> <providing_service_instance_id>",
		Short: "delete the integration between mobile services.",
		Long: `example usage: kubectl plugin mobile delete integration <consuming_service_instance_id> <providing_service_instance_id>
mobile --namespace=myproject delete integration <consuming_service_instance_id> <providing_service_instance_id>
oc plugin mobile delete integration <consuming_service_instance_id> <providing_service_instance_id>
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing arguments.")
			}
			namespace := currentNamespace(cmd.Flags())
			consumerSvcInstName := args[0]
			providerSvcInstName := args[1]

			consumerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(consumerSvcInstName, meta_v1.GetOptions{})
			if err != nil {
				return err
			}
			providerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(providerSvcInstName, meta_v1.GetOptions{})
			if err != nil {
				return err
			}
			consumerSvcName := consumerSvcInst.Labels["serviceName"]
			providerSvcName := providerSvcInst.Labels["serviceName"]
			objectName := objectName(consumerSvcInstName, providerSvcInstName)
			if err := bc.k8Client.SettingsV1alpha1().PodPresets(namespace).Delete(objectName, meta_v1.NewDeleteOptions(0)); err != nil {
				return err
			}
			if err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Delete(objectName, meta_v1.NewDeleteOptions(0)); err != nil {
				return err
			}
			redeploy, err := cmd.PersistentFlags().GetBool("auto-redeploy")
			if err != nil {
				return err
			}
			if !redeploy {
				fmt.Println("you will need to redeploy your service to pick up the changes")
				return nil
			}
			//update the deployment with an annotation
			dep, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Get(consumerSvcName, meta_v1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get deployment for service "+consumerSvcInstName)
			}
			delete(dep.Spec.Template.Labels, providerSvcName)
			if _, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
				return errors.Wrap(err, "failed up update deployment for "+consumerSvcInstName)
			}
			return nil
		},
	}
	cmd.PersistentFlags().Bool("auto-redeploy", false, "--auto-redeploy=true will cause a backing deployment to be rolled out")
	return cmd
}

func (bc *IntegrationCmd) GetIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "get a single integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (bc *IntegrationCmd) ListBindingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integrations",
		Short: "get a list of the current integrations between services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func buildBindParams(from *Service, to *Service) map[string]interface{} {
	var p = map[string]interface{}{}
	p["credentials"] = map[string]string{
		"route":          from.Host,
		"service_secret": to.ID,
	}

	for k, v := range from.Params {
		p[k] = v
	}
	if from.Name == ServiceNameThreeScale {
		p["apicast_route"] = from.Params["apicast_route"]
		p["apicast_token"] = from.Params["token"]
		p["apicast_service_id"] = from.Params["service_id"]
		p["service_route"] = to.Host
		p["service_name"] = to.Name
		p["app_key"] = uuid.NewV4().String()
		p["service_secret"] = to.ID
	} else if from.Name == ServiceNameKeycloak {
		p["service_name"] = to.Name
	}
	return p
}
