# Arsenic [as]

Now with moar go!!!
*******

Arsenic aims to set conventions around how pentest data is stored. It is nothing more than a directory structure and file naming conventions. By itself it is nothing fancy, but when combined with things like [arsenic-hugo](https://github.com/analog-substance/arsenic-hugo), it should make operations fun again!

An example operation directory structure would look like.
```
├── apps
├── bin
├── hosts
│   └── localhost
│       ├── README.md (optional)
│       ├── 00_metadata.md
│       └── recon
│       	├── hostnames.txt
│       	└── ip-addresses.txt
├── recon
│   ├── domains
│   └── leads
├── notes
│   └── example_note.md
├── report
│   ├── findings
│   │   └── first-finding
│   │       ├── 00-metadata.md
│   │       ├── 01-summary.md
│   │       ├── 02-affected_assets.md
│   │       ├── 03-recommendations.md
│   │       ├── 04-references.md
│   │       └── 05-steps_to_reproduce.md
│   ├── sections
│   │   └── README.md
│   ├── social
│   │   └── sample-campaign.md
│   └── static
├── README.md -> report/sections/README.md
├── config.toml
├── arsenic.yaml
└── Makefile
```

## Operation Directory Layout Definitions

### apps/

A free form place to store applications. So far no magic here. Open to suggestions

### bin/

Every operation is different; use this directory for one off operation scripts.

### hosts/

This is where hosts information is stored. Host directories will typically be named after the host's hostname or IPv4/IPv6 address if no hostname exists.

#### hosts/recon

The host recon directory will contain all the recon files for that host only.

### recon/

The recon directory in the operation root will contain all the recon for the operation as a whole.

### report/

Every operation should have findings! This is where to store that information.

## Getting Started

### Prerequisites

To use arsenic, the following are required:
- go v1.16+ (https://go.dev/doc/install or https://github.com/NoF0rte/go-updater)
- aquatone
- nmap
- exploitdb (searchsploit)
- ffuf
- nuclei

#### Optional Prerequisites

To get the best out of arsenic, the following are recommended to be installed:
- hugo (https://gohugo.io/getting-started/installing/)
- npm (https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)

### Installation

Though you are be able to install the arsenic binary by running `go install github.com/analog-substance/arsenic@latest`, you would be missing some key files that have not yet been included in the binary itself. To get the best out of arsenic, run the following:

```bash
git clone https://github.com/analog-substance/arsenic
cd arsenic
go install
```
**Note**: Arsenic is intended to be checked out along side other similarly purposed tools in an `opt/` directory like `$HOME/opt` or `/opt`.

Next, add the following to your shell's rc file:
```bash
source {path_to_arsenic}/arsenic.rc
```
The `arsenic.rc` file automatically adds the `bin/` directory of sibling directories (eg: `opt/arsenic/bin`, `opt/xe/bin`) to your `$PATH`


### Starting an Op

To start an op, run the following:

```bash
arsenic init op_name
```
![Arsenic Init Example](docs/examples/arsenic-init.gif)

#### Customization

If you want to customize the op creation process for whatever reason, there are two ways to do so. The first is by adding custom scripts to the `scripts.init` section of the config file located in your home directory. For more information, on how to do that, click [here](#adding-custom-scripts).

The second way is by creating an init hook script. The `arsenic init` command will run `as-init-op.sh` scripts located at `opt/*/scripts`, where the opt directory is where the Arsenic repository is located. Assuming the Arsenic repository is located at `$HOME/opt/arsenic`, create a script at `$HOME/opt/custom-arsenic/scripts/as-init-op.sh`. Anything in this script will execute when running `arsenic init`.



```bash
echo example.com >> scope-domains.txt
arsenic discover
```

Next we check what hosts we have.
```bash
arsenic analyze
```

If there is something you dont want to scan, blacklist it and re run until you have what you want.


Now you can review things, blacklist things.

```bash
arsenic config -s
```
Then simply add the domain to the `blacklist.domains` array

Once you have everything you want, run:
```bash
arsenic analyze -c
```

This will create your directories in `hosts/`. Now you can run.

```bash
arsenic recon
```

This will probably take a while... but when its done you should have port scans, content discovery, and screen shots.

******

### Config
The configuration file holds many different settings to fine tune commands and scripts run by Arsenic. This is currently the default configuration:
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
scripts-directory: /home/kali/.config/arsenic
wordlist-paths:
- /opt/SecLists
- /usr/share/seclists
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


#### Adding Custom Scripts

TODO

## Collaboration

Working with friends? Not a problem. [arsenic-hugo](https://github.com/analog-substance/arsenic-hugo) should make it easier to see the big picture.

<!-- ### Reviewing Hosts

```bash
export REVIEWER='defektive'

arsenic hosts -u
```
***** -->
