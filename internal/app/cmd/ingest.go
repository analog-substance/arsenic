package cmd

import (
	"github.com/spf13/cobra"
)

// ingestCmd represents the ingest command
var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Import (ingest) output from various tools",
}

func init() {
	rootCmd.AddCommand(ingestCmd)
}
