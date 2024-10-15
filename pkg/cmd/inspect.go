package cmd

import (
	"fmt"
	"github.com/analog-substance/arsenic/pkg/capture"
	"github.com/analog-substance/scopious/pkg/scopious"
	"github.com/spf13/cobra"
)

// inspectCmd represents the ingest command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect data",
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")
		commands := capture.GetWrappedCommands(scopeDir)
		for _, command := range commands {
			fmt.Println(command.Command, command.Args)
		}

	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringP("scope-dir", "s", scopious.DefaultScope, "Scope dir to use")

}
