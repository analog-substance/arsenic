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

		includePublic, _ := cmd.Flags().GetBool("public")
		includePrivate, _ := cmd.Flags().GetBool("private")
		listIPs, _ := cmd.Flags().GetBool("ips")
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

		viewOptions := nmap.ListViewOptions(0)
		if listHostnames {
			if includePublic {
				viewOptions = viewOptions | nmap.ListViewPublicHostnames
			}
			if includePrivate {
				viewOptions = viewOptions | nmap.ListViewPrivateHostnames
			}
		}

		if listIPs {
			if includePublic {
				viewOptions = viewOptions | nmap.ListViewPublicIPs
			}
			if includePrivate {
				viewOptions = viewOptions | nmap.ListViewPrivateIPs
			}
		}

		if viewOptions > 0 {
			nmapView.PrintList(viewOptions)
			return
		}

		tableViewOptions := nmap.TableViewOptions(0)
		if includePublic {
			tableViewOptions = tableViewOptions | nmap.TableViewPublic
		}
		if includePrivate {
			tableViewOptions = tableViewOptions | nmap.TableViewPrivate
		}

		sortBy, _ := cmd.Flags().GetString("sort-by")
		// no options specified
		nmapView.PrintTable(sortBy, tableViewOptions)

	},
}

func init() {
	inspectCmd.AddCommand(inspectHostsCmd)

	inspectHostsCmd.Flags().String("sort-by", "Name;asc", "Sort by the specified column. Format: column[;(asc|dsc)]")
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
