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
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/client-go/pkg/api/v1"
	"net/url"
	"strings"
)

func isClientConfigKey(key string) bool {
	return key == "url" || key == "name" || key == "type" || key == "id"
}

func retrieveHTTPSConfigForServices(svcConfigs []*ServiceConfig, includeInvalidCerts bool) ([]*CertificatePinningHash, error) {
	httpsConfig := make([]*CertificatePinningHash, 0)
	for _, svc := range svcConfigs {
		pinningHash, err := retrieveHTTPSConfigForService(svc, includeInvalidCerts)
		if err != nil {
			return nil, err
		}
		if pinningHash != nil {
			httpsConfig = append(httpsConfig, pinningHash)
		}
	}
	return httpsConfig, nil
}

func retrieveHTTPSConfigForService(svcConfig *ServiceConfig, allowInvalidCert bool) (*CertificatePinningHash, error) {
	// Parse the services URL, if it's not HTTPS then don't attempt to retrieve a cert for it.
	serviceURL, err := url.Parse(svcConfig.URL)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse service URL "+svcConfig.URL)
	}
	if serviceURL.Scheme != "https" {
		return nil, nil
	}

	certificate, err := retrieveCertificateForURL(serviceURL, allowInvalidCert)
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve certificate for service URL "+serviceURL.String())
	}

	hasher := sha256.New()
	_, err = hasher.Write(certificate.RawSubjectPublicKeyInfo)
	if err != nil {
		return nil, errors.Wrap(err, "Could not write public key to buffer for hashing")
	}
	pinningHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	return &CertificatePinningHash{serviceURL.Host, pinningHash}, nil
}

func retrieveCertificateForURL(url *url.URL, allowInvalidCert bool) (*x509.Certificate, error) {
	// If the 443 port is not appended to the URLs host then we should append it or tls.Dial will fail.
	port := "443"
	if url.Port() != "" {
		port = url.Port()
	}
	hostURL := fmt.Sprintf("%s:%s", url.Host, port)

	conn, err := tls.Dial("tcp", hostURL, &tls.Config{
		InsecureSkipVerify: allowInvalidCert,
	})

	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve certificate for URL "+url.String())
	}
	return conn.ConnectionState().PeerCertificates[0], nil
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
