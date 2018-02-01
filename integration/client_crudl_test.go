package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func ValidMobileClientJson(name, clientType, appIdentifier string) func(output []byte, err error) ValidationResult {
	return func(output []byte, err error) ValidationResult {
		var parsed MobileClientJSON
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			return FailureValidation(output, err)
		}
		if parsed.Spec.ClientType != clientType {
			return ValidationResult{
				Success: false,
				Message: []string{fmt.Sprintf("Expected the ClientType to be %s, but got %s", clientType, parsed.Spec.ClientType)},
				Error:   err,
				Output:  output,
			}
		}
		if parsed.Spec.AppIdentifier != appIdentifier {
			return ValidationResult{
				Success: false,
				Message: []string{fmt.Sprintf("Expected the AppIdentifier to be %s, but got %s", appIdentifier, parsed.Spec.AppIdentifier)},
				Error:   err,
				Output:  output,
			}
		}
		if parsed.Spec.Name != name {
			return ValidationResult{
				Success: false,
				Message: []string{fmt.Sprintf("Expected the Name to be %s, but got %s", name, parsed.Spec.Name)},
				Error:   err,
				Output:  output,
			}
		}
		return SuccessValidation(output, err)
	}
}

func TestClientJson(t *testing.T) {

	clientTypes := []string{
		"cordova",
		"iOS",
		"android",
	}

	name := fmt.Sprintf("%s-mobile-crud-test-entity", *prefix)
	appIdentifier := "my.app.org"
	m := ValidatedCmd(*executable, fmt.Sprintf("--namespace=%s", *namespace), "-o=json")
	o := ValidatedCmd("oc", fmt.Sprintf("--namespace=%s", *namespace), "-o=json")
	for _, clientType := range clientTypes {
		t.Run(clientType, func(t *testing.T) {
			expectedId := strings.ToLower(fmt.Sprintf("%s-%s", name, clientType))

			notExists := All(IsErr, ValidRegex(fmt.Sprintf(".*\"%s\" not found.*", expectedId)))
			exists := All(NoErr, ValidMobileClientJson(name, clientType, appIdentifier))

			m.Args("get", "client", expectedId).Should(notExists).Run().Test(t)
			o.Args("get", "mobileclient", expectedId).Should(notExists).Run().Test(t)
			m.Args("create", "client", name, clientType, appIdentifier).Should(exists).Run().Test(t)
			m.Args("get", "client", expectedId).Should(exists).Run().Test(t)
			o.Args("get", "mobileclient", expectedId).Should(exists).Run().Test(t)
			m.Args("delete", "client", expectedId).Should(NoErr).Run().Test(t)
			m.Args("get", "client", expectedId).Should(notExists).Run().Test(t)
			o.Args("get", "mobileclient", expectedId).Should(notExists).Run().Test(t)
		})
	}
}
