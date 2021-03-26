package cmd

import (
	"github.com/defektive/arsenic/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// overrideCmd represents the override command
var overrideCmd = &cobra.Command{
	Use:   "override",
	Short: "Override a phase",
	Long:  `Override a phase`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, phase := range args {
			util.Override(phase)

		}
	},
}

func init() {
	rootCmd.AddCommand(overrideCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// overrideCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// overrideCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
