package cmd

import (
	"fmt"
	"github.com/analog-substance/nex/pkg/nmap"
	"github.com/analog-substance/scopious/pkg/scopious"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
)

// inspectHostsCmd represents the ingest command
var inspectHostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "host data",
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")

		includePublic, _ := cmd.Flags().GetBool("public")
		includePrivate, _ := cmd.Flags().GetBool("private")
		listIPs, _ := cmd.Flags().GetBool("ips")
		listHostnames, _ := cmd.Flags().GetBool("hostnames")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		openOnly, _ := cmd.Flags().GetBool("open")
		upOnly, _ := cmd.Flags().GetBool("up")

		scoper := scopious.FromPath("data")
		scope := scoper.GetScope(scopeDir)

		files, err := filepath.Glob(filepath.Join(scope.Path, "output", "nmap", "*", "*.xml"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(files) == 0 {
			check(fmt.Errorf("no files found"))
		}

		var opts []nmap.Option
		run, err := nmap.XMLMerge(files, opts...)
		check(err)

		nmapView := nmap.NewNmapView(run)

		nmapView.SetFilter(func(hostnames []string, ips []string) bool {
			for _, hostname := range hostnames {
				if !scope.IsDomainInScope(hostname, false) {
					return false
				}
			}
			for _, ip := range ips {
				ip := net.ParseIP(ip)
				if ip == nil {
					continue
				}
				if !scope.IsIPInScope(&ip, false) {
					return false
				}
			}
			return true
		})

		viewOptions := nmap.ViewOptions(0)
		if includePublic {
			viewOptions = viewOptions | nmap.ViewPublic
		}

		if includePrivate {
			viewOptions = viewOptions | nmap.ViewPrivate
		}

		if upOnly {
			viewOptions = viewOptions | nmap.ViewAliveHosts
		}

		if openOnly {
			viewOptions = viewOptions | nmap.ViewOpenPorts
		}

		if jsonOutput {
			err = nmapView.PrintJSON(viewOptions)
			check(err)
			return
		}

		if listHostnames || listIPs {
			if listHostnames {
				viewOptions = viewOptions | nmap.ListHostnames
			}
			if listIPs {
				viewOptions = viewOptions | nmap.ListIPs
			}

			nmapView.PrintList(viewOptions)
			return
		}

		sortBy, _ := cmd.Flags().GetString("sort-by")
		// no options specified
		nmapView.PrintTable(sortBy, viewOptions)

	},
}

func init() {
	inspectCmd.AddCommand(inspectHostsCmd)

	inspectHostsCmd.Flags().String("sort-by", "hostnames;asc", "Sort by the specified column. Format: column[;(asc|dsc)]")
	inspectHostsCmd.Flags().Bool("open", false, "Show only hosts with open ports")
	inspectHostsCmd.Flags().Bool("up", false, "Show only hosts that are up")
	inspectHostsCmd.Flags().Bool("hostnames", false, "Just list hostnames")
	inspectHostsCmd.Flags().Bool("ips", false, "Just list IP addresses")
	inspectHostsCmd.Flags().Bool("private", false, "Only show hosts with private IPs")
	inspectHostsCmd.Flags().Bool("public", false, "Only show hosts with public IPs")
	inspectHostsCmd.Flags().Bool("json", false, "Print JSON")
}

func check(err error) {
	if err != nil {
		fmt.Printf("[!] %v", err)
		os.Exit(1)
	}
}
