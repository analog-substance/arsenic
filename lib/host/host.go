package host

import (
	"encoding/json"
	"fmt"
	"github.com/lair-framework/go-nmap"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// "strings"
	"github.com/ahmetb/go-linq/v3"
	"github.com/defektive/arsenic/lib/util"
)

type Port struct {
	ID int
	Protocol string
	Service string
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
	dir      string
	Metadata *Metadata
}

func InitHost(dir string) Host {
	host := Host{dir: dir}

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
		}
	}

	flags := host.flags()
	ports := host.ports()

	tcpPorts := protoPorts(ports, "tcp")
	udpPorts := protoPorts(ports, "udp")

	reviewStatus := "Reviewed"
	if len(ports) > 0 {
		flags = append(flags, "OpenPorts")
		if metadata.ReviewedBy == "" {
			reviewStatus = "Unreviewed"
		}
	}
	flags = append(flags, reviewStatus)

	metadata.Hostnames = hostnames
	metadata.RootDomains = util.GetRootDomains(hostnames)
	metadata.IPAddresses = ipAddresses
	metadata.TCPPorts = tcpPorts
	metadata.UDPPorts = udpPorts
	metadata.Ports = ports
	metadata.Flags = flags

	host.Metadata = &metadata
	return host
}

func (host Host) SaveMetadata() {
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
	}
}

func (host Host) Hostnames() []string {
	hostnamesFile := fmt.Sprintf("%s/%s", host.dir, "/recon/hostnames.txt")
	hostnames, err := util.ReadLines(hostnamesFile)

	if err != nil {
		return []string{}
	}
	sort.Strings(hostnames)
	return hostnames
}

func (host Host) IPAddresses() []string {
	IPAddressesFile := fmt.Sprintf("%s/%s", host.dir, "/recon/ip-addresses.txt")
	IPAddresses, err := util.ReadLines(IPAddressesFile)

	if err != nil {
		return []string{}
	}
	return IPAddresses
}

func All() []Host {
	allHosts := []Host{}
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

func Get(hostDirsOrHostnames []string) []Host {
	hosts := []Host{}
	for _, hostDir := range getHostDirs() {
		host := InitHost(hostDir)
		hostnames := host.Metadata.Hostnames
		hostnames = append(hostnames, host.Metadata.Name)

		if linq.From(hostDirsOrHostnames).AnyWith(func(hostDirOrHostname interface{}) bool {
			return linq.From(hostnames).AnyWith(func(hostname interface{}) bool {
				return hostDirOrHostname == hostname
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
	return filepath.Join(host.dir, "00_metadata.md")
}

func (host Host) flags() []string {
	flags := []string{}

	globbed, _ := filepath.Glob(fmt.Sprintf("%s/recon/%s", host.dir, "nmap-punched-tcp.*"))
	if len(globbed) > 0 {
		flags = append(flags, "nmap-tcp")
	}

	globbed, _ = filepath.Glob(fmt.Sprintf("%s/recon/%s", host.dir, "nmap-punched-udp.*"))
	if len(globbed) > 0 {
		flags = append(flags, "nmap-udp")
	}

	globbed, _ = filepath.Glob(fmt.Sprintf("%s/recon/%s", host.dir, "gobuster.*"))
	if len(globbed) > 0 {
		flags = append(flags, "Gobuster")
	}

	globbed, _ = filepath.Glob(fmt.Sprintf("%s/recon/%s", host.dir, "aquatone-*"))
	if len(globbed) > 0 {
		flags = append(flags, "Aquatone")
	}

	return flags
}

func (host Host) ports() []Port {
	portMap := make(map[string]Port)
	globbed, _ := filepath.Glob(fmt.Sprintf("%s/recon/%s", host.dir, "nmap-*.xml"))

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
				if port.Service.Name != "tcpwrapped" {
					portMap[fmt.Sprintf("%s/%d", port.Protocol, port.PortId)] = Port{port.PortId, port.Protocol, port.Service.Name}
				}
			}
		}
	}

	ports := []Port{}
	for _, port := range portMap {
		ports = append(ports, port)
	}

	sort.SliceStable(ports, func(i, j int) bool {
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
