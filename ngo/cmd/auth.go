package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/auth"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Add an auth module to a Go Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Run(cmd.Flags())
	},
}

func init() {
	authCmd.Flags().String(auth.FLAG_IDENTIFIER, "", "one of: "+common.FlagValues(auth.AllIdentifiers))
	authCmd.Flags().String(auth.FLAG_LOGIN, "", "one of: "+common.FlagValues(auth.AllLoginMethods))
	authCmd.Flags().Bool(auth.FLAG_MFA, false, "generate MFA (TOTP) endpoints")
	authCmd.Flags().String(auth.FLAG_ACCESS_TOKEN, "", "one of: "+common.FlagValues(auth.AllAccessTokens))
	authCmd.Flags().String(auth.FLAG_STORE, "", "one of: "+common.FlagValues(common.AllInfraTypes)+" (redis only if the project has it)")
}
