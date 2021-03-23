package cmd

import (
	// "fmt"

	"github.com/spf13/cobra"
)

var reconCmd = &cobra.Command{
	Use:   "recon",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		executePhaseScripts("recon")
	},
}

func init() {
	rootCmd.AddCommand(reconCmd)
}
