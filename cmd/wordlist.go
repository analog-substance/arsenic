package cmd

import (
	"fmt"
	"strings"

	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/spf13/cobra"
)

// wordlistCmd represents the wordlist command
var wordlistCmd = &cobra.Command{
	Use:   "wordlist",
	Short: "Generate a wordlist",
	Long:  `Generate a wordlist`,
	// ValidArgs: validWordlistTypes,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return lib.GetValidWordlistTypes(), cobra.ShellCompDirectiveDefault
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if !lib.IsValidWordlistType(args[0]) {
			return fmt.Errorf("invalid argument %q for %q", args[0], cmd.CommandPath())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		wordlistSet := set.NewSet("")
		lib.GenerateWordlist(args[0], &wordlistSet)
		wordlistSet.PrintSorted()
	},
}

func init() {
	rootCmd.AddCommand(wordlistCmd)

	oldUsage := wordlistCmd.UsageFunc()
	wordlistCmd.SetUsageFunc(func(c *cobra.Command) error {
		c.Use = fmt.Sprintf("wordlist (%s)", strings.Join(lib.GetValidWordlistTypes(), "|"))
		return oldUsage(c)
	})

	oldHelp := wordlistCmd.HelpFunc()
	wordlistCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		if !configInitialized {
			initConfig()
		}

		c.Use = fmt.Sprintf("wordlist (%s)", strings.Join(lib.GetValidWordlistTypes(), "|"))
		oldHelp(c, s)
	})
}
