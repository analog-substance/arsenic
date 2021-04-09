package cmd

import (
	"os"
	"github.com/spf13/cobra"
	"github.com/defektive/arsenic/arsenic/lib/util"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Args: cobra.RangeArgs(1, 1),
	Short: "Init a new engagement",
	Long: `Init a new engagement`,
	Run: func(cmd *cobra.Command, args []string) {
		os.MkdirAll(args[0], 0755)
		os.Chdir(args[0])
		util.ExecutePhaseScripts("init", []string{})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
