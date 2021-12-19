package cmd

import (
	"fmt"
	scope2 "github.com/analog-substance/arsenic/lib/scope"

	"github.com/spf13/cobra"
)

// scopeCmd represents the scope command
var scopeCmd = &cobra.Command{
	Use:   "scope",
	Short: "Print all scope",
	Long:  `Print all scope`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, _ := scope2.GetScope("domains")
		ips, _ := scope2.GetScope("ips")

		scope := append(domains, ips...)
		for _, scopeItem := range scope {
			fmt.Println(scopeItem)
		}
	},
}

func init() {
	rootCmd.AddCommand(scopeCmd)
}
