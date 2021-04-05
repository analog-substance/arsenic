package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
		count := len(args)
		switch count {
		case 0:
			printConfig()
		case 1, 2:
			key := args[0]
			if !viper.IsSet(key) {
				fmt.Println("Key not found in config")
				return
			}

			currentValue := viper.Get(key)
			if count == 1 {
				t, err := yaml.Marshal(currentValue)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(t))
				return
			}

			userValue := args[1]
			var newValue interface{}
			var err error
			if _, ok := currentValue.(bool); ok {
				newValue = strings.ToLower(userValue) == "true"
			} else if _, ok := currentValue.(int); ok {
				newValue, err = strconv.Atoi(userValue)
				if err != nil {
					fmt.Printf("Error converting %s to an integer\n", userValue)
					return
				}
			} else if _, ok := currentValue.(string); ok {
				newValue = userValue
			} else {
				fmt.Println("Cannot set keys that are not of type int, string or bool")
				return
			}

			fmt.Println("Writing Config")

			file := viper.ConfigFileUsed()
			if file == "" {
				file = "arsenic.yaml"
			}
			createIfNotExist(file)

			viper.Set(key, newValue)
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
