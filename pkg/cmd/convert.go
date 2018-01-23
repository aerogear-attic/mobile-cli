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
	"strings"

	v1 "k8s.io/client-go/pkg/api/v1"
)

func convertSecretToMobileService(s v1.Secret) *Service {
	params := map[string]string{}
	for key, value := range s.Data {
		if key != "uri" && key != "name" {
			params[key] = string(value)
		}
	}
	external := s.Labels["external"] == "true"
	return &Service{
		Namespace:    s.Labels["namespace"],
		ID:           s.Name,
		External:     external,
		Labels:       s.Labels,
		Name:         strings.TrimSpace(string(s.Data["name"])),
		DisplayName:  strings.TrimSpace(retrieveDisplayNameFromSecret(s)),
		Type:         strings.TrimSpace(string(s.Data["type"])),
		Host:         string(s.Data["uri"]),
		Params:       params,
		Integrations: map[string]*ServiceIntegration{},
	}
}

// If there is no display name in the secret then we will use the service name
func retrieveDisplayNameFromSecret(sec v1.Secret) string {
	if string(sec.Data["displayName"]) == "" {
		return string(sec.Data["name"])
	}
	return string(sec.Data["displayName"])
}
