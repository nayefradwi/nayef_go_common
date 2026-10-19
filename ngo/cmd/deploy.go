package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/deploy"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Scaffold deployment configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy.Run()
	},
}
