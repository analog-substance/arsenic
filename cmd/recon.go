package cmd

import (
	"github.com/defektive/arsenic/lib/util"
	"github.com/spf13/cobra"
)

var reconCmd = &cobra.Command{
	Use:   "recon",
	Short: "Run scripts in the recon phase",
	Long: `Run scripts in the recon phase.

Scripts should determine what hosts it needs to run against.`,
	Run: func(cmd *cobra.Command, args []string) {
		util.ExecutePhaseScripts("recon", []string{})
	},
}

func init() {
	rootCmd.AddCommand(reconCmd)
}
