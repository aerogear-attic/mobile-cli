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
	"strings"

	"github.com/aerogear/mobile-cli/pkg/cmd/output"
	"github.com/aerogear/mobile-crd-client/pkg/apis/mobile/v1alpha1"
	"github.com/aerogear/mobile-crd-client/pkg/apis/servicecatalog/v1beta1"
	mobile "github.com/aerogear/mobile-crd-client/pkg/client/mobile/clientset/versioned"
	"github.com/aerogear/mobile-crd-client/pkg/client/servicecatalog/clientset/versioned"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

type ClientCmd struct {
	*BaseCmd
	mobileClient mobile.Interface
	scClient     versioned.Interface
	k8Client     kubernetes.Interface
}

// NewClientCmd returns a configured ClientCmd ready for use
func NewClientCmd(mobileClient mobile.Interface, scClient versioned.Interface, k8Client kubernetes.Interface, out io.Writer) *ClientCmd {
	return &ClientCmd{mobileClient: mobileClient, scClient: scClient, k8Client: k8Client, BaseCmd: &BaseCmd{Out: output.NewRenderer(out)}}
}

// ListClientsCmd builds the list mobile clients command
func (cc *ClientCmd) ListClientsCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "clients",
		Short: "gets a list of mobile clients represented in the namespace",
		Long:  `get clients allows you to get a list of mobile clients that are represented in your namespace.`,
		Example: `  mobile get clients --namespace=myproject 
  					kubectl plugin mobile get clients
  					oc plugin mobile get clients`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			apps, err := cc.mobileClient.MobileV1alpha1().MobileClients(ns).List(metav1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to list mobile clients")
			}
			outType := outputType(cmd.Flags())
			if err := cc.Out.Render("list"+cmd.Name(), outType, apps); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "mobile clients", outType))
			}
			return nil
		},
	}
	cc.Out.AddRenderer("list"+command.Name(), "table", func(out io.Writer, mobileClientList interface{}) error {
		mClients := mobileClientList.(*v1alpha1.MobileClientList)
		var data [][]string
		for _, mClient := range mClients.Items {
			data = append(data, []string{mClient.Name, mClient.Spec.Name, mClient.Spec.ClientType, mClient.Spec.AppIdentifier})
		}
		table := tablewriter.NewWriter(out)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "ClientType", "AppIdentifier", "ExcludedServices"})
		table.Render()
		return nil
	})
	return command
}

// GetClientCmd builds the get mobileclient command
func (cc *ClientCmd) GetClientCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "client <clientID>",
		Short: "gets a single mobile client in the namespace",
		Long: `get client allows you to view client information for a specific mobile client in your namespace.
			   Run the "mobile get clients" command from this tool to get the client ID.`,
		Example: `  mobile get client <clientID> --namespace=myproject 
  					kubectl plugin mobile get client <clientID>
oc plugin mobile get client <clientID>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			clientID := args[0]
			ns, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			client, err := cc.mobileClient.MobileV1alpha1().MobileClients(ns).Get(clientID, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get mobile client with clientID "+clientID)
			}
			outType := outputType(cmd.Flags())
			if err := cc.Out.Render(cmd.Name(), outType, client); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "mobile client", outType))
			}
			return nil
		},
	}
	cc.Out.AddRenderer(command.Name(), "table", func(out io.Writer, mobileClient interface{}) error {
		mClient := mobileClient.(*v1alpha1.MobileClient)
		var data [][]string
		data = append(data, []string{mClient.Name, mClient.Namespace, mClient.Spec.Name, mClient.Spec.ClientType, mClient.Spec.ApiKey, mClient.Spec.AppIdentifier})
		table := tablewriter.NewWriter(out)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Namespace", "Name", "ClientType", "ApiKey", "AppIdentifier", "Excluded Services"})
		table.Render()
		return nil
	})
	return command
}

// CreateClientCmd builds the create mobileclient command
func (cc *ClientCmd) CreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client <name> <clientType iOS|cordova|android|xamarin> <appIdentifier bundleID|packageName>",
		Short: "create a mobile client representation in your namespace",
		Long: `create client sets up the representation of a mobile application of the specified type in your namespace.
		       This is used to provide a mobile client context for various actions such as creating, starting or stopping mobile client builds.

		       The available client types are android, cordova, iOS and xamarin.

		       When used standalone, a namespace must be specified by providing the --namespace flag.`,
		Example: `  mobile create client <name> <clientType> <appIdentifier> --namespace=myproject 
  					kubectl plugin mobile create client <name> <clientType> <appIdentifier>
					oc plugin mobile create client <name> <clientType> <appIdentifier>`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return cmd.Usage()
			}

			name := args[0]
			clientType := args[1]
			appIdentifier := args[2]

			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}

			if appIdentifier == "" {
				return errors.New("failed validation while creating new mobile client")
			}

			var apbName string
			switch clientType {
			case "android", "iOS", "cordova", "xamarin":
				apbName = clientType + "-app"
			default:
				return errors.New("Unknown client type")

			}

			clientId := strings.ToLower(name + "-" + clientType)
			client, err := cc.mobileClient.MobileV1alpha1().MobileClients(namespace).Get(clientId, metav1.GetOptions{})
			if client.ObjectMeta.UID != "" {
				return errors.New("App with this name already exist for this client type")
			}

			if err != nil && !strings.Contains(err.Error(), fmt.Sprintf("\"%s\" not found", clientId)) {
				return errors.Wrap(err, "failed to check if application name exists")
			}

			//Get available provision parameters from the cluster service plan
			clusterServiceClass, err := findServiceClassByName(cc.scClient, apbName)
			if err != nil {
				return errors.Wrap(err, "failed to find ServiceClass by name")
			}

			validServiceName := clusterServiceClass.Spec.ExternalName
			extMeta := clusterServiceClass.Spec.ExternalMetadata.Raw
			var extServiceClass ExternalServiceMetaData
			if err := json.Unmarshal(extMeta, &extServiceClass); err != nil {
				return errors.Wrap(err, "failed to read ClusterServiceClass")
			}

			secretName := clientId + "-apb-" + "params"
			si := buildServiceInstance(namespace, validServiceName+"-", secretName, *clusterServiceClass)

			if _, err := cc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Create(&si); err != nil {
				return errors.Wrap(err, "failed to create mobile client")
			}
			fmt.Println("Creating Mobile Client")

			parameters := map[string]string{
				"appName":       name,
				"appIdentifier": appIdentifier,
			}

			secretData, err := json.Marshal(parameters)
			if err != nil {
				return errors.Wrap(err, "invalid secret data")
			}

			pSecret := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secretName,
				},
				Data: map[string][]byte{
					"parameters": secretData,
				},
			}

			if _, err := cc.k8Client.CoreV1().Secrets(namespace).Create(&pSecret); err != nil {
				return errors.Wrap(err, "failed to create secret")
			}

			noWait, err := cmd.PersistentFlags().GetBool("no-wait")
			if err != nil {
				return errors.WithStack(err)
			}
			if noWait {
				return nil
			}
			cc.Out.AddRenderer("create"+cmd.Name(), "table", func(writer io.Writer, mobileClient interface{}) error {
				var data [][]string
				mClient := mobileClient.(*v1alpha1.MobileClient)
				data = append(data, []string{mClient.Name, mClient.Spec.Name, mClient.Spec.ClientType, mClient.Spec.AppIdentifier})
				table := tablewriter.NewWriter(writer)
				table.AppendBulk(data)
				table.SetHeader([]string{"ID", "Name", "ClientType", "AppIdentifier"})
				table.Render()
				return nil
			})

			timeout := int64(10 * 60) // ten minutes
			w, err := cc.scClient.ServicecatalogV1beta1().ServiceInstances(namespace).Watch(metav1.ListOptions{TimeoutSeconds: &timeout})

			if err != nil {
				return errors.WithStack(err)
			}
			for {
				select {
				case msg, ok := <-w.ResultChan():
					if !ok {
						fmt.Println("Timedout waiting. It seems to be taking a long time for the Mobile Client to provision. Your Mobile Client service may still be provisioning.")
						return nil
					}
					o := msg.Object.(*v1beta1.ServiceInstance)
					switch msg.Type {
					case watch.Error:
						w.Stop()
						return errors.New("unexpected error watching ServiceInstance " + err.Error())
					case watch.Modified:
						for _, c := range o.Status.Conditions {
							if c.Type == "Ready" && c.Status == "True" {
								w.Stop()

								outType := outputType(cmd.Flags())
								mClient, err := cc.mobileClient.MobileV1alpha1().MobileClients(namespace).Get(clientId, metav1.GetOptions{})
								if err != nil {
									return errors.Wrap(err, "Cant get client post creation, something went wrong")
								}
								if err := cc.Out.Render("create"+cmd.Name(), outType, mClient); err != nil {
									return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "mobile client", outType))
								}
								return nil
							}

							if c.Type == "Failed" {
								w.Stop()
								return errors.New("Failed to provision " + extServiceClass.ServiceName + ". " + c.Message)
							}
						}
					}
				}
			}
		},
	}

	cmd.PersistentFlags().Bool("no-wait", false, "--no-wait will cause the command to exit immediately after a successful response instead of waiting until the service is fully provisioned")
	return cmd
}

// DeleteClientCmd builds the delete mobile client command
func (cc *ClientCmd) DeleteClientCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "client <clientID>",
		Short: "deletes a single mobile client in the namespace",
		Long: `delete client allows you to delete a single mobile client in your namespace.
               Run the "mobile get clients" command from this tool to get the client ID.`,
		Example: `  mobile delete client <clientID> --namespace=myproject 
                    kubectl plugin mobile delete client <clientID>
                    oc plugin mobile delete client <clientID>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var ns string

			if len(args) != 1 {
				return cmd.Usage()
			}
			clientID := args[0]
			ns, err = currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}

			err = cc.mobileClient.MobileV1alpha1().MobileClients(ns).Delete(clientID, &metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get mobile client with clientID "+clientID)
			}
			return nil
		},
	}
	return command
}

// SetClientValueCmd sets value in client
func (cc *ClientCmd) SetClientValueFromJsonCmd() *cobra.Command {
	var patch string
	command := &cobra.Command{
		Use:   "client <clientID>",
		Short: "Patches mobileclient with provided json",
		Long: `set client allows you to patch a mobile client in your namespace.
			   Run the "mobile get clients" command from this tool to get the client ID.`,
		Example: `mobile set client <clientID> --patch='{"spec": {"dmzUrl": "www.dmz.com"}}' --namespace=myproject 
  			      kubectl plugin mobile set client <clientID> --patch='{"spec": {"dmzUrl": "www.dmz.com"}}'
				  oc plugin mobile set client <clientID> --patch='{"spec": {"dmzUrl": "www.dmz.com"}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				res *v1alpha1.MobileClient
				ns  string
			)

			if len(args) != 1 {
				return cmd.Usage()
			}
			clientID := args[0]

			patch, err := cmd.PersistentFlags().GetString("patch")
			if err != nil {
				return errors.Wrap(err, "failed to get patch flag")
			}

			ns, err = currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}

			res, err = cc.mobileClient.MobileV1alpha1().MobileClients(ns).Patch(clientID, types.MergePatchType, []byte(patch))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to patch mobile client with clientID %s", clientID))
			}

			outType := outputType(cmd.Flags())
			if err := cc.Out.Render(cmd.Name(), outType, res); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "mobile client", outType))
			}

			return nil
		},
	}
	command.PersistentFlags().StringVarP(&patch, "patch", "p", "", "patch json to apply")
	return command
}

// SetClientSpecValueCmd sets value in client
func (cc *ClientCmd) SetClientSpecValueCmd() *cobra.Command {
	var (
		clientId  string
		valueName string
		value     string
	)

	command := &cobra.Command{
		Use:   "value",
		Short: "Sets value in mobileclient spec",
		Long: `set client allows you to patch a mobile client in your namespace.
			   Run the "mobile get clients" command from this tool to get the client ID.`,
		Example: `mobile set --client=<clientID> --name=dmzUrl --value=www.example.com --namespace=myproject 
  			      kubectl plugin mobile set --client=<clientID> --name=dmzUrl --value=www.example.com
				  oc plugin mobile set --client=<clientID> --name=dmzUrl --value=www.example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				res *v1alpha1.MobileClient
				ns  string
			)

			clientId, err := cmd.PersistentFlags().GetString("client")
			if err != nil {
				return errors.Wrap(err, "failed to get client flag")
			}

			name, err := cmd.PersistentFlags().GetString("name")
			if err != nil {
				return errors.Wrap(err, "failed to get name flag")
			}

			value, err := cmd.PersistentFlags().GetString("value")
			if err != nil {
				return errors.Wrap(err, "failed to get value flag")
			}

			ns, err = currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}

			var patch string
			if value == "null" || value == "" {
				patch = fmt.Sprintf("{\"spec\": {\"%s\": null}}", name)
			} else {
				patch = fmt.Sprintf("{\"spec\": {\"%s\": \"%v\"}}", name, value)
			}

			res, err = cc.mobileClient.MobileV1alpha1().MobileClients(ns).Patch(clientId, types.MergePatchType, []byte(patch))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to set value in mobile client with clientID %s", clientId))
			}

			if err := cc.Out.Render(cmd.Name(), "json", res); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "mobile client", "json"))
			}

			return nil
		},
	}
	command.PersistentFlags().StringVarP(&clientId, "client", "c", "", "value")
	command.PersistentFlags().StringVarP(&valueName, "name", "n", "", "name")
	command.PersistentFlags().StringVarP(&value, "value", "v", "", "value")
	return command
}
