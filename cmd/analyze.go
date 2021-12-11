package cmd

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/cobra"
)

const (
	analyzeDir       string = "recon/analyze-hosts"
	cfIpResolvDomain        = "zz-cloudfront-net-cdn"
)

var (
	tickChars = []string{"-", "/", "|", "\\"}
	nextTick  = 0
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze discover data and create",
	Long: `Analyze discover data and create hosts.

This will create a single host for hostnames that resolve to the same IPs`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("plz 2 refactor me, cause i am calling a slow shell script... k thx, bye!")
		// create, _ := cmd.Flags().GetBool("create")

		// mode := "dry-run"
		// if create {
		// 	mode = "create"
		// }
		// scriptArgs := []string{mode}
		// util.ExecScript("as-analyze-hosts", scriptArgs)

		util.Mkdirs(fmt.Sprintf("%s/services", analyzeDir), "hosts")

		stringSet := set.NewSet(reflect.TypeOf(""))
		files, _ := filepath.Glob("recon/domains/*/resolv-domains.txt")
		for _, file := range files {
			err := util.ReadLineByLine(file, func(line string) {
				if matched, _ := regexp.MatchString("address", line); matched {
					stringSet.Add(line)
				}
			})

			if err != nil {
				fmt.Println(err)
				return
			}
		}

		spaceRegex := regexp.MustCompile(`\s`)
		lines := stringSet.Slice().([]string)
		sort.Strings(lines)

		domainsByIp := make(map[string]set.Set)
		ipsByDomain := make(map[string]set.Set)
		ipsByIpResolvDomain := make(map[string]set.Set)
		resolvIpsFile := "recon/ips/resolv-ips.txt"
		for _, line := range lines {
			//tick("Reviewing resolved domains")

			split := spaceRegex.Split(line, -1)
			domain := split[0]
			ip := split[len(split)-1]

			ipResolvDomain := ""
			if util.FileExists(resolvIpsFile) {
				re := regexp.MustCompile(fmt.Sprintf("^%s", regexp.QuoteMeta(ip)))
				matches := util.GrepLines(resolvIpsFile, re, 1)
				if matches != nil && strings.Contains(matches[0], "cloudfront.net") {
					ipResolvDomain = cfIpResolvDomain
				}
			}

			if _, ok := domainsByIp[ip]; !ok {
				domainsByIp[ip] = set.NewSet(reflect.TypeOf(""))
			}
			ipDomains := domainsByIp[ip]
			ipDomains.Add(domain)

			if _, ok := ipsByDomain[domain]; !ok {
				ipsByDomain[domain] = set.NewSet(reflect.TypeOf(""))
			}
			domainIps := ipsByDomain[domain]
			domainIps.Add(ip)

			if ipResolvDomain != "" {
				ipDomains.Add(ipResolvDomain)

				if _, ok := ipsByIpResolvDomain[ipResolvDomain]; !ok {
					ipsByIpResolvDomain[ipResolvDomain] = set.NewSet(reflect.TypeOf(""))
				}
				ipResolvDomainIps := ipsByIpResolvDomain[ipResolvDomain]
				ipResolvDomainIps.Add(ip)
			}
		}

		for domain, ipSet := range ipsByDomain {
			domainFile := fmt.Sprintf("%s/resolv-domain-%s.txt", analyzeDir, domain)
			ips := ipSet.Slice().([]string)
			sort.Strings(ips)

			util.WriteLines(domainFile, ips)
		}

		for ip, domainSet := range domainsByIp {
			ipFile := fmt.Sprintf("%s/resolv-ip-%s.txt", analyzeDir, ip)
			domains := domainSet.Slice().([]string)
			sort.Strings(domains)

			util.WriteLines(ipFile, domains)
		}

		for ipResolvDomain, ipSet := range ipsByIpResolvDomain {
			ipResolvDomainFile := fmt.Sprintf("%s/resolve-domain-%s.txt", analyzeDir, ipResolvDomain)
			ips := ipSet.Slice().([]string)
			sort.Strings(ips)

			util.WriteLines(ipResolvDomainFile, ips)
		}

	},
}

func tick(msg string) {
	fmt.Printf("\r[%s] %s", tickChars[nextTick], msg)
	nextTick = (nextTick + 1) % len(tickChars)
}

func normalizeDomain(domain string) string {
	re := regexp.MustCompile(`\.$`)
	return re.ReplaceAllString(domain, "")
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().BoolP("create", "c", false, "really create hosts")
}
