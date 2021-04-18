package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"text/template"

	"github.com/ahmetb/go-linq/v3"
	"github.com/defektive/arsenic/lib/host"
	"github.com/defektive/arsenic/lib/set"
	"github.com/defektive/arsenic/lib/util"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

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

Currently Metadata has the following methods:

- HasPorts
- HasTCPPorts
- HasUDPPorts
- HasFlags
- HasASFlags
- HasUserFlags

`,
	Run: func(cmd *cobra.Command, args []string) {
		hostsArgs, _ := cmd.Flags().GetStringSlice("host")

		query, _ := cmd.Flags().GetString("query")

		var hosts []host.Host
		if len(hostsArgs) > 0 {
			hosts = host.Get(hostsArgs)
		} else if query != "" {
			hostTemplate := template.New("host")
			funcMap := make(template.FuncMap)
			funcMap["in"] = func(s1 []string, s2 ...string) bool {
				// Loop through s1 then s2 and check whether any values of s2 are equal to any values in s1
				return linq.From(s1).AnyWith(func(s1Item interface{}) bool {
					return linq.From(s2).AnyWith(func(s2Item interface{}) bool {
						return s1Item == s2Item
					})
				})
			}

			funcMap["appendMatch"] = func(match host.Host) string {
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

		userFlagsToAdd, _ := cmd.Flags().GetStringSlice("add-flags")
		userFlagsToRemove, _ := cmd.Flags().GetStringSlice("remove-flags")
		updateArsenicFlags, _ := cmd.Flags().GetBool("update")
		jsonOut, _ := cmd.Flags().GetBool("json")

		shouldSave := len(userFlagsToRemove) > 0 || len(userFlagsToAdd) > 0 || updateArsenicFlags

		if shouldSave {
			for _, host := range hosts {
				flagsSet := set.NewSet(reflect.TypeOf(""))
				for _, flag := range host.Metadata.UserFlags {
					if linq.From(userFlagsToRemove).AnyWith(func(item interface{}) bool { return flag == item }) {
						continue
					}
					flagsSet.Add(flag)
				}
				flagsSet.AddRange(userFlagsToAdd)
				host.Metadata.UserFlags = flagsSet.Slice().([]string)
				sort.Strings(host.Metadata.UserFlags)
				host.SaveMetadata()
			}
		}

		if jsonOut {
			json, err := json.MarshalIndent(hosts, "", "  ")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(json))
		} else {
			var lines []string
			linq.From(hosts).SelectT(func(host host.Host) string {
				return host.Metadata.Columnize()
			}).ToSlice(&lines)
			fmt.Println(columnize.SimpleFormat(lines))
		}
	},
}

func init() {
	rootCmd.AddCommand(hostsCmd)
	hostsCmd.Flags().StringSliceP("add-flags", "a", []string{}, "flag(s) to add")
	hostsCmd.Flags().StringSliceP("remove-flags", "r", []string{}, "flag(s) to remove")
	hostsCmd.Flags().BoolP("update", "u", false, "Update arsenic flags")
	hostsCmd.Flags().BoolP("json", "j", false, "Return JSON")
	hostsCmd.Flags().StringSliceP("host", "H", []string{}, "host(s) to add/remove/update flags")
	hostsCmd.RegisterFlagCompletionFunc("host", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return host.AllDirNames(), cobra.ShellCompDirectiveDefault
	})
	hostsCmd.Flags().StringP("query", "q", "", "Query to run. Using Go Template style conditionals.")
}
