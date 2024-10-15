package cmd

import (
	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Run scripts in the discover phase.",
	Long: `Run scripts in the discover phase.

Scripts should determine what hosts it needs to run against.`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		ExecutePhaseScripts("discover", []string{}, dryRun)
	},
}

func init() {
	RootCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().BoolP("dry-run", "d", false, "Dry run")
}
