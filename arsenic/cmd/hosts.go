package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/defektive/arsenic/arsenic/lib/host"
	"github.com/defektive/arsenic/arsenic/lib/set"
	"github.com/defektive/arsenic/arsenic/lib/slice"
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

		var hosts []host.Host
		if len(hostsArgs) > 0 {
			hosts = host.Get(hostsArgs)
		} else {
			hosts = host.All()
		}

		userFlagsToAdd, _ := cmd.Flags().GetStringSlice("add-flags")
		userFlagsToRemove, _ := cmd.Flags().GetStringSlice("remove-flags")
		updateArsenicFlags, _ := cmd.Flags().GetBool("update")

		shouldSave := len(userFlagsToRemove) > 0 || len(userFlagsToAdd) > 0 || updateArsenicFlags

		for _, host := range hosts {
			if  shouldSave {
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

			json, err := json.MarshalIndent(host.Metadata, "", "  ")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(json))
		}
	},
}

func init() {
	rootCmd.AddCommand(hostsCmd)
	hostsCmd.Flags().StringSliceP("add-flags", "a", []string{}, "flag(s) to add")
	hostsCmd.Flags().StringSliceP("remove-flags", "r", []string{}, "flag(s) to remove")
	hostsCmd.Flags().BoolP("update", "u", false, "Update arsenic flags")
	hostsCmd.Flags().StringSliceP("host", "H", []string{}, "host(s) to add/remove/update flags")
	hostsCmd.RegisterFlagCompletionFunc("host", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return host.AllDirNames(), cobra.ShellCompDirectiveDefault
	})
}
