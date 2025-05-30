package cmd

import (
	"context"
	"fmt"
	"github.com/analog-substance/util/fileutil"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/analog-substance/arsenic/pkg/config"
	"github.com/analog-substance/arsenic/pkg/engine"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
)

// scriptCmd represents the serve command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Run arbitrary arsenic scripts",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		argString, _ := cmd.Flags().GetString("script-args")

		var err error
		var scriptArgs []string
		if argString != "" {
			scriptArgs, err = shlex.Split(argString)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		path := filepath.Join(config.Get().Scripts.Directory, name)
		if filepath.Ext(path) != ".tengo" {
			path = path + ".tengo"
		}

		script, err := engine.NewScript(path)
		cobra.CheckErr(err)

		err = script.Run(scriptArgs)
		if err != nil && err != context.Canceled {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(scriptCmd)
	scriptCmd.Flags().StringP("name", "n", "", "Name of the script to run")
	scriptCmd.RegisterFlagCompletionFunc("name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		scriptsDir := config.Get().Scripts.Directory
		if !fileutil.DirExists(scriptsDir) {
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

	scriptCmd.Flags().StringP("script-args", "a", "", "Args to pass to the script")
}
