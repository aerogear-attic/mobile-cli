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
	"k8s.io/client-go/pkg/api/v1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"net/url"
)

func isClientConfigKey(key string) bool {
	return key == "url" || key == "name" || key == "type" || key == "id"
}

func appendCertificatePinningInfoToService(s *ServiceConfig) error {
	serviceURL, err := url.Parse(s.URL)
	if err != nil {
		return err
	}
	if serviceURL.Scheme != "https" {
		return nil
	}
	// TODO: Make the InsecureSkipVerify here configurable. I think there will be times when we don't want to allow auto-pinning to unverified certificates.
	// TODO: Allow for the Host variable to contain a port. So split it and then if there's a port use that, else use 443.
	conn, err := tls.Dial("tcp", serviceURL.Host+":443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	hasher := sha256.New()
	// TODO: Do we want to loop through here? The command here https://developer.mozilla.org/en-US/docs/Web/HTTP/Public_Key_Pinning only returns what we are currently returning.
	hasher.Write(conn.ConnectionState().PeerCertificates[0].RawSubjectPublicKeyInfo)
	s.CertificatePinningHashes = []string{base64.StdEncoding.EncodeToString(hasher.Sum(nil))}
	return nil
}

func convertSecretToMobileService(s v1.Secret) *Service {
	params := map[string]string{}
	for key, value := range s.Data {
		if !isClientConfigKey(key) {
			params[key] = string(value)
		}
	}
	external := s.Labels["external"] == "true"

	return &Service{
		Namespace:    s.Labels["namespace"],
		ID:           s.Name,
		External:     external,
		Labels:       s.Labels,
		Name:         s.Name,
		DisplayName:  strings.TrimSpace(retrieveDisplayNameFromSecret(s)),
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
