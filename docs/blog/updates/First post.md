---
title: Updates are happening
date: 2024-10-16
description: Arsenic improvements are in the works
categories: [ Arsenic ]
tags: [ dev log ]
---

We have learned a lot and changed a lot in how we operate since Arsenic was initially created. We are taking a step back
to rethink how some things work to enable a more streamlined workflow and allow operators more insight and control into
the automations.

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
// random struct brainstorm for data model things 
type Domain struct {
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