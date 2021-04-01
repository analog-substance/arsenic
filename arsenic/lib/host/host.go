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

	// "strings"
	"github.com/defektive/arsenic/arsenic/lib/util"
	"golang.org/x/net/publicsuffix"
)

type Metadata struct {
	Name        string
	Hostnames   []string
	RootDomains []string
	IPAddresses []string
	Flags       []string
	UserFlags   []string
	TCPPorts    []int
	UDPPorts    []int
	ReviewedBy  string
}

type Host struct {
	dir      string
	metadata Metadata
	// hostnames []string
	// ipAddresses []string
}

func InitHost(dir string) Host {
	return Host{dir, defaultMetadata()}
}

func (host Host) SaveMetadata() {
	var metadata Metadata
	changed := false
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
		changed = true
	}

	existing, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	// we have are base metadata loaded
	// lets update it

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
	tcpPorts := host.tcpPorts()
	udpPorts := host.udpPorts()

	reviewStatus := "Reviewed"
	if len(tcpPorts)+len(udpPorts) > 0 {
		flags = append(flags, "OpenPorts")
		if metadata.ReviewedBy == "" {
			reviewStatus = "Unreviewed"
		}
	}
	flags = append(flags, reviewStatus)

	metadata.Hostnames = hostnames
	metadata.RootDomains = getRootDomains(hostnames)
	metadata.IPAddresses = ipAddresses
	metadata.TCPPorts = host.tcpPorts()
	metadata.UDPPorts = host.udpPorts()
	metadata.Flags = flags

	out, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	if string(existing) != string(out) {
		changed = true
	}

	if changed {
		fmt.Printf("Updating %s\n", host.metadataFile())
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

func UpdateFlags() {
	for _, host := range All() {
		host.SaveMetadata()
	}
}

func All() []Host {
	allHosts := []Host{}
	for _, hostDir := range getHostDirs() {
		host := InitHost(hostDir)
		allHosts = append(allHosts, host)
	}
	return allHosts
}

func getHostDirs() []string {
	filePaths := []string{}
	if _, err := os.Stat("hosts"); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir("hosts")
		if err != nil {
			fmt.Println(err)
		}

		for _, file := range files {
			filePaths = append(filePaths, fmt.Sprintf("hosts/%s", file.Name()))
		}
	}

	sort.Strings(filePaths)
	return filePaths
}

func (host Host) metadataFile() string {
	return fmt.Sprintf("%s/%s", host.dir, "00_metadata.md")
}

func (host Host) flags() []string {
	flags := []string{}

	globbed, _ := filepath.Glob(fmt.Sprintf("%s/recon/%s", host.dir, "nmap-punched.*"))
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

func (host Host) tcpPorts() []int {
	re := regexp.MustCompile(`([0-9]+)(/tcp\s+open)`)
	return ports(re, fmt.Sprintf("%s/recon/%s", host.dir, "nmap-punched-tcp.nmap"))
}

func (host Host) udpPorts() []int {
	re := regexp.MustCompile(`([0-9]+)(/udp\s+open)`)
	return ports(re, fmt.Sprintf("%s/recon/%s", host.dir, "nmap-punched-udp.nmap"))
}

func ports(re *regexp.Regexp, nmapFilePath string) []int {
	pre := regexp.MustCompile(`[0-9]+`)
	content, err := ioutil.ReadFile(nmapFilePath)
	if err != nil {
		return []int{}
	}
	portMatches := re.FindAll(content, -1)

	ports := []int{}
	for _, portMatch := range portMatches {
		port, _ := strconv.Atoi(string(pre.Find(portMatch)))
		ports = append(ports, port)
	}
	return ports
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
		ReviewedBy:  "",
	}
}

func getRootDomains(domains []string) []string {
	rootDomainMap := map[string]int{}
	rootDomains := []string{}
	for _, domain := range domains {
		rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(domain)
		if rootDomain != "" {
			rootDomainMap[rootDomain] = 1
		}
	}

	for rootDomain, _ := range rootDomainMap {
		rootDomains = append(rootDomains, rootDomain)
	}
	return rootDomains
}
