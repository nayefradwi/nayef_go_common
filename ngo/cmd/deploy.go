package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/deploy"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Scaffold deployment configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy.Run(cmd.Flags())
	},
}

func init() {
	deployCmd.Flags().String(deploy.FLAG_ENVIRONMENT, "", "environment name, lowercase letters and digits")
	deployCmd.Flags().String(deploy.FLAG_SIZE, "", "one of: "+common.FlagValues(deploy.AllSizes))
	deployCmd.Flags().String(deploy.FLAG_PROVIDER, "", "one of: "+common.FlagValues(deploy.AllProviders)+" (vps only with small)")
}
