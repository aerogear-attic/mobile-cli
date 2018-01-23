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
	"os"

	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

func listServices(ns string, k8Client kubernetes.Interface) []*Service {
	secrets, err := k8Client.CoreV1().Secrets(ns).List(metav1.ListOptions{LabelSelector: "mobile=enabled"})
	if err != nil {
		log.Fatal("failed to get mobile services. Backing secrets error ", err)
	}
	out := []*Service{}
	for _, s := range secrets.Items {
		out = append(out, convertSecretToMobileService(s))
	}
	return out
}

func getService(ns, serviceName string, k8Client kubernetes.Interface) *Service {
	secret, err := k8Client.CoreV1().Secrets(ns).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		log.Fatal("failed to get mobile services. Backing secrets error ", err)
	}
	return convertSecretToMobileService(*secret)
}

type ServiceConfigCmd struct {
	k8client kubernetes.Interface
}

func NewServiceConfigCommand(k8client kubernetes.Interface) *ServiceConfigCmd {
	return &ServiceConfigCmd{
		k8client: k8client,
	}
}

func (scc *ServiceConfigCmd) ListServiceConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serviceconfigs",
		Short: "get a list of deployed mobile enabled services",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			out := listServices(namespace, scc.k8client)
			enc := json.NewEncoder(os.Stdout)
			for _, s := range out {
				s.Capabilities = capabilities[s.Name]
				//non external services are part of the current namespace //TODO maybe should be added to the apbs
				if s.External == false {
					if s.Namespace == "" {
						s.Namespace = namespace
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
			return nil
		},
	}
}

func (scc *ServiceConfigCmd) GetServiceConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serviceconfig",
		Short: "get a mobile aware service definition",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				//log.Println(")
				return errors.Errorf("%v\n%v", "service name is required", cmd.Usage())
			}
			serviceName := args[0]
			if serviceName == "" {
				log.Fatal("service name is required and cannot be empty")
			}
			namespace, err := currentNamespace(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to get namespace")
			}
			svc := getService(namespace, serviceName, scc.k8client)
			enc := json.NewEncoder(os.Stdout)
			svc.Capabilities = capabilities[svc.Type]
			if svc.Capabilities != nil {
				integrations := svc.Capabilities["integrations"]
				for _, v := range integrations {
					isvs := listServices(namespace, scc.k8client)
					if len(isvs) > 0 {
						is := isvs[0]
						enabled := svc.Labels[is.Name] == "true"
						svc.Integrations[v] = &ServiceIntegration{
							ComponentSecret: svc.ID,
							Component:       svc.Type,
							DisplayName:     is.DisplayName,
							Namespace:       namespace,
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
			return nil
		},
	}
}

func (scc *ServiceConfigCmd) CreateServiceConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceconfig",
		Short: "create a new service config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (scc *ServiceConfigCmd) DeleteServiceConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serviceconfig",
		Short: "delete a service config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
