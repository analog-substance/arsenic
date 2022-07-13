#!/usr/bin/pwsh

[CmdletBinding()]
param (
	[Parameter(
		Mandatory=$true,
		ValueFromPipeline=$true,
		Position=0
	)]
	[string]
	$Target,

	[switch]
	$Force
)

Begin {
	function Quick-Scan {
        param(
            [string]
            $OutputDir
        )
		if ($Force -or -not (Test-Path -Path "$OutputDir" -Filter "nmap-punched-quick-tcp.nmap" -PathType Leaf)) {
			nmap --open -T3 -n -p- $Target -oA "$OutputDir/nmap-punched-quick-tcp"
		} else {
			Get-Content "$OutputDir/nmap-punched-quick-tcp.nmap"
		}
	}
    
	function Accurate-Punch {
        param(
            [string]
            $OutputDir
        )

		if ((Test-Path -Path "$OutputDir" -Filter "nmap-punched-tcp.nmap" -PathType Leaf) -and -not $Force) {
			Write-Host "[!] Skipping $Target since it was already done"
			return
		}

		Write-Host "[+] scanning $Target"
        
        $ports = Quick-Scan -OutputDir "$OutputDir" | Select-String "\s+open" | ForEach-Object { $_.Line.Split("/")[0] }
		if ($null -ne $ports) {
			$ports = [string]::Join(",", $ports)
		}
        

		if ([string]::IsNullOrEmpty($ports)) {
			Write-Host "[+] ${Target}: No open ports"
			return
		}

		Write-Host "[+] Version scanning $ports"

		nmap -oA "$OutputDir/nmap-punched-tcp" -n -A -p"$ports" $Target
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
}

Process {
	$targetDir = "hosts/$Target"
	$reconDir = "$targetDir/recon"
	New-Item "$targetDir/loot/passwords", "$reconDir" -ItemType Directory -Force | Out-Null
	Accurate-Punch -OutputDir "$reconDir"
	Fix-NmapXml -Dir "$reconDir"
}