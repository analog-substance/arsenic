package cmd

import (
	"fmt"
	"github.com/analog-substance/arsenic/pkg/capture"
	"github.com/spf13/cobra"
)

// inspectCommandsCmd represents the ingest command
var inspectCommandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "command data",
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")
		commands := capture.GetWrappedCommands(scopeDir)
		for _, command := range commands {
			fmt.Println(command.Command, command.Args)
		}
	},
}

func init() {
	inspectCmd.AddCommand(inspectCommandsCmd)
}
