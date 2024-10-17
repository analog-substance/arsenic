package cmd

import (
	"github.com/analog-substance/arsenic/pkg/capture"
	"github.com/analog-substance/scopious/pkg/scopious"
	"github.com/spf13/cobra"
)

var captureCmd = &cobra.Command{
	Use:     "capture", // lolz typo ftw
	Aliases: []string{"wrapture"},
	Short:   "capture exec",
	Long:    `capture exec`,
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")
		rerun, _ := cmd.Flags().GetBool("rerun")
		capture.InteractiveRun(scopeDir, args, rerun)
	},
}

func init() {
	RootCmd.AddCommand(captureCmd)
	captureCmd.Flags().StringP("scope-dir", "s", scopious.DefaultScope, "Scope dir to use")
	captureCmd.Flags().BoolP("rerun", "R", false, "Rerun")
}
