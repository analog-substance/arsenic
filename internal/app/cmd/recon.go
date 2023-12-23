package cmd

import (
	"github.com/spf13/cobra"
)

var reconCmd = &cobra.Command{
	Use:   "recon",
	Short: "Run scripts in the recon phase",
	Long: `Run scripts in the recon phase.

Scripts should determine what hosts it needs to run against.`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		ExecutePhaseScripts("recon", []string{}, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(reconCmd)
	reconCmd.Flags().BoolP("dry-run", "d", false, "Dry run")
}
