package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/defektive/arsenic/arsenic/lib/util"
	// "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display config information",
	Long: `Display config information.

Helpful to see what scripts would be executed.`,
	Args: cobra.RangeArgs(0, 3),
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 0:
			printConfig()
		case 1:
			key := args[0]
			if !viper.IsSet(key) {
				fmt.Println("Key not found in config")
				return
			}

			t, err := yaml.Marshal(viper.Get(key))
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(t))
		case 2:
			fmt.Println("Writing Config")

			file := viper.ConfigFileUsed()
			if file == "" {
				file = "arsenic.yaml"
			}
			createIfNotExist(file)

			viper.Set(args[0], args[1])
			viper.WriteConfig()
		}
	},
}

func createIfNotExist(fileName string) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
}

func printConfig() {
	t, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Configuration")
	fmt.Println(string(t))
	fmt.Println()

	fmt.Println("--Discover files to be run--")
	for _, scriptConfig := range util.GetScripts("discover") {
		fmt.Printf("%s\n\tenabled: %t\n\torder: %d\n\n", scriptConfig.Script, scriptConfig.Enabled, scriptConfig.Order)
	}
	fmt.Println()

	fmt.Println("--Recon files to be run--")
	for _, scriptConfig := range util.GetScripts("recon") {
		fmt.Printf("%s\n\tenabled: %t\n\torder: %d\n\n", scriptConfig.Script, scriptConfig.Enabled, scriptConfig.Order)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
}
