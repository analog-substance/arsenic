package cmd

import (
	"fmt"

	"github.com/analog-substance/arsenic/lib/util"
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
		allRootDomains, _ := cmd.Flags().GetBool("all-root-domains")
		if rootDomains || allRootDomains {
			domains = util.GetRootDomains(domains, rootDomains)
		}

		for _, scope := range domains {
			fmt.Println(scope)
		}
	},
}

func init() {
	scopeCmd.AddCommand(domainsCmd)

	domainsCmd.Flags().BoolP("root-domains", "r", false, "show only non-blacklisted root domains")
	domainsCmd.Flags().Bool("all-root-domains", false, "show all root domains")
}
