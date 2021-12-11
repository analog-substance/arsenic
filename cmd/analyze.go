package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/analog-substance/arsenic/lib/grep"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/cobra"
)

const (
	analyzeDir       string = "recon/analyze-hosts"
	cfIpResolvDomain string = "zz-cloudfront-net-cdn"
)

var (
	tickChars []string = []string{"-", "/", "|", "\\"}
	nextTick  int      = 0

	ipsByDomain         stringSetMap = make(stringSetMap)
	domainsByIp         stringSetMap = make(stringSetMap)
	ipsByIpResolvDomain stringSetMap = make(stringSetMap)
	serviceByDomain     serviceMap   = make(serviceMap)
)

type stringSetMap map[string]*set.Set

func (ssm stringSetMap) getOrInit(key string) *set.Set {
	strSet, found := ssm[key]
	if !found {
		strSet = set.NewStringSet()
		ssm[key] = strSet
	}
	return strSet
}
func (ssm stringSetMap) addToSet(key string, value string) {
	ssm.getOrInit(key).Add(value)
}
func (ssm stringSetMap) keys() []string {
	var keys []string
	for key := range ssm {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type service struct {
	hostnames   *set.Set
	ipAddresses *set.Set
	diffs       *set.Set
}

func newService() *service {
	return &service{
		hostnames:   set.NewStringSet(),
		ipAddresses: set.NewStringSet(),
		diffs:       set.NewStringSet(),
	}
}

func (svc *service) save(baseDir string) {
	reconDir := filepath.Join(baseDir, "recon")
	util.Mkdir(reconDir)

	if svc.diffs.Length() > 0 {
		util.WriteLines(filepath.Join(baseDir, "domains-with-resolv-differences"), svc.diffs.SortedStringSlice())
	}

	util.WriteLines(filepath.Join(reconDir, "other-hostnames.txt"), svc.hostnames.SortedStringSlice())
	util.WriteLines(filepath.Join(reconDir, "ip-addresses.txt"), svc.ipAddresses.SortedStringSlice())
}

type serviceMap map[string]*service

func (sm serviceMap) getOrInit(key string) *service {
	service, found := sm[key]
	if !found {
		service = newService()
		sm[key] = service
	}
	return service
}
func (sm serviceMap) keys() []string {
	var keys []string
	for key := range sm {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

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
		util.Mkdir(filepath.Join(analyzeDir, "services"), "hosts")

		resolvResults, err := getResolvResults()
		if err != nil {
			fmt.Println(err)
			return
		}

		reviewDomains(resolvResults)
		fmt.Println("\n[+] Domain review complete")

		reviewIps()
		fmt.Println("\n[+] Updating existing hosts")
	},
}

func tick(msg string) {
	fmt.Printf("\r[%s] %s", tickChars[nextTick], msg)
	nextTick = (nextTick + 1) % len(tickChars)
}

func getResolvResults() ([]string, error) {
	stringSet := set.NewStringSet()
	addressRegex := regexp.MustCompile("address")
	files, _ := filepath.Glob("recon/domains/*/resolv-domains.txt")
	for _, file := range files {
		err := grep.LineByLine(file, addressRegex, func(line string) {
			stringSet.Add(line)
		})

		if err != nil {
			return nil, err
		}
	}

	return stringSet.SortedStringSlice(), nil
}

func reviewDomains(resolvResults []string) {
	spaceRegex := regexp.MustCompile(`\s`)

	resolvIpsFile := "recon/ips/resolv-ips.txt"
	for _, result := range resolvResults {
		tick("Reviewing resolved domains")

		split := spaceRegex.Split(result, -1)
		domain := split[0]
		ip := split[len(split)-1]

		ipResolvDomain := ""
		if util.FileExists(resolvIpsFile) {
			re := regexp.MustCompile(fmt.Sprintf("^%s", regexp.QuoteMeta(ip)))
			matches := grep.Matches(resolvIpsFile, re, 1)
			if matches != nil && strings.Contains(matches[0], "cloudfront.net") {
				ipResolvDomain = cfIpResolvDomain
			}
		}

		domainsByIp.addToSet(ip, domain)
		ipsByDomain.addToSet(domain, ip)

		if ipResolvDomain != "" {
			domainsByIp.addToSet(ip, ipResolvDomain)
			ipsByIpResolvDomain.addToSet(ipResolvDomain, ip)
		}
	}

	cfIpSet := ipsByIpResolvDomain[cfIpResolvDomain]
	if cfIpSet.Length() > 0 {
		cfDomainSet := set.NewStringSet()
		cfIps := cfIpSet.StringSlice()
		for _, ip := range cfIps {
			cfDomainSet.AddRange(domainsByIp[ip].StringSlice())
		}
		cfDomains := cfDomainSet.SortedStringSlice()

		util.WriteLines(filepath.Join(analyzeDir, "cloudfront-domains.txt"), cfDomains)

		firstCfDomain := ""
		for _, cfDomain := range cfDomains {
			if firstCfDomain == "" {
				firstCfDomain = cfDomain
			}

			ipSet := ipsByDomain[cfDomain]
			ips := ipSet.StringSlice()
			for _, ip := range ips {
				domainSet := domainsByIp[ip]
				domainSet.AddRange(cfDomains)
			}
		}
	}

	for domain, ips := range ipsByDomain {
		domainFile := fmt.Sprintf("%s/resolv-domain-%s.txt", analyzeDir, domain)

		util.WriteLines(domainFile, ips.SortedStringSlice())
	}

	for ip, domains := range domainsByIp {
		ipFile := fmt.Sprintf("%s/resolv-ip-%s.txt", analyzeDir, ip)

		util.WriteLines(ipFile, domains.SortedStringSlice())
	}

	for ipResolvDomain, ips := range ipsByIpResolvDomain {
		ipResolvDomainFile := fmt.Sprintf("%s/resolve-domain-%s.txt", analyzeDir, ipResolvDomain)

		util.WriteLines(ipResolvDomainFile, ips.SortedStringSlice())
	}
}

func reviewIps() {
	privateIpRegex := regexp.MustCompile(`\b(127\.[0-9]{1,3}\.|10\.[0-9]{1,3}\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[01])\.)[0-9]{1,3}\.[0-9]{1,3}\b`)

	// Start grouping domains based off of their resolved IPs
	ips := domainsByIp.keys()
	for _, ip := range ips {
		// Filter out the private ips
		if privateIpRegex.MatchString(ip) {
			continue
		}

		domains := domainsByIp[ip].SortedStringSlice()

		var svc *service
		var serviceIps []string
		for _, domain := range domains {
			tick("Reviewing resolved IPs")

			if svc != nil {
				domainIps := ipsByDomain[domain].SortedStringSlice()

				// Keep track of the domains that have differences in IPs
				if !util.StringSliceEquals(serviceIps, domainIps) {
					svc.diffs.Add(domain)
				}
			} else {
				svc = serviceByDomain.getOrInit(domain)

				svc.ipAddresses.AddRange(ipsByDomain[domain].StringSlice())
				serviceIps = svc.ipAddresses.SortedStringSlice()
			}

			svc.hostnames.Add(domain)
		}
	}

	domains := serviceByDomain.keys()
	for _, domain := range domains {
		serviceByDomain[domain].save(filepath.Join(analyzeDir, "services", domain))
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
