---
title: Arsenic
linkTitle: Docs
menu: {main: {weight: 20}}
---
> Conventions and automation for offensive operations.
> https://analog-substance.github.io/arsenic/

## Purpose

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
- fast-resolv (https://github.com/defektive/fast-resolv)

#### note on nmap

most scans will require nmap to be run as root or have the appropriate capabilities set on the nmap binary.

```bash
sudo setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip /usr/bin/nmap
```

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
![Arsenic Init Example](/docs/examples/arsenic-init.gif)

#### Customization

If you want to customize the op creation process for whatever reason, there are two ways to do so. The first is by adding custom scripts to the `scripts.init` section of the config file located in your home directory. Refer to the "[Adding Custom Scripts](docs/docs/config.md#adding-custom-scripts)" section of the config documentation for more information.

The second way is by creating an init hook script. The `arsenic init` command will run `as-init-op.sh` scripts located at `opt/*/scripts`, where the opt directory is where the Arsenic repository is located. Assuming the Arsenic repository is located at `$HOME/opt/arsenic`, create a script at `$HOME/opt/custom-arsenic/scripts/as-init-op.sh`. Anything in this script will execute when running `arsenic init`.

### Running an Op

With the op initialized, we must fill out the `scope-domains.txt` and `scope-ips.txt` files with the op's scope. These files contain the hosts that will be used to discover new domains and IPs and will always be regarded as in scope.

```bash
echo example.com >> scope-domains.txt
echo 127.0.0.1 >> scope-ips.txt
```

After the scope has been filled out, we can now run `arsenic discover` which will use the scope to discover subdomains and IP addresses using various tools/services.

![Arsenic Discover](docs/example/arsenic-discover.gif)

To see everything that was discovered, run `arsenic scope`

![Arsenic Discover Scope](docs/example/arsenic-discover-scope.gif)

There may be subdomains and IPs that were discovered but that are not in scope. Refer to the [blacklist](docs/docs/config.md#blacklist) section of the config documentation for more information on how to update the blacklisted domains and IPs. If you do want to re-run the `discover` command after updating the blacklist, remove the `scope-domains-*` and `scope-ips-*` files along with the `recon/domains/*` and `recon/ips/*` directories.

Now that we have discovered more subdomains and IPs, we can use Arsenic to analyze the data and group the hosts by common IP.

```bash
arsenic analyze -c
```

![Arsenic Analyze](docs/example/arsenic-analyze.gif)

This will create your directories in `hosts/`. Now you can run.

```bash
arsenic recon
```

This will probably take a while... but when its done you should have port scans, content discovery, and screen shots.

******

### Config

Refer to the [config](docs/docs/config.md) documentation for more information.

### Tengo Scripting
Currently some of the arsenic scripts are written in the [tengo](https://github.com/d5/tengo) scripting language. These scripts use tengo builtin functions and modules along with custom functions and modules only available to arsenic scripts.

#### References
- [Standard Library](docs/docs/tengo/stdlib.md)
- [Builtin Functions](docs/docs/tengo/builtin.md)
- [Scripting with Arsenic](docs/docs/tengo/scripting.md)

## Collaboration

Working with friends? Not a problem. [arsenic-hugo](https://github.com/analog-substance/arsenic-hugo) should make it easier to see the big picture.

<!-- ### Reviewing Hosts

```bash
export REVIEWER='defektive'

arsenic hosts -u
```
***** -->

## Brainstorming space

### Current state for trying new things


#### Populating scope

Create a test folder
```bash
mkdir ~/arsenic-tutorial
cd ~/arsenic-tutorial
```

init a git repo
```bash
git init
mkdir tmp
echo /tmp >> .gitignore
git add .gitignore
git commit -m "gitignore"
```

Using HackerOne's bug bounty as an example. Pull down scope and convert it JSON with mlr.
```bash
curl -s https://hackerone.com/teams/security/assets/download_csv.csv | mlr --icsv --ojson cat | jq | tee hackerone-scope.json
```

```bash
git add hackerone-scope.json
git commit -m "scope"
```

Get in scope items. Fore demonstration puposes we'll remove anything with a max_severity of low and things with wildcard references (we'll enumerate those later).
```bash
cat hackerone-scope.json | jq '.[]|select(.eligible_for_bounty == "true")|select(.eligible_for_submission == "true")|select(.max_severity != "low") | .identifier' -r | grep -v "\*" | arsenic dev-scope add
```


```bash
git add data
git commit -m "add scope config"
```

Now we should see them by running
```bash
arsenic dev-scope domains
arsenic dev-scope ips

# you can expand the IPs as well
arsenic dev-scope ips -x
```

#### Discovery: Getting IPs

use nmap to get IPs. use `arsenic capture` to capture input and output.

```bash
arsenic capture -- nmap -iL $(arsenic dev-scope get -d) -sL --resolve-all
arsenic capture -- nmap -iL $(arsenic dev-scope get -4) -sL --resolve-all
```


```bash
git add data
git commit -m "expanded scope"
```

#### Discovery: Alive hosts

Save public IPs to a tmp location
```bash
arsenic inspect hosts --ips --public > tmp/public-ips.txt
```

run host discovery
```bash
as-nmap-host-discovery.tengo -f tmp/public-ips.txt -T5
```

```bash
git add data
git commit -m "host discovery"
```

#### Discovery: Port Scans

Explore results
```bash
arsenic inspect hosts --public --up
```

Save them to a temporary file. We are using the IPs here to ensure we do not scan hosts more than once. Since one IP
address can have multiple domains pointing at it.
```bash
arsenic inspect hosts --public --up --ips > tmp/alive-ips.txt
```

Run incremental port scans.
```bash
as-nmap-incremental.tengo -f tmp/alive-ips.txt
```

Wait....

```bash
git add data
git commit -m "port scans"
```

While we wait. Let's go look at the wildcard domains.

```bash
cat hackerone-scope.json | jq '.[]|select(.eligible_for_bounty == "true")|select(.eligible_for_submission == "true")| .identifier' -r | grep "\*"
```

This should return something like:
```txt
https://*.hackerone-ext-content.com
*.vpn.hackerone.net
https://*.hackerone-user-content.com/
```

we'll save the following in a tmp file `tmp/subfinder-targets.txt`
```txt
hackerone-ext-content.com
vpn.hackerone.net
hackerone-user-content.com
```

now lets run `subfinder` and use `arsenic capture`
```bash
arsenic capture -- subfinder -dL tmp/subfinder-targets.txt
```

```bash
git add data
git commit -m "subfinder"
```

now add the results to scope:

```bash
cat data/default/output/subfinder/**/**.json | jq -r '.host' | arsenic dev-scope add
```
check diff

```bash
git diff
```


Lets add the low severity ones in.

```bash
cat hackerone-scope.json | jq '.[]|select(.eligible_for_bounty == "true")|select(.eligible_for_submission == "true")| .identifier' -r | grep -v "\*"  | arsenic dev-scope add
```

Now we can start the process over again. since we used `arsenic capture` only things that haven't been scanned will get scanned.

```bash
arsenic capture -- nmap -iL $(arsenic dev-scope get -d) -sL --resolve-all
arsenic inspect hosts --ips --public > tmp/public-ips.txt
as-nmap-host-discovery.tengo -f tmp/public-ips.txt


```







### Plans and notes

Currently, lots of duplicate scanning can occur if scope is added after the initial discovery phase. it takes a decent
amount of effort to make sure you don't perform duplicate scans.

thoughts on how to best perform initial recon.

1. expand in scope IPs.
2. resolve domains to IPs.
3. host discovery on unique IPs.
4. port scans on discovered hosts.
5. content enumeration on web ports.
6. pull domains from tls certs (go to #2)
7. perform subdomain enumeration (go to #2)
8. search for subdomain takeovers
9. nuclei tech detection
10. nuclei templates

If we keep track of what commands are executed and detect what input is passed in, we can determine if a particular
scope item has had a specific program run against it.

```go

type Domain {
  Value string
}

type IP {
  Version int
  Value string
  Private bool
}

type Host struct {
  Domains []Domain
  IPs []IP
}

type DNSRecord struct {
  Domain Domain
  Type string
  Value string
  TTL int
}

type Port struct {
  IP IP
  Port int
  Protocol string
  Service string
  Fingerprint string
}

type Content struct {
  URL string
  HTTPStatusCode int
}


```