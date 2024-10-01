package cmd

import (
	"github.com/analog-substance/arsenic/pkg/capture"
	"github.com/analog-substance/scopious/pkg/scopious"
	"github.com/spf13/cobra"
)

var captureCmd = &cobra.Command{
	Use:     "wrapture", // lolz typo ftw
	Aliases: []string{"capture"},
	Short:   "capture exec",
	Long:    `capture exec`,
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")
		capture.Run(scopeDir, args)
	},
}

func init() {
	rootCmd.AddCommand(captureCmd)
	captureCmd.Flags().StringP("scope-dir", "s", scopious.DefaultScope, "Scope dir to use")
}
