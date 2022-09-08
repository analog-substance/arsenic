package cmd

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/analog-substance/arsenic/lib/script"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scriptCmd represents the serve command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Run arbitrary arsenic scripts",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		scriptArgs, _ := cmd.Flags().GetStringToString("script-args")

		err := script.Run(name, scriptArgs)
		if err != nil && err != context.Canceled {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)
	scriptCmd.Flags().StringP("name", "n", "", "Name of the script to run")
	scriptCmd.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		scriptsDir := viper.GetString("scripts-directory")
		if !util.DirExists(scriptsDir) {
			return nil, cobra.ShellCompDirectiveError
		}

		var scripts []string
		filepath.WalkDir(scriptsDir, func(path string, d fs.DirEntry, err error) error {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if !d.IsDir() && ext == ".tengo" {
				scriptPath, _ := filepath.Rel(scriptsDir, path)
				scripts = append(scripts, scriptPath)
			}
			return nil
		})

		return scripts, cobra.ShellCompDirectiveDefault
	})
	scriptCmd.MarkFlagRequired("name")

	scriptCmd.Flags().StringToStringP("script-args", "a", make(map[string]string), "Args to pass to the script")
}
