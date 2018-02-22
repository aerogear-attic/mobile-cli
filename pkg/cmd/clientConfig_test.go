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

package cmd_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"regexp"

	"github.com/aerogear/mobile-cli/pkg/apis/mobile/v1alpha1"
	"github.com/aerogear/mobile-cli/pkg/apis/servicecatalog/v1beta1"
	mobile "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned"
	mcFake "github.com/aerogear/mobile-cli/pkg/client/mobile/clientset/versioned/fake"
	"github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned"
	scFake "github.com/aerogear/mobile-cli/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/aerogear/mobile-cli/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
	kt "k8s.io/client-go/testing"
	ktesting "k8s.io/client-go/testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func TestClientConfigCmd_GetClientConfigCmd(t *testing.T) {
	getFakeCbrCmd := func() *cobra.Command {
		return &cobra.Command{
			Use:   "clientconfig <clientID>",
			Short: "get clientconfig returns a client ready filtered configuration of the available services.",
			Long: `get clientconfig
	mobile --namespace=myproject get clientconfig
	kubectl plugin mobile get clientconfig`,
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
	}

	tests := []struct {
		name             string
		k8Client         func() kubernetes.Interface
		mobileClient     func() mobile.Interface
		SvcCatalogClient func() versioned.Interface
		ClusterHost      string
		namespace        string
		args             []string
		cobraCmd         *cobra.Command
		ExpectError      bool
		ErrorPattern     string
		ValidateOut      func(bytes.Buffer) error
	}{
		{
			name: "get client config command with empty namespace",
			k8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			mobileClient: func() mobile.Interface {
				return &mcFake.Clientset{}
			},
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			namespace:    "",
			ClusterHost:  "test",
			args:         []string{"client-id"},
			cobraCmd:     getFakeCbrCmd(),
			ExpectError:  true,
			ErrorPattern: "no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var",
			ValidateOut:  func(out bytes.Buffer) error { return nil },
		},
		{
			name: "get client config command with no services",
			k8Client: func() kubernetes.Interface {
				fakeclient := &kFake.Clientset{}
				fakeclient.AddReactor("list", "secrets", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.SecretList{Items: []v1.Secret{}}, nil
				})
				return fakeclient
			},
			mobileClient: func() mobile.Interface {
				return &mcFake.Clientset{}
			},
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			namespace:   "testing-ns",
			ClusterHost: "test",
			args:        []string{"client-id"},
			cobraCmd:    getFakeCbrCmd(),
			ExpectError: false,
			ValidateOut: func(out bytes.Buffer) error {
				expected := `{
	"version": 1,
	"clusterName": "test",
	"namespace": "testing-ns",
	"clientId": "client-id",
	"services": []
}`
				if strings.TrimSpace(out.String()) != expected {
					return errors.New(fmt.Sprintf("expected: '%v', got: '%v'", expected, strings.TrimSpace(out.String())))
				}
				return nil
			},
		},
		{
			name: "get client config command with services",
			k8Client: func() kubernetes.Interface {
				fakeclient := &kFake.Clientset{}
				fakeclient.AddReactor("list", "secrets", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					secrets := []v1.Secret{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									"mobile": "enabled",
								},
							},
							Data: map[string][]byte{
								"name": []byte("test-service"),
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "keycloak",
								Labels: map[string]string{
									"mobile": "enabled",
								},
							},
							Data: map[string][]byte{
								"name": []byte("keycloak"),
							},
						},
					}
					secretList := &v1.SecretList{
						Items: secrets,
					}
					return true, secretList, nil
				})
				fakeclient.AddReactor("get", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					var config *v1.ConfigMap
					name := action.(ktesting.GetAction).GetName()
					if name == "keycloak" {
						config = &v1.ConfigMap{
							Data: map[string]string{
								"public_installation": "{}",
								"name":                "keycloak",
							},
						}
					}
					if name == "test-service" {
						config = &v1.ConfigMap{
							Data: map[string]string{
								"name": "test-service",
							},
						}
					}
					return true, config, nil
				})
				return fakeclient
			},
			mobileClient: func() mobile.Interface {
				return &mcFake.Clientset{}
			},
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				return fake
			},
			namespace:   "testing-ns",
			ClusterHost: "test",
			args:        []string{"client-id"},
			cobraCmd:    getFakeCbrCmd(),
			ExpectError: false,
			ValidateOut: func(out bytes.Buffer) error {
				expected := `{
	"version": 1,
	"clusterName": "test",
	"namespace": "testing-ns",
	"clientId": "client-id",
	"services": [
		{
			"id": "test-service",
			"name": "test-service",
			"type": "",
			"url": "",
			"config": {}
		},
		{
			"id": "keycloak",
			"name": "keycloak",
			"type": "",
			"url": "",
			"config": {}
		}
	]
}`
				if strings.TrimSpace(out.String()) != expected {
					return errors.New(fmt.Sprintf("expected: '%v', got: '%v'", expected, strings.TrimSpace(out.String())))
				}
				return nil
			},
		},
		{
			name: "get client config command with excluded services",
			k8Client: func() kubernetes.Interface {
				fakeclient := &kFake.Clientset{}
				fakeclient.AddReactor("list", "secrets", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					secrets := []v1.Secret{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-service",
								Labels: map[string]string{
									"mobile": "enabled",
								},
							},
							Data: map[string][]byte{
								"name": []byte("test-service"),
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "keycloak",
								Labels: map[string]string{
									"mobile": "enabled",
								},
							},
							Data: map[string][]byte{
								"name": []byte("keycloak"),
							},
						},
					}
					secretList := &v1.SecretList{
						Items: secrets,
					}
					return true, secretList, nil
				})
				fakeclient.AddReactor("get", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					var config *v1.ConfigMap
					name := action.(ktesting.GetAction).GetName()
					if name == "keycloak" {
						config = &v1.ConfigMap{
							Data: map[string]string{
								"public_installation": "{}",
								"name":                "keycloak",
							},
						}
					}
					if name == "test-service" {
						config = &v1.ConfigMap{
							Data: map[string]string{
								"name": "test-service",
							},
						}
					}
					return true, config, nil
				})
				return fakeclient
			},
			mobileClient: func() mobile.Interface {
				mf := &mcFake.Clientset{}
				mf.AddReactor("get", "mobileclients", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1alpha1.MobileClient{
						Spec: v1alpha1.MobileClientSpec{
							Name:             "test",
							ApiKey:           "testkey",
							ClientType:       "cordova",
							ExcludedServices: []string{"dh-keycloak-fsdfsdfs"},
						},
					}, nil
				})
				return mf
			},
			SvcCatalogClient: func() versioned.Interface {
				fake := &scFake.Clientset{}
				fake.AddReactor("get", "serviceinstances", func(action kt.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1beta1.ServiceInstance{
						ObjectMeta: metav1.ObjectMeta{
							Name: "dh-keycloak-fsdfsdfs",
							Labels: map[string]string{
								"serviceName": "keycloak",
							},
						},
					}, nil
				})
				return fake
			},
			namespace:   "testing-ns",
			ClusterHost: "test",
			args:        []string{"client-id"},
			cobraCmd:    getFakeCbrCmd(),
			ExpectError: false,
			ValidateOut: func(out bytes.Buffer) error {
				expected := `{
	"version": 1,
	"clusterName": "test",
	"namespace": "testing-ns",
	"clientId": "client-id",
	"services": [
		{
			"id": "test-service",
			"name": "test-service",
			"type": "",
			"url": "",
			"config": {}
		}
	]
}`
				if strings.TrimSpace(out.String()) != expected {
					return errors.New(fmt.Sprintf("expected: '%v', got: '%v'", expected, strings.TrimSpace(out.String())))
				}
				return nil
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			ccCmd := cmd.NewClientConfigCmd(tc.k8Client(), tc.mobileClient(), tc.SvcCatalogClient(), tc.ClusterHost, &out)

			got := ccCmd.GetClientConfigCmd()
			if use := got.Use; use != tc.cobraCmd.Use {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().Use = %v, want %v", use, tc.cobraCmd.Use)
			}
			if short := got.Short; short != tc.cobraCmd.Short {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().Short = %v, want %v", short, tc.cobraCmd.Short)
			}

			runE := got.RunE
			tc.cobraCmd.Flags().String("namespace", tc.namespace, "Namespace for software installation")
			err := runE(tc.cobraCmd, tc.args) // args are not used in RunE function
			if tc.ExpectError && err == nil {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().RunE() expected an error but got none")
			}
			if !tc.ExpectError && err != nil {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().RunE() expect no error but got one %v", err)
			}
			if tc.ExpectError && err != nil {
				if m, err := regexp.Match(tc.ErrorPattern, []byte(err.Error())); !m {
					t.Errorf("expected regex %v to match error %v", tc.ErrorPattern, err)
				}
			}
			if err := tc.ValidateOut(out); err != nil {
				t.Errorf("%v\n", err)
			}
		})
	}
}
