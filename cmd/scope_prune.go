package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/analog-substance/arsenic/lib/scope"
	"github.com/spf13/cobra"
)

// scopePruneCmd represents the prune command
var scopePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "read stdin remove out of scope things and print it to stdout.",
	Long:  `read stdin remove out of scope things and print it to stdout.`,
	Run: func(cmd *cobra.Command, args []string) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := scanner.Text()

			rootDomains, _ := cmd.Flags().GetBool("root-domains")

			if scope.IsInScope(input, rootDomains) {
				fmt.Println(input)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	},
}

func init() {
	scopeCmd.AddCommand(scopePruneCmd)

	scopePruneCmd.Flags().BoolP("root-domains", "r", false, "remove domains that belong to a blacklisted root domain, even if they are in the scope-domains.txt")
}
