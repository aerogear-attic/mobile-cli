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
