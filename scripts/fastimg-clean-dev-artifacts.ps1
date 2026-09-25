[CmdletBinding()]
param(
    [switch]$Apply,
    [switch]$IncludeOldRuntime
)

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$backendRoot = Join-Path $repoRoot 'backend'

function Get-TargetDirectories {
    $targets = @()
    $targets += Get-ChildItem -LiteralPath $repoRoot -Force -Directory -Filter '.gocache*' -ErrorAction SilentlyContinue
    $targets += Get-ChildItem -LiteralPath $backendRoot -Force -Directory -Filter '.gocache*' -ErrorAction SilentlyContinue
    $storageCache = Join-Path $backendRoot 'storage/.cache-go-build'
    if (Test-Path -LiteralPath $storageCache) { $targets += Get-Item -LiteralPath $storageCache }
    $fallbackCache = Join-Path $backendRoot '.cache'
    if (Test-Path -LiteralPath $fallbackCache) { $targets += Get-Item -LiteralPath $fallbackCache }
    return @($targets | Sort-Object FullName -Unique)
}

$targets = Get-TargetDirectories
if (-not $Apply) {
    Write-Host '清理预览：以下目录仅在传入 -Apply 时删除。'
    $targets | Select-Object FullName | Format-Table -AutoSize
    Write-Host '不会删除 backend/storage/fastimg、数据库文件或日志。'
    Write-Host '如需继续：pwsh -File scripts/fastimg-clean-dev-artifacts.ps1 -Apply'
    exit 0
}

foreach ($target in $targets) {
    $resolved = (Resolve-Path -LiteralPath $target.FullName).Path
    if ($resolved -notlike "$repoRoot\*") { throw "拒绝删除仓库外路径：$resolved" }
    Remove-Item -LiteralPath $resolved -Recurse -Force
    Write-Host "已删除：$resolved"
}

if ($IncludeOldRuntime) {
    $running = @()
    Get-Process -ErrorAction SilentlyContinue | ForEach-Object {
        try { if ($_.Path) { $running += $_.Path } } catch { }
    }
    Get-ChildItem -LiteralPath (Join-Path $backendRoot '.runtime') -File -Filter '*.exe' -ErrorAction SilentlyContinue |
        Where-Object { $running -notcontains $_.FullName -and $_.Name -ne 'fastimg-dev.exe' } |
        ForEach-Object {
            Remove-Item -LiteralPath $_.FullName -Force
            Write-Host "已删除旧运行产物：$($_.FullName)"
        }
}
