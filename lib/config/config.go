package config

import (
	"path/filepath"
	"strconv"
	"strings"
)

var current *Config

func Get() *Config {
	if current == nil {
		return &Config{}
	}
	return current
}

func Set(c *Config) {
	current = c
}

type Config struct {
	Scripts          Scripts   `yaml:"scripts"`
	Wordlists        Wordlists `yaml:"wordlists"`
	Blacklist        Blacklist `yaml:"blacklist"`
	ScriptsDirectory string    `yaml:"scripts-directory"`
	Discover         Discover  `yaml:"discover"`
	Analyze          Analyze   `yaml:"analyze"`
	Hosts            Hosts     `yaml:"hosts"`
}

type Wordlists struct {
	Paths []string            `yaml:"paths"`
	Types map[string][]string `yaml:"types"`
}

type Blacklist struct {
	RootDomains []string `yaml:"root-domains"`
	Domains     []string `yaml:"domains"`
	IPs         []string `yaml:"ips"`
}

type Discover struct {
	TopTCP        int    `yaml:"top-tcp-count"`
	TopUDP        int    `yaml:"top-udp-count"`
	TimingProfile int    `yaml:"timing-profile"`
	ResolveConf   string `yaml:"resolvconf"`
}

type Analyze struct {
	RequireOpenPorts bool `yaml:"require-open-ports"`
}

type Hosts struct {
	NmapXMLGlob    string          `yaml:"nmap-xml-glob"`
	IgnoreServices []IgnoreService `yaml:"ignore-services"`
}

// Don't know what this should be called, if anyone has any better names, go for it
type IgnoreService struct {
	Name  string `yaml:"name"`
	Ports string `yaml:"ports"`
	Flag  string `yaml:"flag"` // This is so we can add a flag saying that this was ignored to make people aware that ports were ignored
}

func (s IgnoreService) checkPort(port int) bool {
	for _, p := range strings.Split(s.Ports, ",") {
		if strings.Contains(p, "-") {
			min, err := strconv.Atoi(strings.Split(p, "-")[0])
			if err != nil {
				return false
			}

			max, err := strconv.Atoi(strings.Split(p, "-")[1])
			if err != nil {
				return false
			}

			if port >= min && port <= max {
				return true
			}
		} else {
			if strings.EqualFold(p, "all") {
				return true
			}

			converted, err := strconv.Atoi(p)
			if err != nil {
				return false
			}

			if port == converted {
				return true
			}
		}
	}
	return false
}

func (s IgnoreService) ShouldIgnore(service string, port int) bool {
	if !strings.EqualFold(s.Name, service) {
		return false
	}
	return s.checkPort(port)
}

func Default(home string) Config {
	wordlists := Wordlists{
		Paths: []string{
			"/opt/SecLists",
			"/usr/share/seclists",
		},
		Types: map[string][]string{
			"web-content": {
				"Discovery/Web-Content/AdobeCQ-AEM.txt",
				"Discovery/Web-Content/apache.txt",
				"Discovery/Web-Content/Common-DB-Backups.txt",
				"Discovery/Web-Content/Common-PHP-Filenames.txt",
				"Discovery/Web-Content/common.txt",
				"Discovery/Web-Content/confluence-administration.txt",
				"Discovery/Web-Content/default-web-root-directory-linux.txt",
				"Discovery/Web-Content/default-web-root-directory-windows.txt",
				"Discovery/Web-Content/frontpage.txt",
				"Discovery/Web-Content/graphql.txt",
				"Discovery/Web-Content/jboss.txt",
				"Discovery/Web-Content/Jenkins-Hudson.txt",
				"Discovery/Web-Content/nginx.txt",
				"Discovery/Web-Content/oracle.txt",
				"Discovery/Web-Content/quickhits.txt",
				"Discovery/Web-Content/raft-large-directories.txt",
				"Discovery/Web-Content/raft-medium-words.txt",
				"Discovery/Web-Content/reverse-proxy-inconsistencies.txt",
				"Discovery/Web-Content/RobotsDisallowed-Top1000.txt",
				"Discovery/Web-Content/websphere.txt",
			},
			"sqli": {
				"Fuzzing/Databases/sqli.auth.bypass.txt",
				"Fuzzing/Databases/MSSQL.fuzzdb.txt",
				"Fuzzing/Databases/MSSQL-Enumeration.fuzzdb.txt",
				"Fuzzing/Databases/MySQL.fuzzdb.txt",
				"Fuzzing/Databases/NoSQL.txt",
				"Fuzzing/Databases/db2enumeration.fuzzdb.txt",
				"Fuzzing/Databases/Oracle.fuzzdb.txt",
				"Fuzzing/Databases/MySQL-Read-Local-Files.fuzzdb.txt",
				"Fuzzing/Databases/Postgres-Enumeration.fuzzdb.txt",
				"Fuzzing/Databases/MySQL-SQLi-Login-Bypass.fuzzdb.txt",
				"Fuzzing/SQLi/Generic-BlindSQLi.fuzzdb.txt",
				"Fuzzing/SQLi/Generic-SQLi.txt",
				"Fuzzing/SQLi/quick-SQLi.txt",
			},
			"xss": {
				"Fuzzing/XSS/XSS-Somdev.txt",
				"Fuzzing/XSS/XSS-Bypass-Strings-BruteLogic.txt",
				"Fuzzing/XSS/XSS-Jhaddix.txt",
				"Fuzzing/XSS/xss-without-parentheses-semi-colons-portswigger.txt",
				"Fuzzing/XSS/XSS-RSNAKE.txt",
				"Fuzzing/XSS/XSS-Cheat-Sheet-PortSwigger.txt",
				"Fuzzing/XSS/XSS-BruteLogic.txt",
				"Fuzzing/XSS-Fuzzing",
			},
		},
	}

	blacklist := Blacklist{
		RootDomains: []string{
			"1e100.net",
			"akamaitechnologies.com",
			"amazonaws.com",
			"azure.com",
			"azurewebsites.net",
			"azurewebsites.windows.net",
			"c7dc.com",
			"cas.ms",
			"cloudapp.net",
			"cloudfront.net",
			"googlehosted.com",
			"googleusercontent.com",
			"hscoscdn10.net",
			"my.jobs",
			"readthedocs.io",
			"readthedocs.org",
			"sites.hubspot.net",
			"tds.net",
			"wixsite.com",
		},
		Domains: make([]string, 0),
		IPs:     make([]string, 0),
	}

	phases := map[string]Phase{
		"init": {
			Scripts: map[string]Script{
				"as-init-op":      NewScript("as-init-op", 0, 1, true),
				"as-setup-hugo":   NewScript("as-setup-hugo", 100, 1, true),
				"as-init-hooks":   NewScript("as-init-hooks", 200, 1, true),
				"as-init-cleanup": NewScript("as-init-cleanup", 300, 1, true),
			},
		},
		"discover": {
			Scripts: map[string]Script{
				"as-root-domain-recon":             NewScript("as-root-domain-recon", 0, 1, true),
				"as-subdomain-discovery":           NewScript("as-subdomain-discovery", 50, 1, true),
				"as-subdomain-enumeration":         NewScript("as-subdomain-enumeration", 100, 1, true),
				"as-combine-subdomains":            NewScript("as-combine-subdomains", 250, 2, true),
				"as-domains-from-domain-ssl-certs": NewScript("as-domains-from-domain-ssl-certs", 275, 1, true),
				"as-dns-resolution":                NewScript("as-dns-resolution", 300, 2, true),
				"as-ip-recon":                      NewScript("as-ip-recon", 400, 2, true),
				"as-domains-from-ip-ssl-certs":     NewScript("as-domains-from-ip-ssl-certs", 500, 2, true),
				"as-ip-resolution":                 NewScript("as-ip-resolution", 600, 2, true),
				"as-http-screenshot-domains":       NewScript("as-http-screenshot-domains", 700, 1, true),
			},
		},
		"recon": {
			Scripts: map[string]Script{
				"as-port-scan-tcp":         NewScript("as-port-scan-tcp", 0, 1, true),
				"as-content-discovery":     NewScript("as-content-discovery", 100, 1, true),
				"as-http-screenshot-hosts": NewScript("as-http-screenshot-hosts", 200, 1, true),
				"as-port-scan-udp":         NewScript("as-port-scan-udp", 300, 1, true),
			},
		},
		"hunt": {
			Scripts: map[string]Script{
				"as-takeover-aquatone":   NewScript("as-takeover-aquatone", 0, 1, true),
				"as-searchsploit":        NewScript("as-searchsploit", 100, 1, true),
				"as-nuclei-technologies": NewScript("as-nuclei-technologies", 200, 1, true),
				"as-nuclei-cves":         NewScript("as-nuclei-cves", 300, 1, true),
			},
		},
	}

	scripts := Scripts{
		Directory: filepath.Join(home, ".config", "arsenic"),
		Phases:    phases,
	}

	discover := Discover{
		TopTCP:        30,
		TopUDP:        30,
		TimingProfile: 4,
		ResolveConf:   "",
	}

	analyze := Analyze{
		RequireOpenPorts: true,
	}

	hosts := Hosts{
		NmapXMLGlob: "nmap-*-??p.xml",
		IgnoreServices: []IgnoreService{
			{
				Name:  "msrpc",
				Ports: "40000-65535",
				Flag:  "ignored::ephemeral-msrpc",
			},
			{
				Name:  "tcpwrapped",
				Ports: "all",
			},
			{
				Name:  "unknown",
				Ports: "all",
			},
		},
	}

	return Config{
		Wordlists: wordlists,
		Blacklist: blacklist,
		Scripts:   scripts,
		Discover:  discover,
		Analyze:   analyze,
		Hosts:     hosts,
	}
}
