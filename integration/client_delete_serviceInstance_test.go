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
	appsv1 "k8s.io/client-go/pkg/apis/apps/v1beta1"
	extv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const deleteServicetestPath = "deleteServiceInstanceTestData/"

func TestDeleteServiceInstance(t *testing.T) {

	fhSyncServer := &ProvisionServiceParams{
		ServiceName: "fh-sync-server",
		Namespace:   fmt.Sprintf("--namespace=%s", *namespace),
		Params: []string{
			"-p MONGODB_USER_NAME=fhsync",
			"-p MONGODB_USER_PASSWORD=fhsyncpass",
			"-p MONGODB_ADMIN_PASSWORD=pass",
		},
	}

	createInstance(t, fhSyncServer)
	fhSyncID, err := getInstanceID(fhSyncServer)

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
			name:          "Successful Delete",
			args:          []string{"delete", "serviceinstance", fhSyncID, "--namespace=" + *namespace},
			expectedError: nil,
			validate: func(t *testing.T) {
				time.Sleep(time.Second * 10)
				instance, err := getInstance(fhSyncServer)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				for _, condition := range instance.Status.Conditions {
					if condition.Type == "Ready" && condition.Status != "False" {
						t.Fatalf("Expected ready condition to be false and got %v", condition.Status)

					}
				}

				pods := &v1.PodList{}
				if err = getResource("pods", pods); err != nil {
					t.Fatalf("unexpected error %v", err)

				}
				for _, pod := range pods.Items {
					if strings.Contains(pod.ObjectMeta.Name, "fh-sync-server") {
						for _, condition := range pod.Status.Conditions {
							if condition.Type == "Ready" && condition.Status != "False" {
								t.Fatalf("Expected fh-sync-server pod (%v) ready condition to be false and got %v", pod.ObjectMeta.Name, condition.Status)
							}
						}
					}
				}

				deployments := &appsv1.DeploymentList{}
				if err = getResource("deployments", deployments); err != nil {
					t.Fatalf("unexpected error %v", err)

				}
				for _, dp := range deployments.Items {
					if strings.Contains(dp.ObjectMeta.Name, "fh-sync-server") {
						for _, condition := range dp.Status.Conditions {
							if condition.Type == "Available" && condition.Status != "False" {
								t.Fatalf("Expected fh-sync-server deployment (%v) available condition to be false and got %v", dp.ObjectMeta.Name, condition.Status)
							}
						}
					}
				}

				routes := &extv1.IngressList{}
				if err = getResource("routes", routes); err != nil {
					t.Fatalf("unexpected error %v", err)

				}
				for _, route := range routes.Items {
					if strings.Contains(route.ObjectMeta.Name, "fh-sync-server") {
						t.Fatalf("Expected route to be removed but got: %v route", route.ObjectMeta.Name)
					}
				}

				configmaps := &v1.ConfigMapList{}
				fmt.Println(configmaps)
				if err = getResource("configmaps", configmaps); err != nil {
					t.Fatalf("unexpected error %v", err)

				}
				for _, cm := range configmaps.Items {
					if strings.Contains(cm.ObjectMeta.Name, "fh-sync-server") {
						t.Fatalf("Expected configmap to be removed but got: %v configmap", cm.ObjectMeta.Name)
					}
				}

				secrets := &v1.SecretList{}
				if err = getResource("secrets", secrets); err != nil {
					t.Fatalf("unexpected error %v", err)
				}

				for _, secret := range secrets.Items {
					if strings.Contains(secret.ObjectMeta.Name, "fh-sync-server") {

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
		t.Fatalf("Failed to create service instance %s: %v, with output %v", si.ServiceName, err, output)
	}
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

	return &siList.Items[0], nil
}

func getInstanceID(si *ProvisionServiceParams) (id string, err error) {
	instance, err := getInstance(si)
	if err != nil {
		return "", err
	}
	return instance.ObjectMeta.Name, nil
}

func getResource(resourceType string, dest interface{}) (err error) {
	args := []string{"get", resourceType, "-o=json"}
	cmd := exec.Command("oc", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return json.Unmarshal(output, dest)
}
