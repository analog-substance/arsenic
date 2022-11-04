package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/analog-substance/arsenic/lib/engine"
	"github.com/analog-substance/arsenic/lib/util"
)

var cfgFile string
var configInitialized bool = false

var rootCmd = &cobra.Command{
	Use:     "arsenic",
	Version: "v0.2.0",
	Short:   "Pentesting conventions",
	Long: `Arsenic - Pentest Conventions


`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		script := engine.NewScript(args[0])

		err := script.Run(args[1:])
		if err != nil && err != context.Canceled {
			panic(err)
		}
	},
}

func Execute() {
	if len(os.Args) > 1 && strings.Contains(os.Args[1], fmt.Sprintf("%c", os.PathSeparator)) {
		rootCmd.DisableFlagParsing = true
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "the arsenic.yaml config file")
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
	defaultInitScripts := make(map[string]util.ScriptConfig)

	defaultInitScripts["as-init-op"] = util.NewScriptConfig("as-init-op", 0, 1, true)
	defaultInitScripts["as-setup-hugo"] = util.NewScriptConfig("as-setup-hugo", 100, 1, true)
	defaultInitScripts["as-init-hooks"] = util.NewScriptConfig("as-init-hooks", 200, 1, true)
	defaultInitScripts["as-init-cleanup"] = util.NewScriptConfig("as-init-cleanup", 300, 1, true)

	defaultDiscoverScripts["as-root-domain-recon"] = util.NewScriptConfig("as-root-domain-recon", 0, 1, true)
	defaultDiscoverScripts["as-subdomain-discovery"] = util.NewScriptConfig("as-subdomain-discovery", 50, 1, true)
	defaultDiscoverScripts["as-subdomain-enumeration"] = util.NewScriptConfig("as-subdomain-enumeration", 100, 1, true)
	defaultDiscoverScripts["as-combine-subdomains"] = util.NewScriptConfig("as-combine-subdomains", 250, 2, true)
	defaultDiscoverScripts["as-dns-resolution"] = util.NewScriptConfig("as-dns-resolution", 300, 2, true)
	defaultDiscoverScripts["as-domains-from-domain-ssl-certs"] = util.NewScriptConfig("as-domains-from-domain-ssl-certs", 200, 1, true)
	defaultDiscoverScripts["as-ip-recon"] = util.NewScriptConfig("as-ip-recon", 400, 2, true)
	defaultDiscoverScripts["as-domains-from-ip-ssl-certs"] = util.NewScriptConfig("as-domains-from-ip-ssl-certs", 500, 2, true)
	defaultDiscoverScripts["as-ip-resolution"] = util.NewScriptConfig("as-ip-resolution", 600, 2, true)

	defaultDiscoverScripts["as-http-screenshot-domains"] = util.NewScriptConfig("as-http-screenshot-domains", 700, 1, true)

	defaultReconScripts["as-port-scan-tcp"] = util.NewScriptConfig("as-port-scan-tcp", 0, 1, true)
	defaultReconScripts["as-content-discovery"] = util.NewScriptConfig("as-content-discovery", 100, 1, true)
	defaultReconScripts["as-http-screenshot-hosts"] = util.NewScriptConfig("as-http-screenshot-hosts", 200, 1, true)
	defaultReconScripts["as-port-scan-udp"] = util.NewScriptConfig("as-port-scan-udp", 300, 1, true)

	defaultHuntScripts["as-takeover-aquatone"] = util.NewScriptConfig("as-takeover-aquatone", 0, 1, true)
	defaultHuntScripts["as-searchsploit"] = util.NewScriptConfig("as-searchsploit", 100, 1, true)
	defaultHuntScripts["as-nuclei-technologies"] = util.NewScriptConfig("as-nuclei-technologies", 200, 1, true)
	defaultHuntScripts["as-nuclei-cves"] = util.NewScriptConfig("as-nuclei-cves", 300, 1, true)

	defaultScripts := make(map[string]map[string]util.ScriptConfig)
	defaultScripts["init"] = defaultInitScripts
	defaultScripts["discover"] = defaultDiscoverScripts
	defaultScripts["recon"] = defaultReconScripts
	defaultScripts["hunt"] = defaultHuntScripts

	wordlists := make(map[string][]string)
	wordlists["web-content"] = []string{
		"Discovery/Web-Content/AdobeCQ-AEM.txt",
		"Discovery/Web-Content/apache.txt",
		"Discovery/Web-Content/Common-DB-Backups.txt",
		"Discovery/Web-Content/Common-PHP-Filenames.txt",
		"Discovery/Web-Content/common.txt",
		"Discovery/Web-Content/confluence-administration.txt",
		"Discovery/Web-Content/default-web-root-directory-linux.txt",
		"Discovery/Web-Content/default-web-root-directory-windows.txt",
		"Discovery/Web-Content/frontpage.txt",
		"Discovery/Web-Content/graphql.txt",
		"Discovery/Web-Content/jboss.txt",
		"Discovery/Web-Content/Jenkins-Hudson.txt",
		"Discovery/Web-Content/nginx.txt",
		"Discovery/Web-Content/oracle.txt",
		"Discovery/Web-Content/quickhits.txt",
		"Discovery/Web-Content/raft-large-directories.txt",
		"Discovery/Web-Content/raft-medium-words.txt",
		"Discovery/Web-Content/reverse-proxy-inconsistencies.txt",
		"Discovery/Web-Content/RobotsDisallowed-Top1000.txt",
		"Discovery/Web-Content/websphere.txt",
	}

	wordlists["sqli"] = []string{
		"Fuzzing/Databases/sqli.auth.bypass.txt",
		"Fuzzing/Databases/MSSQL.fuzzdb.txt",
		"Fuzzing/Databases/MSSQL-Enumeration.fuzzdb.txt",
		"Fuzzing/Databases/MySQL.fuzzdb.txt",
		"Fuzzing/Databases/NoSQL.txt",
		"Fuzzing/Databases/db2enumeration.fuzzdb.txt",
		"Fuzzing/Databases/Oracle.fuzzdb.txt",
		"Fuzzing/Databases/MySQL-Read-Local-Files.fuzzdb.txt",
		"Fuzzing/Databases/Postgres-Enumeration.fuzzdb.txt",
		"Fuzzing/Databases/MySQL-SQLi-Login-Bypass.fuzzdb.txt",
		"Fuzzing/SQLi/Generic-BlindSQLi.fuzzdb.txt",
		"Fuzzing/SQLi/Generic-SQLi.txt",
		"Fuzzing/SQLi/quick-SQLi.txt",
	}

	wordlists["xss"] = []string{
		"Fuzzing/XSS/XSS-Somdev.txt",
		"Fuzzing/XSS/XSS-Bypass-Strings-BruteLogic.txt",
		"Fuzzing/XSS/XSS-Jhaddix.txt",
		"Fuzzing/XSS/xss-without-parentheses-semi-colons-portswigger.txt",
		"Fuzzing/XSS/XSS-RSNAKE.txt",
		"Fuzzing/XSS/XSS-Cheat-Sheet-PortSwigger.txt",
		"Fuzzing/XSS/XSS-BruteLogic.txt",
		"Fuzzing/XSS-Fuzzing",
	}

	blacklistedRootDomains := []string{
		"1e100.net",
		"akamaitechnologies.com",
		"amazonaws.com",
		"azure.com",
		"azurewebsites.net",
		"azurewebsites.windows.net",
		"c7dc.com",
		"cas.ms",
		"cloudapp.net",
		"cloudfront.net",
		"googlehosted.com",
		"googleusercontent.com",
		"hscoscdn10.net",
		"my.jobs",
		"readthedocs.io",
		"readthedocs.org",
		"sites.hubspot.net",
		"tds.net",
		"wixsite.com",
	}

	ignoreServices := []util.IgnoreService{
		{
			Name:  "msrpc",
			Ports: "40000-65535",
			Flag:  "ignored::ephemeral-msrpc",
		},
		{
			Name:  "tcpwrapped",
			Ports: "all",
		},
		{
			Name:  "unknown",
			Ports: "all",
		},
	}

	setConfigDefault("ignore-services", ignoreServices)
	setConfigDefault("blacklist.root-domains", blacklistedRootDomains)
	setConfigDefault("blacklist.domains", []string{})
	setConfigDefault("blacklist.ips", []string{})
	setConfigDefault("scripts-directory", filepath.Join(home, ".config", "arsenic"))
	setConfigDefault("scripts", defaultScripts)
	setConfigDefault("wordlists", wordlists)
	setConfigDefault("wordlist-paths", []string{
		"/opt/SecLists",
		"/usr/share/seclists",
	})

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

	configInitialized = true
}

// If no config file exists, all possible keys in the defaults
// need to be registered with viper otherwise viper will only think
// the keys explicitly set via viper.SetDefault() exist.
func setConfigDefault(key string, value interface{}) {
	valueType := reflect.TypeOf(value)
	valueValue := reflect.ValueOf(value)

	if valueType.Kind() == reflect.Map {
		iter := valueValue.MapRange()
		for iter.Next() {
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			setConfigDefault(fmt.Sprintf("%s.%s", key, k), v)
		}
	} else if valueType.Kind() == reflect.Struct {
		numFields := valueType.NumField()
		for i := 0; i < numFields; i++ {
			structField := valueType.Field(i)
			fieldValue := valueValue.Field(i)

			setConfigDefault(fmt.Sprintf("%s.%s", key, structField.Name), fieldValue.Interface())
		}
	} else {
		viper.SetDefault(key, value)
	}
}
