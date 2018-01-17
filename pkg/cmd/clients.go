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
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Long: `Example: mobile get clients
mobile --namespace=myproject get clients
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
			data = append(data, []string{mClient.Name, mClient.Spec.Name, mClient.Spec.ClientType})
		}
		table := tablewriter.NewWriter(out)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "ClientType"})
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
		Long: `Example: mobile --namespace=myproject get client <clientID>
kubectl plugin mobile get client <clientID>
oc plugin mobile get client <clientID>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(args, len(args))
			if len(args) != 1 {
				return errors.New("missing argument <clientID>")
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
		data = append(data, []string{mClient.Name, mClient.Namespace, mClient.Spec.Name, mClient.Spec.ClientType, mClient.Spec.ApiKey})
		table := tablewriter.NewWriter(out)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Namespace", "Name", "ClientType", "ApiKey"})
		table.Render()
		return nil
	})
	return command
}

// CreateClientCmd builds the create mobileclient command
func (cc *ClientCmd) CreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client <name> <clientType iOS|cordova|android>",
		Short: "create a mobile client representation in your namespace",
		Long: `Sets up the representation of a mobile application of the specified type in your namespace.
This is used to provide a mobile client context for various actions such as creating, starting or stopping mobile client builds.

The available client types are android, cordova and iOS. 

When used standalone, a namespace must be specified by providing the --namespace flag.`,
		Example: `  mobile create client <name> <clientType> --namespace=myproject 
  kubectl plugin mobile create client <name> <clientType>
  oc plugin mobile create client <name> <clientType>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("expected a name and a clientType")
			}
			name := args[0]
			clientType := args[1]
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
					Labels: map[string]string{},
				},
				Spec: v1alpha1.MobileClientSpec{Name: name, ApiKey: appKey, ClientType: clientType},
			}
			switch app.Spec.ClientType {
			case "android":
				app.Labels["icon"] = "fa-android"
				break
			case "iOS":
				app.Labels["icon"] = "fa-apple"
				break
			case "cordova":
				app.Labels["icon"] = "icon-cordova"
				break
			}
			app.Name = name + "-" + strings.ToLower(app.Spec.ClientType)
			if err := input.ValidateMobileClient(app); err != nil {
				return errors.Wrap(err, "Failed validation while creating new mobile client")
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
		data = append(data, []string{mClient.Name, mClient.Spec.Name, mClient.Spec.ClientType})
		table := tablewriter.NewWriter(writer)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "ClientType"})
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
		Long: `Example: mobile --namespace=myproject delete client <clientID>
kubectl plugin mobile delete client <clientID>
oc plugin mobile delete client <clientID>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var ns string

			if len(args) != 1 {
				return errors.New("expected a clientID argument to be passed " + cmd.Use)
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
