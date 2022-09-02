package cmd

import (
	"context"

	"github.com/analog-substance/arsenic/lib/script"
	"github.com/spf13/cobra"
)

// scriptCmd represents the serve command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Run arbitrary arsenic scripts",
	Run: func(cmd *cobra.Command, args []string) {
		// name, _ := cmd.Flags().GetString("name")
		scriptArgs, _ := cmd.Flags().GetStringToString("script-args")

		err := script.Run("/opt/arsenic/scripts/recon/as-content-discovery.tengo", scriptArgs)
		if err != nil && err != context.Canceled {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)
	// scriptCmd.Flags().StringP("name", "n", "", "Name of the script to run")
	// scriptCmd.MarkFlagRequired("name")

	scriptCmd.Flags().StringToStringP("script-args", "a", make(map[string]string), "Args to pass to the script")
}
