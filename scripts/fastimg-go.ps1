[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$Command,

    [Parameter(ValueFromRemainingArguments = $true, Position = 1)]
    [string[]]$Arguments
)

$repoRoot = Split-Path -Parent $PSScriptRoot
$backendRoot = Join-Path $repoRoot 'backend'
$localCacheRoot = [Environment]::GetFolderPath('LocalApplicationData')
if ([string]::IsNullOrWhiteSpace($localCacheRoot)) {
    $localCacheRoot = Join-Path $repoRoot 'backend/.cache'
}
$goCache = if ($env:FASTIMG_GOCACHE_ROOT) {
    $env:FASTIMG_GOCACHE_ROOT
} else {
    Join-Path $localCacheRoot 'FastImg/go-build'
}

try {
    New-Item -ItemType Directory -Force -Path $goCache -ErrorAction Stop | Out-Null
} catch {
    $goCache = Join-Path $repoRoot 'backend/.cache/go-build'
    New-Item -ItemType Directory -Force -Path $goCache -ErrorAction Stop | Out-Null
    Write-Warning "无法写入用户缓存目录，回退到仓库内唯一缓存：$goCache"
}
$previousGoCache = $env:GOCACHE
$env:GOCACHE = $goCache
$exitCode = 1

Push-Location $backendRoot
try {
    Write-Host "FastImg Go cache: $goCache"
    & go $Command @Arguments
    $exitCode = $LASTEXITCODE
} finally {
    Pop-Location
    if ($null -eq $previousGoCache) {
        Remove-Item Env:GOCACHE -ErrorAction SilentlyContinue
    } else {
        $env:GOCACHE = $previousGoCache
    }
}

exit $exitCode
