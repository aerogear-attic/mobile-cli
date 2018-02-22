package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
)

const deleteServicetestPath = "deleteServiceInstanceTestData/"

func TestDeleteServiceInstance(t *testing.T) {
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
	if err := deleteAllMobileEnabledSecrets(*namespace); err != nil {
		t.Log("failed to clean up secrets")
	}
	defer func() {
		if err := deleteAllMobileEnabledSecrets(*namespace); err != nil {
			t.Log("failed to clean up secrets")
		}
	}()
	createInstance(t, fhSyncServer)
	fhSyncID, err := getInstanceID(fhSyncServer)
	defer func() {
		deleteServiceInstance(t, fhSyncID, *namespace)
	}()

	if err != nil {
		t.Fatalf("failed to get fh-syncID: %v", err)
	}

	tests := []struct {
		name          string
		fixture       string
		args          []string
		validate      func(t *testing.T)
		expectedError error
	}{
		{
			name:          "missing arguments",
			fixture:       "missing-args.golden",
			args:          []string{"delete", "serviceinstance", "", "--namespace=" + *namespace},
			expectedError: errors.New("exit status 1"),
		},
		{
			name:          "Invalid instance ID",
			fixture:       "invalid-instanceId.golden",
			args:          []string{"delete", "serviceinstance", "fdhfgnfgnfg", "--namespace=" + *namespace},
			expectedError: errors.New("exit status 1"),
		},
		{
			name:          "Successful Delete serviceinstance",
			args:          []string{"delete", "serviceinstance", fhSyncID, "--namespace=" + *namespace},
			expectedError: nil,
			validate: func(t *testing.T) {
				expectedStateFound := false
				for i := 0; i < 10; i++ {
					time.Sleep(time.Second * 1)
					instance, err := getInstance(fhSyncServer)
					if err != nil {
						t.Fatalf("unexpected error %v", err)
					}
					for _, condition := range instance.Status.Conditions {
						if condition.Type == "Ready" && condition.Status == "False" {
							expectedStateFound = true
							break
						}
					}
				}
				if !expectedStateFound {
					t.Fatalf("expected to find the serviceinstance in a non ready state but it remained ready")
				}

				secrets := &v1.SecretList{}
				if err = getResource("secrets", *namespace, secrets); err != nil {
					t.Fatalf("unexpected error when getting secret %v", err)
				}
				for _, secret := range secrets.Items {
					if strings.Contains(secret.ObjectMeta.Name, "fh-sync-server-apb-params") {

						t.Fatalf("Expected secret to be removed but got: %v secret", secret.ObjectMeta.Name)
					}
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

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
			if test.fixture != "" {
				actual := strings.TrimSpace(string(output))
				expected := strings.TrimSpace(LoadSnapshot(t, deleteServicetestPath+test.fixture))
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

func createInstance(t *testing.T, si *ProvisionServiceParams) {
	args := []string{"create", "serviceinstance", si.ServiceName, si.Namespace}
	args = append(args, si.Params...)
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create service instance %s: %v, with output %v", si.ServiceName, err, string(output))
	}
}

func deleteServiceInstance(t *testing.T, sid, namespace string) {
	args := []string{"delete", "serviceinstance", sid, "-n=" + namespace}
	cmd := exec.Command("oc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to delete service instance %s: %v, with output %v", sid, err, string(output))
	}
}

func deleteResource(resourceType, name, namespace string) error {
	args := []string{"delete", resourceType, name, "-n=" + namespace}
	cmd := exec.Command("oc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v error output %s", err, string(output))
	}
	return nil
}

func deleteAllMobileEnabledSecrets(namespace string) error {
	args := []string{"delete", "secret", "-l", "mobile=enabled", "-n=" + namespace}
	cmd := exec.Command("oc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v error output %s", err, string(output))
	}
	return nil
}

func getInstance(si *ProvisionServiceParams) (instance *v1beta1.ServiceInstance, err error) {
	args := []string{"get", "serviceinstances", si.ServiceName, si.Namespace, "-o=json"}
	cmd := exec.Command(*executable, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	siList := &v1beta1.ServiceInstanceList{}
	if err = json.Unmarshal(output, siList); err != nil {
		return nil, err
	}

	if len(siList.Items) == 0 {
		return nil, errors.New("no matching instances found")
	}
	return &siList.Items[0], nil
}

func getInstanceID(si *ProvisionServiceParams) (id string, err error) {
	instance, err := getInstance(si)
	if err != nil {
		return "", err
	}
	return instance.ObjectMeta.Name, nil
}

func getResource(resourceType, namespace string, dest interface{}) (err error) {
	args := []string{"get", resourceType, "-n=" + namespace, "-o=json"}
	cmd := exec.Command("oc", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v error output %s", err, string(output))
	}
	return json.Unmarshal(output, dest)
}
