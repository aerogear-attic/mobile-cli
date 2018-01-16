package cmd

import (
	"reflect"
	"testing"

	"regexp"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
)

func TestNewClientConfigCmd(t *testing.T) {
	getFakeK8Client := func() kubernetes.Interface {
		return &kFake.Clientset{}
	}
	fakeK8Client := getFakeK8Client()
	tests := []struct {
		name     string
		k8Client kubernetes.Interface
		want     *ClientConfigCmd
	}{
		{
			name:     "new client config command",
			k8Client: fakeK8Client,
			want: &ClientConfigCmd{
				k8Client: fakeK8Client,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewClientConfigCmd(tc.k8Client); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewClientConfigCmd() = %v, want %v", got, tc.want)
			}
		})
	}
}

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
