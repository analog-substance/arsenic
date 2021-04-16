package cmd

import (
	"github.com/defektive/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// huntCmd represents the hunt command
var huntCmd = &cobra.Command{
	Use:   "hunt",
	Short: "Find interesting things",
	Long:  `Find interesting things`,
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		util.ExecutePhaseScripts("hunt", []string{}, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(huntCmd)
	huntCmd.Flags().BoolP("dry-run", "d", false, "Dry run")
}
