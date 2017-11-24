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
	"encoding/json"
	"os"

	"log"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO move to the secret data read when discovering the services
//TODO need to come up with a better way of representing this
var capabilities = map[string]map[string][]string{
	"fh-sync-server": map[string][]string{
		"capabilities": {"data storage, data syncronisation"},
		"integrations": {ServiceNameKeycloak, IntegrationAPIKeys, ServiceNameThreeScale},
	},
	"keycloak": map[string][]string{
		"capabilities": {"authentication, authorisation"},
		"integrations": {},
	},
	"mcp-mobile-keys": map[string][]string{
		"capabilities": {"access apps"},
		"integrations": {},
	},
	"3scale": map[string][]string{
		"capabilities": {"authentication, authorization"},
		"integrations": {},
	},
	"custom": map[string][]string{
		"capabilities": {""},
		"integrations": {""},
	},
}

func listServices() []*Service {
	secrets, err := clientset.CoreV1().Secrets(currentNamespace()).List(metav1.ListOptions{LabelSelector: "mobile=enabled"})
	if err != nil {
		log.Fatal("failed to get mobile services. Backing secrets error ", err)
	}
	out := []*Service{}
	for _, s := range secrets.Items {
		out = append(out, convertSecretToMobileService(s))
	}
	return out
}

func getService(serviceName string) *Service {
	secret, err := clientset.CoreV1().Secrets(currentNamespace()).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		log.Fatal("failed to get mobile services. Backing secrets error ", err)
	}
	return convertSecretToMobileService(*secret)
}

// servicesCmd represents the services command
var getServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "get a list of deployed mobile enabled services",
	Run: func(cmd *cobra.Command, args []string) {
		out := listServices()
		enc := json.NewEncoder(os.Stdout)
		for _, s := range out {
			s.Capabilities = capabilities[s.Name]
			//non external services are part of the current namespace //TODO maybe should be added to the apbs
			if s.External == false {
				if s.Namespace == "" {
					s.Namespace = currentNamespace()
				}
				s.Writable = true
			}
			//if s.External {
			//	perm, err := authChecker.Check("deployments", s.Namespace, client)
			//	if err != nil {
			//		return nil, errors.Wrap(err, "error checking access permissions")
			//	}
			//	s.Writable = perm
			//}
		}
		if err := enc.Encode(out); err != nil {
			log.Fatal("failed to json encode output ", err)
		}
	},
}

var getServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "get a mobile aware service definition",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Println("service name is required", cmd.Usage())
			return
		}
		serviceName := args[0]
		if serviceName == "" {
			log.Fatal("service name is required and cannot be empty")
		}
		svc := getService(serviceName)
		enc := json.NewEncoder(os.Stdout)
		svc.Capabilities = capabilities[svc.Type]
		if svc.Capabilities != nil {
			integrations := svc.Capabilities["integrations"]
			for _, v := range integrations {
				isvs := listServices()
				if len(isvs) > 0 {
					is := isvs[0]
					enabled := svc.Labels[is.Name] == "true"
					svc.Integrations[v] = &ServiceIntegration{
						ComponentSecret: svc.ID,
						Component:       svc.Type,
						DisplayName:     is.DisplayName,
						Namespace:       currentNamespace(),
						Service:         is.ID,
						Enabled:         enabled,
					}
				}
			}
		}
		svc.Writable = true
		//if svc.External {
		//	perm, err := authChecker.Check("deployments", svc.Namespace, client)
		//	if err != nil {
		//		return nil, errors.Wrap(err, "error checking access permissions")
		//	}
		//	svc.Writable = perm
		//}
		if err := enc.Encode(svc); err != nil {
			log.Fatal("failed to json encode output ", err)
		}
	},
}

func init() {
	getCmd.AddCommand(getServicesCmd)
	getCmd.AddCommand(getServiceCmd)

	// servicesCmd.PersistentFlags().String("foo", "", "A help for foo")

}
