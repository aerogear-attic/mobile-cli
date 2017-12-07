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
