package cmd

import (
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

		listPublicIPs, _ := cmd.Flags().GetBool("pub-ips")
		listPrivateIPs, _ := cmd.Flags().GetBool("priv-ips")
		listIPs, _ := cmd.Flags().GetBool("ips")
		listPublicHostnames, _ := cmd.Flags().GetBool("pub-hostnames")
		listPrivateHostnames, _ := cmd.Flags().GetBool("priv-hostnames")
		listHostnames, _ := cmd.Flags().GetBool("hostnames")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		openOnly, _ := cmd.Flags().GetBool("open")
		upOnly, _ := cmd.Flags().GetBool("up")

		files, err := filepath.Glob(fmt.Sprintf("data/%s/output/nmap/*/*.xml", scopeDir))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(files) == 0 {
			check(fmt.Errorf("no files found"))
		}

		var opts []nmap.Option
		if upOnly {
			opts = append(opts, nmap.WithUpOnly())
		}
		if openOnly {
			opts = append(opts, nmap.WithOpenOnly())
		}

		run, err := nmap.XMLMerge(files, opts...)
		check(err)

		nmapView := nmap.NewNmapView(run)

		if jsonOutput {
			err = nmapView.PrintJSON()
			check(err)
			return
		}

		viewOptions := nmap.ViewOptions(0)
		if listHostnames {
			listPublicHostnames = true
			listPrivateHostnames = true
		}

		if listIPs {
			listPublicIPs = true
			listPrivateIPs = true
		}

		if listPublicHostnames {
			viewOptions = viewOptions | nmap.ViewListPublicHostnames
		}
		if listPrivateHostnames {
			viewOptions = viewOptions | nmap.ViewListPrivateHostnames
		}
		if listPublicIPs {
			viewOptions = viewOptions | nmap.ViewListPublicIPs
		}
		if listPrivateIPs {
			viewOptions = viewOptions | nmap.ViewListPrivateIPs
		}

		if viewOptions > 0 {
			nmapView.PrintList(viewOptions)
			return
		}

		sortBy, _ := cmd.Flags().GetString("sort-by")
		// no options specified
		nmapView.PrintTable(sortBy)

	},
}

func init() {
	inspectCmd.AddCommand(inspectHostsCmd)

	inspectHostsCmd.Flags().String("sort-by", "Name;asc", "Sort by the specified column. Format: column[;(asc|dsc)]")
	inspectHostsCmd.Flags().Bool("open", false, "Show only hosts with open ports")
	inspectHostsCmd.Flags().Bool("up", false, "Show only hosts that are up")
	inspectHostsCmd.Flags().Bool("pub-hostnames", false, "Just print public hostnames")
	inspectHostsCmd.Flags().Bool("priv-hostnames", false, "Just print private hostnames")
	inspectHostsCmd.Flags().Bool("hostnames", false, "Just print hostnames")
	inspectHostsCmd.Flags().Bool("pub-ips", false, "Just print public IP addresses")
	inspectHostsCmd.Flags().Bool("priv-ips", false, "Just print private IP addresses")
	inspectHostsCmd.Flags().Bool("ips", false, "Just print IP addresses")
	inspectHostsCmd.Flags().Bool("json", false, "Print JSON")
}

func check(err error) {
	if err != nil {
		fmt.Printf("[!] %v", err)
		os.Exit(1)
	}
}
