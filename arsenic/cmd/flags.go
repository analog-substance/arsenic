package cmd

import (
	"sort"

	"github.com/defektive/arsenic/arsenic/lib/host"
	"github.com/defektive/arsenic/arsenic/lib/util"
	"github.com/spf13/cobra"
)

// flagsCmd represents the flags command
var flagsCmd = &cobra.Command{
	Use:   "flags",
	Short: "Add, Update, and delete flags",
	Long: `Add, Update, and delete flags.

Flags are neat`,
	Run: func(cmd *cobra.Command, args []string) {
		hostsArgs, _ := cmd.Flags().GetStringSlice("host")

		var hosts []host.Host
		if len(hostsArgs) > 0 {
			hosts = host.Get(hostsArgs)
		} else {
			hosts = host.All()
		}

		userFlagsToAdd, _ := cmd.Flags().GetStringSlice("add")
		userFlagsToRemove, _ := cmd.Flags().GetStringSlice("remove")

		for _, host := range hosts {
			flagsSet := util.NewStringSet()
			for _, flag := range host.Metadata.UserFlags {
				if util.Any(userFlagsToRemove, func(item interface{}) bool { return flag == item }) {
					continue
				}
				flagsSet.Add(flag)
			}
			flagsSet.AddRange(userFlagsToAdd)
			host.Metadata.UserFlags = flagsSet.Slice()
			sort.Strings(host.Metadata.UserFlags)

			host.SaveMetadata()
		}
	},
}

func init() {
	rootCmd.AddCommand(flagsCmd)
	flagsCmd.Flags().StringSliceP("add", "a", []string{}, "flag(s) to add")
	flagsCmd.Flags().StringSliceP("remove", "r", []string{}, "flag(s) to remove")
	flagsCmd.Flags().StringSlice("host", []string{}, "host(s) to add/remove/update flags")
}
