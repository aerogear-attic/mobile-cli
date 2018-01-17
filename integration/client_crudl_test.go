package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type MobileClientSpec struct {
	Name       string
	ApiKey     string
	ClientType string
}

type MobileClientJson struct {
	Spec MobileClientSpec
}

func VMobileClientJson(name string, clientType string) func(output []byte, err error) (bool, []string) {
	return func(output []byte, err error) (bool, []string) {
		var parsed MobileClientJson
		errJson := json.Unmarshal([]byte(output), &parsed)
		if errJson != nil {
			return false, []string{fmt.Sprintf("%s", err)}
		}
		if parsed.Spec.ClientType != clientType {
			return false, []string{fmt.Sprintf("Expected the ClientType to be %s, but got %s", clientType, parsed.Spec.ClientType)}
		}
		if parsed.Spec.Name != name {
			return false, []string{fmt.Sprintf("Expected the Name to be %s, but got %s", name, parsed.Spec.Name)}
		}
		return true, []string{}
	}
}

func TestClientJson(t *testing.T) {

	clientTypes := []string{
		"cordova",
		"iOS",
		"android",
	}

	name := fmt.Sprintf("%s-mobile-crud-test-entity", *prefix)
	m := ValidatedCmd(*executable, fmt.Sprintf("--namespace=%s", *namespace), "-o=json")
	oc := ValidatedCmd("oc", fmt.Sprintf("--namespace=%s", *namespace), "-o=json")
	for _, clientType := range clientTypes {
		t.Run(clientType, func(t *testing.T) {
			expectedId := strings.ToLower(fmt.Sprintf("%s-%s", name, clientType))

			notExists := All(VIsErr, VRegex(fmt.Sprintf(".*\"%s\" not found.*", expectedId)))
			exists := All(VNoErr, VMobileClientJson(name, clientType))

			m.Add("get", "client", expectedId).Complying(notExists).Run(t)
			oc.Add("get", "mobileclient", expectedId).Complying(notExists).Run(t)
			m.Add("create", "client", name, clientType).Complying(exists).Run(t)
			m.Add("get", "client", expectedId).Complying(exists).Run(t)
			oc.Add("get", "mobileclient", expectedId).Complying(exists).Run(t)
			m.Add("delete", "client", expectedId).Complying(VNoErr).Run(t)
			m.Add("get", "client", expectedId).Complying(notExists).Run(t)
			oc.Add("get", "mobileclient", expectedId).Complying(notExists).Run(t)
		})
	}
}
