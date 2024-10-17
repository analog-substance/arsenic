package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/ahmetb/go-linq/v3"
	"github.com/analog-substance/arsenic/pkg/host"
	"github.com/analog-substance/arsenic/pkg/util"
	"github.com/analog-substance/util/set"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

func defaultFuncMap() template.FuncMap {
	funcMap := make(template.FuncMap)

	funcMap["in"] = func(s1 []string, s2 ...string) bool {
		// Loop through s1 then s2 and check whether any values of s2 are equal to any values in s1
		return linq.From(s1).AnyWith(func(s1Item interface{}) bool {
			return linq.From(s2).AnyWith(func(s2Item interface{}) bool {
				return s1Item == s2Item
			})
		})
	}

	funcMap["join"] = func(sep string, v interface{}) string {
		return strings.Join(util.ToStringSlice(v), sep)
	}

	return funcMap
}

// hostsCmd represents the flags command
var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "View, query, and flag hosts",
	Long: `View, query, and flag hosts

Show unreviewed hosts:

  $ arsenic hosts -q '.HasFlags "Unreviewed"'

Show hosts that have Gobuster results:

  $ arsenic hosts -q '.HasFlags "Gobuster"'

Show hosts with the root domain example.com:

  $ arsenic hosts -q 'in .RootDomains "example.com"'

Show hosts with ports 22 or 2022:

  $ arsenic hosts -q '.HasPorts 22 2022'

Show hosts who are in a CIDR block

  $ arsenic hosts -q '.InCIDR "10.1.1.0/24"'

Metadata:
	Methods:
	- HasPorts(ports ...int) bool
	- HasAnyPort() bool
	- HasTCPPorts(ports ...int) bool
	- HasAnyTCPPort() bool
	- HasUDPPorts(ports ...int) bool
	- HasAnyUDPPort() bool
	- HasFlags(flags ...string) bool
	- HasAllFlags(flags ...string) bool
	- HasASFlags(flags ...string) bool
	- HasAllASFlags(flags ...string) bool
	- HasUserFlags(flags ...string) bool
	- HasAllUserFlags(flags ...string) bool
	- HasAnyHostname() bool
	- InCIDR(cidrStr string) bool

	Fields:
	- Name        string
	- Hostnames   []string
	- RootDomains []string
	- IPAddresses []string
	- Flags       []string
	- UserFlags   []string
	- TCPPorts    []int
	- UDPPorts    []int
	- Ports       []Port
	- ReviewedBy  string

Port:
	Fields:
	- ID       int // The port number
	- Protocol string
	- Service  string
`,
	Run: func(cmd *cobra.Command, args []string) {
		hostsArgs, _ := cmd.Flags().GetStringSlice("host")

		query, _ := cmd.Flags().GetString("query")

		var hosts []*host.Host
		if len(hostsArgs) > 0 {
			hosts = host.Get(hostsArgs...)
		} else if query != "" {
			hostTemplate := template.New("host")
			funcMap := defaultFuncMap()

			funcMap["appendMatch"] = func(match *host.Host) string {
				hosts = append(hosts, match)
				return ""
			}

			templateString := fmt.Sprintf(`{{range $host := .}}{{with .Metadata}}{{if %s}}{{appendMatch $host}}{{end}}{{end}}{{end}}`, query)
			_, err := hostTemplate.Funcs(funcMap).Parse(templateString)
			if err != nil {
				cmd.PrintErrln(err)
				return
			}

			err = hostTemplate.Execute(util.NoopWriter{}, host.All())
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
		} else {
			hosts = host.All()
		}

		addReviewedBy := false
		reviewerFlag, _ := cmd.Flags().GetString("reviewed-by")

		if cmd.Flags().Lookup("reviewed-by").Changed && (len(hostsArgs) > 0 || len(query) > 0) {
			addReviewedBy = true
		}
		reviewer := util.GetReviewer(reviewerFlag)
		userFlagsToAdd, _ := cmd.Flags().GetStringSlice("add-flags")
		userFlagsToRemove, _ := cmd.Flags().GetStringSlice("remove-flags")
		namesToAdd, _ := cmd.Flags().GetStringSlice("add-names")
		namesToRemove, _ := cmd.Flags().GetStringSlice("remove-names")
		updateArsenicFlags, _ := cmd.Flags().GetBool("update")
		jsonOut, _ := cmd.Flags().GetBool("json")
		pathsOut, _ := cmd.Flags().GetBool("paths")

		shouldSave := len(userFlagsToRemove) > 0 || len(userFlagsToAdd) > 0 || len(namesToAdd) > 0 || len(namesToRemove) > 0 || updateArsenicFlags || addReviewedBy

		if shouldSave {
			if (len(namesToAdd) > 0 || len(namesToRemove) > 0) && len(hosts) == 1 {
				host := hosts[0]
				noHostnames := len(host.Metadata.Hostnames) == 0

				hostnames := set.NewSet("")
				for _, hostname := range host.Metadata.Hostnames {
					if linq.From(namesToRemove).AnyWith(func(item interface{}) bool { return strings.EqualFold(hostname, item.(string)) }) {
						continue
					}
					hostnames.Add(hostname)
				}
				hostnames.AddRange(namesToAdd)

				host.Metadata.Hostnames = hostnames.SortedStringSlice()
				if noHostnames && len(host.Metadata.Hostnames) > 0 {
					host.Metadata.Name = host.Metadata.Hostnames[0]
					newDir := filepath.Join("hosts", host.Metadata.Name)

					err := os.Rename(host.Dir, newDir)
					if err != nil {
						fmt.Println(err)
					}

					host.Dir = newDir
				}
			}

			for _, host := range hosts {
				flagsSet := set.NewSet("")
				for _, flag := range host.Metadata.UserFlags {
					if linq.From(userFlagsToRemove).AnyWith(func(item interface{}) bool { return flag == item }) {
						continue
					}
					flagsSet.Add(flag)
				}
				flagsSet.AddRange(userFlagsToAdd)
				host.Metadata.UserFlags = flagsSet.StringSlice()
				sort.Strings(host.Metadata.UserFlags)
				if addReviewedBy {
					host.SetReviewedBy(reviewer)
				}
				host.SaveMetadata()
			}
		}

		format, _ := cmd.Flags().GetString("format")
		protocols, _ := cmd.Flags().GetStringSlice("protocols")
		if len(protocols) > 0 {
			hostURLs := set.NewSet("")
			for _, host := range hosts {
				hostURLs.AddRange(host.URLs())
			}

			validHostURLs := set.NewSet("")
			for _, hostURL := range hostURLs.SortedStringSlice() {
				for _, proto := range protocols {
					if strings.HasPrefix(hostURL, proto) || proto == "all" {
						validHostURLs.Add(hostURL)
					}
				}
			}
			if jsonOut {
				outObj := map[string][]string{"urls": validHostURLs.SortedStringSlice()}
				json, err := json.MarshalIndent(outObj, "", "  ")
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(json))
			} else {
				validHostURLs.WriteSorted(os.Stdout)
			}
		} else if jsonOut {
			json, err := json.MarshalIndent(hosts, "", "  ")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(json))
		} else if pathsOut {
			for _, host := range hosts {
				fmt.Println(host.Dir)
			}
		} else if format != "" {
			t := template.New("format")
			funcMap := defaultFuncMap()

			_, err := t.Funcs(funcMap).Parse(format)
			if err != nil {
				cmd.PrintErrln(err)
				return
			}

			for _, host := range hosts {
				buf := new(bytes.Buffer)
				err = t.Execute(buf, host.Metadata)
				if err != nil {
					cmd.PrintErrln(err)
					return
				}

				line := buf.String()
				if line != "" {
					if strings.HasSuffix(line, "\n") {
						fmt.Print(line)
					} else {
						fmt.Println(line)
					}
				}
			}
		} else {
			var lines []string
			linq.From(hosts).SelectT(func(host *host.Host) string {
				return host.Metadata.Columnize()
			}).ToSlice(&lines)
			fmt.Println(columnize.SimpleFormat(lines))
		}
	},
}

func init() {
	RootCmd.AddCommand(hostsCmd)
	hostsCmd.Flags().StringSliceP("add-flags", "a", []string{}, "flag(s) to add")
	hostsCmd.Flags().StringSliceP("remove-flags", "r", []string{}, "flag(s) to remove")
	hostsCmd.Flags().StringSliceP("protocols", "p", []string{}, "print protocol strings")
	hostsCmd.Flags().StringSlice("add-names", []string{}, "Hostname(s) to add")
	hostsCmd.Flags().StringSlice("remove-names", []string{}, "Hostname(s) to remove")
	hostsCmd.Flags().BoolP("update", "u", false, "Update arsenic flags")
	hostsCmd.Flags().BoolP("json", "j", false, "Return JSON")
	hostsCmd.Flags().Bool("paths", false, "Return only the path to each hosts directory")
	hostsCmd.Flags().StringSliceP("host", "H", []string{}, "host(s) to add/remove/update flags")
	hostsCmd.RegisterFlagCompletionFunc("host", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return host.AllDirNames(), cobra.ShellCompDirectiveDefault
	})
	hostsCmd.Flags().StringP("query", "q", "", "Query to run. Using Go Template style conditionals.")
	hostsCmd.Flags().StringP("format", "f", "", "Go template format to apply to each matched host's metadata")
	hostsCmd.Flags().StringP("reviewed-by", "R", "operator", "Set the reviewer. -R=reviewer or reads from $AS_REVIEWER, and $USER.")
	hostsCmd.Flags().Lookup("reviewed-by").NoOptDefVal = "operator"
}
