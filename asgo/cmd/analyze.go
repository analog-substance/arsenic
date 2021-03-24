package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze discovery data and create",
	Long: `Analyze discovery data and create hosts.

This will create a single host for hostnames that resolve to the same IPs`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("plz 2 refactor me, cause i am calling a slow shell script... k thx, bye!")
		execScript("determine_services.sh")
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
