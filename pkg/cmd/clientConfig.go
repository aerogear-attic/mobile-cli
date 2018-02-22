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

	mobile "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned"
	sc "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	"github.com/aerogear/mobile-cli/pkg/cmd/output"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClientConfigCmd executes the retrieval and display of the client config
type ClientConfigCmd struct {
	*BaseCmd
	k8Client     kubernetes.Interface
	mobileClient mobile.Interface
	scClient     sc.Interface
	clusterHost  string
}

// NewClientConfigCmd creates and returns a ClientConfigCmd object
func NewClientConfigCmd(k8Client kubernetes.Interface, mobileClient mobile.Interface, scClient sc.Interface, clusterHost string, out io.Writer) *ClientConfigCmd {
	return &ClientConfigCmd{
		k8Client:     k8Client,
		clusterHost:  clusterHost,
		mobileClient: mobileClient,
		scClient:     scClient,
		BaseCmd:      &BaseCmd{Out: output.NewRenderer(out)},
	}
}

// GetClientConfigCmd returns a cobra command object for getting client configs
func (ccc *ClientConfigCmd) GetClientConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientconfig <clientID>",
		Short: "get clientconfig returns a client ready filtered configuration of the available services.",
		Long: `get clientconfig
mobile --namespace=myproject get clientconfig
kubectl plugin mobile get clientconfig`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var ns string
			var err error
			ret := []*ServiceConfig{}
			convertors := map[string]SecretConvertor{
				"fh-sync-server": &syncSecretConvertor{},
				"keycloak":       &keycloakSecretConvertor{},
			}
			if len(args) != 1 {
				return cmd.Usage()
			}
			clientID := args[0]
			ns, err = currentNamespace(cmd.Flags())
			if err != nil {
				return err
			}

			client, err := ccc.mobileClient.MobileV1alpha1().MobileClients(ns).Get(clientID, metav1.GetOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to get mobile client with clientID "+clientID)
			}

			ms := listServices(ns, ccc.k8Client)
			for _, svc := range ms {
				var svcConfig *ServiceConfig
				var err error
				includedService := true
				for _, excluded := range client.Spec.ExcludedServices {
					catalogService, err := ccc.scClient.ServicecatalogV1beta1().ServiceInstances(ns).Get(excluded, v1.GetOptions{})
					if kerror.IsNotFound(err) {
						continue
					} else if err != nil {
						return err
					}
					if strings.TrimSpace(catalogService.Labels["serviceName"]) == strings.TrimSpace(svc.Name) {
						includedService = false
						break
					}
				}
				configMap, err := ccc.k8Client.CoreV1().ConfigMaps(ns).Get(svc.Name, v1.GetOptions{})
				// ignoring not found as some services will not have this public configmap
				if err != nil {
					if kerror.IsNotFound(err) {
						continue
					}
					return errors.Wrap(err, "unable to create config. Failed to get service "+svc.Name+" configmap")
				}
				if _, ok := convertors[svc.Name]; !ok {
					convertor := defaultSecretConvertor{}
					if svcConfig, err = convertor.Convert(svc.ID, configMap.Data); err != nil {
						return err
					}
				} else {
					// we can only convert what is available
					convertor := convertors[svc.Name]
					if svcConfig, err = convertor.Convert(svc.ID, configMap.Data); err != nil {
						return err
					}
				}
				if includedService {
					ret = append(ret, svcConfig)
				}
			}

			outputJSON := ServiceConfigs{
				Version:     1,
				Services:    ret,
				Namespace:   ns,
				ClientID:    clientID,
				ClusterName: ccc.clusterHost,
			}
			if err := ccc.Out.Render("get"+cmd.Name(), outputType(cmd.Flags()), outputJSON); err != nil {
				return errors.Wrap(err, fmt.Sprintf(output.FailedToOutPutInFormat, "ServiceConfig", outputType(cmd.Flags())))
			}
			return nil
		},
	}

	ccc.Out.AddRenderer("get"+cmd.Name(), "table", func(writer io.Writer, serviceConfigs interface{}) error {
		serviceConfigList := serviceConfigs.(ServiceConfigs)
		var data [][]string
		if serviceConfigList.ClientID != "" {
			data = append(data, []string{"Client ID", serviceConfigList.ClientID, "", ""})
		}
		for _, service := range serviceConfigList.Services {
			data = append(data, []string{service.ID, service.Name, service.Type, service.URL})
		}
		table := tablewriter.NewWriter(writer)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.AppendBulk(data)
		table.SetHeader([]string{"ID", "Name", "Type", "URL"})
		table.Render()
		return nil
	})
	return cmd
}
