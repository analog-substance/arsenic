package cmd

import (
	"fmt"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"

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


	defaultDiscoverScripts := make(map[string]util.ScriptConfig)
	defaultReconScripts := make(map[string]util.ScriptConfig)
	defaultHuntScripts := make(map[string]util.ScriptConfig)

	defaultDiscoverScripts["as-subdomain-discovery"] = util.ScriptConfig{"as-subdomain-discovery", 0, true}
	defaultDiscoverScripts["as-subdomain-enumeration"] = util.ScriptConfig{"as-subdomain-enumeration", 100, true}
	defaultDiscoverScripts["as-domains-from-domain-ssl-certs"] = util.ScriptConfig{"as-domains-from-domain-ssl-certs", 200, true}
	defaultDiscoverScripts["as-dns-resolution"] = util.ScriptConfig{"as-dns-resolution", 300, true}
	defaultDiscoverScripts["as-ip-recon"] = util.ScriptConfig{"as-ip-recon", 400, true}
	defaultDiscoverScripts["as-domains-from-ip-ssl-certs"] = util.ScriptConfig{"as-domains-from-ip-ssl-certs", 500, true}
	defaultDiscoverScripts["as-ip-resolution"] = util.ScriptConfig{"as-ip-resolution", 600, true}
	defaultDiscoverScripts["as-http-screenshot-domains"] = util.ScriptConfig{"as-http-screenshot-domains", 700, true}

	defaultReconScripts["as-port-scan-tcp"] = util.ScriptConfig{"as-port-scan-tcp", 0, true}
	defaultReconScripts["as-content-discovery"] = util.ScriptConfig{"as-content-discovery", 100, true}
	defaultReconScripts["as-http-screenshot-hosts"] = util.ScriptConfig{"as-http-screenshot-hosts", 200, true}
	defaultReconScripts["as-port-scan-udp"] = util.ScriptConfig{"as-port-scan-udp", 300, true}

	defaultHuntScripts["as-takeover-aquatone"] = util.ScriptConfig{"as-takeover-aquatone", 0, true}
	defaultHuntScripts["as-searchsploit"] = util.ScriptConfig{"as-searchsploit", 100, true}

	// viper.SetDefault("DiscoverScripts", defaultDiscoverScripts)
	// viper.SetDefault("ReconScripts", defaultReconScripts)

	defaultScripts := make(map[string]map[string]util.ScriptConfig)
	defaultScripts["discover"] = defaultDiscoverScripts
	defaultScripts["recon"] = defaultReconScripts
	defaultScripts["hunt"] = defaultHuntScripts

	viper.SetDefault("scripts", defaultScripts)
	viper.SetDefault("sec_lists_path", "/opt/SecLists")

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

}
