[CmdletBinding()]
param(
    [switch]$RunNetworkScan,
    [string]$BackendPath,
    [string]$FrontendPath
)

$ErrorActionPreference = 'Continue'
if ([string]::IsNullOrWhiteSpace($BackendPath)) { $BackendPath = Join-Path $PSScriptRoot '..\backend' }
if ([string]::IsNullOrWhiteSpace($FrontendPath)) { $FrontendPath = Join-Path $PSScriptRoot '..\admin' }
$failed = $false

Write-Host 'FastImg dependency vulnerability scan (does not modify go.mod, lockfiles, or node_modules).'

$govulncheck = Get-Command govulncheck -ErrorAction SilentlyContinue
if ($govulncheck) {
    Push-Location $BackendPath
    try {
        & $govulncheck.Source ./...
        if ($LASTEXITCODE -ne 0) { $failed = $true }
    } finally { Pop-Location }
} else {
    Write-Warning 'govulncheck was not found; Go dependency scanning was not run. Install the official golang.org/x/vuln tool and retry.'
    $failed = $true
}

$pnpm = Get-Command pnpm -ErrorAction SilentlyContinue
if (-not $pnpm) {
    Write-Warning 'pnpm was not found; frontend production dependency scanning was not run.'
    $failed = $true
} elseif (-not $RunNetworkScan) {
    Write-Warning 'Network access is disabled by default, so pnpm audit was not run. Pass -RunNetworkScan in an approved network environment.'
    $failed = $true
} else {
    Push-Location $FrontendPath
    try {
        & $pnpm.Source audit --prod --audit-level high
        if ($LASTEXITCODE -ne 0) { $failed = $true }
    } finally { Pop-Location }
}

if ($failed) {
    Write-Error 'Dependency gate failed: a scanner is missing, a scan was skipped, or high/critical issues were found.'
    exit 2
}
Write-Host 'Dependency vulnerability scan passed.'
