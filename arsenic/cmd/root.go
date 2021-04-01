package cmd

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "arsenic",
	Short: "Arsenic - Pentest Conventions",
	Long: `Arsenic - Pentest Conventions


`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.arsenic.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	defaultVarDirs := []string{}
	defaultVarDirs = append(defaultVarDirs, "/opt/arsenic/var/")
	defaultVarDirs = append(defaultVarDirs, fmt.Sprintf("%s/opt/arsenic/var/", home))
	defaultVarDirs = append(defaultVarDirs, fmt.Sprintf("%s/as/var/", cwd))

	viper.SetDefault("varDirs", defaultVarDirs)
	viper.SetDefault("secListsPath", "/opt/SecLists")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name ".arsenic" (without extension).
		viper.AddConfigPath(cwd)
		viper.AddConfigPath(home)
		viper.SetConfigName(".arsenic")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Using default configuration")
			fmt.Println()
		} else {
			fmt.Println("Error Using config file:", viper.ConfigFileUsed())
		}
	}
}
