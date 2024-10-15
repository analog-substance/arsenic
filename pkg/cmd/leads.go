package cmd

import (
	"github.com/spf13/cobra"
)

var leadsCmd = &cobra.Command{
	Use:   "leads",
	Short: "Leads from other sources",
	Long: `Import leads from other sources
`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(leadsCmd)

}
