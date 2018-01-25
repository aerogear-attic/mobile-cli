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

	"net/http"

	"io"

	"github.com/pkg/errors"
)

var renderers = map[string]func(writer io.Writer, data interface{}) error{}

//Service represents a serverside application that mobile application will interact with
type Service struct {
	ID           string                         `json:"id"`
	Name         string                         `json:"name"`
	DisplayName  string                         `json:"displayName"`
	Namespace    string                         `json:"namespace"`
	Host         string                         `json:"host"`
	Description  string                         `json:"description"`
	Type         string                         `json:"type"`
	Capabilities map[string][]string            `json:"capabilities"`
	Params       map[string]string              `json:"params"`
	Labels       map[string]string              `json:"labels"`
	Integrations map[string]*ServiceIntegration `json:"integrations"`
	External     bool                           `json:"external"`
	Writable     bool                           `json:"writable"`
}

type ExternalServiceMetaData struct {
	Dependencies        []string `json:"dependencies"`
	DisplayName         string   `json:"displayName"`
	DocumentationURL    string   `json:"documentationUrl"`
	ImageURL            string   `json:"imageUrl"`
	ProviderDisplayName string   `json:"providerDisplayName"`
	ServiceName         string   `json:"serviceName"`
}

type ServiceIntegration struct {
	Enabled         bool   `json:"enabled"`
	Component       string `json:"component"`
	Service         string `json:"service"`
	Namespace       string `json:"namespace"`
	ComponentSecret string `json:"componentSecret"`
	DisplayName     string `json:"displayName"`
}

const (
	ServiceNameKeycloak   = "keycloak"
	ServiceNameThreeScale = "3scale"
	ServiceNameSync       = "fh-sync-server"
	ServiceNameMobileCICD = "aerogear-digger"
	ServiceNameCustom     = "custom"
	IntegrationAPIKeys    = "mcp-mobile-keys"
)

// SecretConvertor converts a kubernetes secret into a mobile.ServiceConfig
type SecretConvertor interface {
	Convert(s *Service) (*ServiceConfig, error)
}

type ServiceConfigs struct {
	Services  []*ServiceConfig `json:"services"`
	Namespace string           `json:"namespace"`
}

type ServiceConfig struct {
	Config map[string]interface{} `json:"config"`
	Name   string                 `json:"name"`
}

// defaultSecretConvertor will provide a default secret to config conversion
type defaultSecretConvertor struct{}

type ignoredFields []string

func (i ignoredFields) Contains(field string) bool {
	for _, f := range i {
		if field == f {
			return true
		}
	}
	return false
}

var ignored = ignoredFields{"password", "token"}

//Convert a kubernetes secret to a mobile.ServiceConfig
func (dsc defaultSecretConvertor) Convert(s *Service) (*ServiceConfig, error) {
	config := map[string]interface{}{}
	headers := map[string]string{}
	config["uri"] = s.Host
	config["name"] = s.Name
	for k, v := range s.Params {
		if !ignored.Contains(k) {
			config[k] = string(v)
		}
	}
	config["headers"] = headers
	return &ServiceConfig{
		Config: config,
		Name:   s.Name,
	}, nil
}

type keycloakSecretConvertor struct{}

//Convert a kubernetes keycloak secret into a keycloak mobile.ServiceConfig
func (ksc keycloakSecretConvertor) Convert(s *Service) (*ServiceConfig, error) {
	config := map[string]interface{}{}
	headers := map[string]string{}
	err := json.Unmarshal([]byte(s.Params["public_installation"]), &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall keycloak configuration ")
	}
	config["headers"] = headers
	return &ServiceConfig{
		Config: config,
		Name:   string(s.Name),
	}, nil
}

type syncSecretConvertor struct{}

//Convert a kubernetes Sync Server secret into a keycloak mobile.ServiceConfig
func (scc syncSecretConvertor) Convert(s *Service) (*ServiceConfig, error) {
	config := map[string]interface{}{
		"uri": s.Host,
	}
	headers := map[string]string{}

	acAppID, acAppIDExists := s.Params["apicast_app_id"]
	acAppKey, acAppKeyExists := s.Params["apicast_app_key"]
	acRoute, acRouteExists := s.Params["apicast_route"]
	if acAppIDExists && acAppKeyExists && acRouteExists {
		headers["app_id"] = acAppID
		headers["app_key"] = acAppKey
		config["uri"] = acRoute
	}
	config["headers"] = headers

	return &ServiceConfig{
		Config: config,
		Name:   s.Name,
	}, nil
}

type SCCInterface interface {
	BindToService(bindableService, targetSvcName string, bindingParams map[string]interface{}, bindableServiceNamespace, targetSvcNamespace string) error
	UnBindFromService(bindableService, targetSvcName, bindableServiceNamespace string) error
	AddMobileApiKeys(targetSvcName, namespace string) error
	RemoveMobileApiKeys(targetSvcName, namespace string) error
}

type ExternalHTTPRequester interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

type IncorrectParameterFormat struct {
	context string
}

func (ip IncorrectParameterFormat) Error() string {
	return "param was incorrect format context: " + ip.context
}

// NewIncorrectParameterFormat returns an error type for incorrectly formatted parameters
func NewIncorrectParameterFormat(context string) IncorrectParameterFormat {
	return IncorrectParameterFormat{context: context}
}

// IsIncorrectParameterFormatErr checks the err to see is it a IncorrectParameterFormat error
func IsIncorrectParameterFormatErr(err error) bool {
	_, ok := err.(IncorrectParameterFormat)
	return ok
}
