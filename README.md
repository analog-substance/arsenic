# Arsenic [as]
Now with moar go!!!
*******

Arsenic aims to set conventions around how pentest data is stored. It is nothing more than a directory structure and file naming conventions. By itself it is nothing fancy, but when combined with things like [arsenic-hugo](https://github.com/analog-substance/arsenic-hugo), it should make operations fun again!

An example operation directory structure would look like.
```
├── apps
├── bin
├── config.toml
├── hosts
│   └── 127.0.0.1
│       ├── README.md
│       └── recon
├── Makefile
├── README.md -> report/sections/README.md
├── recon
│   └── domains
└── report
    ├── findings
    │   └── first-finding
    │       ├── 00-metadata.md
    │       ├── affected_assets.md
    │       ├── recommendations.md
    │       ├── references.md
    │       ├── steps_to_reproduce.md
    │       └── summary.md
    ├── sections
    │   └── README.md
    └── static
```

## Operation Directory Layout Definitions

### apps/
A free form place to store applications. So far no magic here. Open to suggestions

### bin/
Every operation is different; use this directory for one off operation scripts.

### hosts/
This is where hosts information is stored. Host directories will typically be named after the host IPv4 or IPv6 address, but a hostname should work (untested). Every host should have README.md and hostname files, as well as recon and loot directories.

### recon/
The recon directory in the operation root will contain all the recon for the operation as a whole. when doing host specific recon it should be in the host's recon directory.

### report/

Every operation should have findings! This is where to store that information.

## Getting Started

to start an op:

```bash
arsenic init opname

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

## Collaboration

Working with friends? Not a problem. [arsenic-hugo](https://github.com/analog-substance/arsenic-hugo) should make it easier to see the big picture.

### Reviewing Hosts

```bash
export REVIEWER='defektive'

ar-mark-reviewed.sh 192.168.0.1 # this will set reviewer = 'defektive' in the README for the host
arsenic hosts -u
```
*****
## Installation

Oh goodie!! This is all still very much a WIP, but it should get better as it gets used more. Assuming you already have `go` setup and your `arsenic` is in `$HOME/opt` or `/opt`, to get started:
```bash
cd arsenic
go install
```
Now you should be able to run `arsenic`. If not something is wrong.

### Suggested Installation

- Arsenic is intended to be checked out along side other similarly purposed tools in an `opt/` directory.
- Arsenic will automatically add the `bin/` directory of sibling directories (eg: `opt/arsenic/bin`, `opt/xe/bin`) to your `$PATH`

*******
## Requirements

- curl
- aquatone
- ncat
- nmap
- awk
- sed
- grep
- figlet
- exploitdb (searchsploit)
