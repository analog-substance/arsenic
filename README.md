# Arsenic

Arsenic aims to set conventions around how pentest data is stored. It is nothing more than a directory structure and file naming conventions. By itself it is nothing fancy, but when combined with things like [arsenic-hugo](https://github.com/defektive/arsenic-hugo) or [xenon](https://github.com/defektive/xenon) it should make operations fun again!

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

Helper scripts for running engagements.

```bash
source path/to/arsenic.rc
cd op_dir
echo 192.168.0.1/24 > scope-ips-initial.txt
arsenic
```

## Suggested Installation

- Arsenic is intended to be checked out along side other similarly purposed tools in an `opt/` directory.
- Arsenic will automatically add the `bin/` directory of sibling directories (eg: `opt/arsenic/bin`, `opt/xe/bin`) to your `$PATH`

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
