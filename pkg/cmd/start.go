package cmd

import (
	"github.com/spf13/cobra"
)

func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "start clientbuild",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
}
