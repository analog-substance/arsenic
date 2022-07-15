#!/usr/bin/pwsh

[CmdletBinding()]
param (
	[Parameter(
		Mandatory=$true,
		ValueFromPipeline=$true,
		Position=0
	)]
	[string]
	$URL,

	[int]
	$RecursionDepth = 0,

	[string[]]
	$Headers,

	[int]
	$Delay,

	[string]
	$Extensions,

	[string]
	$MatchCodes = "all",

	[string]
	$MatchLines,

	[string]
	$MatchSize,

	[string]
	$MatchWords,

	[string]
	$MatchRegex,

	[string]
	$FilterCodes = "404",

	[string]
	$FilterLines,

	[string]
	$FilterSize,

	[string]
	$FilterWords,

	[string]
	$FilterRegex,

	[string]
	$Method = "GET",

	[string]
	$Proxy,

	[string]
	$Output,

	[string]
	$Wordlist,

	[switch]
	$DisableAutoCalibrate
)

Begin {
	# TODO: Implement the +- append/removing of filter codes
	function Add-Args {
        foreach ($a in $args) {
			$Script:ffufArgs += "$a"
		}
	}

	$Script:ffufArgs = @("-v", "-se", "-X", "$Method", "-mc", "$MatchCodes", "-fc", "$FilterCodes", "-of", "json", "-w", "$Wordlist")
	if ($RecursionDepth -gt 0) {
		Add-Args -recursion -recursion-depth $RecursionDepth
	}

	foreach ($header in $Headers) {
		Add-Args -H "$header"
	}

	if ($Delay -gt 0) {
		Add-Args -p "$Delay"
	}

	if (![string]::IsNullOrEmpty($Extensions)) {
		Add-Args -e "$Extensions"
	}

	if (!$DisableAutoCalibrate) {
		Add-Args "-ac"
	}

	if (![string]::IsNullOrEmpty($MatchLines)) {
		Add-Args -ml "$MatchLines"
	}

	if (![string]::IsNullOrEmpty($MatchSize)) {
		Add-Args -ms "$MatchSize"
	}

	if (![string]::IsNullOrEmpty($MatchWords)) {
		Add-Args -mw "$MatchWords"
	}

	if (![string]::IsNullOrEmpty($MatchRegex)) {
		Add-Args -mr "$MatchRegex"
	}

	if (![string]::IsNullOrEmpty($FilterLines)) {
		Add-Args -fl "$FilterLines"
	}

	if (![string]::IsNullOrEmpty($FilterSize)) {
		Add-Args -fs "$FilterSize"
	}

	if (![string]::IsNullOrEmpty($FilterWords)) {
		Add-Args -fw "$FilterWords"
	}

	if (![string]::IsNullOrEmpty($FilterRegex)) {
		Add-Args -fr "$FilterRegex"
	}

	if (![string]::IsNullOrEmpty($Proxy)) {
		Add-Args -x "$Proxy"
	}

    $outputOverride = $false
    if (![string]::IsNullOrEmpty($Output)) {
		$outputOverride = $true
	}
}

Process {
	$wordlistName = [System.IO.Path]::GetFileNameWithoutExtension($Wordlist)

	$uri = [System.Uri]$URL
	$hostname = $uri.Host
	if (!$outputOverride) {
		$uriPath = $uri.AbsolutePath.Trim("/").Replace("/", ".")
		if (![string]::IsNullOrEmpty($uriPath)) {
			$uriPath = ".$uriPath"
		}
        
        $port = ""
        if ($uri.Port -ne 80 -and $uri.Port -ne 443) {
            $port = ".$($uri.Port)"
        }

		$Output = "ffuf.$Method.$($uri.Scheme).$hostname$port$uriPath.$wordlistName.json"
	}

	$outputPath = "recon\$Output"

	$hostDir = ""
	if (Test-Path -Path hosts) {
		$hostDir = arsenic.exe hosts -H "$hostname" --paths | Select-Object -First 1
		if ([string]::IsNullOrWhiteSpace($hostDir)) {
			$hostDir = "hosts/$hostname"
		}

		mkdir "$hostDir\recon" -Force | Out-Null
		$outputPath = "$hostDir\$outputPath"
	}

	# TODO: Add code to add either the hostname or the ip to the hostnames.txt or ip-address.txt files

	$URL = $URL.TrimEnd("/")
    Write-Output "[+] Running ffuf on $URL"

	if ($url -notmatch ".*FUZZ.*") {
		$URL = "$URL/FUZZ"
	}
    
	& ffuf.exe -u "$URL" -o "$outputPath" $Script:ffufArgs | Out-Host

	cat $outputPath | ConvertFrom-Json | ConvertTo-Json | Out-File -FilePath "$outputPath.new" -Encoding utf8
    mv -Force "$outputPath.new" $outputPath
}