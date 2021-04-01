package cmd

import (
	"fmt"
	"os"
	"log"

	"github.com/defektive/arsenic/arsenic/lib/util"
	// "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display config information",
	Long: `Display config information.

Helpful to see what scripts would be executed.`,
	Run: func(cmd *cobra.Command, args []string) {
		writeCfg, _ := cmd.Flags().GetBool("write")
		getCfg, _ := cmd.Flags().GetString("get")

		if getCfg != "" {
			fmt.Println(viper.GetString(getCfg))
			return
		}

		t, err := yaml.Marshal(viper.AllSettings())
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Configuration")
		fmt.Println(string(t))
		fmt.Println()

		fmt.Println("discover files to be run")
		for _, scriptConfig := range util.GetScripts("discover") {
			fmt.Printf("%s\n\tenabled: %t\n\torder: %d\n\n", scriptConfig.Script, scriptConfig.Enabled, scriptConfig.Order)
		}

		fmt.Println("Recon files to be run")
		for _, scriptConfig := range util.GetScripts("discover") {
			fmt.Printf("%s\n\tenabled: %t\n\torder: %d\n\n", scriptConfig.Script, scriptConfig.Enabled, scriptConfig.Order)
		}

		if writeCfg {
			fmt.Println("Writing Config")
			createIfNotExist(".arsenic.yaml")
			viper.WriteConfig()
		}
	},
}

func createIfNotExist (fileName string) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
			file, err := os.Create(fileName)
			if err != nil {
					log.Fatal(err)
			}
			defer file.Close()
	}
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolP("write", "w", false, "write config")
	configCmd.Flags().StringP("get", "g", "", "get config value")
}
