package cmd

import (
	"fmt"
	"os"
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
	cfIpResolvDomain string = "zz-cloudfront-net-cdn"
)

var (
	tickChars                     = []string{"-", "/", "|", "\\"}
	nextTick       int            = 0
	privateIpRegex *regexp.Regexp = regexp.MustCompile(`\b(127\.[0-9]{1,3}\.|10\.[0-9]{1,3}\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[01])\.)[0-9]{1,3}\.[0-9]{1,3}\b`)

	ipsByDomain         map[string][]string = make(map[string][]string)
	domainsByIp         map[string][]string = make(map[string][]string)
	ipsByIpResolvDomain map[string][]string = make(map[string][]string)
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze discover data and create",
	Long: `Analyze discover data and create hosts.

This will create a single host for hostnames that resolve to the same IPs`,
	Run: func(cmd *cobra.Command, args []string) {
		// create, _ := cmd.Flags().GetBool("create")

		// mode := "dry-run"
		// if create {
		// 	mode = "create"
		// }
		// scriptArgs := []string{mode}
		// util.ExecScript("as-analyze-hosts", scriptArgs)

		os.RemoveAll(analyzeDir)
		util.Mkdirs(fmt.Sprintf("%s/services", analyzeDir), "hosts")

		domainLines, err := getDomainLines()
		if err != nil {
			fmt.Println(err)
			return
		}

		reviewDomains(domainLines)
		fmt.Println("\n[+] Domain review complete")

		reviewIps()
		fmt.Println("\n[+] Updating existing hosts")
	},
}

func tick(msg string) {
	fmt.Printf("\r[%s] %s", tickChars[nextTick], msg)
	nextTick = (nextTick + 1) % len(tickChars)
}

func getDomainLines() ([]string, error) {
	stringSet := set.NewSet(reflect.TypeOf(""))
	files, _ := filepath.Glob("recon/domains/*/resolv-domains.txt")
	for _, file := range files {
		err := util.ReadLineByLine(file, func(line string) {
			if matched, _ := regexp.MatchString("address", line); matched {
				stringSet.Add(line)
			}
		})

		if err != nil {
			return nil, err
		}
	}

	return stringSet.SortedStringSlice(), nil
}

func reviewDomains(domainLines []string) {
	spaceRegex := regexp.MustCompile(`\s`)

	domainSetByIp := make(map[string]set.Set)
	ipSetByDomain := make(map[string]set.Set)
	ipSetByIpResolvDomain := make(map[string]set.Set)
	resolvIpsFile := "recon/ips/resolv-ips.txt"
	for _, line := range domainLines {
		tick("Reviewing resolved domains")

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

		if _, ok := domainSetByIp[ip]; !ok {
			domainSetByIp[ip] = set.NewSet(reflect.TypeOf(""))
		}
		ipDomains := domainSetByIp[ip]
		ipDomains.Add(domain)

		if _, ok := ipSetByDomain[domain]; !ok {
			ipSetByDomain[domain] = set.NewSet(reflect.TypeOf(""))
		}
		domainIps := ipSetByDomain[domain]
		domainIps.Add(ip)

		if ipResolvDomain != "" {
			ipDomains.Add(ipResolvDomain)

			if _, ok := ipSetByIpResolvDomain[ipResolvDomain]; !ok {
				ipSetByIpResolvDomain[ipResolvDomain] = set.NewSet(reflect.TypeOf(""))
			}
			ipResolvDomainIps := ipSetByIpResolvDomain[ipResolvDomain]
			ipResolvDomainIps.Add(ip)
		}
	}

	cfIpSet := ipSetByIpResolvDomain[cfIpResolvDomain]
	if cfIpSet.Length() > 0 {
		cfDomainSet := set.NewSet(reflect.TypeOf(""))
		cfIps := cfIpSet.StringSlice()
		for _, ip := range cfIps {
			domainSet := domainSetByIp[ip]
			cfDomainSet.AddRange(domainSet.StringSlice())
		}
		cfDomains := cfDomainSet.SortedStringSlice()

		util.WriteLines(fmt.Sprintf("%s/cloudfront-domains.txt", analyzeDir), cfDomains)

		firstCfDomain := ""
		for _, cfDomain := range cfDomains {
			if firstCfDomain == "" {
				firstCfDomain = cfDomain
			}

			ipSet := ipSetByDomain[cfDomain]
			ips := ipSet.StringSlice()
			for _, ip := range ips {
				domainSet := domainSetByIp[ip]
				domainSet.AddRange(cfDomains)
			}
		}
	}

	for domain, ipSet := range ipSetByDomain {
		domainFile := fmt.Sprintf("%s/resolv-domain-%s.txt", analyzeDir, domain)

		ips := ipSet.SortedStringSlice()
		ipsByDomain[domain] = ips

		util.WriteLines(domainFile, ips)
	}

	for ip, domainSet := range domainSetByIp {
		ipFile := fmt.Sprintf("%s/resolv-ip-%s.txt", analyzeDir, ip)

		domains := domainSet.SortedStringSlice()
		domainsByIp[ip] = domains

		util.WriteLines(ipFile, domains)
	}

	for ipResolvDomain, ipSet := range ipSetByIpResolvDomain {
		ipResolvDomainFile := fmt.Sprintf("%s/resolve-domain-%s.txt", analyzeDir, ipResolvDomain)

		ips := ipSet.SortedStringSlice()
		ipsByIpResolvDomain[ipResolvDomain] = ips

		util.WriteLines(ipResolvDomainFile, ips)
	}
}

func reviewIps() {
	var ips []string
	for ip := range domainsByIp {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	for _, ip := range ips {
		domains := domainsByIp[ip]

		// Filter out the private ips
		if privateIpRegex.MatchString(ip) {
			continue
		}

		firstDomain := ""
		diffSet := set.NewSet(reflect.TypeOf(""))
		for _, domain := range domains {
			tick("Reviewing resolved IPs")

			if firstDomain != "" {
				firstDomainIps := ipsByDomain[firstDomain]
				domainIps := ipsByDomain[domain]
				if !util.StringSliceEquals(firstDomainIps, domainIps) {
					diffSet.Add(domain)
				}
			} else {
				firstDomain = domain
			}
		}
		domainReconDir := fmt.Sprintf("%s/services/%s/recon", analyzeDir, firstDomain)
		util.Mkdirs(domainReconDir)

		diffs := diffSet.SortedStringSlice()
		if len(diffs) > 0 {
			util.WriteLines(fmt.Sprintf("%s/services/%s/domains-with-resolv-differences", analyzeDir, firstDomain), diffs)
		}

		domainSet := set.NewSet(reflect.TypeOf(""))
		domainSet.AddRange(domains)
		domainSet.AddRange(diffs)

		util.WriteLines(fmt.Sprintf("%s/other-hostnames.txt", domainReconDir), domainSet.SortedStringSlice())
		util.WriteLines(fmt.Sprintf("%s/ip-addresses.txt", domainReconDir), ipsByDomain[firstDomain])
	}
}

// func normalizeDomain(domain string) string {
// 	re := regexp.MustCompile(`\.$`)
// 	return re.ReplaceAllString(domain, "")
// }

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().BoolP("create", "c", false, "really create hosts")
}
