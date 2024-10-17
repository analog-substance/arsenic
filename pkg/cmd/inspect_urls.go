package cmd

import (
	"fmt"
	"github.com/analog-substance/nex/pkg/nmap"
	"github.com/analog-substance/scopious/pkg/scopious"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// inspectHostsCmd represents the ingest command
var inspectURLsCmd = &cobra.Command{
	Use:   "urls",
	Short: "print URLs",
	Run: func(cmd *cobra.Command, args []string) {
		scopeDir, _ := cmd.Flags().GetString("scope-dir")
		includePublic, _ := cmd.Flags().GetBool("public")
		includePrivate, _ := cmd.Flags().GetBool("private")

		protocolPrefix, _ := cmd.Flags().GetString("protocol")
		//includePublic, _ := cmd.Flags().GetBool("public")
		//includePrivate, _ := cmd.Flags().GetBool("private")

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

		urls := nmapView.GetURLs(protocolPrefix, viewOptions)

		fmt.Println(strings.Join(urls, "\n"))

	},
}

func init() {
	inspectCmd.AddCommand(inspectURLsCmd)

	inspectURLsCmd.Flags().Bool("hostnames", false, "Just list hostnames")
	inspectURLsCmd.Flags().Bool("private", false, "Only show hosts with private IPs")
	inspectURLsCmd.Flags().Bool("public", false, "Only show hosts with public IPs")
	inspectURLsCmd.Flags().Bool("ips", false, "Just list IP addresses")
	inspectURLsCmd.Flags().StringP("protocol", "p", "", "protocol prefix")
}
