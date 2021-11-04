package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/ahmetb/go-linq/v3"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
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

		addReviewedBy := false
		reviewerFlag, _ := cmd.Flags().GetString("reviewed-by")

		if cmd.Flags().Lookup("reviewed-by").Changed && (len(hostsArgs) > 0 || len(query) > 0) {
			addReviewedBy = true
		}
		reviewer := getReviewer(reviewerFlag)
		userFlagsToAdd, _ := cmd.Flags().GetStringSlice("add-flags")
		userFlagsToRemove, _ := cmd.Flags().GetStringSlice("remove-flags")
		updateArsenicFlags, _ := cmd.Flags().GetBool("update")
		jsonOut, _ := cmd.Flags().GetBool("json")
		pathsOut, _ := cmd.Flags().GetBool("paths")

		shouldSave := len(userFlagsToRemove) > 0 || len(userFlagsToAdd) > 0 || updateArsenicFlags || addReviewedBy

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
				if addReviewedBy {
					host.Metadata.ReviewedBy = reviewer
				}
				host.SaveMetadata()
			}
		}

		protocols, _ := cmd.Flags().GetStringSlice("protocols")
		if len(protocols) > 0 {
			var hostURLs []string
			for _, host := range hosts {
				hostURLs = append(hostURLs, host.URLs()...)
			}

			for _, hostURL := range hostURLs {
				for _, proto := range protocols {
					if strings.HasPrefix(hostURL, proto) || proto == "all" {
						fmt.Println(hostURL)
					}
				}
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
		} else {
			var lines []string
			linq.From(hosts).SelectT(func(host host.Host) string {
				return host.Metadata.Columnize()
			}).ToSlice(&lines)
			fmt.Println(columnize.SimpleFormat(lines))
		}
	},
}

func getReviewer(reviewerFlag string) string {
	if reviewerFlag == "operator" {
		envReviewer := os.Getenv("AS_REVIEWER")
		envUser := os.Getenv("USER")
		if len(envReviewer) > 0 {
			reviewerFlag = envReviewer
		} else if len(envUser) > 0 {
			reviewerFlag = envUser
		}
	}

	return reviewerFlag
}

func init() {
	rootCmd.AddCommand(hostsCmd)
	hostsCmd.Flags().StringSliceP("add-flags", "a", []string{}, "flag(s) to add")
	hostsCmd.Flags().StringSliceP("remove-flags", "r", []string{}, "flag(s) to remove")
	hostsCmd.Flags().StringSliceP("protocols", "p", []string{}, "print protocol strings")
	hostsCmd.Flags().BoolP("update", "u", false, "Update arsenic flags")
	hostsCmd.Flags().BoolP("json", "j", false, "Return JSON")
	hostsCmd.Flags().Bool("paths", false, "Return only the path to each hosts directory")
	hostsCmd.Flags().StringSliceP("host", "H", []string{}, "host(s) to add/remove/update flags")
	hostsCmd.RegisterFlagCompletionFunc("host", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return host.AllDirNames(), cobra.ShellCompDirectiveDefault
	})
	hostsCmd.Flags().StringP("query", "q", "", "Query to run. Using Go Template style conditionals.")
	hostsCmd.Flags().StringP("reviewed-by", "R", "operator", "Set the reviewer. -R=reviewer or reads from $AS_REVIEWER, and $USER.")
	hostsCmd.Flags().Lookup("reviewed-by").NoOptDefVal = "operator"
}
