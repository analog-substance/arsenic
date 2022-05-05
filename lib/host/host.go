package host

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/analog-substance/arsenic/lib/scope"

	"github.com/lair-framework/go-nmap"

	"github.com/ahmetb/go-linq/v3"
	"github.com/analog-substance/arsenic/lib/util"
)

const (
	reviewedFlag   string = "Reviewed"
	unreviewedFlag string = "Unreviewed"
)

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

func (md Metadata) HasPorts(ports ...int) bool {
	return md.HasTCPPorts(ports...) || md.HasUDPPorts(ports...)
}

func (md Metadata) HasTCPPorts(ports ...int) bool {
	return containsInt(md.TCPPorts, ports...)
}

func (md Metadata) HasUDPPorts(ports ...int) bool {
	return containsInt(md.UDPPorts, ports...)
}

func (md Metadata) HasFlags(flags ...string) bool {
	return md.HasASFlags(flags...) || md.HasUserFlags(flags...)
}

func (md Metadata) HasASFlags(flags ...string) bool {
	return containsStr(md.Flags, flags...)
}

func (md Metadata) HasUserFlags(flags ...string) bool {
	return containsStr(md.UserFlags, flags...)
}

func (md Metadata) Columnize() string {
	return fmt.Sprintf("%s | %s | %s\n", md.Name, strings.Join(md.Flags, ","), strings.Join(md.UserFlags, ","))
}

type Host struct {
	Dir      string
	Metadata *Metadata
}

func InitHost(dir string) *Host {
	host := &Host{Dir: dir}
	var metadata Metadata
	if _, err := os.Stat(host.metadataFile()); !os.IsNotExist(err) {
		jsonFile, err := os.Open(host.metadataFile())
		if err != nil {
			fmt.Println(err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &metadata)
	} else {
		metadata = defaultMetadata()
		metadata.changed = true
	}

	existing, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Println(err)
		return host
	}
	metadata.existing = string(existing)

	hostnames := host.Hostnames()
	ipAddresses := host.IPAddresses()
	if metadata.Name == "unknown" || len(metadata.Name) == 0 {
		if len(hostnames) > 0 {
			metadata.Name = hostnames[0]
		} else if len(ipAddresses) == 1 {
			metadata.Name = ipAddresses[0]
		} else {
			// If the host doesn't have hostnames.txt or ip-addresses.txt, lets use the name of the host directory
			metadata.Name = filepath.Base(dir)
			hostnames = append(hostnames, metadata.Name)
		}
	}

	flags := host.flags()
	ports := host.ports()

	tcpPorts := protoPorts(ports, "tcp")
	udpPorts := protoPorts(ports, "udp")

	reviewStatus := reviewedFlag
	if len(ports) > 0 {
		flags = append(flags, "OpenPorts")
		if metadata.ReviewedBy == "" {
			reviewStatus = unreviewedFlag
		}
	}
	flags = append(flags, reviewStatus)

	metadata.Hostnames = hostnames
	metadata.RootDomains = scope.GetRootDomains(hostnames, true)
	metadata.IPAddresses = ipAddresses
	metadata.TCPPorts = tcpPorts
	metadata.UDPPorts = udpPorts
	metadata.Ports = ports
	metadata.Flags = flags

	host.Metadata = &metadata
	return host
}

func (host Host) SaveMetadata() {
	reconDir := filepath.Join(host.Dir, "recon")
	util.Mkdir(reconDir)

	out, err := json.MarshalIndent(host.Metadata, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	if string(host.Metadata.existing) != string(out) {
		host.Metadata.changed = true
	}

	if host.Metadata.changed {
		err = ioutil.WriteFile(host.metadataFile(), out, 0644)
		if err != nil {
			fmt.Println(err)
		}

		err = util.WriteLines(host.hostnamesFile(), host.Metadata.Hostnames)
		if err != nil {
			fmt.Println(err)
		}

		err = util.WriteLines(host.ipAddressesFile(), host.Metadata.IPAddresses)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (host Host) Hostnames() []string {
	hostnamesFile := fmt.Sprintf("%s/%s", host.Dir, "/recon/hostnames.txt")
	hostnames, err := util.ReadLines(hostnamesFile)

	if err != nil || len(hostnames) == 0 {
		return []string{}
	}

	sort.Strings(hostnames)
	return hostnames
}

func (host Host) URLs() []string {
	URLMap := map[string]bool{}
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

		URLPort := fmt.Sprintf(":%d", port.ID)
		if proto == "http" && port.ID == 80 || proto == "https" && port.ID == 443 {
			URLPort = ""
		}

		URLMap[fmt.Sprintf("%s://%s%s", proto, host.Metadata.Name, URLPort)] = true

		if strings.HasPrefix(proto, "http") {
			// we have an http or https port, we should llo through hostnames
			for _, hostname := range host.Metadata.Hostnames {
				URLMap[fmt.Sprintf("%s://%s%s", proto, hostname, URLPort)] = true
			}
		}
	}

	URLs := []string{}
	for URL := range URLMap {
		URLs = append(URLs, URL)
	}
	return URLs
}

func (host Host) IPAddresses() []string {
	IPAddressesFile := fmt.Sprintf("%s/%s", host.Dir, "/recon/ip-addresses.txt")
	IPAddresses, err := util.ReadLines(IPAddressesFile)

	if err != nil {
		return []string{}
	}
	return IPAddresses
}

func (host Host) isReviewed() bool {
	return host.Metadata.ReviewedBy != "" &&
		linq.From(host.Metadata.Flags).Contains(reviewedFlag)
}

func (host Host) SetReviewedBy(reviewer string) {
	if reviewer == "" {
		return
	}

	host.Metadata.ReviewedBy = reviewer
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
		files, err := filepath.Glob(filepath.Join(host.Dir, glob))
		if err != nil {
			return nil, err
		}

		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

func All() []*Host {
	allHosts := []*Host{}
	for _, hostDir := range getHostDirs() {
		host := InitHost(hostDir)
		allHosts = append(allHosts, host)
	}
	return allHosts
}

func AllDirNames() []string {
	var hosts []string
	for _, hostDir := range getHostDirs() {
		host := InitHost(hostDir)
		hostnames := host.Metadata.Hostnames
		hostnames = append(hostnames, host.Metadata.Name)
		hosts = append(hosts, hostnames...)
	}
	return hosts
}

func Get(hostDirsOrHostnames ...string) []*Host {
	hosts := []*Host{}
	for _, hostDir := range getHostDirs() {
		host := InitHost(hostDir)
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

func GetByIp(ips ...string) []*Host {
	hosts := []*Host{}
	for _, hostDir := range getHostDirs() {
		host := InitHost(hostDir)
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

func getHostDirs() []string {
	filePaths := []string{}
	if _, err := os.Stat("hosts"); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir("hosts")
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			filePaths = append(filePaths, filepath.Join("hosts", file.Name()))
		}
	}

	sort.Strings(filePaths)
	return filePaths
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
	flags := []string{}

	checkGlob := func(glob string) bool {
		globbed, _ := filepath.Glob(filepath.Join(host.Dir, "recon", glob))
		return len(globbed) > 0
	}

	if checkGlob("nmap-punched-tcp.*") {
		flags = append(flags, "nmap-tcp")
	}

	if checkGlob("nmap-punched-udp.*") {
		flags = append(flags, "nmap-udp")
	}

	if checkGlob("gobuster.*") {
		flags = append(flags, "Gobuster")
	}

	if checkGlob("ffuf.*") {
		flags = append(flags, "Ffuf")
	}

	if checkGlob("aquatone-*") {
		flags = append(flags, "Aquatone")
	}

	checkGlob = func(glob string) bool {
		globbed, _ := filepath.Glob(filepath.Join(host.Dir, "loot", glob))
		return len(globbed) > 0
	}

	if checkGlob("hashes*") {
		flags = append(flags, "Loot::hashes")
	}

	return flags
}

func (host Host) ports() []Port {
	portMap := make(map[string]Port)
	globbed, _ := filepath.Glob(fmt.Sprintf("%s/recon/%s", host.Dir, "nmap-*-??p.xml"))

	for _, file := range globbed {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		nmapRun, err := nmap.Parse(data)
		if err != nil {
			continue
		}

		for _, host := range nmapRun.Hosts {
			for _, port := range host.Ports {

				if port.State.State != "closed" && port.State.State != "filtered" {
					service := port.Service.Name

					if strings.HasPrefix(service, "http") && port.Service.Tunnel == "ssl" || port.PortId == 443 {
						service = "https"
					}

					if port.PortId == 80 {
						service = "http"
					}
					portMap[fmt.Sprintf("%s/%d", port.Protocol, port.PortId)] = Port{port.PortId, port.Protocol, service}
				}
			}
		}
	}

	re := regexp.MustCompile(`(?m)^([0-9]+)\s*(.*)$`)
	globbed, _ = filepath.Glob(fmt.Sprintf("%s/recon/%s", host.Dir, "??p-ports.txt"))
	for _, file := range globbed {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		protocol := "tcp"
		if file == "udp-ports.txt" {
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
