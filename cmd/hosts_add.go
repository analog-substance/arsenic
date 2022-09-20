package cmd

import (
	"fmt"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/spf13/cobra"
)

// hostsAddCmd represents the add command
var hostsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new host",
	Run: func(cmd *cobra.Command, args []string) {
		ips, _ := cmd.Flags().GetStringSlice("ips")
		hostnames, _ := cmd.Flags().GetStringSlice("hostnames")

		allHostnames := append(hostnames, ips...)
		hosts := host.Get(allHostnames...)
		if len(hosts) > 0 {
			// Since host.Get() can get empty host directories
			// we need to make sure the hosts are valid
			for _, h := range hosts {
				if len(h.Metadata.IPAddresses) != 0 {
					fmt.Println("[!] Host already exists")
					return
				}
			}
		}

		h, err := host.AddHost(hostnames, ips)
		if err != nil {
			fmt.Println(err)
			return
		}

		if h != nil {
			fmt.Printf("[+] Host %s added\n", h.Metadata.Name)
		}
	},
}

func init() {
	hostsCmd.AddCommand(hostsAddCmd)

	hostsAddCmd.Flags().StringSliceP("ips", "i", []string{}, "IP addresses for the host")
	hostsAddCmd.MarkFlagRequired("ips")

	hostsAddCmd.Flags().StringSliceP("hostnames", "H", []string{}, "Hostnames for the host")
}
