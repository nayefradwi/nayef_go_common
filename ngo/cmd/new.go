package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/new"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Scaffold a new Go service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return new.Run()
	},
}
