package cmd_test

import (
	"testing"

	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	scFake "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/aerogear/mobile-cli/pkg/cmd"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
	kv1 "k8s.io/client-go/pkg/api/v1"
	kbeta "k8s.io/client-go/pkg/apis/apps/v1beta1"
	ktesting "k8s.io/client-go/testing"
)

func TestIntegrationCmd_CreateIntegrationCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ExpectError      bool
		ValidateErr      func(t *testing.T, err error)
		Args             []string
		Flags            []string
	}{
		{
			Name: "test returns error if missing arguments",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "missing arguments: integration <consuming_service_instance_id> <providing_service_instance_id>"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{},
			Flags: []string{},
		},
		{
			Name: "test returns error if flags not set",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to get namespace: no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{},
		},
		{
			Name: "returns error when deployment cannot be found",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "keycloak",
						},
					}, nil
				})
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fh-sync-server",
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				fake := &kFake.Clientset{}
				fake.AddReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kbeta.Deployment{
						Spec: kbeta.DeploymentSpec{
							Template: kv1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{},
								},
							},
						},
					}, errors.New("failed to get deployment")
				})
				return fake
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to get deployment for service keycloak: failed to get deployment"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{"--namespace=test", "--auto-redeploy=true", "--no-wait=true"},
		},
		{
			Name: "returns error when deployment cannot be updated",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "keycloak",
						},
					}, nil
				})
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fh-sync-server",
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				fake := &kFake.Clientset{}
				fake.AddReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kbeta.Deployment{
						Spec: kbeta.DeploymentSpec{
							Template: kv1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{},
								},
							},
						},
					}, nil
				})
				fake.AddReactor("update", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed to update deployment")
				})
				return fake
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to update deployment for service keycloak: failed to update deployment"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{"--namespace=test", "--auto-redeploy=true", "--no-wait=true"},
		},
		{
			Name: "should pass when serviceinstances exist and auto-redeploy & no-wait are set",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "keycloak",
						},
					}, nil
				})
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fh-sync-server",
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				fake := &kFake.Clientset{}
				fake.AddReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kbeta.Deployment{
						Spec: kbeta.DeploymentSpec{
							Template: kv1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{},
								},
							},
						},
					}, nil
				})
				return fake
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{"--namespace=test", "--auto-redeploy=true", "--no-wait=true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			root := cmd.NewRootCmd()
			integrationCmd := cmd.NewIntegrationCmd(tc.SvcCatalogClient(), tc.K8Client())
			createCmd := integrationCmd.CreateIntegrationCmd()
			root.AddCommand(createCmd)
			if err := createCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse command flags", err)
			}
			err := createCmd.RunE(createCmd, tc.Args)

			if err != nil && !tc.ExpectError {
				t.Fatal("did not expect an error but gone one:", err)
			}
			if err == nil && tc.ExpectError {
				t.Fatal("expected an error but got none")
			}
			if tc.ValidateErr != nil {
				tc.ValidateErr(t, err)
			}
		})
	}
}

func TestIntegrationCmd_DeleteIntegrationCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ExpectError      bool
		ValidateErr      func(t *testing.T, err error)
		Args             []string
		Flags            []string
	}{
		{
			Name: "test returns error if missing arguments",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "missing arguments."
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{},
			Flags: []string{},
		},
		{
			Name: "test returns error if flags not set",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to get namespace: no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{},
		},
		{
			Name: "returns error when deployment cannot be found",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "keycloak",
						},
					}, nil
				})
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fh-sync-server",
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				fake := &kFake.Clientset{}
				fake.AddReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kbeta.Deployment{
						Spec: kbeta.DeploymentSpec{
							Template: kv1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{},
								},
							},
						},
					}, errors.New("failed to get deployment")
				})
				return fake
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to get deployment for service keycloak: failed to get deployment"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{"--namespace=test", "--auto-redeploy=true"},
		},
		{
			Name: "returns error when deployment cannot be updated",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "keycloak",
						},
					}, nil
				})
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fh-sync-server",
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				fake := &kFake.Clientset{}
				fake.AddReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kbeta.Deployment{
						Spec: kbeta.DeploymentSpec{
							Template: kv1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{},
								},
							},
						},
					}, nil
				})
				fake.AddReactor("update", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed to update deployment")
				})
				return fake
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to update deployment for service keycloak: failed to update deployment"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{"--namespace=test", "--auto-redeploy=true"},
		},
		{
			Name: "should pass when serviceinstances exist and auto-redeploy is set",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "keycloak",
						},
					}, nil
				})
				fake.AddReactor("get", "serviceinstances", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fh-sync-server",
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				fake := &kFake.Clientset{}
				fake.AddReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &kbeta.Deployment{
						Spec: kbeta.DeploymentSpec{
							Template: kv1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: map[string]string{},
								},
							},
						},
					}, nil
				})
				return fake
			},
			Args:  []string{"keycloak", "fh-sync-server"},
			Flags: []string{"--namespace=test", "--auto-redeploy=true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			root := cmd.NewRootCmd()
			integrationCmd := cmd.NewIntegrationCmd(tc.SvcCatalogClient(), tc.K8Client())
			deleteCmd := integrationCmd.DeleteIntegrationCmd()
			root.AddCommand(deleteCmd)
			if err := deleteCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse command flags", err)
			}
			err := deleteCmd.RunE(deleteCmd, tc.Args)

			if err != nil && !tc.ExpectError {
				t.Fatal("did not expect an error but gone one:", err)
			}
			if err == nil && tc.ExpectError {
				t.Fatal("expected an error but got none")
			}
			if tc.ValidateErr != nil {
				tc.ValidateErr(t, err)
			}
		})
	}
}

func TestIntegrationCmd_GetIntegrationCmd(t *testing.T) {
	cases := []struct {
		Name string
	}{
		{},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
		})
	}
}

func TestIntegrationCmd_ListIntegrationCmd(t *testing.T) {
	cases := []struct {
		Name             string
		SvcCatalogClient func() versioned.Interface
		K8Client         func() kubernetes.Interface
		ExpectError      bool
		ValidateErr      func(t *testing.T, err error)
		Flags            []string
	}{
		{
			Name: "test returns error if flags not set",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			ExpectError: true,
			ValidateErr: func(t *testing.T, err error) {
				expectedErr := "failed to get namespace: no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var"
				if err.Error() != expectedErr {
					t.Fatalf("expected error to be '%s' but got '%v'", expectedErr, err)
				}
			},
			Flags: []string{},
		},
		{
			Name: "should list service bindings as expected",
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("list", "servicebindings", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceBindingList{
						Items: []v1beta1.ServiceBinding{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "fh-sync-server",
								},
							},
						},
					}, nil
				})
				return fake
			},
			K8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			Flags: []string{"--namespace=test", "-o=json"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			root := cmd.NewRootCmd()
			integrationCmd := cmd.NewIntegrationCmd(tc.SvcCatalogClient(), tc.K8Client())
			listCmd := integrationCmd.ListIntegrationsCmd()
			root.AddCommand(listCmd)
			if err := listCmd.ParseFlags(tc.Flags); err != nil {
				t.Fatal("failed to parse flags ", err)
			}
			err := listCmd.RunE(listCmd, []string{})

			if err != nil && !tc.ExpectError {
				t.Fatal("did not expect an error but gone one:", err)
			}
			if err == nil && tc.ExpectError {
				t.Fatal("expected an error but got none")
			}
			if tc.ValidateErr != nil {
				tc.ValidateErr(t, err)
			}
		})
	}
}
