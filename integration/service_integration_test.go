package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
)

const integrationTestPath = "createIntegrationTestData/"

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

	fhSyncServer := &ProvisionServiceParams{
		ServiceName: "fh-sync-server",
		Namespace:   fmt.Sprintf("--namespace=%s", *namespace),
		Params: []string{
			"-p MONGODB_USER_NAME=fhsync",
			"-p MONGODB_USER_PASSWORD=fhsyncpass",
			"-p MONGODB_ADMIN_PASSWORD=pass",
		},
	}

	keycloak := &ProvisionServiceParams{
		ServiceName: "keycloak",
		Namespace:   fmt.Sprintf("--namespace=%s", *namespace),
		Params: []string{
			"-p ADMIN_NAME=admin",
			"-p ADMIN_PASSWORD=pass",
		},
	}

	// create fh-sync-server service instance
	CreateInstance(t, fhSyncServer)
	fhSyncID := GetInstanceID(t, fhSyncServer)

	// create keycloak service instance
	CreateInstance(t, keycloak)
	keycloakID := GetInstanceID(t, keycloak)

	tests := []struct {
		name          string
		fixture       string
		args          []string
		validate      func(t *testing.T)
		expectedError error
	}{
		{
			name:          "create integrations missing arguments",
			fixture:       "missing-args.golden",
			args:          []string{"create", "integration", "", "", "--namespace=" + *namespace},
			expectedError: errors.New("exit status 1"),
		},
		{
			name:          "create integration returns ready status",
			expectedError: nil,
			args:          []string{"create", "integration", fhSyncID, keycloakID, "--namespace=" + *namespace},
			validate: func(t *testing.T) {
				sb, err := GetBinding()
				if err != nil {
					t.Fatalf("error retrieving binding: %v\n", err)
				}

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
		{
			name:          "delete integrations missing arguments",
			fixture:       "missing-args.golden",
			args:          []string{"delete", "integration", "", "", "--namespace=" + *namespace},
			expectedError: errors.New("exit status 1"),
		},
		{
			name:          "delete integration",
			expectedError: nil,
			args:          []string{"delete", "integration", fhSyncID, keycloakID, "--namespace=" + *namespace},
			validate: func(t *testing.T) {
				expectedError := "Could not find servicebindings"
				_, err := GetBinding()
				if err == nil {
					t.Fatalf("expected: '%v', got nil", expectedError)
				} else if err.Error() != expectedError {
					t.Fatalf("expected: '%v', got: '%v'", expectedError, err.Error())
				}

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create service binding
			cmd := exec.Command(*executable, test.args...)
			output, err := cmd.CombinedOutput()
			if err != nil && test.expectedError == nil {
				t.Fatalf("Failed to create binding: '%s', with error: '%s'", output, err.Error())
			}
			if err == nil && test.expectedError != nil {
				t.Fatalf("expected error: '%s', got: nil, output: '%s'", test.expectedError.Error(), output)
			}
			if err != nil && err.Error() != test.expectedError.Error() {
				t.Fatalf("expected error: '%s', got: '%s', output: '%s'", test.expectedError.Error(), err.Error(), output)
			}
			if *update {
				WriteSnapshot(t, integrationTestPath+test.fixture, output)
			}

			if test.fixture != "" {
				actual := strings.TrimSpace(string(output))
				expected := strings.TrimSpace(LoadSnapshot(t, integrationTestPath+test.fixture))
				if actual != expected {
					t.Fatalf("actual = \n'%s', expected = \n'%s'", actual, expected)
				}
			}

			if test.validate != nil {
				test.validate(t)
			}
		})
	}
}

func CreateInstance(t *testing.T, si *ProvisionServiceParams) {
	args := []string{"create", "serviceinstance", si.ServiceName, si.Namespace}
	args = append(args, si.Params...)
	t.Logf("executing: %v\n", *executable+strings.Join(args, " "))
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create service instance '%s': '%s' with error: '%s'", si.ServiceName, output, err)
	}

	fmt.Println(string(output))
}

func GetInstanceID(t *testing.T, si *ProvisionServiceParams) (id string) {
	args := []string{"get", "serviceinstances", si.ServiceName, si.Namespace, "-o=json"}
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get instances for %s: %s", si.ServiceName, output)
	}

	siList := &v1beta1.ServiceInstanceList{}
	if err = json.Unmarshal(output, siList); err != nil {
		t.Fatal("Unexpected error unmarshalling service instance list:", err)
	}

	return siList.Items[0].ObjectMeta.Name
}

func GetBinding() (*v1beta1.ServiceBinding, error) {
	args := []string{"get", "integrations", "--namespace=" + *namespace, "-o=json"}
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	sbList := &v1beta1.ServiceBindingList{}
	if err = json.Unmarshal(output, sbList); err != nil {
		return nil, err
	}

	if len(sbList.Items) == 0 {
		return nil, errors.New("Could not find servicebindings")
	}

	return &sbList.Items[0], nil
}
