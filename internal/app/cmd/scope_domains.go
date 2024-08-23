package cmd

import (
	"fmt"

	"github.com/analog-substance/arsenic/lib/scope"

	"github.com/spf13/cobra"
)

// scopeDomainsCmd represents the domains command
var scopeDomainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Print domains in scope",
	Long: `Print domains in scope

This will prune blacklisted domains, and blacklisted root domains.
`,
	Run: func(cmd *cobra.Command, args []string) {
		domains, _ := scope.GetScope("domains")

		rootDomains, _ := cmd.Flags().GetBool("root-domains")
		allRootDomains, _ := cmd.Flags().GetBool("all-root-domains")
		if rootDomains || allRootDomains {
			domains = scope.GetRootDomains(domains, rootDomains)
		}

		for _, scope := range domains {
			fmt.Println(scope)
		}
	},
}

func init() {
	//scopeCmd.AddCommand(scopeDomainsCmd)
	//
	//scopeDomainsCmd.Flags().BoolP("root-domains", "r", false, "show only non-blacklisted root domains")
	//scopeDomainsCmd.Flags().Bool("all-root-domains", false, "show all root domains")
}
