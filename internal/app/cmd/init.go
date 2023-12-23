package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Args:  cobra.ExactArgs(1),
	Short: "Init a new engagement",
	Long:  `Init a new engagement`,
	Run: func(cmd *cobra.Command, args []string) {
		os.MkdirAll(args[0], 0755)
		os.Chdir(args[0])
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		ExecutePhaseScripts("init", []string{}, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("dry-run", "d", false, "Dry run")
}
