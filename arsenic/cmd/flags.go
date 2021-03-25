package cmd

import (
	"github.com/defektive/arsenic/arsenic/lib/host"
	"github.com/spf13/cobra"
)

// flagsCmd represents the flags command
var flagsCmd = &cobra.Command{
	Use:   "flags",
	Short: "Add, Update, and delete flags",
	Long: `Add, Update, and delete flags.

Flags are neat`,
	Run: func(cmd *cobra.Command, args []string) {
		host.UpdateFlags()
	},
}

func init() {
	rootCmd.AddCommand(flagsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// flagsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// flagsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
