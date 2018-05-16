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
	"io"
	"strings"

	"github.com/aerogear/mobile-cli/pkg/apis/mobile/v1alpha1"
	mobile "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned"
	"github.com/aerogear/mobile-cli/pkg/cmd/input"
	"github.com/aerogear/mobile-cli/pkg/cmd/output"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ClientCmd struct {
	*BaseCmd
	mobileClient mobile.Interface
}

// NewClientCmd returns a configured ClientCmd ready for use
func NewClientCmd(mobileClient mobile.Interface, out io.Writer) *ClientCmd {
	return &ClientCmd{mobileClient: mobileClient, BaseCmd: &BaseCmd{Out: output.NewRenderer(out)}}
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
			appKey := uuid.NewV4().String()

			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			app := &v1alpha1.MobileClient{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MobileClient",
					APIVersion: "mobile.k8s.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: v1alpha1.MobileClientSpec{Name: name, ApiKey: appKey, ClientType: clientType, AppIdentifier: appIdentifier},
			}
			switch app.Spec.ClientType {
			case "android":
				app.Annotations["icon"] = "fa fa-android"
				break
			case "iOS":
				app.Annotations["icon"] = "fa fa-apple"
				break
			case "cordova":
				app.Annotations["icon"] = "font-icon icon-cordova"
				break
			case "xamarin":
				app.Annotations["icon"] = "font-icon icon-xamarin"
				break
			default:
				return errors.New("Unknown client type")
			}
			app.Name = name + "-" + strings.ToLower(app.Spec.ClientType)
			if err := input.ValidateMobileClient(app); err != nil {
				return errors.Wrap(err, "failed validation while creating new mobile client")
			}
			createdClient, err := cc.mobileClient.MobileV1alpha1().MobileClients(namespace).Create(app)
			if err != nil {
				return errors.Wrap(err, "failed to create mobile client")
			}
			outType := outputType(cmd.Flags())
			if err := cc.Out.Render("create"+cmd.Name(), outType, createdClient); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "mobile client", outType))
			}
			return nil
		},
	}

	cc.Out.AddRenderer("create"+cmd.Name(), "table", func(writer io.Writer, mobileClient interface{}) error {
		mClient := mobileClient.(*v1alpha1.MobileClient)
		var data [][]string
		data = append(data, []string{mClient.Name, mClient.Spec.Name, mClient.Spec.ClientType, mClient.Spec.AppIdentifier})
		table := tablewriter.NewWriter(writer)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "ClientType", "AppIdentifier"})
		table.Render()
		return nil
	})

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
