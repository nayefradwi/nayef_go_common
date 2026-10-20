package cmd

import (
	"github.com/nayefradwi/nayef_go_common/ngo/internal/common"
	"github.com/nayefradwi/nayef_go_common/ngo/internal/new"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Scaffold a new Go service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return new.Run(cmd.Flags())
	},
}

func init() {
	newCmd.Flags().String(common.FLAG_NAME, "", "service name, lowercase letters and digits")
	newCmd.Flags().String(common.FLAG_SERVICE_TYPE, "", "one of: "+common.FlagValues(common.AllServiceTypes))
	newCmd.Flags().StringSlice(common.FLAG_INFRA_TYPES, nil, "comma separated: "+common.FlagValues(common.AllInfraTypes))
	newCmd.Flags().StringSlice(common.FLAG_FEATURES, nil, "comma separated: "+common.FlagValues(common.AllFeatures))
}
