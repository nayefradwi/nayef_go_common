package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/add"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add features to Go Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return add.Run()
	},
}
