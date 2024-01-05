package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/fileutil"
	"github.com/spf13/cobra"
)

type naabuResult struct {
	IP   string `json:"ip"`
	Port struct {
		Port     int  `json:"Port"`
		Protocol int  `json:"Protocol"`
		TLS      bool `json:"TLS"`
	} `json:"Port"`
}

// ingestNaabuCmd represents the naabu command
var ingestNaabuCmd = &cobra.Command{
	Use:   "naabu files...",
	Short: "Import naabu port scan output, creating/updating hosts with the open ports",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostMap := make(map[string][]string)
		for _, file := range args {
			lines, err := fileutil.ReadLineByLine(file)
			if err != nil {
				return err
			}

			if filepath.Ext(file) == ".json" {
				for line := range lines {
					var result naabuResult
					err = json.Unmarshal([]byte(line), &result)
					if err != nil {
						return err
					}

					hostMap[result.IP] = append(hostMap[result.IP], fmt.Sprintf("%d", result.Port.Port))
				}
			} else {
				for line := range lines {
					parts := strings.Split(line, ":")
					host := parts[0]
					port := parts[1]

					hostMap[host] = append(hostMap[host], port)
				}
			}

		}

		var err error
		for name, ports := range hostMap {
			h := host.GetFirst(name)
			if h == nil {
				var hostnames []string
				var ips []string

				if util.IsIp(name) {
					ips = append(ips, name)
				} else {
					hostnames = append(hostnames, name)
				}

				fmt.Printf("[+] Adding host %s\n", name)

				h, err = host.AddHost(hostnames, ips)
				if err != nil {
					return err
				}
			} else {
				fmt.Printf("[+] Updating host %s\n", name)
			}

			portsFile := filepath.Join(h.Dir, "recon", "naabu-tcp-ports.txt")

			portSet := set.NewStringSet(ports)
			if fileutil.FileExists(portsFile) {
				p, err := fileutil.ReadLines(portsFile)
				if err != nil {
					return err
				}

				portSet.AddRange(p)
			}

			allPorts := portSet.StringSlice()
			sort.Slice(allPorts, func(i, j int) bool {
				iInt, _ := strconv.Atoi(allPorts[i])
				jInt, _ := strconv.Atoi(allPorts[j])
				return iInt < jInt
			})

			err = fileutil.WriteLines(portsFile, allPorts)
			if err != nil {
				return err
			}

			err = h.SyncMetadata(host.SyncOptions{
				Ports: true,
			})
			if err != nil {
				return err
			}

			h.SaveMetadata()
		}

		return nil
	},
}

func init() {
	ingestCmd.AddCommand(ingestNaabuCmd)
}
