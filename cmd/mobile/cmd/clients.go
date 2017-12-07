package cmd

import (
	"os"

	"fmt"
	"time"

	"github.com/aerogear/mobile-cli/pkg/apis/mobile/v1alpha1"
	mobile "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned"
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
func NewClientCmd(mobileClient mobile.Interface) *ClientCmd {
	return &ClientCmd{mobileClient: mobileClient, BaseCmd: &BaseCmd{Out: OutPutFactory{out: os.Stdout}}}
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
			apps, err := cc.mobileClient.MobileV1alpha1().MobileClients(currentNamespace(cmd.Flags())).List(metav1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to list mobile clients")
			}
			outType := outputType(cmd.Flags())
			if err := cc.Out.Output(cmd.Name(), outType, apps); err != nil {
				return errors.Wrap(err, fmt.Sprintf(failedToOutPutInFormat, "mobile clients", outType))
			}
			return nil
		},
	}

	templates[command.Name()] = `{{range $k,$v := .Items}}Name: {{$v.ObjectMeta.Name}} ID: {{$v.ObjectMeta.Name}} Type: {{$v.Spec.ClientType}}

{{end}}`

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
			if len(args) != 1 {
				return errors.New("missing argument <clientID>")
			}
			clientID := args[0]
			client, err := cc.mobileClient.MobileV1alpha1().MobileClients(currentNamespace(cmd.Flags())).Get(clientID, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get mobile client with clientID "+clientID)
			}
			outType := outputType(cmd.Flags())
			if err := cc.Out.Output(cmd.Name(), outType, client); err != nil {
				return errors.Wrap(err, fmt.Sprintf(failedToOutPutInFormat, "mobile client", outType))
			}
			return nil
		},
	}
	templates[command.Name()] = `
Name: {{.Spec.Name}}  ID: {{.ObjectMeta.Name}}  ClientType: {{.Spec.ClientType}}
`
	return command
}

// CreateClientCmd builds the create mobileclient command
func (cc *ClientCmd) CreateClientCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "client <name> <clientType iOS|cordova|android>",
		Short: "create a mobile client representation in your namespace",
		Long: `Sets up the representation of a mobile application in your namespace. This used to provide a mobile client context for various
actions such as doing mobile client builds,
mobile --namespace=myproject create client <name> <clientType>
kubectl plugin mobile create client <name> <clientType>
oc plugin mobile create client <name> <clientType>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("expected a name and clientType " + cmd.Use)
			}
			name := args[0]
			clientType := args[1]
			appKey := uuid.NewV4().String()
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
			app.Name = name + "-" + fmt.Sprintf("%v", time.Now().Unix())
			if err := validateMobileClient(app); err != nil {
				return errors.Wrap(err, "Failed validation while creating new mobile client")
			}
			createdClient, err := cc.mobileClient.MobileV1alpha1().MobileClients(currentNamespace(cmd.Flags())).Create(app)
			if err != nil {
				return errors.Wrap(err, "failed to create mobile client")
			}
			outType := outputType(cmd.Flags())
			if err := cc.Out.Output(cmd.Name(), outType, createdClient); err != nil {
				return errors.Wrap(err, fmt.Sprintf(failedToOutPutInFormat, "mobile client", outType))
			}
			return nil
		},
	}
	templates[command.Name()] = `Name: {{.Spec.Name}} | Type: {{.Spec.ClientType}}
`
	return command
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
			if len(args) != 1 {
				return errors.New("expected a clientID argument to be passed " + cmd.Use)
			}
			clientID := args[0]
			err := cc.mobileClient.MobileV1alpha1().MobileClients(currentNamespace(cmd.Flags())).Delete(clientID, &metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get mobile client with clientID "+clientID)
			}
			return nil
		},
	}
	return command
}
