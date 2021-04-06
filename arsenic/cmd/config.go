package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
		case 1, 2: // 1 arg = get config value, 2 args = set config value
			key := args[0]
			if !viper.IsSet(key) {
				fmt.Println("Key not found in config")
				return
			}

			currentValue := viper.Get(key)
			if count == 1 { // If only one argument, just display the current config value
				subKeysOnly, _ := cmd.Flags().GetBool("sub-keys")
				if subKeysOnly {
					if valueMap, ok := currentValue.(map[string]interface{}); ok {
						for key := range valueMap {
							fmt.Println(key)
						}
					} else {
						fmt.Println("No sub-keys")
					}
					return
				}

				t, err := yaml.Marshal(currentValue)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Print(string(t))
				return
			}

			newValue, err := convertToConfigType(currentValue, args[1])
			if err != nil {
				fmt.Println(err)
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

func convertToConfigType(currentValue interface{}, userValue string) (interface{}, error) {
	// Currently our config values are either bool, int, or string
	if _, ok := currentValue.(bool); ok {
		value, err := strconv.ParseBool(userValue)
		if err != nil {
			return nil, fmt.Errorf("error converting %s to a bool\n%v", userValue, err)
		}
		return value, nil
	} else if _, ok := currentValue.(int); ok {
		value, err := strconv.Atoi(userValue)
		if err != nil {
			return nil, fmt.Errorf("error converting %s to an integer\n%v", userValue, err)
		}
		return value, nil
	} else if _, ok := currentValue.(string); ok {
		return userValue, nil
	}
	return nil, fmt.Errorf("cannot set keys that are not of type int, string or bool")
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
	configCmd.Flags().BoolP("sub-keys", "k", false, "display only the sub-keys")
}
