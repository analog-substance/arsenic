package cmd

import (
	"fmt"
	"github.com/lair-framework/go-nmap"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/analog-substance/arsenic/lib/scope"

	"github.com/ahmetb/go-linq/v3"
	"github.com/analog-substance/arsenic/lib/grep"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/cobra"
)

const (
	analyzeDir       string = "recon/analyze-hosts"
	cfIpResolvDomain string = "zz-cloudfront-net-cdn"
	cfIpResolvIP     string = "13.249.205.47"
	akIpResolvDomain string = "zz-akamaiedge-net-cdn"
	akIpResolvIP     string = "104.87.84.116"
)

var (
	tickChars = []string{"-", "/", "|", "\\"}
	nextTick  = 0

	ipsByDomain         = make(stringSetMap)
	domainsByIp         = make(stringSetMap)
	ipsByIpResolvDomain = make(stringSetMap)
	serviceByDomain     = make(serviceMap)
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
		create, _ := cmd.Flags().GetBool("create")
		nmap, _ := cmd.Flags().GetBool("nmap")
		keepPrivateIPs, _ := cmd.Flags().GetBool("private-ips")

		// mode := "dry-run"
		// if create {
		// 	mode = "create"
		// }
		// scriptArgs := []string{mode}
		// util.ExecScript("as-analyze-hosts", scriptArgs)

		os.RemoveAll(analyzeDir)
		util.Mkdir(filepath.Join(analyzeDir, "services"), "hosts")

		if nmap {
			getDiscoverNmaps()
		} else {

			resolvResults, err := getResolvResults()
			if err != nil {
				fmt.Println(err)
				return
			}

		reviewDomains(resolvResults)
		fmt.Println("\n[+] Domain review complete")

		reviewIps(keepPrivateIPs)
		fmt.Println("\n[+] IP review complete")

		domains := serviceByDomain.keys()
		for _, domain := range domains {
			if !scope.IsInScope(domain, false) {
				continue
			}

			service := serviceByDomain[domain]
			var h *host.Host

			// find host by domain
			hosts := host.Get(domain)
			if len(hosts) == 0 {
				// we have a new domain, lets see if it's IP is in use anywhere...
				if ips, ok := ipsByDomain[domain]; ok {
					hosts = host.GetByIp(ips.StringSlice()...)
				}
			}

			if len(hosts) == 0 {
				// Still no host, lets create a new one
				fmt.Printf("[+] Creating new service %s\n", domain)
				h = host.InitHost(filepath.Join("hosts", domain))
			} else {
				h = hosts[0]
				if h.Metadata.Name != domain {
					fmt.Printf("[+] Adding domains to %s from %s\n", h.Metadata.Name, domain)
				} else {
					fmt.Printf("[+] Updating existing %s\n", domain)
				}
			}

			for _, hostname := range h.Metadata.Hostnames {
				if scope.IsInScope(hostname, false) {
					service.hostnames.Add(hostname)
				}
			}
			h.Metadata.Hostnames = service.hostnames.SortedStringSlice()

			for _, IPAddr := range h.Metadata.IPAddresses {
				if scope.IsInScope(IPAddr, false) {
					service.ipAddresses.Add(IPAddr)
				}
			}
			h.Metadata.IPAddresses = service.ipAddresses.SortedStringSlice()

			if create || nmap {
				h.SaveMetadata()
			}
		}

		fmt.Println("\n[+] Domain processing complete")

		if !nmap {
			fmt.Println("\n[+] IP processing started")

			scopeIps, err := util.ReadLines("scope-ips.txt")
			if err != nil {
				fmt.Println(err)
				return
			}
			linq.From(scopeIps).
				ForEach(func(i interface{}) {
					ip := i.(string)

					if len(host.GetByIp(ip)) > 0 {
						return
					}

					fmt.Printf("[+] Creating new service %s\n", ip)
					if create {
						h := host.InitHost(filepath.Join("hosts", ip))
						h.Metadata.Hostnames = make([]string, 0)
						h.Metadata.RootDomains = make([]string, 0)

						h.SaveMetadata()
					}
				})

			fmt.Println("\n[+] IP processing complete")
		}
	},
}

func tick(msg string) {
	fmt.Printf("\r[%s] %s", tickChars[nextTick], msg)
	nextTick = (nextTick + 1) % len(tickChars)
}

func getResolvResults() ([]string, error) {
	stringSet := set.NewStringSet()
	addressRegex := regexp.MustCompile("address|is an alias for")
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

func getDiscoverNmaps() {
	files, _ := filepath.Glob("recon/nmap-*.xml")
	requireOpenPorts := viper.GetBool("analyze.require-open-ports")
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		nmapRun, err := nmap.Parse(data)
		if err != nil {
			continue
		}

		for _, nmapHost := range nmapRun.Hosts {
			inScope := false
			for _, addr := range nmapHost.Addresses {
				if scope.IsInScope(addr.Addr, false) {
					inScope = true
					break
				}
			}

			if inScope == false {
				for _, hostname := range nmapHost.Hostnames {
					if scope.IsInScope(hostname.Name, false) {
						inScope = true
						break
					}
				}
			}

			if inScope == false {
				continue
			}

			if !requireOpenPorts || len(nmapHost.Ports) > 0 {
				var svc *service

				if len(nmapHost.Hostnames) > 0 {
					svc = serviceByDomain.getOrInit(nmapHost.Hostnames[0].Name)
				} else {
					svc = serviceByDomain.getOrInit(nmapHost.Addresses[0].Addr)
				}

				for _, s := range nmapHost.Hostnames {
					svc.hostnames.Add(s)
				}

				for _, s := range nmapHost.Addresses {
					svc.ipAddresses.Add(s)
				}
			}
		}
	}
}

func reviewDomains(resolvResults []string) {
	spaceRegex := regexp.MustCompile(`\s`)
	domainCDNAliasMap := map[string]string{}
	resolvIpsFile := "recon/ips/resolv-ips.txt"

	// loop through and get aliases first
	for _, result := range resolvResults {
		tick("Reviewing resolved domains")

		split := spaceRegex.Split(result, -1)
		domain := split[0]
		ip := split[len(split)-1]

		if !scope.IsInScope(domain, false) {
			//fmt.Printf("\nIgnoring %s\n", domain)
			continue
		}

		if strings.Contains(result, "is an alias") {
			if strings.Contains(ip, "cloudfront.net") {
				domainCDNAliasMap[domain] = cfIpResolvIP
			}
			if strings.Contains(ip, "akamaiedge.net") {
				domainCDNAliasMap[domain] = akIpResolvIP
			}

		} else if util.FileExists(resolvIpsFile) {
			// I doubt we ever get to here, need to do more testing
			re := regexp.MustCompile(fmt.Sprintf("^%s", regexp.QuoteMeta(ip)))
			matches := grep.Matches(resolvIpsFile, re, 1)
			if matches != nil && strings.Contains(matches[0], "cloudfront.net") {
				domainCDNAliasMap[domain] = cfIpResolvIP
			}
			if matches != nil && strings.Contains(matches[0], "akamaiedge.net") {
				domainCDNAliasMap[domain] = akIpResolvIP
			}
		}
	}

	for _, result := range resolvResults {
		tick("Reviewing resolved domains")

		split := spaceRegex.Split(result, -1)
		domain := split[0]
		ip := split[len(split)-1]

		if !scope.IsInScope(domain, false) {
			//fmt.Printf("\nIgnoring %s\n", domain)
			continue
		}

		if strings.Contains(result, "is an alias") {
			// ignore aliases since the last fragment is not an IP...
			continue
		}

		//ipResolvDomain := ""
		if aliasCDNIP, ok := domainCDNAliasMap[domain]; ok {
			// we know it is an alias lets use our alias domain...
			//ipResolvDomain = aliasDomain
			ip = aliasCDNIP
			//domainsByIp.addToSet(ip, ipResolvDomain)
			//ipsByIpResolvDomain.addToSet(ipResolvDomain, ip)
			//continue
		}

		domainsByIp.addToSet(ip, domain)
		ipsByDomain.addToSet(domain, ip)
		//
		//if ipResolvDomain != "" {
		//} else {
		//
		//}
	}

	createCDNRefs(cfIpResolvDomain)
	createCDNRefs(akIpResolvDomain)

	for domain, ips := range ipsByDomain {
		domainFile := fmt.Sprintf("%s/resolv-domain-%s.txt", analyzeDir, domain)

		util.WriteLines(domainFile, ips.SortedStringSlice())
	}

	for ip, domains := range domainsByIp {
		ipFile := fmt.Sprintf("%s/resolv-ip-%s.txt", analyzeDir, ip)

		util.WriteLines(ipFile, domains.SortedStringSlice())
	}

	for ipResolvDomain, ips := range ipsByIpResolvDomain {
		ipResolvDomainFile := fmt.Sprintf("%s/resolv-domain-%s.txt", analyzeDir, ipResolvDomain)

		util.WriteLines(ipResolvDomainFile, ips.SortedStringSlice())
	}
}

func createCDNRefs(CDNDomain string) {
	CDNIPSet := ipsByIpResolvDomain[CDNDomain]
	if CDNIPSet != nil && CDNIPSet.Length() > 0 {
		CDNDomainSet := set.NewStringSet()
		CDNIPs := CDNIPSet.StringSlice()
		for _, ip := range CDNIPs {
			CDNDomainSet.AddRange(domainsByIp[ip].StringSlice())
		}
		CDNDomains := CDNDomainSet.SortedStringSlice()

		util.WriteLines(filepath.Join(analyzeDir, fmt.Sprintf("%s-domains.txt", CDNDomain)), CDNDomains)

		//firstCfDomain := ""
		for _, CDNDomain := range CDNDomains {
			// if firstCfDomain == "" {
			// 	firstCfDomain = cfDomain
			// }

			ipSet := ipsByDomain[CDNDomain]
			if ipSet == nil {
				continue
			}

			ips := ipSet.StringSlice()
			for _, ip := range ips {
				domainSet := domainsByIp[ip]
				domainSet.AddRange(CDNDomains)
			}
		}
	}
}

func reviewIps(keepPrivateIPs bool) {
	privateIpRegex := regexp.MustCompile(`\b(127\.[0-9]{1,3}\.|10\.[0-9]{1,3}\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[01])\.)[0-9]{1,3}\.[0-9]{1,3}\b`)

	// Start grouping domains based off of their resolved IPs
	ips := domainsByIp.keys()
	for _, ip := range ips {
		// Filter out the private ips
		if !keepPrivateIPs && privateIpRegex.MatchString(ip) {
			continue
		}

		domains := domainsByIp[ip].SortedStringSlice()

		var svc *service
		var serviceIps []string
		for _, domain := range domains {
			tick("Reviewing resolved IPs")

			if svc != nil {
				domainIpSet := ipsByDomain[domain]
				if domainIpSet == nil {
					continue
				}
				domainIps := domainIpSet.SortedStringSlice()

				// Keep track of the domains that have differences in IPs
				if !util.StringSliceEquals(serviceIps, domainIps) {
					svc.diffs.Add(domain)
				} else {
					svc.hostnames.Add(domain)
				}
			} else {
				svc = serviceByDomain.getOrInit(domain)

				svc.ipAddresses.AddRange(ipsByDomain[domain].StringSlice())
				serviceIps = svc.ipAddresses.SortedStringSlice()
				svc.hostnames.Add(domain)
			}
		}
	}

	domains := serviceByDomain.keys()
	for _, domain := range domains {
		serviceByDomain[domain].save(filepath.Join(analyzeDir, "services", domain))
	}
}

func analyzeNmaps() {

}

// func normalizeDomain(domain string) string {
// 	re := regexp.MustCompile(`\.$`)
// 	return re.ReplaceAllString(domain, "")
// }

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().BoolP("create", "c", false, "really create hosts")
	analyzeCmd.Flags().Bool("private-ips", false, "keep private IPs")
	analyzeCmd.Flags().Bool("nmap", false, "import hosts from recon/nmap-*.xml files")
}
