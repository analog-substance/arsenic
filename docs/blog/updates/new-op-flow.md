---
title: New op flow
date: 2024-10-17
description: Arsenic improvements are in the works
categories: [Arsenic]
tags: [dev log]
---

Some progress is being made. It does exist on `main` and in current release builds, but it is rapidly changing based on
usage. Here is a quick example on how things work.

## Populating scope

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
cat hackerone-scope.json | jq '.[]|select(.eligible_for_bounty == "true")|select(.eligible_for_submission == "true")|select(.max_severity != "low") | .identifier' -r | grep -v "\*" | arsenic scopious add
```


```bash
git add data
git commit -m "add scope config"
```

Now we should see them by running
```bash
arsenic scopious domains
arsenic scopious ips

# you can expand the IPs as well
arsenic scopious ips -x
```

## Discovery: Getting IPs

use nmap to get IPs. use `arsenic capture` to capture input and output.

```bash
arsenic capture -- nmap -iL $(arsenic scopious get -d) -sL --resolve-all
arsenic capture -- nmap -iL $(arsenic scopious get -4) -sL --resolve-all
```


```bash
git add data
git commit -m "expanded scope"
```

## Discovery: Alive hosts

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

## Discovery: Port Scans

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

## Add more scope

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
cat data/default/output/subfinder/**/**.json | jq -r '.host' | arsenic scopious add
```
check diff

```bash
git diff
```

```bash
git add data
git commit -m "subfinder results added to scope"
```

## Add more scope part II

Let's add the low severity items we ignored at the bv.

```bash
cat hackerone-scope.json | jq '.[]|select(.eligible_for_bounty == "true")|select(.eligible_for_submission == "true")| .identifier' -r | grep -v "\*"  | arsenic scopious add
```

## Repeat previous commands

Now we can start the process over again. since we used `arsenic capture` only things that haven't been scanned will get scanned.

```bash
arsenic capture -- nmap -iL $(arsenic scopious get -d) -sL --resolve-all
arsenic inspect hosts --ips --public > tmp/public-ips.txt
as-nmap-host-discovery.tengo -f tmp/public-ips.txt
```
