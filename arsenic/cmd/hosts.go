package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/defektive/arsenic/arsenic/lib/host"
	"github.com/defektive/arsenic/arsenic/lib/set"
	"github.com/defektive/arsenic/arsenic/lib/slice"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

// hostsCmd represents the flags command
var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "manage hosts",
	Long: `Add, Update, and delete flags from hosts
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
				return slice.Any(s1, func(s1Item interface{}) bool {
					return slice.Any(s2, func(s2Item interface{}) bool {
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
				panic(err)
			}
			err = hostTemplate.Execute(os.Stdout, host.All())
			if err != nil {
				fmt.Println(err)
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

		for _, host := range hosts {
			if shouldSave {
				flagsSet := set.NewSet(reflect.TypeOf(""))
				for _, flag := range host.Metadata.UserFlags {
					if slice.Any(userFlagsToRemove, func(item interface{}) bool { return flag == item }) {
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
			for _, host := range hosts {
				lines = append(lines, fmt.Sprintf("%s | %s | %s\n", host.Metadata.Name, strings.Join(host.Metadata.Flags, ","), strings.Join(host.Metadata.UserFlags, ",")))
			}
			fmt.Println(columnize.SimpleFormat(lines))
		}
	},
}

func init() {
	rootCmd.AddCommand(hostsCmd)
	hostsCmd.Flags().StringSliceP("add-flags", "a", []string{}, "flag(s) to add")
	hostsCmd.Flags().StringSliceP("remove-flags", "r", []string{}, "flag(s) to remove")
	hostsCmd.Flags().BoolP("update", "u", false, "Update arsenic flags")
	hostsCmd.Flags().BoolP("json", "j", false, "Update arsenic flags")
	hostsCmd.Flags().StringSliceP("host", "H", []string{}, "host(s) to add/remove/update flags")
	hostsCmd.RegisterFlagCompletionFunc("host", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return host.AllDirNames(), cobra.ShellCompDirectiveDefault
	})
	hostsCmd.Flags().StringP("query", "q", "", "the host query")
}
