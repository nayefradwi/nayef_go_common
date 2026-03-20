package cmd

import (
	"fmt"
	"os"

	"github.com/nayefradwi/nayef_go_common/ngo/internal/log"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:           "ngo",
	Short:         "Bootstrap Go backend services using nayef_go_common",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug output")
	cobra.OnInitialize(func() {
		if verbose {
			log.SetVerbose()
		}
	})
}
