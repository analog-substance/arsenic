package cmd

import (
	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/spf13/cobra"
)

// wordlistCmd represents the wordlist command
var wordlistCmd = &cobra.Command{
	Use:       "wordlist (web-content)",
	Short:     "Generate a wordlist",
	Long:      `Generate a wordlist`,
	ValidArgs: []string{"web-content", "sqli", "xss"},
	Args:      cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wordlistSet := set.NewSet("")
		lib.GenerateWordlist(args[0], &wordlistSet)
		wordlistSet.PrintSorted()
	},
}

func init() {
	rootCmd.AddCommand(wordlistCmd)
}
