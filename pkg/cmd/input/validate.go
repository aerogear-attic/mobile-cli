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

package input

import (
	"strings"

	"github.com/aerogear/mobile-cli/pkg/apis/mobile/v1alpha1"
	"github.com/pkg/errors"
)

func ValidateMobileClient(client *v1alpha1.MobileClient) error {
	if !ValidClients.Contains(client.Spec.ClientType) {
		return errors.New("invalid clientType " + client.Spec.ClientType + " valid clientTypes are " + strings.Join(ValidClients, ","))
	}
	return nil
}

type validClients []string

func (vc validClients) Contains(client string) bool {
	for _, c := range vc {
		if c == client {
			return true
		}
	}
	return false
}

var ValidClients = validClients{"iOS", "android", "cordova"}
