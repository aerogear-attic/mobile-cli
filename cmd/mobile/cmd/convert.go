package cmd

import (
	"strings"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func convertConfigMapToMobileApp(m v1.ConfigMap) *App {
	return &App{
		ID:          m.Name,
		Name:        m.Data["name"],
		DisplayName: m.Data["displayName"],
		ClientType:  m.Data["clientType"],
		APIKey:      m.Data["apiKey"],
		Labels:      m.Labels,
		Description: m.Data["description"],
		MetaData: map[string]string{
			"icon":    m.Annotations["icon"],
			"created": m.Annotations["created"],
		},
	}
}
func convertMobileAppToConfigMap(app *App) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: app.ID,
			Labels: map[string]string{
				"group": "mobileapp",
				"name":  app.Name,
			},
			Annotations: map[string]string{
				"icon":    app.MetaData["icon"],
				"created": app.MetaData["created"],
			},
		},
		Data: map[string]string{
			"name":        app.Name,
			"displayName": app.DisplayName,
			"clientType":  app.ClientType,
			"apiKey":      app.APIKey,
			"description": app.Description,
		},
	}
}

func convertMobileServiceToSecret(ms *Service) *v1.Secret {
	data := map[string][]byte{}
	labels := map[string]string{
		"group":     "mobile",
		"namespace": ms.Namespace,
	}
	for k, v := range ms.Labels {
		labels[k] = v
	}
	data["uri"] = []byte(ms.Host)
	data["name"] = []byte(ms.Name)
	data["displayName"] = []byte(ms.DisplayName)
	data["type"] = []byte(ms.Type)
	for k, v := range ms.Params {
		data[k] = []byte(v)
	}
	return &v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Labels: labels,
			Name:   ms.ID,
		},
		Data: data,
	}
}

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
