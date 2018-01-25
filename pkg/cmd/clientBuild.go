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

import "github.com/spf13/cobra"

type ClientBuildsCmd struct{}

func NewClientBuildsCmd() *ClientBuildsCmd {
	return &ClientBuildsCmd{}
}

func (cbc *ClientBuildsCmd) GetClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuilds",
		Short: "get clientbuilds for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (cbc *ClientBuildsCmd) ListClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuild",
		Short: "get a specific clientbuild for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (cbc *ClientBuildsCmd) CreateClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuild",
		Short: "create a build for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (cbc *ClientBuildsCmd) DeleteClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuild",
		Short: "delete a build for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (cbc *ClientBuildsCmd) StopClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuild",
		Short: "stop a build for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func (cbc *ClientBuildsCmd) StartClientBuildsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clientbuild",
		Short: "start a build for a mobile client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
