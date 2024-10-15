package cmd

import (
	"github.com/analog-substance/scopious/pkg/scopious"
	"github.com/spf13/cobra"
)

// inspectCmd represents the ingest command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect data",
}

func init() {
	RootCmd.AddCommand(inspectCmd)
	inspectCmd.PersistentFlags().StringP("scope-dir", "s", scopious.DefaultScope, "Scope dir to use")
}
