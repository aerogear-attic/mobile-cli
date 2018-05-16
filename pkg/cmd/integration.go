// Copyright Red Hat, Inc., and individual contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"

	"io"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/aerogear/mobile-cli/pkg/cmd/output"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	kalpha "k8s.io/client-go/pkg/apis/settings/v1alpha1"
)

type IntegrationCmd struct {
	*BaseCmd
	scClient sc.Interface
	k8Client kubernetes.Interface
}

func NewIntegrationCmd(scClient sc.Interface, k8Client kubernetes.Interface, out io.Writer) *IntegrationCmd {
	return &IntegrationCmd{scClient: scClient, k8Client: k8Client, BaseCmd: &BaseCmd{Out: output.NewRenderer(out)}}
}

func createBindingObject(consumer, provider, bindingName, instance string, bindParams *ServiceParams, secretName string) (*v1beta1.ServiceBinding, error) {
	parsedBindParams := buildBindParams(bindParams)
	pdata, err := json.Marshal(parsedBindParams)

	if err != nil {
		return nil, err
	}
	b := &v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        bindingName,
			Annotations: map[string]string{"consumer": consumer, "provider": provider},
		},
		Spec: v1beta1.ServiceBindingSpec{
			ServiceInstanceRef: v1beta1.LocalObjectReference{Name: instance},
			Parameters:         &runtime.RawExtension{Raw: pdata},
			SecretName:         secretName,
		},
	}
	return b, nil
}

func podPreset(objectName, secretName, providerSvcName, consumerSvcName string) *kalpha.PodPreset {
	podPreset := kalpha.PodPreset{
		ObjectMeta: metav1.ObjectMeta{
			Name: objectName,
			Labels: map[string]string{
				"group":   "mobile",
				"service": providerSvcName,
			},
		},
		Spec: kalpha.PodPresetSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run":           consumerSvcName,
					providerSvcName: "enabled",
				},
			},
			Volumes: []v1.Volume{
				{
					Name: providerSvcName,
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      providerSvcName,
					MountPath: "/etc/secrets/" + providerSvcName,
				},
			},
		},
	}
	return &podPreset
}

func (bc *IntegrationCmd) getServiceNameFromServiceInst(si *v1beta1.ServiceInstance) (string, error) {
	serviceClass, err := bc.
		scClient.
		ServicecatalogV1beta1().
		ClusterServiceClasses().
		Get(si.Spec.ClusterServiceClassRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	serviceMeta := map[string]interface{}{}
	if err := json.Unmarshal(serviceClass.
		Spec.
		ExternalMetadata.
		Raw, &serviceMeta); err != nil {
		return "", errors.WithStack(err)
	}
	serviceName := serviceMeta["serviceName"].(string)
	return serviceName, nil
}

// CreateIntegrationCmd will create a binding from the provider services and setup a pod preset for the consumer service
func (bc *IntegrationCmd) CreateIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration <consuming_service_instance_id> <providing_service_instance_id>",
		Short: "integrate certain mobile services together. mobile get services will show you what can be integrated.",
		Long: `create integration creates a ServiceBinding from one mobile services and injects it into the consuming service via a pod preset in your namespace and optionally
redeploys the consuming service.
To get the IDs of your consuming/providing service instances, run the "mobile get serviceinstances <serviceName>" command from this tool.

If both the --no-wait and --auto-redeploy flags are set to true, --auto-redeploy will override --no-wait.`,
		Example: `  mobile create integration <consuming_service_instance_id> <providing_service_instance_id> --namespace=myproject
  kubectl plugin mobile create integration <consuming_service_instance_id> <providing_service_instance_id>
  oc plugin mobile create integration <consuming_service_instance_id> <providing_service_instance_id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cmd.Usage()
			}
			quiet, err := cmd.Flags().GetBool("quiet")
			if err != nil {
				return errors.Wrap(err, "failed to get quiet flag")
			}
			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			consumerSvcInstName := args[0]
			providerSvcInstName := args[1]
			providerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(providerSvcInstName, metav1.GetOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			consumerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(consumerSvcInstName, metav1.GetOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			consumerServiceName, err := bc.getServiceNameFromServiceInst(consumerSvcInst)
			if err != nil {
				return errors.WithStack(err)
			}
			providerServiceName, err := bc.getServiceNameFromServiceInst(providerSvcInst)
			if err != nil {
				return errors.WithStack(err)
			}
			// Get available bind parameters from the provider cluster service plan
			clusterServiceClass, err := findServiceClassByName(bc.scClient, providerServiceName)
			if err != nil {
				return errors.WithStack(err)
			}
			clusterServicePlan, err := findServicePlanByNameAndClass(bc.scClient, "default", clusterServiceClass.Name)
			if err != nil {
				return errors.WithStack(err)
			}

			bindParams := &ServiceParams{}
			if err := json.Unmarshal(clusterServicePlan.Spec.ServiceBindingCreateParameterSchema.Raw, bindParams); err != nil {
				return errors.WithStack(err)
			}

			flagParams, err := cmd.Flags().GetStringArray("params")
			if err != nil {
				return errors.WithStack(err)
			}

			// Get bind parameters value from user input
			bindParams, err = GetParams(flagParams, bindParams)
			if err != nil {
				return errors.WithStack(err)
			}

			objectName := objectName(consumerSvcInstName, providerSvcInstName)
			preset := podPreset(objectName, objectName, providerServiceName, consumerServiceName)

			// Create Pod Preset for service
			if _, err := bc.k8Client.SettingsV1alpha1().PodPresets(namespace).Create(preset); err != nil {
				return errors.Wrap(err, "failed to create pod preset for service ")
			}
			// prepare and create our binding
			binding, err := createBindingObject(consumerServiceName, providerServiceName, objectName, providerSvcInstName, bindParams, objectName)
			if err != nil {
				return errors.WithStack(err)
			}
			sb, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Create(binding)
			if err != nil {
				return errors.WithStack(err)
			}
			// check if a redeploy was asked for
			redeploy, err := cmd.PersistentFlags().GetBool("auto-redeploy")
			if err != nil {
				return errors.WithStack(err)
			}
			noWait, err := cmd.PersistentFlags().GetBool("no-wait")
			if err != nil {
				return errors.WithStack(err)
			}
			if noWait && !redeploy {
				fmt.Println("you will need to redeploy your service/pod to pick up the changes")
				return nil
			}
			w, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Watch(metav1.ListOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			for u := range w.ResultChan() {
				o := u.Object.(*v1beta1.ServiceBinding)
				if o.Name != sb.Name {
					continue
				}
				switch u.Type {
				case watch.Error:
					w.Stop()
					return errors.New("unexpected error watching service binding " + err.Error())
				case watch.Modified:
					for _, c := range o.Status.Conditions {
						if !quiet {
							fmt.Println("status: " + c.Message)
						}
						if c.Type == "Ready" && c.Status == "True" {
							w.Stop()
						}
						if c.Type == "Failed" {
							w.Stop()
							return errors.New("Failed to create integration: " + c.Message)
						}
					}
				case watch.Deleted:
					w.Stop()
				}
			}
			// once the binding is finished update the deployment to cause a redeploy
			if redeploy {

				// update the deployment with an annotation
				dep, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Get(consumerServiceName, metav1.GetOptions{})
				if err != nil {
					return errors.Wrap(err, "failed to get deployment for service "+consumerSvcInstName)
				}
				dep.Spec.Template.Labels[providerServiceName] = "enabled"
				if _, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
					return errors.Wrap(err, "failed to update deployment for service "+consumerSvcInstName)
				}
			}

			return nil
		},
	}
	cmd.PersistentFlags().Bool("no-wait", false, "--no-wait will cause the command to exit immediately after a successful response instead of waiting until the binding is complete")
	cmd.PersistentFlags().Bool("auto-redeploy", false, "--auto-redeploy=true will cause a backing deployment to be rolled out")
	cmd.PersistentFlags().StringArrayP("params", "p", []string{}, "set the parameters needed to set up the integration programatically rather than being prompted for them: -p PARAM1=val -p PARAM2=val2")

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
			if len(args) != 2 {
				return cmd.Usage()
			}
			quiet, err := cmd.Flags().GetBool("quiet")
			if err != nil {
				return errors.Wrap(err, "failed to get quiet flag")
			}
			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			consumerSvcInstName := args[0]
			providerSvcInstName := args[1]
			providerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(providerSvcInstName, metav1.GetOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			consumerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(consumerSvcInstName, metav1.GetOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			consumerServiceName, err := bc.getServiceNameFromServiceInst(consumerSvcInst)
			if err != nil {
				return errors.WithStack(err)
			}
			providerServiceName, err := bc.getServiceNameFromServiceInst(providerSvcInst)
			if err != nil {
				return errors.WithStack(err)
			}
			objectName := objectName(consumerSvcInstName, providerSvcInstName)
			if err := bc.k8Client.SettingsV1alpha1().PodPresets(namespace).Delete(objectName, metav1.NewDeleteOptions(0)); err != nil {
				return errors.WithStack(err)
			}
			if err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Delete(objectName, metav1.NewDeleteOptions(0)); err != nil {
				return errors.WithStack(err)
			}
			redeploy, err := cmd.PersistentFlags().GetBool("auto-redeploy")
			if err != nil {
				return errors.WithStack(err)
			}
			noWait, err := cmd.PersistentFlags().GetBool("no-wait")
			if err != nil {
				return errors.WithStack(err)
			}
			if noWait && !redeploy {
				fmt.Sprintln(bc.Out, "you will need to redeploy your service to pick up the changes")
				return nil
			}

			w, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Watch(metav1.ListOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			for u := range w.ResultChan() {
				o := u.Object.(*v1beta1.ServiceBinding)
				if o.Name != objectName {
					continue
				}
				switch u.Type {
				case watch.Error:
					w.Stop()
					return errors.New("unexpected error watching service binding " + err.Error())
				case watch.Modified:
					for _, c := range o.Status.Conditions {
						if !quiet {
							fmt.Println("status: " + c.Message)
						}
						if c.Type == "Ready" && c.Status == "True" {
							w.Stop()
						}
						if c.Type == "Failed" {
							w.Stop()
							return errors.New("Failed to create integration: " + c.Message)
						}
					}
				case watch.Deleted:
					w.Stop()
				}
			}

			if !quiet {
				fmt.Printf("Completed deletion of ServiceBinding %v\n", objectName)
			}

			if redeploy {
				//update the deployment with an annotation
				dep, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Get(consumerServiceName, metav1.GetOptions{})
				if err != nil {
					return errors.Wrap(err, "service "+consumerSvcInstName)
				}
				delete(dep.Spec.Template.Labels, providerServiceName)
				if _, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
					return errors.Wrap(err, "failed to update deployment for service "+consumerSvcInstName)
				}
			}
			return nil
		},
	}
	cmd.PersistentFlags().Bool("no-wait", false, "--no-wait will cause the command to exit immediately after a successful response instead of waiting until the binding is complete")
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

func (bc *IntegrationCmd) ListIntegrationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integrations",
		Short: "get a list of the current integrations between services",
		RunE: func(cmd *cobra.Command, args []string) error {
			// list services bincinbx show their annotation values
			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			sbList, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).List(metav1.ListOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			outType := outputType(cmd.Flags())
			if err := bc.Out.Render("list"+cmd.Name(), outType, sbList); err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	}
	bc.Out.AddRenderer("list"+cmd.Name(), "table", func(out io.Writer, dataList interface{}) error {
		bindingList := dataList.(*v1beta1.ServiceBindingList)
		var data [][]string
		for _, b := range bindingList.Items {
			data = append(data, []string{b.Spec.ExternalID, b.Name, b.Annotations["provider"], b.Annotations["consumer"]})
		}
		table := tablewriter.NewWriter(out)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "Provider", "Consumer"})
		table.Render()
		return nil

	})
	return cmd
}

// TODO review how we build params. This is still POC
func buildBindParams(bindParams *ServiceParams) map[string]interface{} {
	params := map[string]interface{}{}

	for k, v := range bindParams.Properties {
		params[k] = v["value"]
	}

	return params
}
