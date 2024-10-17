---
title: Analyze
description: Analyze discover data and create
---

## Synopsis

Analyze discover data and create hosts.

This will create a single host for hostnames that resolve to the same IPs

```
arsenic analyze [flags]
```

## Options

```
  -c, --create         really create hosts
  -h, --help           help for analyze
  -i, --ignore-scope   ignore scope
      --nmap           import hosts from recon/nmap-*.xml files
      --private-ips    keep private IPs
  -u, --update         only update existing hosts, dont create new ones
```

## Options inherited from parent commands

```
      --config string   the arsenic.yaml config file
      --debug           the arsenic.yaml config file
```

## SEE ALSO

* [arsenic](arsenic.md)	 - Pentesting conventions

###### Auto generated by spf13/cobra on 17-Oct-2024