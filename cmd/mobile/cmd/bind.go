package cmd

import (
	"log"

	sc "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

type BindCmd struct {
	scClient sc.Interface
	k8Client kubernetes.Interface
}

func NewBindCmd(scClient sc.Interface, k8Client kubernetes.Interface) *BindCmd {
	return &BindCmd{scClient: scClient, k8Client: k8Client}
}

func (bc *BindCmd) CreateBindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binding",
		Short: "bind mobile services that integrate together",
		Long: `example usage: kubectl plugin mobile create binding <client_service> <bindable_service>
mobile --namespace=myproject create binding <client_service> <bindable_service>
oc plugin mobile create binding <client_service> <bindable_service>
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing arguments: " + cmd.Use)
			}
			namespace := currentNamespace(cmd.Flags())
			// TODO remove need for this
			sc := serviceCatalogClient{
				k8client: bc.k8Client,
				scClient: bc.scClient,
			}
			from := args[0]
			to := args[1]

			fromSvc := getService(namespace, from, bc.k8Client)
			toSvc := getService(namespace, to, bc.k8Client)
			bindParams := buildBindParams(fromSvc, toSvc)
			if err := sc.BindToService(from, to, bindParams, namespace, namespace); err != nil {
				return errors.Wrap(err, "failed to bind to service ")
			}
			return nil
		},
	}
	return cmd
}

func (bc *BindCmd) DeleteBindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "binding",
		Short: "disintegrate mobile services together",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				log.Fatal("expected a service to bind from and to ")
			}
			from := args[0]
			to := args[1]
			sc := serviceCatalogClient{}
			if err := sc.UnBindFromService(from, to, currentNamespace(cmd.Flags())); err != nil {
				log.Fatal(err)
			}
		},
	}

}

func (bc *BindCmd) GetBindingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binding",
		Short: "get a single binding",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (bc *BindCmd) ListBindingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bindings",
		Short: "get a list of bindings",
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
