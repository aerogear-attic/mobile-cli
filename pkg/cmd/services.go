package cmd

import (
	"os"

	"fmt"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			scList, err := sc.scClient.ServicecatalogV1beta1().ClusterServiceClasses().List(metav1.ListOptions{})
			fmt.Println("sclist", scList, err)
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
			extServiceClass := map[string]string{}
			if err := json.Unmarshal(extMeta, &extServiceClass); err != nil {
				return err
			}
			data = append(data, []string{item.Spec.ExternalName, extServiceClass["serviceName"], extServiceClass["integrations"], item.Name})
		}
		table := tablewriter.NewWriter(writer)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "Integrations", "Class"})
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
		return nil, errors.New("failed to find and service classes for " + name)
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
	return nil, nil

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

	return nil, nil
}

type instanceCreateParams struct {
	AdditionalProperties bool                         `json:"additionalProperties"`
	Properties           map[string]map[string]string `json:"properties"`
	Required             []string                     `json:"required"`
	Type                 string                       `json:"type"`
}

func (sc *ServicesCmd) CreateServiceInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceinstance <serviceName>",
		Short: `create a running instance of the given service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("expected the name of a service to provision")
			}
			serviceName := args[0]
			ns := currentNamespace(cmd.Flags())
			clusterServiceClass, err := findServiceClassByName(sc.scClient, serviceName)
			if err != nil || clusterServiceClass == nil {
				msg := "failed to find a service class associated with that name "
				if err != nil {
					msg += err.Error()
				}
				return errors.New(msg)
			}
			clusterServicePlan, err := findServicePlanByNameAndClass(sc.scClient, "default", clusterServiceClass.Name)
			if err != nil {
				return err
			}

			if clusterServicePlan == nil {
				return errors.New("failed to find service plan with name default for service " + serviceName)
			}

			params := &instanceCreateParams{}

			if err := json.Unmarshal(clusterServicePlan.Spec.ServiceInstanceCreateParameterSchema.Raw, params); err != nil {
				return err
			}
			scanner := bufio.NewScanner(os.Stdin)
			for k, v := range params.Properties {
				fmt.Println("Set value for " + k + " default value: " + v["default"])
				scanner.Scan()
				//
				val := scanner.Text()
				if val == "" {
					val = v["default"]
				}
				v["value"] = val
				params.Properties[k] = v
				fmt.Println("set value for " + k + " to : " + val)
			}

			validServiceName := clusterServiceClass.Spec.ExternalName
			sid := uuid.NewV4().String()
			extMeta := clusterServiceClass.Spec.ExternalMetadata.Raw
			var extServiceClass ExternalServiceMetaData
			if err := json.Unmarshal(extMeta, &extServiceClass); err != nil {
				return err
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
				return err
			}

			pSecret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: validServiceName + "-" + "params",
				},
			}
			pSecret.Data = map[string][]byte{}
			parameters := map[string]string{}

			for k, v := range params.Properties {
				parameters[k] = v["value"]
			}
			secretData, err := json.Marshal(parameters)
			if err != nil {
				return err
			}
			pSecret.Data["parameters"] = secretData
			if _, err := sc.k8Client.CoreV1().Secrets(ns).Create(&pSecret); err != nil {
				return err
			}
			noWait, err := cmd.PersistentFlags().GetBool("no-wait")
			if err != nil {
				return err
			}
			if noWait {
				return nil
			}
			w, err := sc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).Watch(metav1.ListOptions{LabelSelector: "id=" + sid})
			if err != nil {
				return err
			}
			for u := range w.ResultChan() {
				switch u.Type {
				case watch.Modified:
					o := u.Object.(*v1beta1.ServiceInstance)
					lastOp := o.Status.LastOperation
					if nil != lastOp {
						//fmt.Println("last operation " + *lastOp)
					}
					for _, c := range o.Status.Conditions {
						fmt.Println(c.Message)
						if c.Type == "Ready" && c.Status == "True" {
							w.Stop()
						}
					}
				}
			}

			return nil
		},
	}
	cmd.PersistentFlags().Bool("no-wait", false, "--no-wait will cause the command to exit immediately instead of waiting for the service to be provisioned")
	return cmd
}

func (sc *ServicesCmd) DeleteServiceInstanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serviceinstance <serviceInstanceID>",
		Short: "deletes a service instance and other objects created when provisioning the services instance such as pod presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			//delete service instance
			//delete params secret
			if len(args) != 1 {
				return errors.New("expected a serviceInstanceID")
			}
			ns := currentNamespace(cmd.Flags())
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
		Short: "get a list of provisioned serviceInstances based on the service name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("no service name passed")
			}
			serviceName := args[0]
			ns := currentNamespace(cmd.Flags())
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
