package cmd

import (
	"github.com/spf13/cobra"
		"github.com/defektive/arsenic/arsenic/lib/util"
)

// huntCmd represents the hunt command
var huntCmd = &cobra.Command{
	Use:   "hunt",
	Short: "Find interesting things",
	Long: `Find interesting things`,
	Run: func(cmd *cobra.Command, args []string) {
		util.ExecutePhaseScripts("hunt")
	},
}

func init() {
	rootCmd.AddCommand(huntCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// huntCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// huntCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
