package host

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/lib/scope"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/fileutil"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/viper"

	"github.com/analog-substance/nmap/v3"

	"github.com/ahmetb/go-linq/v3"
	"github.com/analog-substance/arsenic/lib/util"
)

const (
	reviewedFlag   string = "Reviewed"
	unreviewedFlag string = "Unreviewed"
)

// SyncOptions represents the different parts of host metadata to sync
type SyncOptions struct {
	IPAddresses bool
	Hostnames   bool
	Ports       bool
	Flags       bool
}

type Port struct {
	ID       int
	Protocol string
	Service  string
}

type Metadata struct {
	Name        string
	Hostnames   []string
	RootDomains []string
	IPAddresses []string
	Flags       []string
	UserFlags   []string
	TCPPorts    []int
	UDPPorts    []int
	Ports       []Port
	ReviewedBy  string
	existing    string
	changed     bool
}

func containsInt(i1 []int, i2 ...int) bool {
	for _, i1v := range i1 {
		for _, v := range i2 {
			if i1v == v {
				return true
			}
		}
	}
	return false
}

func containsStr(i1 []string, i2 ...string) bool {
	for _, i1v := range i1 {
		for _, v := range i2 {
			if i1v == v {
				return true
			}
		}
	}
	return false
}

func containsAllStr(i1 []string, i2 ...string) bool {
	for _, v := range i2 {
		exists := false
		for _, i1v := range i1 {
			if i1v == v {
				exists = true
				break
			}
		}

		if !exists {
			return false
		}
	}
	return true
}

func (md Metadata) HasPorts(ports ...int) bool {
	return md.HasTCPPorts(ports...) || md.HasUDPPorts(ports...)
}

func (md Metadata) HasAnyPort() bool {
	return len(md.Ports) > 0
}

func (md Metadata) HasTCPPorts(ports ...int) bool {
	return containsInt(md.TCPPorts, ports...)
}

func (md Metadata) HasAnyTCPPort() bool {
	return len(md.TCPPorts) > 0
}

func (md Metadata) HasUDPPorts(ports ...int) bool {
	return containsInt(md.UDPPorts, ports...)
}

func (md Metadata) HasAnyUDPPort() bool {
	return len(md.UDPPorts) > 0
}

func (md Metadata) HasFlags(flags ...string) bool {
	return md.HasASFlags(flags...) || md.HasUserFlags(flags...)
}

func (md Metadata) HasAllFlags(flags ...string) bool {
	allFlags := append(md.Flags, md.UserFlags...)
	return containsAllStr(allFlags, flags...)
}

func (md Metadata) HasASFlags(flags ...string) bool {
	return containsStr(md.Flags, flags...)
}

func (md Metadata) HasAllASFlags(flags ...string) bool {
	return containsAllStr(md.Flags, flags...)
}

func (md Metadata) HasUserFlags(flags ...string) bool {
	return containsStr(md.UserFlags, flags...)
}

func (md Metadata) HasAllUserFlags(flags ...string) bool {
	return containsAllStr(md.UserFlags, flags...)
}

func (md Metadata) HasAnyHostname() bool {
	return len(md.Hostnames) > 0
}

func (md Metadata) InCIDR(cidrStr string) bool {
	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}

	for _, ip := range md.IPAddresses {
		if cidr.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}

func (md Metadata) Columnize() string {
	return fmt.Sprintf("%s | %s | %s\n", md.Name, strings.Join(md.Flags, ","), strings.Join(md.UserFlags, ","))
}

func (md *Metadata) AddFlags(flags ...string) {
	flagSet := set.NewStringSet(md.Flags)
	flagSet.AddRange(flags)
	md.Flags = flagSet.SortedStringSlice()
}

func (md *Metadata) RemoveFlags(flags ...string) {
	var filteredFlags []string
	linq.From(md.Flags).Where(func(i interface{}) bool {
		return !linq.From(flags).AnyWith(func(j interface{}) bool {
			return i == j
		})
	}).ToSlice(&filteredFlags)
	md.Flags = filteredFlags
}

type Host struct {
	Dir      string
	Metadata *Metadata
}

func NewHost(dir string) *Host {
	metadata := defaultMetadata()
	return &Host{
		Dir:      dir,
		Metadata: &metadata,
	}
}

func (h *Host) SyncMetadata(options SyncOptions) error {
	var metadata Metadata
	if fileutil.FileExists(h.metadataFile()) {
		bytes, err := os.ReadFile(h.metadataFile())
		if err != nil {
			return err
		}

		json.Unmarshal(bytes, &metadata)
	} else if h.Metadata != nil {
		metadata = *h.Metadata
	} else {
		metadata = defaultMetadata()
		metadata.changed = true
	}

	existing, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	metadata.existing = string(existing)

	h.Metadata = &metadata

	hostnames := h.Metadata.Hostnames
	if options.Hostnames {
		hostnames = h.Hostnames()
	}

	ipAddresses := h.Metadata.IPAddresses
	if options.IPAddresses {
		ipAddresses = h.IPAddresses()
	}

	if metadata.Name == "unknown" || len(metadata.Name) == 0 {
		if len(hostnames) > 0 {
			metadata.Name = hostnames[0]
		} else if len(ipAddresses) == 1 {
			metadata.Name = ipAddresses[0]
		} else {
			// If the host doesn't have hostnames.txt or ip-addresses.txt, lets use the name of the host directory
			metadata.Name = filepath.Base(h.Dir)

			if util.IsIp(metadata.Name) && options.IPAddresses {
				ipAddresses = append(ipAddresses, metadata.Name)
			} else if options.Hostnames {
				hostnames = append(hostnames, metadata.Name)
			}
		}
	}

	ports := metadata.Ports
	tcpPorts := metadata.TCPPorts
	udpPorts := metadata.UDPPorts
	if options.Ports {
		ports = h.ports()
		tcpPorts = protoPorts(ports, "tcp")
		udpPorts = protoPorts(ports, "udp")
	}

	if options.Flags {
		reviewStatus := reviewedFlag
		metadata := h.Metadata

		metadata.AddFlags(h.flags()...)

		if len(tcpPorts) > 0 {
			metadata.AddFlags("open-tcp")
			if metadata.ReviewedBy == "" {
				reviewStatus = unreviewedFlag
			}
		}

		if len(udpPorts) > 0 {
			metadata.AddFlags("open-udp")
			if metadata.ReviewedBy == "" {
				reviewStatus = unreviewedFlag
			}
		}

		if len(ports) > 0 {
			metadata.AddFlags("OpenPorts")
			if metadata.ReviewedBy == "" {
				reviewStatus = unreviewedFlag
			}
		}

		metadata.RemoveFlags(reviewedFlag, unreviewedFlag)
		metadata.AddFlags(reviewStatus)
	}

	metadata.Hostnames = hostnames
	metadata.RootDomains = scope.GetRootDomains(hostnames, true)
	metadata.IPAddresses = ipAddresses

	metadata.Ports = ports
	metadata.TCPPorts = tcpPorts
	metadata.UDPPorts = udpPorts

	return nil
}

func InitHost(dir string) *Host {
	//profiler.Timer("InitHost")
	host := &Host{Dir: dir}
	err := host.SyncMetadata(SyncOptions{
		IPAddresses: true,
		Hostnames:   true,
		Ports:       true,
		Flags:       true,
	})

	if err != nil {
		fmt.Println(err)
		return host
	}

	return host
}

func AddHost(hostnames []string, ips []string) (*Host, error) {
	name := ""
	if len(hostnames) > 0 {
		name = hostnames[0]
	} else if len(ips) > 0 {
		name = ips[0]
	} else {
		return nil, nil
	}

	dir := filepath.Join("hosts", name)
	host := NewHost(dir)

	host.Metadata.Hostnames = hostnames
	host.Metadata.IPAddresses = ips
	host.Metadata.changed = true

	err := host.SyncMetadata(SyncOptions{
		IPAddresses: false,
		Hostnames:   false,
		Ports:       true,
		Flags:       true,
	})
	if err != nil {
		return nil, err
	}
	host.SaveMetadata()

	return host, nil
}

func (host Host) SaveMetadata() {
	reconDir := filepath.Join(host.Dir, "recon")
	fileutil.MkdirAll(reconDir)

	out, err := json.MarshalIndent(host.Metadata, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	if string(host.Metadata.existing) != string(out) {
		host.Metadata.changed = true
	}

	if host.Metadata.changed {
		err = os.WriteFile(host.metadataFile(), out, 0644)
		if err != nil {
			fmt.Println(err)
		}

		err = fileutil.WriteLines(host.hostnamesFile(), host.Metadata.Hostnames)
		if err != nil {
			fmt.Println(err)
		}

		err = fileutil.WriteLines(host.ipAddressesFile(), host.Metadata.IPAddresses)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (host Host) Hostnames() []string {
	hostnamesFile := filepath.Join(host.Dir, "recon/hostnames.txt")
	hostnames, err := fileutil.ReadLines(hostnamesFile)

	if err != nil || len(hostnames) == 0 {
		return []string{}
	}

	hostnameSet := set.NewStringSet(hostnames)
	return hostnameSet.SortedStringSlice()
}

func (host Host) URLs() []string {
	urlSet := set.NewStringSet()
	httpProtocolRe := regexp.MustCompile(`^https?`)
	for _, port := range host.Metadata.Ports {
		proto := port.Service

		if port.ID == 443 {
			proto = "https"
		} else if port.ID == 80 {
			proto = "http"
		} else if httpProtocolRe.MatchString(port.Service) {
			proto = httpProtocolRe.FindString(port.Service)
		}

		urlPort := fmt.Sprintf(":%d", port.ID)
		if proto == "http" && port.ID == 80 || proto == "https" && port.ID == 443 {
			urlPort = ""
		}

		urlSet.Add(fmt.Sprintf("%s://%s%s", proto, host.Metadata.Name, urlPort))

		if strings.HasPrefix(proto, "http") {
			// we have an http or https port, we should loop through hostnames
			for _, hostname := range host.Metadata.Hostnames {
				urlSet.Add(fmt.Sprintf("%s://%s%s", proto, hostname, urlPort))
			}
		}
	}

	return urlSet.StringSlice()
}

func (host Host) IPAddresses() []string {
	ipAddressesFile := filepath.Join(host.Dir, "recon/ip-addresses.txt")
	ipAddresses, err := fileutil.ReadLines(ipAddressesFile)

	if err != nil || len(ipAddresses) == 0 {
		return []string{}
	}

	ipSet := set.NewStringSet(ipAddresses)
	return ipSet.SortedStringSlice()
}

func (host Host) isReviewed() bool {
	return host.Metadata.ReviewedBy != "" &&
		linq.From(host.Metadata.Flags).Contains(reviewedFlag)
}

func (host Host) SetReviewedBy(reviewer string) {
	host.Metadata.ReviewedBy = reviewer

	if reviewer == "" {
		linq.From(host.Metadata.Flags).
			Where(func(i interface{}) bool {
				return i != reviewedFlag && i != unreviewedFlag
			}).
			Append(unreviewedFlag).
			ToSlice(&host.Metadata.Flags)
		return
	}

	if !host.isReviewed() {
		linq.From(host.Metadata.Flags).
			Where(func(i interface{}) bool {
				return i != unreviewedFlag
			}).
			Append(reviewedFlag).
			ToSlice(&host.Metadata.Flags)
	}
}

func (host Host) Files(globs ...string) ([]string, error) {
	var allFiles []string
	for _, glob := range globs {
		files, err := doublestar.FilepathGlob(filepath.Join(host.Dir, glob))
		if err != nil {
			return nil, err
		}

		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

func All() []*Host {
	allHosts := []*Host{}
	for _, host := range getHosts() {
		allHosts = append(allHosts, host)
	}
	return allHosts
}

func AllDirNames() []string {
	var hosts []string
	for _, host := range getHosts() {
		hostnames := host.Metadata.Hostnames
		hostnames = append(hostnames, host.Metadata.Name)
		hosts = append(hosts, hostnames...)
	}
	return hosts
}

func Get(hostDirsOrHostnames ...string) []*Host {
	hosts := []*Host{}
	for _, host := range getHosts() {
		hostnames := host.Metadata.Hostnames
		hostnames = append(hostnames, host.Metadata.Name)
		hostnames = append(hostnames, host.Metadata.IPAddresses...)

		if linq.From(hostDirsOrHostnames).AnyWith(func(hostDirOrHostname interface{}) bool {
			return linq.From(hostnames).AnyWith(func(hostname interface{}) bool {
				return strings.EqualFold(hostDirOrHostname.(string), hostname.(string))
			})
		}) {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

func GetFirst(hostDirsOrHostnames ...string) *Host {
	var foundHost *Host
	for _, host := range getHosts() {
		hostnames := host.Metadata.Hostnames
		hostnames = append(hostnames, host.Metadata.Name)
		hostnames = append(hostnames, host.Metadata.IPAddresses...)

		if linq.From(hostDirsOrHostnames).AnyWith(func(hostDirOrHostname interface{}) bool {
			return linq.From(hostnames).AnyWith(func(hostname interface{}) bool {
				return strings.EqualFold(hostDirOrHostname.(string), hostname.(string))
			})
		}) {
			foundHost = host
			break
		}
	}
	return foundHost
}

func GetByIp(ips ...string) []*Host {
	hosts := []*Host{}
	for _, host := range getHosts() {
		hostIps := host.Metadata.IPAddresses

		if linq.From(ips).AnyWith(func(ip interface{}) bool {
			return linq.From(hostIps).AnyWith(func(hostIp interface{}) bool {
				return ip == hostIp
			})
		}) {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

var hostDirs []string

func getHostDirs() []string {
	//defer profiler.Timer("getHostDirs")()
	if hostDirs != nil {
		return hostDirs
	}

	hostDirs = []string{}
	hostMap = make(map[string]*Host)
	if _, err := os.Stat("hosts"); !os.IsNotExist(err) {
		files, err := os.ReadDir("hosts")
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			hostDir := filepath.Join("hosts", file.Name())
			hostDirs = append(hostDirs, hostDir)
		}
	}

	sort.Strings(hostDirs)
	return hostDirs
}

var hostMap map[string]*Host

func intiHostWorker(hosts chan string, res chan *Host) {
	for hostDir := range hosts {
		res <- InitHost(hostDir)
	}
}

var hostMapMutex sync.Mutex

func getHosts() map[string]*Host {
	//defer profiler.Timer("getHosts")()
	if hostMap != nil {
		return hostMap
	}

	workers := make(chan string, 20)
	c := make(chan *Host)
	for i := 0; i < cap(workers); i++ {
		go intiHostWorker(workers, c)
	}

	hostDirs := getHostDirs()
	go func() {
		for _, host := range hostDirs {
			workers <- host
		}
	}()

	for i := 0; i < len(hostDirs); i++ {
		host := <-c

		hostMapMutex.Lock()
		hostMap[host.Dir] = host
		hostMapMutex.Unlock()
	}

	close(workers)
	close(c)

	return hostMap
}

func (host Host) metadataFile() string {
	return filepath.Join(host.Dir, "00_metadata.md")
}
func (host Host) hostnamesFile() string {
	return filepath.Join(host.Dir, "recon", "hostnames.txt")
}
func (host Host) ipAddressesFile() string {
	return filepath.Join(host.Dir, "recon", "ip-addresses.txt")
}

func (host Host) flags() []string {
	// Make the auto flags configurable
	flags := []string{}

	var globbed []string
	checkGlob := func(glob string) bool {
		globbed, _ = filepath.Glob(filepath.Join(host.Dir, "recon", glob))
		return len(globbed) > 0
	}

	if checkGlob("nmap-punched-tcp.*") {
		flags = append(flags, "nmap-tcp-svc")
	}

	if checkGlob("nmap-*-tcp.*") {
		flags = append(flags, "nmap-tcp")
	}

	if checkGlob("nmap-punched-udp.*") {
		flags = append(flags, "nmap-udp")
	}

	hasResults := false
	hasGobuster := checkGlob("gobuster.*")
	if hasGobuster {
		flags = append(flags, "web-content::gobuster")

		results, _ := gocdp.SmartParseFiles(globbed)
		if len(results) > 0 {
			hasResults = true
		}
	}

	hasFfuf := checkGlob("ffuf.*")
	if hasFfuf {
		flags = append(flags, "web-content::ffuf")

		results, _ := gocdp.SmartParseFiles(globbed)
		if !hasResults && len(results) > 0 {
			hasResults = true
		}
	}

	hasDirb := checkGlob("dirb.*")
	if hasDirb {
		flags = append(flags, "web-content::dirb")

		results, _ := gocdp.SmartParseFiles(globbed)
		if !hasResults && len(results) > 0 {
			hasResults = true
		}
	}

	if hasGobuster || hasFfuf || hasDirb {
		flags = append(flags, "web-content")
	}

	if hasResults {
		flags = append(flags, "web-content::results")
	}

	if checkGlob("aquatone-*") {
		flags = append(flags, "Aquatone")
	}

	checkGlob = func(glob string) bool {
		globbed, _ = filepath.Glob(filepath.Join(host.Dir, "loot", glob))
		return len(globbed) > 0
	}

	if checkGlob("hashes*") {
		flags = append(flags, "Loot::hashes")
	}

	return flags
}

func (host Host) ports() []Port {
	var ignoreServices []util.IgnoreService
	viper.UnmarshalKey("ignore-services", &ignoreServices)

	portMap := make(map[string]Port)
	globbed, _ := filepath.Glob(fmt.Sprintf("%s/recon/%s", host.Dir, viper.GetString("hosts.nmap-xml-glob")))
	quick := []string{}

	tcpPorts := false
	for _, file := range globbed {
		if strings.Contains(file, "quick") {
			quick = append(quick, file)
			continue
		}

		nmapRun := &nmap.Run{}
		err := nmapRun.FromFile(file)
		if err != nil {
			continue
		}

		for _, nmapHost := range nmapRun.Hosts {
			for _, port := range nmapHost.Ports {

				if port.State.State != "closed" && port.State.State != "filtered" {
					ignore := false
					for _, svc := range ignoreServices {
						if svc.ShouldIgnore(port.Service.Name, int(port.ID)) {
							ignore = true
							if svc.Flag != "" {
								host.Metadata.AddFlags(svc.Flag)
							}
							break
						}
					}

					if ignore {
						continue
					}

					if port.Service.Product != "" {
						host.Metadata.AddFlags(fmt.Sprintf("SVC::%s", port.Service.Product))
					}

					service := port.Service.Name

					if strings.HasPrefix(service, "http") && port.Service.Tunnel == "ssl" || port.ID == 443 {
						service = "https"
					}

					if port.ID == 80 {
						service = "http"
					}
					portMap[fmt.Sprintf("%s/%d", port.Protocol, port.ID)] = Port{int(port.ID), port.Protocol, service}
					if !tcpPorts && port.Protocol == "tcp" {
						tcpPorts = true
					}
				}
			}
		}
	}

	if !tcpPorts && len(quick) > 0 && !host.Metadata.HasFlags("ignore::nmap-quick") {
		for _, file := range quick {
			nmapRun := &nmap.Run{}
			err := nmapRun.FromFile(file)
			if err != nil {
				continue
			}

			for _, nmapHost := range nmapRun.Hosts {
				for _, port := range nmapHost.Ports {

					if port.State.State != "closed" && port.State.State != "filtered" {
						ignore := false
						for _, svc := range ignoreServices {
							if svc.ShouldIgnore(port.Service.Name, int(port.ID)) {
								ignore = true
								if svc.Flag != "" {
									host.Metadata.AddFlags(svc.Flag)
								}
								break
							}
						}

						if ignore {
							continue
						}

						service := port.Service.Name

						if strings.HasPrefix(service, "http") && port.Service.Tunnel == "ssl" || port.ID == 443 {
							service = "https"
						}

						if port.ID == 80 {
							service = "http"
						}
						portMap[fmt.Sprintf("%s/%d", port.Protocol, port.ID)] = Port{int(port.ID), port.Protocol, service}
						if !tcpPorts && port.Protocol == "tcp" {
							tcpPorts = true
						}
					}
				}
			}
		}
	}

	re := regexp.MustCompile(`(?m)^([0-9]+)\s*?(.*)$`)
	globbed, _ = filepath.Glob(fmt.Sprintf("%s/recon/%s", host.Dir, "*??p-ports.txt"))
	for _, file := range globbed {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		protocol := "tcp"
		if strings.HasSuffix(file, "udp-ports.txt") {
			protocol = "udp"
		}

		dataString := string(data)
		linesMatches := re.FindAllStringSubmatch(dataString, -1)
		for _, matches := range linesMatches {
			portId, _ := strconv.Atoi(matches[1])
			key := fmt.Sprintf("%s/%d", protocol, portId)
			if _, ok := portMap[key]; !ok {
				portMap[key] = Port{
					ID:       portId,
					Protocol: protocol,
					Service:  matches[2],
				}
			}
		}
	}

	ports := []Port{}
	for _, port := range portMap {
		ports = append(ports, port)
	}

	sort.SliceStable(ports, func(i, j int) bool {
		if ports[i].ID == ports[j].ID {
			return ports[i].Protocol == "tcp"
		}

		return ports[i].ID < ports[j].ID
	})

	return ports
}

func protoPorts(ports []Port, proto string) []int {
	retPorts := []int{}
	for _, port := range ports {
		if port.Protocol == proto {
			retPorts = append(retPorts, port.ID)
		}
	}
	return retPorts
}

func defaultMetadata() Metadata {
	return Metadata{
		Name:        "unknown",
		Hostnames:   []string{},
		IPAddresses: []string{},
		Flags:       []string{},
		UserFlags:   []string{},
		TCPPorts:    []int{},
		UDPPorts:    []int{},
		Ports:       []Port{},
		ReviewedBy:  "",
	}
}
