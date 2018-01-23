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

package cmd

import (
	"testing"

	"regexp"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ccCmd := &ClientConfigCmd{
				k8Client: tc.k8Client(),
			}

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
