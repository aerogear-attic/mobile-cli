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

	"os"

	"io"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/aerogear/mobile-cli/pkg/cmd/output"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
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

func NewIntegrationCmd(scClient sc.Interface, k8Client kubernetes.Interface) *IntegrationCmd {
	return &IntegrationCmd{scClient: scClient, k8Client: k8Client, BaseCmd: &BaseCmd{Out: output.NewRenderer(os.Stdout)}}
}

func createBindingObject(consumer, provider, bindingName, instance string, params map[string]interface{}, secretName string) (*v1beta1.ServiceBinding, error) {
	pdata, err := json.Marshal(params)
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

func (bc *IntegrationCmd) CreateIntegrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration <consuming_service_instance_id> <providing_service_instance_id>",
		Short: "integrate certain mobile services together",
		Long: `create integration allows you to create a binding between mobile services in your namespace.
To get the IDs of your consuming/providing service instances, run the "mobile get serviceinstances <serviceName>" command from this tool.

If both the --no-wait and --auto-redeploy flags are set to true, --auto-redeploy will override --no-wait.`,
		Example: `  mobile create integration <consuming_service_instance_id> <providing_service_instance_id> --namespace=myproject
  kubectl plugin mobile create integration <consuming_service_instance_id> <providing_service_instance_id>
  oc plugin mobile create integration <consuming_service_instance_id> <providing_service_instance_id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cmd.Usage()
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
			// todo remove the need for these by updating the apbs to read the secrets themselves (https://github.com/feedhenry/keycloak-apb/issues/37)
			consumerSvc := getService(namespace, consumerSvcInst.Labels["serviceName"], bc.k8Client) // the consumer service
			providerSvc := getService(namespace, providerSvcInst.Labels["serviceName"], bc.k8Client) // the provider service
			bindParams := buildBindParams(providerSvc, consumerSvc)
			objectName := objectName(consumerSvcInstName, providerSvcInstName)
			binding, err := createBindingObject(consumerSvc.Name, providerSvc.Name, objectName, providerSvcInst.Name, bindParams, objectName)
			if err != nil {
				return errors.WithStack(err)
			}
			sb, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Create(binding)
			if err != nil {
				return errors.WithStack(err)
			}
			preset := podPreset(objectName, objectName, providerSvc.Name, consumerSvc.Name)
			if _, err := bc.k8Client.SettingsV1alpha1().PodPresets(namespace).Create(preset); err != nil {
				return errors.Wrap(err, "failed to create pod preset for service ")
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
				fmt.Println("you will need to redeploy your service/pod to pick up the changes")
				return nil
			}

			if redeploy {
				id := sb.Spec.ExternalID
				w, err := bc.scClient.ServicecatalogV1beta1().ServiceBindings(namespace).Watch(metav1.ListOptions{LabelSelector: "id=" + id})
				if err != nil {
					return errors.WithStack(err)
				}
				for u := range w.ResultChan() {
					o := u.Object.(*v1beta1.ServiceBinding)
					switch u.Type {
					case watch.Error:
						w.Stop()
						return errors.New("unexpected error watching service binding " + err.Error())
					case watch.Modified:
						for _, c := range o.Status.Conditions {
							fmt.Println("status: " + c.Message)
							if c.Type == "Ready" && c.Status == "True" {
								w.Stop()
							}
							if c.Type == "Failed" {
								w.Stop()
								return errors.New("Failed to create integration: " + c.Message)
							}
						}
					}
				}

				// update the deployment with an annotation
				dep, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Get(consumerSvc.Name, metav1.GetOptions{})
				if err != nil {
					return errors.Wrap(err, "failed to get deployment for service "+consumerSvcInstName)
				}
				dep.Spec.Template.Labels[providerSvc.Name] = "enabled"
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
			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			consumerSvcInstName := args[0]
			providerSvcInstName := args[1]

			consumerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(consumerSvcInstName, metav1.GetOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			providerSvcInst, err := bc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(providerSvcInstName, metav1.GetOptions{})
			if err != nil {
				return errors.WithStack(err)
			}
			consumerSvcName := consumerSvcInst.Labels["serviceName"]
			providerSvcName := providerSvcInst.Labels["serviceName"]
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
			if !redeploy {
				fmt.Println("you will need to redeploy your service to pick up the changes")
				return nil
			}
			//update the deployment with an annotation
			dep, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Get(consumerSvcName, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get deployment for service "+consumerSvcInstName)
			}
			delete(dep.Spec.Template.Labels, providerSvcName)
			if _, err := bc.k8Client.AppsV1beta1().Deployments(namespace).Update(dep); err != nil {
				return errors.Wrap(err, "failed to update deployment for service "+consumerSvcInstName)
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

// TODO remove need for this by retrieving the params from the secret in the APB
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
