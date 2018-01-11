// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"log"

	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

type ClientConfigCmd struct {
	k8Client kubernetes.Interface
}

func NewClientConfigCmd(k8Client kubernetes.Interface) *ClientConfigCmd {
	return &ClientConfigCmd{
		k8Client: k8Client,
	}
}

func (ccc *ClientConfigCmd) GetClientConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clientconfig",
		Short: "get clientconfig returns a client ready filtered configuration of the available services.",
		Long: `get clientconfig
mobile --namespace=myproject get clientconfig
kubectl plugin mobile get clientconfig`,
		Run: func(cmd *cobra.Command, args []string) {
			ret := []*ServiceConfig{}

			convertors := map[string]SecretConvertor{
				"fh-sync-server": &syncSecretConvertor{},
				"keycloak":       &keycloakSecretConvertor{},
			}
			namespace := currentNamespace(cmd.Flags())
			ms := listServices(namespace, ccc.k8Client)
			for _, svc := range ms {
				var svcConifg *ServiceConfig
				var err error
				if _, ok := convertors[svc.Name]; !ok {

					convertor := defaultSecretConvertor{}
					svcConifg, err = convertor.Convert(svc)
					if err != nil {
						//bail out here as now our config may not be compelete?
						log.Fatal("failed to convert to sdk config ", err)
					}
				} else {
					// we can only convert what is available
					convertor := convertors[svc.Name]
					svcConifg, err = convertor.Convert(svc)
					if err != nil {
						//bail out here as now our config may not be compelete?
						log.Fatal("failed to convert to sdkconfig ", err)
					}
				}
				ret = append(ret, svcConifg)
			}

			outputJSON := ServiceConfigs{
				Services: ret,
				Namespace:     namespace,
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(outputJSON); err != nil {
				log.Fatal("failed to encode sdk config ", err)
			}

		},
	}
}
