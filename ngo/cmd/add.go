package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/add"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add features to Go Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return add.Run(cmd.Flags())
	},
}

func init() {
	addCmd.Flags().String(common.FLAG_NAME, "", "feature name, lowercase letters and digits")
	addCmd.Flags().StringSlice(common.FLAG_INFRA_TYPES, nil, "comma separated, from the project's infra: "+common.FlagValues(common.AllInfraTypes))
	addCmd.Flags().StringSlice(common.FLAG_FEATURES, nil, "comma separated, from the project's features: "+common.FlagValues(common.AllFeatures))
}
