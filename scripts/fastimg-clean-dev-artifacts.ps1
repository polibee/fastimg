[CmdletBinding()]
param(
    [switch]$Apply,
    [switch]$IncludeOldRuntime,
    [switch]$IncludeStorageBin
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
    $runtimeRoot = Join-Path $backendRoot '.runtime'
    $runtimeStems = @($running | Where-Object { $_.StartsWith($runtimeRoot + '\', [System.StringComparison]::OrdinalIgnoreCase) } | ForEach-Object { [IO.Path]::GetFileNameWithoutExtension($_) })
    Get-ChildItem -LiteralPath $runtimeRoot -File -ErrorAction SilentlyContinue |
        Where-Object {
            $keep = $false
            foreach ($stem in $runtimeStems) { if ($_.Name.StartsWith($stem, [System.StringComparison]::OrdinalIgnoreCase)) { $keep = $true } }
            -not $keep
        } |
        ForEach-Object {
            try {
                Remove-Item -LiteralPath $_.FullName -Force -ErrorAction Stop
                Write-Host "已删除旧运行产物：$($_.FullName)"
            } catch {
                Write-Host "跳过被占用运行产物：$($_.FullName)"
            }
        }
}

if ($Apply -and $IncludeStorageBin) {
    $running = @()
    Get-Process -ErrorAction SilentlyContinue | ForEach-Object {
        try { if ($_.Path) { $running += $_.Path } } catch { }
    }
    $binRoot = Join-Path $backendRoot 'storage/bin'
    $binStems = @($running | Where-Object { $_.StartsWith($binRoot + '\', [System.StringComparison]::OrdinalIgnoreCase) } | ForEach-Object { [IO.Path]::GetFileNameWithoutExtension($_) })
    Get-ChildItem -LiteralPath $binRoot -File -ErrorAction SilentlyContinue |
        Where-Object {
            $keep = $false
            foreach ($stem in $binStems) { if ($_.Name.StartsWith($stem, [System.StringComparison]::OrdinalIgnoreCase)) { $keep = $true } }
            -not $keep
        } |
        ForEach-Object {
            try {
                Remove-Item -LiteralPath $_.FullName -Force -ErrorAction Stop
                Write-Host "已删除 storage/bin 构建产物：$($_.FullName)"
            } catch {
                Write-Host "跳过被占用 storage/bin 文件：$($_.FullName)"
            }
        }
}
