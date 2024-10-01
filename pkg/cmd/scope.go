package cmd

import (
	scopiusCmd "github.com/analog-substance/scopious/pkg/cmd"
)

func init() {
	scopiusCmd.RootCmd.Use = "scope"
	rootCmd.AddCommand(scopiusCmd.RootCmd)
}
