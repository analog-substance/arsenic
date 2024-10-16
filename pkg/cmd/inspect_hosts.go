package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/analog-substance/nex/pkg/nmap"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

// inspectHostsCmd represents the ingest command
var inspectHostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "host data",
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")
		nmapXMLFiles, err := filepath.Glob(fmt.Sprintf("data/%s/output/nmap/*/*.xml", scopeDir))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(nmapXMLFiles) == 0 {
			fmt.Println("No nmap.xml files found")
			os.Exit(1)
		}

		var opts []nmap.Option
		//if upOnly {
		//opts = append(opts, nmap.WithUpOnly())
		//}
		//if openOnly {
		//opts = append(opts, nmap.WithOpenOnly())
		//}

		run, err := nmap.XMLMerge(nmapXMLFiles, opts...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		output, err := json.MarshalIndent(run.Hosts, "", "  ")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func init() {
	inspectCmd.AddCommand(inspectHostsCmd)
}
