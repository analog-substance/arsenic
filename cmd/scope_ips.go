package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ipsCmd represents the ips command
var ipsCmd = &cobra.Command{
	Use:   "ips",
	Short: "Print in scope IP addresses.",
	Long:  `Print in scope IP addresses.`,
	Run: func(cmd *cobra.Command, args []string) {
		ips, _ := getScope("ips")
		for _, scope := range ips {
			fmt.Println(scope)
		}
	},
}

func init() {
	scopeCmd.AddCommand(ipsCmd)

}
