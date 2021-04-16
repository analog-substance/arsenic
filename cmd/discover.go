package cmd

import (
	"github.com/defektive/arsenic/lib/util"
	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Run scripts in the discover phase.",
	Long: `Run scripts in the discover phase.

Scripts should determine what hosts it needs to run against.`,
	Run: func(cmd *cobra.Command, args []string) {
		util.ExecutePhaseScripts("discover", []string{})
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)
}
