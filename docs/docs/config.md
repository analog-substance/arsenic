---
title: Config
description: Config
weight: 10
---

# Config

The configuration file holds many different settings to fine tune commands and scripts run by Arsenic.

## Analyze

The `analyze` sections contains options to fine to the `arsenic analyze` command.

```yaml
analyze:
  require-open-ports: true
```

- **require-open-ports**: Whether to require open ports when `arsenic analyze` is used against an Nmap host discovery
  scan

## Blacklist

The blacklist section contains the domains and IPs to filter out of the scope. This really only matters if subdomain
discovery is in scope. If it is out of scope, meaning you must stick to what is in the `scope-domains.txt` and
`scope-ips.txt` files, this can be skipped.

```yaml
blacklist:
  domains: [ ]
  ips: [ ]
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

- **`domains`**: An array of regex strings to filter from the discovered domains. Refer
  to https://pkg.go.dev/regexp/syntax for the supported syntax.
- **`ips`**: An array of IP addresses to filter from the discovered IPs.
- **`root-domains`**: An array of root domains to filter from the discovered domains.

**Note:** The domains and IPs in `scope-domains.txt` and `scope-ips.txt` will never be filtered. This is to ensure
client provided information is always included in scope.

## Discover

The `discover` section contains options to fine tune some of the scripts in the `discover` phase

```yaml
discover:
  resolvconf: ""
  timing-profile: 4
  top-tcp-count: 30
  top-udp-count: 30
```

- **`resolvconf`**: The path of the resolv-conf file to use for DNS resolution
- **`timing-profile`**: The Nmap timing profile to use for host discovery
- **`top-tcp-count`**: The number of top TCP ports to use for host discovery
- **`top-udp-count`**: The number of top UDP ports to use for host discovery

## Hosts

The `hosts` sections contains configurations to fine tune the `arsenic hosts` command. In the, hopefully, near future,
the automatic flags will be configurable here.

```yaml
hosts:
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
  nmap-xml-glob: nmap-*-??p.xml
```

- **`ignore-services`**: A mapping of ports to Nmap services to ignore when parsing a host's Nmap scans.
    - **`name`**: The name of the Nmap service
    - **`ports`**: The ports/port ranges of the service
    - **`flag`**: An optional flag to add to the host if the service and ports match

**Note:** Both the ports and name must match in order for the service to be ignored.

- **`nmap-xml-glob`**: The glob to use to search for the Nmap XML scans for a host. The Nmap XML scans are parsed to get
  port and service data

## Scripts

The `scripts` section contains phases and global script options.

```yaml
scripts:
  directory: $HOME/.config/arsenic
  phases:
    ...
```

- **`directory`**: This is the location where future `tengo` scripts will be installed. This will make it so the Arsenic
  repository won't need to be cloned on your machine in order to get the most out of Arsenic.

### Phases

Phases are simply scripts grouped together for a single purpose. By default, Arsenic supports four separate phases:
init, discover, recon, and hunt. These phases normally map to the different phases of pen-testing with the exception of
init.

Phases have the following schema:

```yaml
phase-name:
  args: string
  scripts:
    name:
      args: string
      count: int
      enabled: boolean
      order: int
      script: string
```

- **`phase-name`**: The name of the phase. Must be unique since it is a YAML key and not a value
    - **`args`**: The global arguments to pass to each script in the phase. This can be useful, for example, when all
      the phase scripts support a proxy argument and you need to run the scripts in the phase through a proxy
    - **`scripts`**: The scripts in the phase
        - **`name`**: The name of the script. Must be unique since it is a YAML key and not a value. Most times it is
          the same as `script`.
            - **`args`**: The arguments to pass to the script. These are appended to the phase arguments if any exist
            - **`count`:** The number of times to run the script. Most often it is set to 1. Some scripts in the
              discover phase run multiple times to ensure domains have been gathered or resolved.
            - **`order`:** This is a number that will determine when in the phase this executes. Lower numbers execute
              first.
            - **`script`:** The absolute or relative path of the script to run. Scripts within your PATH only need the
              name of the script.

#### Init

These scripts aren't part of a phase of testing. They are run when `arsenic init` is called to create the op.

```yaml
scripts:
  phases:
    init:
      args: ""
      scripts:
        as-init-cleanup:
          args: ""
          count: 1
          enabled: true
          order: 300
          script: as-init-cleanup
        as-init-hooks:
          args: ""
          count: 1
          enabled: true
          order: 200
          script: as-init-hooks
        as-init-op:
          args: ""
          count: 1
          enabled: true
          order: 0
          script: as-init-op
        as-setup-hugo:
          args: ""
          count: 1
          enabled: true
          order: 100
          script: as-setup-hugo
```

- **`as-init-cleanup`**: Runs common cleanup tasks.
- **`as-init-hooks`**: Calls `as-init-op.sh` files for custom op initialization. Will change in the future to run a
  tengo script so as to be run on any OS.
- **`as-init-op`**: Creates the necessary directory/file structure for the op.
- **`as-setup-hugo`**: Sets up the op for use with `hugo`.

#### Discover

The discover phase is for actively and passively discovering new subdomains and IPs. The scripts run during this phase
generally only require the domains and IPs contained in the scoping files.

```yaml 
scripts:
  phases:
    discover:
      args: ""
      scripts:
        as-combine-subdomains:
          args: ""
          count: 2
          enabled: true
          order: 250
          script: as-combine-subdomains
        as-dns-resolution:
          args: ""
          count: 2
          enabled: true
          order: 300
          script: as-dns-resolution
        as-domains-from-domain-ssl-certs:
          args: ""
          count: 1
          enabled: true
          order: 275
          script: as-domains-from-domain-ssl-certs
        as-domains-from-ip-ssl-certs:
          args: ""
          count: 2
          enabled: true
          order: 500
          script: as-domains-from-ip-ssl-certs
        as-http-screenshot-domains:
          args: ""
          count: 1
          enabled: true
          order: 700
          script: as-http-screenshot-domains
        as-ip-recon:
          args: ""
          count: 2
          enabled: true
          order: 400
          script: as-ip-recon
        as-ip-resolution:
          args: ""
          count: 2
          enabled: true
          order: 600
          script: as-ip-resolution
        as-root-domain-recon:
          args: ""
          count: 1
          enabled: true
          order: 0
          script: as-root-domain-recon
        as-subdomain-discovery:
          args: ""
          count: 1
          enabled: true
          order: 50
          script: as-subdomain-discovery
        as-subdomain-enumeration:
          args: ""
          count: 1
          enabled: true
          order: 100
          script: as-subdomain-enumeration
```

- **`as-combine-subdomains`**: Combines all discovered subdomains for each in scope root domain, removing duplicates and
  blacklisted domains.
- **`as-dns-resolution`**: Runs DNS resolution for each root domain's `subdomains.txt` file created from the
  `as-combine-subdomains` script.
- **`as-domains-from-domain-ssl-certs`**: Retrieves subdomains from SSL/TLS certificates for the hosts in each root
  domain's `subdomains.txt` file created from the `as-combine-subdomains` script.
- **`as-domains-from-ip-ssl-certs`**: Retrieves subdomains from SSL/TLS certificates for all discovered IPs.
- **`as-http-screenshot-domains`**: Runs `aquatone` on all discovered domains to take screenshots of the web pages
  found.
- **`as-ip-recon`**: Gathers the discovered IPs, runs an nmap ping scan and organizes them into different files based on
  IP version and private/public ranges.
- **`as-ip-resolution`**: Runs reverse DNS resolution for all discovered IPs in the public ranges.
- **`as-root-domain-recon`**: Creates a `subdomains-discovered.txt` file from `scope-domains*.txt`, runs `whois` and
  queries different DNS records for each in scope root domain.
- **`as-subdomain-discovery`**: Queries `https://crt.sh` and runs amass enum and intel for each in scope root domain
- **`as-subdomain-enumeration`**: Currently doesn't do anything. Probably should use `gobuster` or something else to
  discover subdomains using a wordlist.

**Note:** Some of these scripts do rely on each other. Disabling one might cause errors. This hopefully will change in
the future.

#### Recon

The recon phase is for running active recon against the discovered hosts. Currently active recon consists of TCP/UDP
Nmap scans, web content discovery, and `aquatone` screenshots.

```yaml
scripts:
  phases:
    recon:
      args: ""
      scripts:
        as-content-discovery:
          args: ""
          count: 1
          enabled: true
          order: 100
          script: as-content-discovery
        as-http-screenshot-hosts:
          args: ""
          count: 1
          enabled: true
          order: 200
          script: as-http-screenshot-hosts
        as-port-scan-tcp:
          args: ""
          count: 1
          enabled: true
          order: 0
          script: as-port-scan-tcp
        as-port-scan-udp:
          args: ""
          count: 1
          enabled: true
          order: 300
          script: as-port-scan-udp
```

- **`as-content-discovery`**: Runs content enumeration scans with `ffuf` on all hosts with web services.
- **`as-http-screenshot-hosts`**: Takes screenshots using `aquatone` of the content discovered from
  `as-content-discovery` that returned a 200 status code.
- **`as-port-scan-tcp`**: Runs full TCP Nmap scans for all discovered hosts.
- **`as-port-scan-udp`**: Runs full UDP Nmap scans for all discovered hosts. These scans have been configured for speed,
  due to the nature of UDP scanning.

#### Hunt

The hunt phase is kind of like the recon phase except its to "hunt" for potential vulnerabilities. Scripts in this phase
directly use the recon data by passing it to different tools like `searchsploit` and `nuclei`.

```yaml
scripts:
  phases:
    hunt:
      args: ""
      scripts:
        as-nuclei-cves:
          args: ""
          count: 1
          enabled: true
          order: 300
          script: as-nuclei-cves
        as-nuclei-technologies:
          args: ""
          count: 1
          enabled: true
          order: 200
          script: as-nuclei-technologies
        as-searchsploit:
          args: ""
          count: 1
          enabled: true
          order: 100
          script: as-searchsploit
        as-takeover-aquatone:
          args: ""
          count: 1
          enabled: true
          order: 0
          script: as-takeover-aquatone
```

- **`as-nuclei-cves`**: Finds common CVE vulnerabilities for all hosts with web services.
- **`as-nuclei-technologies`**: Determines the technology stack of all hosts with web services.
- **`as-searchsploit`**: Passes the Nmap scan data directly to `searchsploit` for each host.
- **`as-takeover-aquatone`**: Searches through the aquatone scans for each host to determine whether possible domain
  takeovers were found.

#### Adding Custom Scripts

It is possible to add custom scripts to be run during the different phases. Adding custom scripts to run during the
different phases is as simple as adding the script entry YAML under the desired phase.

## Wordlists

The `wordlist` section contains the options for generating different types of wordlists

```yaml
wordlists:
  paths:
    - /opt/SecLists
    - /usr/share/seclists
  types:
    sqli:
      - Fuzzing/Databases/sqli.auth.bypass.txt
      - Fuzzing/Databases/MSSQL.fuzzdb.txt
      - Fuzzing/Databases/MSSQL-Enumeration.fuzzdb.txt
      - Fuzzing/Databases/MySQL.fuzzdb.txt
      - Fuzzing/Databases/NoSQL.txt
      - Fuzzing/Databases/db2enumeration.fuzzdb.txt
      - Fuzzing/Databases/Oracle.fuzzdb.txt
      - Fuzzing/Databases/MySQL-Read-Local-Files.fuzzdb.txt
      - Fuzzing/Databases/Postgres-Enumeration.fuzzdb.txt
      - Fuzzing/Databases/MySQL-SQLi-Login-Bypass.fuzzdb.txt
      - Fuzzing/SQLi/Generic-BlindSQLi.fuzzdb.txt
      - Fuzzing/SQLi/Generic-SQLi.txt
      - Fuzzing/SQLi/quick-SQLi.txt
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
    xss:
      - Fuzzing/XSS/XSS-Somdev.txt
      - Fuzzing/XSS/XSS-Bypass-Strings-BruteLogic.txt
      - Fuzzing/XSS/XSS-Jhaddix.txt
      - Fuzzing/XSS/xss-without-parentheses-semi-colons-portswigger.txt
      - Fuzzing/XSS/XSS-RSNAKE.txt
      - Fuzzing/XSS/XSS-Cheat-Sheet-PortSwigger.txt
      - Fuzzing/XSS/XSS-BruteLogic.txt
      - Fuzzing/XSS-Fuzzing
```

- **`paths`**: An array of file paths to be used when creating the wordlists from the `types` section of the config.
- **`types`**: Sets of different wordlist file paths to be combined to generate wordlists for specific purposes.

Currently, there are three default wordlist types: sqli, web-content, and xss. Other wordlists paths can be added to the
existing ones or used to create new types of wordlists.
