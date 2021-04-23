package cmd

import (
	"fmt"
	"github.com/defektive/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// domainsCmd represents the domains command
var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Print domains in scope",
	Long: `Print domains in scope

This will prune blacklisted domains, and blacklisted root domains.
`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, _ := getScope("domains")

		rootDomains, _ := cmd.Flags().GetBool("root-domains")
		if rootDomains {
			domains = util.GetRootDomains(domains)
		}

		for _, scope := range domains {
			fmt.Println(scope)
		}
	},
}

func init() {
	scopeCmd.AddCommand(domainsCmd)

	domainsCmd.Flags().BoolP("root-domains", "r", false, "show only root domains")
}
