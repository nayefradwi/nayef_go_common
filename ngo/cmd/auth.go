package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Add an auth module to a Go Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Run()
	},
}
