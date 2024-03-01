package cmd

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/analog-substance/arsenic/lib/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const defaultConfigFile = "arsenic.yaml"

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
			saveNew, _ := cmd.Flags().GetBool("save")
			if saveNew {
				saveConfig(true)
				return
			}
			subKeysOnly, _ := cmd.Flags().GetBool("sub-keys")
			if subKeysOnly {
				printKeys(viper.AllSettings())
				return
			}
			printConfig()
		case 1, 2: // 1 arg = get/remove/reset config value, 2 args = set config value
			key := args[0]

			currentValue := viper.Get(key)
			if count == 1 { // If only one argument, display current value/keys or remove/reset key
				if !viper.IsSet(key) {
					fmt.Println("Key not found in config")
					return
				}

				subKeysOnly, _ := cmd.Flags().GetBool("sub-keys")
				if subKeysOnly {
					printKeys(currentValue)
					return
				}

				removeOrReset, _ := cmd.Flags().GetBool("remove-reset")
				if removeOrReset {
					parentKey, childKey := splitIntoParentChild(key)
					parentValue := viper.GetStringMap(parentKey)
					if parentKey == "" {
						parentValue = viper.AllSettings()
					}

					delete(parentValue, childKey)
					viper.Set(parentKey, parentValue)
					if parentKey == "" {
						overwriteInMemConfig(parentValue)
					} else {
						overwriteInMemConfig(viper.AllSettings())
					}
					saveConfig(false)
					return
				}

				if arrayValue, ok := currentValue.([]string); ok {
					for _, value := range arrayValue {
						fmt.Println(value)
					}
					return
				}

				t, err := yaml.Marshal(currentValue)
				if err != nil {
					fmt.Println(err)
					return
				}

				str := strings.TrimSpace(string(t))
				if str == `""` {
					str = ""
				}

				fmt.Println(str)
				return
			}

			newValue, err := matchConfigType(currentValue, args[1])
			if err != nil {
				fmt.Println(err)
				return
			}

			// If parentKey is empty, we are setting a root key
			// If parentValue is nil, we are setting a nested key which has not been set
			parentKey, _ := splitIntoParentChild(key)
			parentValue := viper.Get(parentKey)
			if _, ok := parentValue.(map[string]interface{}); !ok &&
				parentKey != "" && parentValue != nil {
				fmt.Printf("Key '%s' is not a key/value object", parentKey)
				return
			}

			viper.Set(key, newValue)
			saveConfig(false)
		}
	},
}

func saveConfig(saveNew bool) {
	fmt.Println("Writing Config")

	fileName := viper.ConfigFileUsed()
	if saveNew || fileName == "" {
		fileName = defaultConfigFile
	}

	viper.WriteConfigAs(fileName)
}

func setOrRefreshConfig() error {
	if !configInitialized {
		initConfig()
	}

	var c config.Config
	err := viper.Unmarshal(&c)
	if err != nil {
		return err
	}

	config.Set(&c)

	return nil
}

func overwriteInMemConfig(configMap map[string]interface{}) error {
	encodedConfig, _ := yaml.Marshal(configMap)
	return viper.ReadConfig(bytes.NewReader(encodedConfig))
}

// Based on the current value of the config, attempts to return the user's value as the right type
func matchConfigType(currentValue interface{}, userValue string) (interface{}, error) {
	// Currently our config values are either bool, int, string, or []string
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
	} else if slice, ok := currentValue.([]interface{}); ok {
		// If the slice is a string slice or is empty, we can append to it
		if (len(slice) > 0 && reflect.TypeOf(slice[0]).String() == "string") ||
			len(slice) == 0 {
			userValues := strings.Split(userValue, ",")
			for _, value := range userValues {
				slice = append(slice, value)
			}
			return slice, nil
		}
	} else if currentValue == nil { // If currentValue is nil, we are setting a new key and we must guess the value type
		intValue, err := strconv.Atoi(userValue)
		if err == nil {
			return intValue, nil
		}

		boolValue, err := strconv.ParseBool(userValue)
		if err == nil {
			return boolValue, nil
		}

		return userValue, nil
	}
	return nil, fmt.Errorf("cannot set keys that are not of type int, string, []string or bool")
}

func printConfig() {
	t, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(t))
}

func printKeys(value interface{}) {
	if valueMap, ok := value.(map[string]interface{}); ok {
		for key := range valueMap {
			fmt.Println(key)
		}
	} else {
		fmt.Println("No sub-keys")
	}
}

func splitIntoParentChild(key string) (string, string) {
	split := strings.Split(key, ".")
	keyCount := len(split)
	if keyCount == 1 {
		return "", key
	}

	return strings.Join(split[:keyCount-1], "."), split[keyCount-1]
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolP("save", "s", false, "save the current configuration")
	configCmd.Flags().BoolP("sub-keys", "k", false, "display only the sub-keys")
	configCmd.Flags().BoolP("remove-reset", "r", false, "remove key from the config or reset to default")
}
