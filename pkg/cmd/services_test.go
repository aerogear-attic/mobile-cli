package cmd_test

import (
	"bytes"
	"testing"

	"encoding/json"

	"fmt"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	scFake "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/aerogear/mobile-cli/pkg/cmd"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestServicesCmd_DeleteServiceInstanceCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ExpectError      bool
		ValidateOutput   func(t *testing.T, output []byte)
		ValidateErr      func(t *testing.T, err error)
		Flags            []string
		Args             []string
	}{

		{
			Name: "test if no service instance id passed that error returned",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected an error but did not get one")
				}
				if err.Error() != "expected a serviceInstanceID" {
					t.Fatalf("expected error to be %s but got %v", "expected a serviceInstanceID", err)
				}
			},
			Flags: []string{"--namespace=test", "-o=json"},
			Args:  []string{},
		},

		{
			Name: "test if error occurs getting service instance that an error is returned",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("error in get")
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected an error but did not get one")
				}
				if err.Error() != "error in get" {
					t.Fatalf("expected error to be %s but got %v", "error in get", err)
				}
			},
			Flags: []string{"--namespace=test", "-o=json"},
			Args:  []string{"someid"},
		},
		{
			Name: "test if error occurs deleting service instance that an error is returned",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, nil
				})
				fake.AddReactor("delete", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("error in delete")
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected an error but did not get one")
				}
				if err.Error() != "error in delete" {
					t.Fatalf("expected error to be %s but got %v", "error in delete", err)
				}
			},
			Flags: []string{"--namespace=test", "-o=json"},
			Args:  []string{"someid"},
		},
		{
			Name: "test successful delete",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{GenerateName: "test"},
					}, nil
				})
				fake.AddReactor("delete", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: false,
			ValidateErr: func(t *testing.T, err error) {
				if err != nil {
					t.Fatalf("expected no error but got %v", err)
				}
			},
			Flags: []string{"--namespace=test", "-o=json"},
			Args:  []string{"someid"},
		},

		{
			Name: "test error on missing namespace",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected an error but didn't got one")
				}
				if err.Error() != "failed to get namespace: no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var" {
					t.Fatalf("Expected error 'failed to get namespace: no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var' but got %v", err)
				}
			},
			Flags: []string{"-o=json"},
			Args:  []string{"someid"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var out bytes.Buffer
			root := cmd.NewRootCmd()
			deleteClient := cmd.NewDeleteComand()
			serviceCmd := cmd.NewServicesCmd(tc.SvcCatalogClient(), tc.K8Client(), &out)
			deleteServiceInstCmd := serviceCmd.DeleteServiceInstanceCmd()
			deleteClient.AddCommand(deleteServiceInstCmd)
			root.AddCommand(deleteClient)
			if err := deleteServiceInstCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse flags ", err)
			}
			err := deleteServiceInstCmd.RunE(deleteServiceInstCmd, tc.Args)
			if err != nil && !tc.ExpectError {
				t.Fatal("did not expect an error but gone one ", err)
			}
			if err == nil && tc.ExpectError {
				t.Fatal("expected an error but got none")
			}
			if tc.ValidateOutput != nil {
				tc.ValidateOutput(t, out.Bytes())
			}
			if tc.ValidateErr != nil {
				tc.ValidateErr(t, err)
			}
		})
	}
}

func TestServicesCmd_ListServicesCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ExpectError      bool
		Validate         func(t *testing.T, output []byte)
		Flags            []string
	}{
		{
			Name:  "test list services as expected",
			Flags: []string{"-o=json"},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ClusterServiceClassList{
						Items: []v1beta1.ClusterServiceClass{
							{
								Spec: v1beta1.ClusterServiceClassSpec{
									Tags: []string{"mobile-service", "other"},
								},
							},
							{
								Spec: v1beta1.ClusterServiceClassSpec{
									Tags: []string{},
								},
							},
						},
					}, nil
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			Validate: func(t *testing.T, data []byte) {
				var list = &v1beta1.ClusterServiceClassList{}
				if err := json.Unmarshal(data, list); err != nil {
					t.Fatal("failed to unmarshal data", err)
				}

				if nil == list {
					t.Fatal("expected a list but got nil")
				}
				if len(list.Items) != 1 {
					t.Fatalf("expected only one item in the list but got %v", len(list.Items))
				}
			},
		},
		{
			Name:  "test list services returns error on failure",
			Flags: []string{"-o=json"},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed for some reason")
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var out bytes.Buffer
			serviceCmd := cmd.NewServicesCmd(tc.SvcCatalogClient(), tc.K8Client(), &out)
			listCmd := serviceCmd.ListServicesCmd()
			err := listCmd.RunE(listCmd, tc.Flags)
			if err != nil && !tc.ExpectError {
				t.Fatal("did not expect an error but gone one ", err)
			}
			if err == nil && tc.ExpectError {
				t.Fatal("expected an error but got none")
			}
			if tc.Validate != nil {
				tc.Validate(t, out.Bytes())
			}

		})
	}
}

func TestServicesCmd_CreateServiceInstanceCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ValidateErr      func(t *testing.T, err error)
		ValidateOut      func(t *testing.T, output []byte)
		Args             []string
		Flags            []string
	}{
		{
			Name: "should fail when no service class found",
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to find serviceclass with name: keycloak"
				if err == nil {
					t.Fatal("expected an error but got none")
				}
				if err.Error() != expectedErr {
					t.Fatalf("expected the error to be %s but was %v", expectedErr, err)
				}
			},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			Flags: []string{"--namespace=test"},
			Args:  []string{"keycloak"},
		},
		{
			Name: "should fail when no service plan found",
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to find serviceplan associated with the serviceclass test"
				if err == nil {
					t.Fatal("expected an error but got none")
				}
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be %s but got %v", expectedErr, err)
				}
			},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					externalData := cmd.ExternalServiceMetaData{
						ServiceName: "keycloak",
					}
					data, _ := json.Marshal(externalData)
					return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "test"},
							Spec: v1beta1.ClusterServiceClassSpec{
								ExternalMetadata: &runtime.RawExtension{Raw: data},
							},
						},
					}}, nil
				})
				fakeClient.AddReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {

					return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{{}}}, nil
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			Flags: []string{"--namespace=test"},
			Args:  []string{"keycloak"},
		},
		{
			Name: "should fail when missing param and exit without waiting",
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "missing required parameter ADMIN_NAME"
				if err == nil {
					t.Fatal("expected an error but got none")
				}
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be %s but got %v", expectedErr, err)
				}
			},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					externalData := cmd.ExternalServiceMetaData{
						ServiceName: "keycloak",
					}
					data, _ := json.Marshal(externalData)
					return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "test"},
							Spec: v1beta1.ClusterServiceClassSpec{
								ExternalMetadata: &runtime.RawExtension{Raw: data},
							},
						},
					}}, nil
				})
				fakeClient.AddReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					params := &cmd.InstanceCreateParams{Required: []string{"ADMIN_NAME"}, Properties: map[string]map[string]interface{}{"ADMIN_NAME": {"value": ""}}}
					b, _ := json.Marshal(params)
					return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{{
						Spec: v1beta1.ClusterServicePlanSpec{ServiceInstanceCreateParameterSchema: &runtime.RawExtension{Raw: b}, ClusterServiceClassRef: v1beta1.ClusterObjectReference{Name: "test"}, ExternalName: "default"},
					},
					},
					}, nil
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			Flags: []string{"--namespace=test", "-ptest=test", "--no-wait=true"},
			Args:  []string{"keycloak"},
		},
		{
			Name: "should not fail when finds service class service plan and all params present and no wait flag set",
			ValidateErr: func(t *testing.T, err error) {
				if err != nil {
					t.Fatalf("did not expect an error but got one %v ", err)
				}
			},
			SvcCatalogClient: func() versioned.Interface {
				fakeClient := &scFake.Clientset{}
				fakeClient.AddReactor("list", "clusterserviceclasses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					externalData := cmd.ExternalServiceMetaData{
						ServiceName: "keycloak",
					}
					data, _ := json.Marshal(externalData)
					return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "test"},
							Spec: v1beta1.ClusterServiceClassSpec{
								ExternalMetadata: &runtime.RawExtension{Raw: data},
							},
						},
					}}, nil
				})
				fakeClient.AddReactor("list", "clusterserviceplans", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					params := &cmd.InstanceCreateParams{Required: []string{"ADMIN_NAME"}, Properties: map[string]map[string]interface{}{"ADMIN_NAME": {"value": "", "default": "admin"}}}
					b, _ := json.Marshal(params)
					return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{{
						Spec: v1beta1.ClusterServicePlanSpec{ServiceInstanceCreateParameterSchema: &runtime.RawExtension{Raw: b}, ClusterServiceClassRef: v1beta1.ClusterObjectReference{Name: "test"}, ExternalName: "default"},
					},
					},
					}, nil
				})
				return fakeClient
			},
			K8Client: func() kubernetes.Interface {
				fakeClient := &kFake.Clientset{}
				return fakeClient
			},
			Flags: []string{"--namespace=test", "-pADMIN_NAME=test", "-pADMIN_PASSWORD=test", "--no-wait=true"},
			Args:  []string{"keycloak"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var out bytes.Buffer
			//need root cmd to allow parsing shared flags
			root := cmd.NewRootCmd()
			serviceCmd := cmd.NewServicesCmd(tc.SvcCatalogClient(), tc.K8Client(), &out)
			createCmd := serviceCmd.CreateServiceInstanceCmd()
			root.AddCommand(createCmd)
			if err := createCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse command flags", err)
			}
			err := createCmd.RunE(createCmd, tc.Args)
			if tc.ValidateErr != nil {
				tc.ValidateErr(t, err)
			}

			if tc.ValidateOut != nil {
				tc.ValidateOut(t, out.Bytes())
			}

		})
	}
}
