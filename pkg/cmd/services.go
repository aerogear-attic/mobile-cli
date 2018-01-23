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
	"fmt"
	"os"
	"strings"

	"io"

	"encoding/json"

	"bufio"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"sort"

	"github.com/aerogear/mobile-cli/pkg/cmd/output"
	"github.com/satori/go.uuid"
	"k8s.io/apimachinery/pkg/watch"
)

type ServicesCmd struct {
	*BaseCmd
	scClient versioned.Interface
	k8Client kubernetes.Interface
}

func NewServicesCmd(scClient versioned.Interface, k8Client kubernetes.Interface, out io.Writer) *ServicesCmd {
	return &ServicesCmd{scClient: scClient, k8Client: k8Client, BaseCmd: &BaseCmd{Out: output.NewRenderer(out)}}
}

func (sc *ServicesCmd) ListServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "get mobile aware services that can be provisioned to your namespace",
		Long:  `get services allows you to get a list of services that can be provisioned in your namespace.`,
		Example: `  mobile get services --namespace=myproject 
  kubectl plugin mobile get services
  oc plugin mobile get services`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scList, err := sc.scClient.ServicecatalogV1beta1().ClusterServiceClasses().List(metav1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to list service classes")
			}
			outPutType := outputType(cmd.Flags())
			all, err := cmd.PersistentFlags().GetBool("all")
			if err != nil {
				return errors.Wrap(err, "failed to get all flag")
			}
			if nil == scList {
				return errors.New("no serviceclasses returned")
			}
			if !all {
				tempList := &v1beta1.ClusterServiceClassList{}
				for _, item := range scList.Items {
					for _, tag := range item.Spec.Tags {
						if tag == "mobile-service" {
							tempList.Items = append(tempList.Items, item)
						}
					}
				}
				scList = tempList
			}
			if err := sc.Out.Render("list"+cmd.Name(), outPutType, scList); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "serviceclass", outPutType))
			}
			return nil
		},
	}

	// add our table output renderer
	sc.Out.AddRenderer("list"+cmd.Name(), "table", func(writer io.Writer, serviceClasses interface{}) error {
		scL := serviceClasses.(*v1beta1.ClusterServiceClassList)
		var data [][]string
		for _, item := range scL.Items {
			extMeta := item.Spec.ExternalMetadata.Raw
			extServiceClass := map[string]interface{}{}
			if err := json.Unmarshal(extMeta, &extServiceClass); err != nil {
				return err
			}
			serviceName := ""
			integrations := ""

			clusterServicePlan, err := findServicePlanByNameAndClass(sc.scClient, "default", item.Name)
			if err != nil {
				return err
			}

			params := &InstanceCreateParams{}
			if err := json.Unmarshal(clusterServicePlan.Spec.ServiceInstanceCreateParameterSchema.Raw, &params); err != nil {
				return err
			}
			var createParams []string
			for k := range params.Properties {
				createParams = append(createParams, k)
			}

			sort.Strings(createParams)
			if v, ok := extServiceClass["serviceName"].(string); ok {
				serviceName = v
			}
			if v, ok := extServiceClass["integrations"].(string); ok {
				integrations = v
			}

			data = append(data, []string{serviceName, integrations, strings.Join(createParams, ",\n")})
		}
		table := tablewriter.NewWriter(writer)
		table.AppendBulk(data)
		table.SetHeader([]string{"Name", "Integrations", "Parameters"})
		table.Render()
		return nil
	})
	cmd.PersistentFlags().Bool("all", false, "--all return all services not just mobile aware ones")
	return cmd
}

func findServiceClassByName(scClient versioned.Interface, name string) (*v1beta1.ClusterServiceClass, error) {
	mobileServices, err := scClient.ServicecatalogV1beta1().ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if mobileServices == nil || len(mobileServices.Items) == 0 {
		return nil, errors.New("failed to find serviceclass with name: " + name)
	}

	for _, item := range mobileServices.Items {
		var extData ExternalServiceMetaData
		rawData := item.Spec.ExternalMetadata.Raw
		if err := json.Unmarshal(rawData, &extData); err != nil {
			return nil, err
		}
		if extData.ServiceName == name {
			return &item, nil
		}
	}
	return nil, errors.New("failed to find serviceclass with name: " + name)
}

func findServicePlanByNameAndClass(scClient versioned.Interface, planName, serviceClassName string) (*v1beta1.ClusterServicePlan, error) {
	plans, err := scClient.ServicecatalogV1beta1().ClusterServicePlans().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, item := range plans.Items {
		if item.Spec.ClusterServiceClassRef.Name == serviceClassName && item.Spec.ExternalName == planName {
			return &item, nil
		}
	}

	return nil, errors.New("failed to find serviceplan associated with the serviceclass " + serviceClassName)
}

func requiredParam(instParams InstanceCreateParams, key string) bool {
	for _, r := range instParams.Required {
		if r == key {
			return true
		}
	}
	return false
}

func parseParams(keyVals []string) (map[string]string, error) {
	params := map[string]string{}
	for _, p := range keyVals {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			return nil, NewIncorrectParameterFormat("key value pairs are needed failed to find one: " + p)
		}
		params[strings.TrimSpace(kv[0])] = kv[1]
	}
	return params, nil
}

type InstanceCreateParams struct {
	AdditionalProperties bool                              `json:"additionalProperties"`
	Properties           map[string]map[string]interface{} `json:"properties"`
	Required             []string                          `json:"required"`
	Type                 string                            `json:"type"`
}

func (sc *ServicesCmd) CreateServiceInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceinstance <serviceName>",
		Short: `create a running instance of the given service`,
		Long: `create service instance allows you to create a running instance of a service in your namespace. 
Run the "mobile get services" command from this tool to see which services are available for provisioning.`,
		Example: `  mobile create serviceinstance <serviceName> --namespace=myproject 
  kubectl plugin mobile create serviceinstance <serviceName>
  oc plugin mobile create serviceinstance <serviceName>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("expected the name of a service to provision")
			}
			// find our serviceclass and plan
			serviceName := args[0]

			ns, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}

			clusterServiceClass, err := findServiceClassByName(sc.scClient, serviceName)
			if err != nil {
				return errors.WithStack(err)
			}
			clusterServicePlan, err := findServicePlanByNameAndClass(sc.scClient, "default", clusterServiceClass.Name)
			if err != nil {
				return errors.WithStack(err)
			}
			// handle the params
			params, err := cmd.Flags().GetStringArray("params")
			if err != nil {
				return errors.WithStack(err)
			}
			parsedParams, err := parseParams(params)
			if err != nil {
				return errors.WithStack(err)
			}
			instParams := &InstanceCreateParams{}
			if err := json.Unmarshal(clusterServicePlan.Spec.ServiceInstanceCreateParameterSchema.Raw, instParams); err != nil {
				return errors.WithStack(err)
			}

			if len(parsedParams) > 0 {
				for k, v := range instParams.Properties {
					defaultVal := v["default"]
					if pVal, ok := parsedParams[k]; !ok && requiredParam(*instParams, k) || requiredParam(*instParams, k) && pVal == "" {
						if defaultVal != nil {
							//use default
							v["value"] = defaultVal
							continue
						}
						return errors.New(fmt.Sprintf("missing required parameter %s", k))
					}
					v["value"] = parsedParams[k]
					instParams.Properties[k] = v
				}
			} else {
				scanner := bufio.NewScanner(os.Stdin)
				for k, v := range instParams.Properties {
					questionFormat := "Set value for %s [default value: %s required: %v]"

					fmt.Println(fmt.Sprintf(questionFormat, k, v["default"], requiredParam(*instParams, k)))
					scanner.Scan()
					//
					val := scanner.Text()
					if val == "" {
						val = v["default"].(string)
					}
					v["value"] = val
					instParams.Properties[k] = v
					fmt.Println("set value for " + k + " to : " + val)
				}
			}

			validServiceName := clusterServiceClass.Spec.ExternalName
			sid := uuid.NewV4().String()
			extMeta := clusterServiceClass.Spec.ExternalMetadata.Raw
			var extServiceClass ExternalServiceMetaData
			if err := json.Unmarshal(extMeta, &extServiceClass); err != nil {
				return errors.WithStack(err)
			}

			si := v1beta1.ServiceInstance{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "servicecatalog.k8s.io/v1beta1",
					Kind:       "ServiceInstance",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels:       map[string]string{"id": sid, "serviceName": extServiceClass.ServiceName},
					Namespace:    ns,
					GenerateName: validServiceName + "-",
				},
				Spec: v1beta1.ServiceInstanceSpec{
					PlanReference: v1beta1.PlanReference{
						ClusterServiceClassExternalName: clusterServiceClass.Spec.ExternalName,
					},
					ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
						Name: clusterServiceClass.Name,
					},
					ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
						Name: "default",
					},
					ParametersFrom: []v1beta1.ParametersFromSource{
						{
							SecretKeyRef: &v1beta1.SecretKeyReference{
								Name: validServiceName + "-" + "params",
								Key:  "parameters"},
						},
					},
				},
			}
			if _, err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).Create(&si); err != nil {
				return errors.WithStack(err)
			}

			pSecret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: validServiceName + "-" + "params",
				},
			}
			pSecret.Data = map[string][]byte{}
			parameters := map[string]string{}

			for k, v := range instParams.Properties {
				if v, ok := v["value"]; ok && v != nil {
					parameters[k] = v.(string)
				}
			}
			secretData, err := json.Marshal(parameters)
			if err != nil {
				return errors.WithStack(err)
			}
			pSecret.Data["parameters"] = secretData
			if _, err := sc.k8Client.CoreV1().Secrets(ns).Create(&pSecret); err != nil {
				return errors.WithStack(err)
			}

			noWait, err := cmd.PersistentFlags().GetBool("no-wait")
			if err != nil {
				return errors.WithStack(err)
			}
			if noWait {
				return nil
			}
			w, err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).Watch(metav1.ListOptions{LabelSelector: "id=" + sid})
			if err != nil {
				return errors.WithStack(err)
			}
			for u := range w.ResultChan() {
				o := u.Object.(*v1beta1.ServiceInstance)
				switch u.Type {
				case watch.Error:
					w.Stop()
					return errors.New("unexpected error watching ServiceInstance " + err.Error())
				case watch.Modified:
					for _, c := range o.Status.Conditions {
						fmt.Println("status: " + c.Message)
						if c.Type == "Ready" && c.Status == "True" {
							w.Stop()
						}
					}
				}
			}

			return nil
		},
	}
	cmd.PersistentFlags().Bool("no-wait", false, "--no-wait will cause the command to exit immediately after a successful response instead of waiting until the service is fully provisioned")
	cmd.PersistentFlags().StringArrayP("params", "p", []string{}, "set the parameters  needed to set up the service programatically rather than being prompted for them: -p PARAM1=val -p PARAM2=val2")
	return cmd
}

func (sc *ServicesCmd) DeleteServiceInstanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serviceinstance <serviceInstanceID>",
		Short: "deletes a service instance and other objects created when provisioning the services instance, such as pod presets",
		Long: `delete serviceinstance allows you to delete a service instance and other objects created when provisioning the services instance, such as pod presets. 
Run the "mobile get serviceinstances" command from this tool to see which service instances are available for deleting.`,
		Example: `  mobile delete serviceinstance <serviceInstanceID> --namespace=myproject 
  kubectl plugin mobile delete serviceinstance <serviceInstanceID>
  oc plugin mobile delete serviceinstance <serviceInstanceID>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			//delete service instance
			//delete params secret
			if len(args) != 1 {
				return errors.New("expected a serviceInstanceID")
			}
			ns, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			sid := args[0]
			// Retrieve the service instance in full so we can build the secret name
			serviceInstance, err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).Get(sid, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).Delete(sid, &metav1.DeleteOptions{}); err != nil {
				return err
			}
			secretName := serviceInstance.ObjectMeta.GenerateName + "params"
			return sc.k8Client.CoreV1().Secrets(ns).Delete(secretName, &metav1.DeleteOptions{})
		},
	}
}

func (sc *ServicesCmd) ListServiceInstCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceinstances <serviceName>",
		Short: "get a list of provisioned service instances based on the service name.",
		Long:  `get serviceinstances allows you to get a list of provisioned service instances in your namespace, based on the service name.`,
		Example: `  mobile get serviceinstances <serviceName> --namespace=myproject 
  kubectl plugin mobile get serviceinstances <serviceName>
  oc plugin mobile get serviceinstances <serviceName>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("no service name passed")
			}
			serviceName := args[0]
			ns, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			si, err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).List(metav1.ListOptions{LabelSelector: "serviceName=" + serviceName})
			if err != nil {
				return err
			}
			outType := outputType(cmd.Flags())
			if err := sc.Out.Render("list"+cmd.Name(), outType, si); err != nil {
				return err
			}

			return nil
		},
	}
	sc.Out.AddRenderer("list"+cmd.Name(), "table", func(writer io.Writer, serviceInstances interface{}) error {
		scL := serviceInstances.(*v1beta1.ServiceInstanceList)
		var data [][]string
		for _, item := range scL.Items {
			data = append(data, []string{item.Spec.ClusterServiceClassExternalName, item.Name})
		}
		table := tablewriter.NewWriter(writer)
		table.AppendBulk(data)
		table.SetHeader([]string{"Name", "ID"})
		table.Render()
		return nil
	})
	return cmd
}
