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

If you want to customize the op creation process for whatever reason, there are two ways to do so. The first is by adding custom scripts to the `scripts.init` section of the config file located in your home directory. Refer to the "[Adding Custom Scripts](docs/config.md#adding-custom-scripts)" section of the config documentation for more information.

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

Refer to the [config](docs/config.md) documentation for more information.

## Collaboration

Working with friends? Not a problem. [arsenic-hugo](https://github.com/analog-substance/arsenic-hugo) should make it easier to see the big picture.

<!-- ### Reviewing Hosts

```bash
export REVIEWER='defektive'

arsenic hosts -u
```
***** -->
