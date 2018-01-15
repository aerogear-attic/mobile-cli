package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	kFake "k8s.io/client-go/kubernetes/fake"
)

type args struct {
	k8Client kubernetes.Interface
}

func TestNewClientConfigCmd(t *testing.T) {
	k8Client := &kFake.Clientset{}
	tests := []struct {
		name string
		args args
		want *ClientConfigCmd
	}{
		{
			name: "new client config command",
			args: args{k8Client: k8Client},
			want: &ClientConfigCmd{
				k8Client: k8Client,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewClientConfigCmd(tc.args.k8Client); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewClientConfigCmd() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClientConfigCmd_GetClientConfigCmd(t *testing.T) {
	fakeK8Client := &kFake.Clientset{}
	fakeCbrCmd := &cobra.Command{
		Use:   "clientconfig",
		Short: "get clientconfig returns a client ready filtered configuration of the available services.",
		Long: `get clientconfig
		mobile --namespace=myproject get clientconfig
		kubectl plugin mobile get clientconfig`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "get client config command",
			args: args{k8Client: fakeK8Client},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ccCmd := &ClientConfigCmd{
				k8Client: tc.args.k8Client,
			}

			got := ccCmd.GetClientConfigCmd()
			if use := got.Use; !reflect.DeepEqual(use, fakeCbrCmd.Use) {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().Use = %v, want %v", use, fakeCbrCmd.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, fakeCbrCmd.Short) {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().Short = %v, want %v", short, fakeCbrCmd.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(fakeCbrCmd.RunE)) {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(fakeCbrCmd.RunE))
			}
			runE := got.RunE
			fakeCbrCmd.Flags().String("namespace", "", "Namespace for software installation")
			err := runE(fakeCbrCmd, nil) // args are not used in RunE function
			if err == nil {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().RunE() should fail with error")
			}
			fakeCbrCmd.Flags().Set("namespace", "tesing-ns")
			err = runE(fakeCbrCmd, nil) // args are not used in RunE function
			if err != nil {
				t.Errorf("ClientConfigCmd.GetClientConfigCmd().RunE() should not fail with error")
			}
		})
	}
}
