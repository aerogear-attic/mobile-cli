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
	"testing"

	"regexp"

	"github.com/aerogear/mobile-cli/pkg/cmd"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/pkg/api/v1"
)

func TestClientConfigCmd_GetClientConfigCmd(t *testing.T) {
	getFakeCbrCmd := func() *cobra.Command {
		return &cobra.Command{
			Use:   "clientconfig",
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
		name         string
		k8Client     func() kubernetes.Interface
		namespace    string
		cobraCmd     *cobra.Command
		ExpectError  bool
		ErrorPattern string
	}{
		{
			name: "get client config command with empty namespace",
			k8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			namespace:    "",
			cobraCmd:     getFakeCbrCmd(),
			ExpectError:  true,
			ErrorPattern: "no namespace present. Cannot continue. Please set the --namespace flag or the KUBECTL_PLUGINS_CURRENT_NAMESPACE env var",
		},
		{
			name: "get client config command with testing-ns namespace",
			k8Client: func() kubernetes.Interface {
				return &kFake.Clientset{}
			},
			namespace:   "testing-ns",
			cobraCmd:    getFakeCbrCmd(),
			ExpectError: false,
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
								"name":                []byte("keycloak"),
								"public_installation": []byte("{}"),
							},
						},
					}
					secretList := &v1.SecretList{
						Items: secrets,
					}
					return true, secretList, nil
				})
				return fakeclient
			},
			namespace:   "testing-ns",
			cobraCmd:    getFakeCbrCmd(),
			ExpectError: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			ccCmd := cmd.NewClientConfigCmd(tc.k8Client(), &out)

			got := ccCmd.GetClientConfigCmd()
			if use := got.Use; use != tc.cobraCmd.Use {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().Use = %v, want %v", use, tc.cobraCmd.Use)
			}
			if short := got.Short; short != tc.cobraCmd.Short {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().Short = %v, want %v", short, tc.cobraCmd.Short)
			}

			runE := got.RunE
			tc.cobraCmd.Flags().String("namespace", tc.namespace, "Namespace for software installation")
			err := runE(tc.cobraCmd, nil) // args are not used in RunE function
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
		})
	}
}
