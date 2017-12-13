package cmd

import (
	"github.com/spf13/cobra"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "stop clientbuild",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
