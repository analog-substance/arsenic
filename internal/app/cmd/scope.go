package cmd

import (
	scopiusCmd "github.com/analog-substance/scopious/pkg/cmd"
)

//// scopeCmd represents the scope command
//var scopeCmd = &cobra.Command{
//	Use:   "scope",
//	Short: "Print all scope",
//	Long:  `Print all scope`,
//	Run: func(cmd *cobra.Command, args []string) {
//		domains, _ := scope.GetScope("domains")
//		ips, _ := scope.GetScope("ips")
//
//		allScope := append(domains, ips...)
//		for _, scopeItem := range allScope {
//			fmt.Println(scopeItem)
//		}
//	},
//}

func init() {
	scopiusCmd.RootCmd.Use = "scope"
	rootCmd.AddCommand(scopiusCmd.RootCmd)
}
