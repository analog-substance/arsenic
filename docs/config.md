# Config
The configuration file holds many different settings to fine tune commands and scripts run by Arsenic.

## Blacklist

The blacklist section contains the domains and IPs to filter out of the scope. This really only matters if subdomain discovery is in scope. If it is out of scope, meaning you must stick to what is in the `scope-domains.txt` and `scope-ips.txt` files, this can be skipped.

```yaml
blacklist:
  domains: []
  ips: []
  root-domains:
  - 1e100.net
  - akamaitechnologies.com
  - amazonaws.com
  - azure.com
  - azurewebsites.net
  - azurewebsites.windows.net
  - c7dc.com
  - cas.ms
  - cloudapp.net
  - cloudfront.net
  - googlehosted.com
  - googleusercontent.com
  - hscoscdn10.net
  - my.jobs
  - readthedocs.io
  - readthedocs.org
  - sites.hubspot.net
  - tds.net
  - wixsite.com
```

- **`domains`**: An array of regex strings to filter from the discovered domains. Refer to https://pkg.go.dev/regexp/syntax for the supported syntax.
- **`ips`**: An array of IP addresses to filter from the discovered IPs.
- **`root-domains`**: An array of root domains to filter from the discovered domains.

**Note:** The domains and IPs in `scope-domains.txt` and `scope-ips.txt` will never be filtered. This is to ensure client provided information is always included in scope.

## Ignore Services

The `ignore-services` section is a mapping of ports to nmap services to ignore when parsing a host's nmap scans.

```yaml
ignore-services:
- name: msrpc
  ports: 40000-65535
  flag: ignored::ephemeral-msrpc
- name: tcpwrapped
  ports: all
  flag: ""
- name: unknown
  ports: all
  flag: ""
```
- **`name`**: The name of the nmap service
- **`ports`**: The ports/port ranges of the service
- **`flag`**: An optional flag to add to the host if the service and ports match

**Note:** Both the ports and name must match in order for the service to be ignored.

## Scripts

The `scripts` section contains the different scripts to run for different phases.

### Init

These scripts aren't part of a phase of testing. They are run when `arsenic init` is called to create the op.

```yaml
scripts:
	init:
		as-init-cleanup:
			count: 1
			enabled: true
			order: 300
			script: as-init-cleanup
		as-init-hooks:
			count: 1
			enabled: true
			order: 200
			script: as-init-hooks
		as-init-op:
			count: 1
			enabled: true
			order: 0
			script: as-init-op
		as-setup-hugo:
			count: 1
			enabled: true
			order: 100
			script: as-setup-hugo
```
- **as-init-cleanup**: Runs common cleanup tasks.
- **as-init-hooks**: Calls `as-init-op.sh` files for custom op initialization.
- **as-init-op**: Creates the necessary directory/file structure for the op.
- **as-setup-hugo**: Sets up the op for use with `hugo`.

### Discover

The discover phase is for actively and passively discovering new subdomains and IPs. The scripts run during this phase generally only require the domains and IPs contained in the scoping files.

```yaml 
scripts:
	discover:
		as-combine-subdomains:
			count: 2
			enabled: true
			order: 250
			script: as-combine-subdomains
		as-dns-resolution:
			count: 2
			enabled: true
			order: 300
			script: as-dns-resolution
		as-domains-from-domain-ssl-certs:
			count: 1
			enabled: true
			order: 200
			script: as-domains-from-domain-ssl-certs
		as-domains-from-ip-ssl-certs:
			count: 2
			enabled: true
			order: 500
			script: as-domains-from-ip-ssl-certs
		as-http-screenshot-domains:
			count: 1
			enabled: true
			order: 700
			script: as-http-screenshot-domains
		as-ip-recon:
			count: 2
			enabled: true
			order: 400
			script: as-ip-recon
		as-ip-resolution:
			count: 2
			enabled: true
			order: 600
			script: as-ip-resolution
		as-root-domain-recon:
			count: 1
			enabled: true
			order: 0
			script: as-root-domain-recon
		as-subdomain-discovery:
			count: 1
			enabled: true
			order: 50
			script: as-subdomain-discovery
		as-subdomain-enumeration:
			count: 1
			enabled: true
			order: 100
			script: as-subdomain-enumeration
```

- **as-combine-subdomains**: Combines all discovered subdomains for each in scope root domain, removing duplicates and blacklisted domains.
- **as-dns-resolution**: Runs DNS resolution for each root domain's `subdomains.txt` file created from the `as-combine-subdomains` script.
- **as-domains-from-domain-ssl-certs**: Retrieves subdomains from SSL/TLS certificates for the hosts in each root domain's `subdomains.txt` file created from the `as-combine-subdomains` script.
- **as-domains-from-ip-ssl-certs**: Retrieves subdomains from SSL/TLS certificates for all discovered IPs.
- **as-http-screenshot-domains**: Runs `aquatone` on all discovered domains to take screenshots of the web pages found.
- **as-ip-recon**: Gathers the discovered IPs, runs an nmap ping scan and organizes them into different files based on IP version and private/public ranges.
- **as-ip-resolution**: Runs reverse DNS resolution for all discovered IPs in the public ranges.
- **as-root-domain-recon**: Creates a `subdomains-discovered.txt` file from `scope-domains*.txt`, runs `whois` and queries different DNS records for each in scope root domain.
- **as-subdomain-discovery**: Queries `https://crt.sh` and runs amass enum and intel for each in scope root domain
- **as-subdomain-enumeration**: Currently doesn't do anything. Probably should use `gobuster` or something else to discover subdomains using a wordlist.

**Note:** Some of these scripts do rely on each other. Disabling one might cause errors. This hopefully will change in the future.

### Recon

The recon phase is for running active recon against the discovered hosts. Currently active recon consists of TCP/UDP nmap scans, web content discovery, and `aquatone` screenshots.

```yaml
scripts:
	recon:
		as-content-discovery:
			count: 1
			enabled: true
			order: 100
			script: as-content-discovery
		as-http-screenshot-hosts:
			count: 1
			enabled: true
			order: 200
			script: as-http-screenshot-hosts
		as-port-scan-tcp:
			count: 1
			enabled: true
			order: 0
			script: as-port-scan-tcp
		as-port-scan-udp:
			count: 1
			enabled: true
			order: 300
			script: as-port-scan-udp
```
- **as-content-discovery**: Runs content enumeration scans with `ffuf` on all hosts with web services.
- **as-http-screenshot-hosts**: Takes screenshots using `aquatone` of the content discovered from `as-content-discovery` that returned a 200 status code.
- **as-port-scan-tcp**: Runs full TCP nmap scans for all discovered hosts.
- **as-port-scan-udp**: Runs full UDP nmap scans for all discovered hosts. These scans have been configured for speed, due to the nature of UDP scanning.

### Hunt

The hunt phase is kind of like the recon phase except its to "hunt" for potential vulnerabilities. Scripts in this phase directly use the recon data by passing it to different tools like `searchsploit` and `nuclei`.

```yaml
scripts:
	hunt:
		as-nuclei-cves:
			count: 1
			enabled: true
			order: 300
			script: as-nuclei-cves
		as-nuclei-technologies:
			count: 1
			enabled: true
			order: 200
			script: as-nuclei-technologies
		as-searchsploit:
			count: 1
			enabled: true
			order: 100
			script: as-searchsploit
		as-takeover-aquatone:
			count: 1
			enabled: true
			order: 0
			script: as-takeover-aquatone
```
- **as-nuclei-cves**: Finds common CVE vulnerabilities for all hosts with web services.
- **as-nuclei-technologies**: Determines the technology stack of all hosts with web services.
- **as-searchsploit**: Passes the nmap scan data directly to `searchsploit` for each host.
- **as-takeover-aquatone**: Searches through the aquatone scans for each host to determine whether possible domain takeovers were found.

### Adding Custom Scripts

It is possible to add custom scripts to be run during the different phases. The scripts for each phase, follow this schema:
```yaml
name:
	count: int
	enabled: boolean
	order: int
	script: string
```
- **name:** The name of the script. This does need to be unique within the phase. Most times it is the same as `script`.
- **count:** The number of times to run the script. Most often it is set to 1. Some scripts in the discover phase run multiple times to ensure domains have been gathered or resolved.
- **order:** This is a number that will determine when in the phase this executes. Lower numbers execute first.
- **script:** The absolute or relative path of the script to run. Scripts within your PATH only need the name of the script.

Adding custom scripts to run during the different phases is as simple as adding the script entry YAML under the desired phase.

## Scripts Directory

This is the location where future `tengo` scripts will be installed. This will make it so the Arsenic repository won't need to be cloned on your machine in order to get the most out of Arsenic.
```yaml
scripts-directory: $HOME/.config/arsenic
```

## Wordlist Paths

The `wordlist-paths` section is an array of file paths to be used when creating the wordlists from the wordlists section of the config.

```yaml
wordlist-paths:
- /opt/SecLists
- /usr/share/seclists
```

## Wordlists

The `wordlist` section contains lists of different wordlists to be combined to generated wordlists for specific purposes.

```yaml
wordlists:
  web-content:
  - Discovery/Web-Content/AdobeCQ-AEM.txt
  - Discovery/Web-Content/apache.txt
  - Discovery/Web-Content/Common-DB-Backups.txt
  - Discovery/Web-Content/Common-PHP-Filenames.txt
  - Discovery/Web-Content/common.txt
  - Discovery/Web-Content/confluence-administration.txt
  - Discovery/Web-Content/default-web-root-directory-linux.txt
  - Discovery/Web-Content/default-web-root-directory-windows.txt
  - Discovery/Web-Content/frontpage.txt
  - Discovery/Web-Content/graphql.txt
  - Discovery/Web-Content/jboss.txt
  - Discovery/Web-Content/Jenkins-Hudson.txt
  - Discovery/Web-Content/nginx.txt
  - Discovery/Web-Content/oracle.txt
  - Discovery/Web-Content/quickhits.txt
  - Discovery/Web-Content/raft-large-directories.txt
  - Discovery/Web-Content/raft-medium-words.txt
  - Discovery/Web-Content/reverse-proxy-inconsistencies.txt
  - Discovery/Web-Content/RobotsDisallowed-Top1000.txt
  - Discovery/Web-Content/websphere.txt
```

Currently Arsenic can only generate the `web-content` wordlist. In the future, it will support being able to generate any wordlist defined in this section.
