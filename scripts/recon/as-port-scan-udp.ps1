#!/usr/bin/pwsh

[CmdletBinding()]
param (
	[ValidateSet("list", "scan")]
	[string]
	$Action = "scan",

	[string]
	$Target
)

$ErrorActionPreference = 'Stop'

function Needs-Scanning {
	param (
		[string]
		$Target
	)

	return -not (Test-Path -Path "hosts/$Target/recon/nmap-punched-udp*")
}

function Fix-NmapXml {
	param (
		[string]
		$Dir
	)
	Get-ChildItem -Path "$Dir" -Recurse -Filter '*nmap*.xml' | ForEach-Object {
		(Get-Content $_.FullName) -replace '[^"]*nmap\.xsl', "/static/nmap.xsl" | Out-File $_.FullName
	}
}

function Scan-Target {
	if ([string]::IsNullOrEmpty($Target)) {
		return
	}

	Write-Host "[+] Port Scan / UDP / $Target / checking"

	if (Needs-Scanning -Target $Target) {
		$outputDir = "hosts/$Target/recon"

		Write-Host "[-] Port Scan / UDP / $Target / preparing"

		New-Item -Path "$outputDir" -ItemType Directory -Force

		# TODO: gitLock "hosts/$Target/recon/nmap-punched-udp.nmap" "UDP port scan lock: $Target"

		Write-Host "[-] Port Scan / UDP / $Target / running"
		nmap -oA "$outputDir/nmap-punched-udp" -sUV -p- -Pn --max-rtt-timeout 100ms --min-rate 1000 --version-intensity 0 "$Target"
		Fix-NmapXml -Dir "$outputDir"

		# TODO: gitCommit "$outputDir" "UDP port scan complete: $Target"

		Write-Host "[-] Port Scan / UDP / $Target / complete"
	}
	
	$nextTarget = Get-Hosts
	if ([string]::IsNullOrEmpty($nextTarget)) {
		# if grep lock hosts/*/recon/nmap-punched-udp.nmap | grep :lock > /dev/null; then
		# 	_warn "other UDP port scans are still running... lets wait before continuing"
		# 	exit 1
		# fi
	} else {
		Invoke-Expression -Command "$PSCommandPath -Action `"scan`" -Target $nextTarget"
	}
}

function Get-Hosts {
	$excludeHosts = Get-ChildItem -Path hosts -Recurse -File -Filter 'nmap-punched-udp*' `
		| Select-Object -ExpandProperty DirectoryName -Unique `
		| ForEach-Object {
			[regex]::Matches($_, “hosts[\\/]([^\\/]+)[\\/]recon”).Groups[1].Value
		}

	$allHosts = Get-ChildItem -Path hosts -Directory | Select-Object -ExpandProperty Name

	Compare-Object -ReferenceObject $allHosts -DifferenceObject $excludeHosts `
		| Select-Object -ExpandProperty InputObject `
		| Sort-Object {Get-Random}
}

if ($Action -eq "list") {
	Get-Hosts
	exit
}

if ([string]::IsNullOrEmpty($Target)) {
	$Target = Get-Hosts | Select-Object -First
	Write-Host "[!] Auto selected $Target"
}

if ([string]::IsNullOrEmpty($Target)) {
	exit
}

Scan-Target
$outputDir = "hosts/$Target"
$reconDir = "$outputDir/recon"
New-Item "$outputDir/loot/passwords", "$reconDir" -ItemType Directory -Force | Out-Null
Scan-Target -OutputDir "$reconDir"
Get-ChildItem -Path "$reconDir" -Recurse -Filter '*nmap*.xml' | ForEach-Object {
	(Get-Content $_.FullName) -replace '[^"]*nmap\.xsl', "/static/nmap.xsl" | Out-File $_.FullName
}