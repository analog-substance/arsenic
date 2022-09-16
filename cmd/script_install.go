package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/analog-substance/arsenic/scripts"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scriptInstallCmd represents the install command
var scriptInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install arsenic scripts locally",
	Run: func(cmd *cobra.Command, args []string) {
		installPath, _ := cmd.Flags().GetString("path")

		if strings.HasPrefix(installPath, "~") {
			home, _ := homedir.Dir()
			installPath = strings.Replace(installPath, "~", home, 1)
		}

		if installPath != viper.GetString("scripts-directory") {
			viper.Set("scripts-directory", installPath)
			saveConfig()
		}

		dirs, err := scripts.All.ReadDir(".")
		if err != nil {
			panic(err)
		}

		fmt.Printf("[+] Installing scripts to %s\n", installPath)
		for _, dir := range dirs {
			files, err := scripts.All.ReadDir(dir.Name())
			if err != nil {
				panic(err)
			}

			err = os.MkdirAll(filepath.Join(installPath, dir.Name()), 0755)
			if err != nil {
				panic(err)
			}

			for _, f := range files {
				scriptPath := filepath.Join(dir.Name(), f.Name())
				data, err := scripts.All.ReadFile(scriptPath)
				if err != nil {
					panic(err)
				}

				err = os.WriteFile(filepath.Join(installPath, scriptPath), data, 0644)
				if err != nil {
					panic(err)
				}
			}
		}
	},
}

func init() {
	scriptCmd.AddCommand(scriptInstallCmd)

	home, _ := homedir.Dir()
	scriptInstallCmd.Flags().StringP("path", "p", filepath.Join(home, ".config", "arsenic"), "Path where the scripts will be installed")
}
