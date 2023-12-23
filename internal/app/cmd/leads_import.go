package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/analog-substance/arsenic/lib/lead"
	"github.com/analog-substance/fileutil"
	nessus "github.com/reapertechlabs/go_nessus"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Leads",
	Long: `Import leads from other sources
`,
	Run: func(cmd *cobra.Command, args []string) {

		filesToImport, _ := cmd.Flags().GetStringSlice("file")
		nessusMode, _ := cmd.Flags().GetBool("nessus")

		if nessusMode {

			findings := map[string]*lead.NessusFinding{}

			for _, file := range filesToImport {
				fileutil.FileExists(file)

				fileContents, err := os.ReadFile(file)

				if err != nil {
					log.Fatalln(err)
				}

				nessusData, err := nessus.Parse(fileContents)

				if err != nil {
					log.Fatalln(err)
				}

				for _, host := range nessusData.Report.ReportHosts {
					for _, item := range host.ReportItems {
						finding, ok := findings[item.PluginID]

						a := lead.AffectedAsset{
							Name:         host.Name,
							Port:         item.Port,
							SvcName:      item.SvcName,
							Protocol:     item.Protocol,
							PluginOutput: item.PluginOutput,

							AffectedHost: host,
						}

						if !ok {
							findings[item.PluginID] = &lead.NessusFinding{ReportItem: item, AffectedAssets: []lead.AffectedAsset{a}}
						} else {
							finding.AffectedAssets = append(finding.AffectedAssets, a)
						}
					}
				}

			}

			fmt.Printf("Found %d findings\n", len(findings))

			summary := map[string]int{}

			for _, finding := range findings {

				_, ok := summary[finding.ReportItem.RiskFactor]
				if !ok {
					summary[finding.ReportItem.RiskFactor] = 1
				} else {
					summary[finding.ReportItem.RiskFactor] = 1 + summary[finding.ReportItem.RiskFactor]
				}
				fmt.Printf("[%s] %s %s\n", finding.ReportItem.RiskFactor, finding.ReportItem.PluginName, finding.ReportItem.PluginID)
				fmt.Printf("Affected hosts: %d\n", len(finding.AffectedAssets))

				lead := lead.FromNessusFinding(finding)
				lead.Save()
			}

			for s, v := range summary {
				fmt.Println(s, v)
			}

			//b, err := json.MarshalIndent(findings, "", "  ")
			//if err != nil {
			//	log.Fatal(err)
			//}
			//
			//fmt.Println(string(b))
		}
	},
}

func init() {
	leadsCmd.AddCommand(importCmd)
	importCmd.Flags().BoolP("nessus", "n", false, "Nessus import mode")
	importCmd.Flags().StringSliceP("file", "f", []string{}, "files(s) to import")

}
