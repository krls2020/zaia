#!/usr/bin/env pwsh
# Copyright (c) 2024-2025 Zerops s.r.o. All rights reserved. MIT license.

$ErrorActionPreference = 'Stop'

if ($v) {
  $Version = "v${v}"
}
if ($Args.Length -eq 1) {
  $Version = $Args.Get(0)
}

$ZaiaInstall = $env:ZAIA_INSTALL
$BinDir = if ($ZaiaInstall) {
  "${ZaiaInstall}\bin"
} else {
  "${Home}\.zerops\bin"
}

$ZaiaExe = "$BinDir\zaia.exe"
$Target = 'win-x64'

$DownloadUrl = if (!$Version) {
  "https://github.com/krls2020/zaia/releases/latest/download/zaia-${Target}.exe"
} else {
  "https://github.com/krls2020/zaia/releases/download/${Version}/zaia-${Target}.exe"
}

if (!(Test-Path $BinDir)) {
  New-Item $BinDir -ItemType Directory | Out-Null
}

curl.exe -Lo $ZaiaExe $DownloadUrl

$User = [System.EnvironmentVariableTarget]::User
$Path = [System.Environment]::GetEnvironmentVariable('Path', $User)
if (!(";${Path};".ToLower() -like "*;${BinDir};*".ToLower())) {
  [System.Environment]::SetEnvironmentVariable('Path', "${Path};${BinDir}", $User)
  $Env:Path += ";${BinDir}"
}

Write-Output ""
Write-Output "ZAIA was installed successfully to ${ZaiaExe}"
Write-Output "Run 'zaia --help' to get started"
Write-Output "Stuck? Join our Discord https://discord.com/invite/WDvCZ54"
