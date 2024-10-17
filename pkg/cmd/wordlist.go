package cmd

import (
	"fmt"
	"strings"

	"github.com/analog-substance/arsenic/pkg"
	"github.com/analog-substance/util/set"
	"github.com/spf13/cobra"
)

// wordlistCmd represents the wordlist command
var wordlistCmd = &cobra.Command{
	Use:   "wordlist",
	Short: "Generate a wordlist",
	Long:  `Generate a wordlist`,
	// ValidArgs: validWordlistTypes,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return pkg.GetValidWordlistTypes(), cobra.ShellCompDirectiveDefault
	},
	Args: func(cmd *cobra.Command, args []string) error {
		setOrRefreshConfig()

		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		if !pkg.IsValidWordlistType(args[0]) {
			return fmt.Errorf("invalid argument %q for %q", args[0], cmd.CommandPath())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		wordlistSet := set.NewSet("")
		pkg.GenerateWordlist(args[0], &wordlistSet)
		wordlistSet.PrintSorted()
	},
}

func init() {
	RootCmd.AddCommand(wordlistCmd)

	oldUsage := wordlistCmd.UsageFunc()
	wordlistCmd.SetUsageFunc(func(c *cobra.Command) error {
		c.Use = fmt.Sprintf("wordlist (%s)", strings.Join(pkg.GetValidWordlistTypes(), "|"))
		return oldUsage(c)
	})

	oldHelp := wordlistCmd.HelpFunc()
	wordlistCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		setOrRefreshConfig()

		c.Use = fmt.Sprintf("wordlist (%s)", strings.Join(pkg.GetValidWordlistTypes(), "|"))
		oldHelp(c, s)
	})
}
