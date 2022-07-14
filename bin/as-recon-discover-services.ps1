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
	$UDP,

	[switch]
	$QuickOnly,

	[switch]
	$Force
)

Begin {
	function Quick-TCPScan {
        param(
            [string]
            $OutputDir
        )
		if ($Force -or -not (Test-Path -Path "$OutputDir\nmap-punched-quick-tcp.nmap")) {
			nmap --open -T3 -n -p- $Target -oA "$OutputDir/nmap-punched-quick-tcp"
		} else {
			Get-Content "$OutputDir/nmap-punched-quick-tcp.nmap"
		}
	}
    
	function Accurate-TCPPunch {
        param(
            [string]
            $OutputDir
        )
        $FixXml = $false

		if ((Test-Path -Path "$OutputDir\nmap-punched-tcp.nmap") -and -not $Force) {
			Write-Output "[!] Skipping $Target since it was already done"
			return
		}

		Write-Output "[+] scanning $Target"
        
        $ports = Quick-TCPScan -OutputDir "$OutputDir" | Select-String "\s+open" | ForEach-Object { $_.Line.Split("/")[0] }
		if ($null -ne $ports) {
			$ports = [string]::Join(",", $ports)
		}

        $FixXml = $true
        

		if ([string]::IsNullOrEmpty($ports)) {
			Write-Output "[+] ${Target}: No open ports"
			return
		}

		if ($QuickOnly) {
			return
		}

		Write-Output "[+] Version scanning $ports"

		nmap -oA "$OutputDir/nmap-punched-tcp" -n -A -p"$ports" $Target
	}

	function Accurate-UDPPunch {
		param(
            [string]
            $OutputDir
        )

        $FixXml = $false

		if ((Test-Path -Path "$OutputDir\nmap-punched-udp.nmap") -and -not $Force) {
			Write-Output "[!] Skipping $Target since it was already done"
			return
		}

        $FixXml = $true

		Write-Output "[+] scanning $Target"
		nmap -oA "$OutputDir/nmap-punched-udp" -sUV -Pn -n -p- --max-rtt-timeout 100ms --min-rate 1000 --version-intensity 0 $Target
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

    $FixXML = $false
}

Process {
	$targetDir = "hosts/$Target"
	$reconDir = "$targetDir/recon"
	New-Item "$targetDir/loot/passwords", "$reconDir" -ItemType Directory -Force | Out-Null
	if ($UDP) {
		Accurate-UDPPunch -OutputDir "$reconDir"
	} else {
		Accurate-TCPPunch -OutputDir "$reconDir"
	}

	if ($FixXML) {
        Fix-NmapXml -Dir "$reconDir"
    }
}