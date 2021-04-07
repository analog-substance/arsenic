package cmd

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/defektive/arsenic/arsenic/lib/util"
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "the arsenic.yaml config file")

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

	defaultDiscoverScripts := make(map[string]interface{})
	defaultReconScripts := make(map[string]interface{})
	defaultHuntScripts := make(map[string]interface{})

	defaultDiscoverScripts["as-subdomain-discovery"] = util.NewScriptConfig("as-subdomain-discovery", 0, true)
	defaultDiscoverScripts["as-subdomain-enumeration"] = util.NewScriptConfig("as-subdomain-enumeration", 100, true)
	defaultDiscoverScripts["as-domains-from-domain-ssl-certs"] = util.NewScriptConfig("as-domains-from-domain-ssl-certs", 200, true)
	defaultDiscoverScripts["as-dns-resolution"] = util.NewScriptConfig("as-dns-resolution", 300, true)
	defaultDiscoverScripts["as-ip-recon"] = util.NewScriptConfig("as-ip-recon", 400, true)
	defaultDiscoverScripts["as-domains-from-ip-ssl-certs"] = util.NewScriptConfig("as-domains-from-ip-ssl-certs", 500, true)
	defaultDiscoverScripts["as-ip-resolution"] = util.NewScriptConfig("as-ip-resolution", 600, true)
	defaultDiscoverScripts["as-http-screenshot-domains"] = util.NewScriptConfig("as-http-screenshot-domains", 700, true)

	defaultReconScripts["as-port-scan-tcp"] = util.NewScriptConfig("as-port-scan-tcp", 0, true)
	defaultReconScripts["as-content-discovery"] = util.NewScriptConfig("as-content-discovery", 100, true)
	defaultReconScripts["as-http-screenshot-hosts"] = util.NewScriptConfig("as-http-screenshot-hosts", 200, true)
	defaultReconScripts["as-port-scan-udp"] = util.NewScriptConfig("as-port-scan-udp", 300, true)

	defaultHuntScripts["as-takeover-aquatone"] = util.NewScriptConfig("as-takeover-aquatone", 0, true)
	defaultHuntScripts["as-searchsploit"] = util.NewScriptConfig("as-searchsploit", 100, true)

	// viper.SetDefault("DiscoverScripts", defaultDiscoverScripts)
	// viper.SetDefault("ReconScripts", defaultReconScripts)

	defaultScripts := make(map[string]interface{})
	defaultScripts["discover"] = defaultDiscoverScripts
	defaultScripts["recon"] = defaultReconScripts
	defaultScripts["hunt"] = defaultHuntScripts

	// viper.SetDefault("scripts", defaultScripts)
	setConfigDefault("scripts", defaultScripts)
	setConfigDefault("sec-lists-path", "/opt/SecLists")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name "arsenic" (without extension).
		viper.AddConfigPath(cwd)
		viper.AddConfigPath(home)
		viper.SetConfigName("arsenic")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.ReadInConfig()
}

// If no config file exists, all possible keys in the defaults
// need to be registered with viper otherwise viper will only think
// scripts and sec-lists-path are valid keys
func setConfigDefault(key string, value interface{}) {
	if valueMap, ok := value.(map[string]interface{}); ok {
		for k, v := range valueMap {
			setConfigDefault(fmt.Sprintf("%s.%s", key, k), v)
		}
	} else if mappable, ok := value.(util.Mappable); ok {
		valueMap := mappable.ToMap()
		for k, v := range valueMap {
			setConfigDefault(fmt.Sprintf("%s.%s", key, k), v)
		}
	} else {
		viper.SetDefault(key, value)
	}
}
