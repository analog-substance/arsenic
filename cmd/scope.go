package cmd

import (
	"fmt"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// scopeCmd represents the scope command
var scopeCmd = &cobra.Command{
	Use:   "scope",
	Short: "Print all scope",
	Long:  `Print all scope`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, _ := util.GetScope("domains")
		ips, _ := util.GetScope("ips")

		scope := append(domains, ips...)
		for _, scopeItem := range scope {
			fmt.Println(scopeItem)
		}
	},
}

func init() {
	rootCmd.AddCommand(scopeCmd)
}
