package integration

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
)

const integrationTestPath = "createIntegrationTestData/"

func TestCreateIntegration(t *testing.T) {
	fhSyncServer := &ProvisionServiceParams{
		Name:      "fh-sync-server",
		Namespace: fmt.Sprintf("--namespace=%s", *namespace),
		Params: []string{
			"-p MONGODB_USER_NAME=fhsync",
			"-p MONGODB_USER_PASSWORD=fhsyncpass",
			"-p MONGODB_ADMIN_PASSWORD=pass",
		},
	}

	keycloak := &ProvisionServiceParams{
		Name:      "keycloak",
		Namespace: fmt.Sprintf("--namespace=%s", *namespace),
		Params: []string{
			"-p ADMIN_NAME=admin",
			"-p ADMIN_PASSWORD=pass",
		},
	}

	// create fh-sync-server service instance
	createInstance(t, fhSyncServer)
	fhSyncID := getInstanceID(t, fhSyncServer)

	// create keycloak service instance
	createInstance(t, keycloak)
	keycloakID := getInstanceID(t, keycloak)

	tests := []struct {
		name       string
		fhSyncID   string
		keycloakID string
		fixture    string
		validate   func(t *testing.T, sb *v1beta1.ServiceBinding)
	}{
		{
			name:       "missing/incorrect arguments",
			fhSyncID:   "",
			keycloakID: "",
			fixture:    "missing-args.golden",
		},
		{
			name:       "create integration returns ready status",
			fhSyncID:   fhSyncID,
			keycloakID: keycloakID,
			validate: func(t *testing.T, sb *v1beta1.ServiceBinding) {
				expectedType := "Ready"
				expectedStatus := "True"

				if actualType := string(sb.Status.Conditions[0].Type); actualType != expectedType {
					t.Fatalf("Expected condition type to be '%s' but got '%s'", expectedType, actualType)
				}

				if actualStatus := string(sb.Status.Conditions[0].Status); actualStatus != expectedStatus {
					t.Fatalf("Expected condition status to be '%s' but got '%s'", expectedStatus, actualStatus)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create service binding
			args := []string{"create", "integration", test.fhSyncID, test.keycloakID, "--namespace=" + *namespace}
			cmd := exec.Command(*executable, args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to create binding: %v", err)
			}
			if *update {
				WriteSnapshot(t, integrationTestPath+test.fixture, output)
			}

			if test.fixture != "" {
				actual := string(output)
				expected := LoadSnapshot(t, integrationTestPath+test.fixture)
				if actual != expected {
					t.Fatalf("actual = \n%s, expected = \n%s", actual, expected)
				}
			}

			sb := getBinding(t)
			if test.validate != nil {
				test.validate(t, sb)
			}
		})
	}
}

func createInstance(t *testing.T, si *ProvisionServiceParams) {
	args := []string{"create", "serviceinstance", si.Name, si.Namespace}
	args = append(args, si.Params...)
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create service instance %s: %v", si.Name, err)
	}

	fmt.Println(string(output))
}

func getInstanceID(t *testing.T, si *ProvisionServiceParams) (id string) {
	args := []string{"get", "serviceinstances", si.Name, si.Namespace, "-o=json"}
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get instances for %s: %v", si.Name, err)
	}

	siList := &v1beta1.ServiceInstanceList{}
	if err = json.Unmarshal(output, siList); err != nil {
		t.Fatal("Unexpected error unmarshalling service instance list:", err)
	}

	return siList.Items[0].ObjectMeta.Name
}

func getBinding(t *testing.T) *v1beta1.ServiceBinding {
	args := []string{"get", "integrations", "--namespace=" + *namespace, "-o=json"}
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal("Failed to get integrations:", err)
	}

	sbList := &v1beta1.ServiceBindingList{}
	if err = json.Unmarshal(output, sbList); err != nil {
		t.Fatal("Unexpected error unmarshalling service bindings list:", err)
	}

	return &sbList.Items[0]
}
