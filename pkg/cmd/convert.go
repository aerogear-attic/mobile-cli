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
