package cmd

import (
	"fmt"

	"github.com/analog-substance/arsenic/pkg/scope"
	"github.com/spf13/cobra"
)

// scopeCmd represents the scope command
var scopeCmd = &cobra.Command{
	Use:   "scope",
	Short: "Print all scope",
	Long:  `Print all scope`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, _ := scope.GetScope("domains")
		ips, _ := scope.GetScope("ips")

		allScope := append(domains, ips...)
		for _, scopeItem := range allScope {
			fmt.Println(scopeItem)
		}
	},
}

func init() {
	RootCmd.AddCommand(scopeCmd)
}
