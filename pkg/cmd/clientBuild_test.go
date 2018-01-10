package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		want *ClientBuildsCmd
	}{
		{
			name: "test new client builds command",
			want: &ClientBuildsCmd{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewClientBuildsCmd(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewClientBuildsCmd() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClientBuildsCmd_GetClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		cbc  *ClientBuildsCmd
		want *cobra.Command
	}{
		{
			name: "test get client builds command",
			cbc:  &ClientBuildsCmd{},
			want: &cobra.Command{
				Use:   "clientbuilds",
				Short: "get clientbuilds for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cbc.GetClientBuildsCmd()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.GetClientBuildsCmd().Use = %v, want %v", use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.GetClientBuildsCmd().Short = %v, want %v", short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.GetClientBuildsCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.GetClientBuildsCmd().RunE = %v, want %v", runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}

func TestClientBuildsCmd_ListClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		cbc  *ClientBuildsCmd
		want *cobra.Command
	}{
		{
			name: "test list client builds command",
			cbc:  &ClientBuildsCmd{},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "get a specific clientbuild for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cbc.ListClientBuildsCmd()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.ListClientBuildsCmd().Use = %v, want %v", use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.ListClientBuildsCmd().Short = %v, want %v", short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.ListClientBuildsCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.ListClientBuildsCmd().RunE = %v, want %v", runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}

func TestClientBuildsCmd_CreateClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		cbc  *ClientBuildsCmd
		want *cobra.Command
	}{
		{
			name: "test create client builds command",
			cbc:  &ClientBuildsCmd{},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "create a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cbc.CreateClientBuildsCmd()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.CreateClientBuildsCmd().Use = %v, want %v", use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.CreateClientBuildsCmd().Short = %v, want %v", short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.CreateClientBuildsCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.CreateClientBuildsCmd().RunE = %v, want %v", runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}

func TestClientBuildsCmd_DeleteClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		cbc  *ClientBuildsCmd
		want *cobra.Command
	}{
		{
			name: "test create client builds command",
			cbc:  &ClientBuildsCmd{},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "delete a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//cbc := &ClientBuildsCmd{}
			got := tc.cbc.DeleteClientBuildsCmd()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.DeleteClientBuildsCmd().Use = %v, want %v", use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.DeleteClientBuildsCmd().Short = %v, want %v", short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.DeleteClientBuildsCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.DeleteClientBuildsCmd().RunE = %v, want %v", runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}

func TestClientBuildsCmd_StopClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		cbc  *ClientBuildsCmd
		want *cobra.Command
	}{
		{
			name: "test create client builds command",
			cbc:  &ClientBuildsCmd{},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "stop a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cbc.StopClientBuildsCmd()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.StopClientBuildsCmd().Use = %v, want %v", use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.StopClientBuildsCmd().Short = %v, want %v", short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.StopClientBuildsCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.StopClientBuildsCmd().RunE = %v, want %v", runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}

func TestClientBuildsCmd_StartClientBuildsCmd(t *testing.T) {
	tests := []struct {
		name string
		cbc  *ClientBuildsCmd
		want *cobra.Command
	}{
		{
			name: "test create client builds command",
			cbc:  &ClientBuildsCmd{},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "start a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//cbc := &ClientBuildsCmd{}
			got := tc.cbc.StartClientBuildsCmd()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.StartClientBuildsCmd().Use = %v, want %v", use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.StartClientBuildsCmd().Short = %v, want %v", short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.StartClientBuildsCmd().RunE = %v, want %v", reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.StartClientBuildsCmd().RunE = %v, want %v", runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}
