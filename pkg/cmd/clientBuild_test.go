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
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func funcName(v reflect.Value, stripFm bool) (pkg, name string) {
	var pkgRes, nameRes string
	if v.Kind() == reflect.Func { // check if function is passed
		name = runtime.FuncForPC(v.Pointer()).Name()
		if i := strings.LastIndex(name, "."); i != -1 {
			pkgRes, nameRes = name[:i], name[i+1:]
		}
	}

	if stripFm { // remove -fm suffix if true
		fmPos := strings.LastIndex(nameRes, "-fm")
		nameRes = nameRes[:fmPos]
	}

	return pkgRes, nameRes
}

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
		name           string
		cbcClosureFunc func(cbc *ClientBuildsCmd) func() *cobra.Command
		want           *cobra.Command
	}{
		{
			name: "test get client builds command",
			cbcClosureFunc: func(cbc *ClientBuildsCmd) func() *cobra.Command {
				return cbc.GetClientBuildsCmd
			},
			want: &cobra.Command{
				Use:   "clientbuilds",
				Short: "get clientbuilds for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
		{
			name: "test list client builds command",
			cbcClosureFunc: func(cbc *ClientBuildsCmd) func() *cobra.Command {
				return cbc.ListClientBuildsCmd
			},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "get a specific clientbuild for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
		{
			name: "test create client builds command",
			cbcClosureFunc: func(cbc *ClientBuildsCmd) func() *cobra.Command {
				return cbc.CreateClientBuildsCmd
			},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "create a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
		{
			name: "test delete client builds command",
			cbcClosureFunc: func(cbc *ClientBuildsCmd) func() *cobra.Command {
				return cbc.DeleteClientBuildsCmd
			},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "delete a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
		{
			name: "test stop client builds command",
			cbcClosureFunc: func(cbc *ClientBuildsCmd) func() *cobra.Command {
				return cbc.StopClientBuildsCmd
			},
			want: &cobra.Command{
				Use:   "clientbuild",
				Short: "stop a build for a mobile client",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
		{
			name: "test start client builds command",
			cbcClosureFunc: func(cbc *ClientBuildsCmd) func() *cobra.Command {
				return cbc.StartClientBuildsCmd
			},
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
			cbc := &ClientBuildsCmd{}
			gotFunc := tc.cbcClosureFunc(cbc)
			_, gotFuncName := funcName(reflect.ValueOf(gotFunc), true)

			got := gotFunc()
			if use := got.Use; !reflect.DeepEqual(use, tc.want.Use) {
				t.Errorf("ClientBuildsCmd.%v().Use = %v, want %v", gotFuncName, use, tc.want.Use)
			}
			if short := got.Short; !reflect.DeepEqual(short, tc.want.Short) {
				t.Errorf("ClientBuildsCmd.%v().Short = %v, want %v", gotFuncName, short, tc.want.Short)
			}
			if runE := got.RunE; !reflect.DeepEqual(reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE)) {
				t.Errorf("ClientBuildsCmd.%v().RunE = %v, want %v", gotFuncName, reflect.TypeOf(runE), reflect.TypeOf(tc.want.RunE))
			}
			if runE := got.RunE; !reflect.DeepEqual(runE(nil, nil), tc.want.RunE(nil, nil)) {
				t.Errorf("ClientBuildsCmd.%v().RunE = %v, want %v", gotFuncName, runE(nil, nil), tc.want.RunE(nil, nil))
			}
		})
	}
}
