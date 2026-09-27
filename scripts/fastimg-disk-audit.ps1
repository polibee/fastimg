[CmdletBinding()]
param(
    [double]$WarnGB = 8,
    [switch]$Apply
)

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path

function Get-DirectoryGB([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path)) { return 0.0 }
    $sum = (Get-ChildItem -LiteralPath $Path -Force -File -Recurse -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum).Sum
    if ($null -eq $sum) { return 0.0 }
    return [math]::Round(([double]$sum / 1GB), 2)
}

$rows = @()
function Add-Row([string]$Path, [string]$Class, [bool]$Rebuildable, [bool]$Protected) {
    if (Test-Path -LiteralPath $Path) {
        $script:rows += [pscustomobject]@{
            Class = $Class
            GB = Get-DirectoryGB $Path
            Path = $Path
            Rebuildable = $Rebuildable
            Protected = $Protected
        }
    }
}

Get-ChildItem -LiteralPath $repoRoot -Force -Directory -Filter '.gocache*' -ErrorAction SilentlyContinue | ForEach-Object { Add-Row $_.FullName 'Go build cache' $true $false }
$backendRoot = Join-Path $repoRoot 'backend'
Get-ChildItem -LiteralPath $backendRoot -Force -Directory -Filter '.gocache*' -ErrorAction SilentlyContinue | ForEach-Object { Add-Row $_.FullName 'Go build cache' $true $false }
Add-Row (Join-Path $backendRoot 'storage/.cache-go-build') 'Go build cache' $true $false
Add-Row (Join-Path $backendRoot '.cache') 'Go build cache' $true $false
Add-Row (Join-Path $repoRoot 'admin/node_modules') 'Frontend dependencies' $true $false
Add-Row (Join-Path $backendRoot '.runtime') 'Runtime artifacts; inspect process first' $false $false
Add-Row (Join-Path $backendRoot 'storage/bin') 'Build artifacts; inspect manually' $false $false
Add-Row (Join-Path $backendRoot 'storage/fastimg') 'Protected media data' $false $true
Add-Row (Join-Path $backendRoot 'storage/logs') 'Protected runtime logs' $false $true

if ($rows.Count -gt 0) {
    $rows | Sort-Object GB -Descending | Format-Table Class, GB, Protected, Path -AutoSize
}
$rebuildable = ($rows | Where-Object { $_.Rebuildable -and -not $_.Protected } | Measure-Object -Property GB -Sum).Sum
if ($null -eq $rebuildable) { $rebuildable = 0.0 }
Write-Host ("Rebuildable candidates total: {0:N2} GB" -f [double]$rebuildable)

if ([double]$rebuildable -gt $WarnGB) {
    Write-Warning ("Rebuildable candidates exceed {0:N2} GB. Review the list before cleanup." -f $WarnGB)
}

if (-not $Apply) {
    Write-Host 'Read-only audit complete. Pass -Apply to remove rebuildable, unprotected candidates.'
    exit 0
}

$protectedMedia = (Resolve-Path -LiteralPath (Join-Path $backendRoot 'storage/fastimg') -ErrorAction SilentlyContinue).Path
foreach ($row in ($rows | Where-Object { $_.Rebuildable -and -not $_.Protected })) {
    $resolved = (Resolve-Path -LiteralPath $row.Path -ErrorAction Stop).Path
    if ($protectedMedia -and ($resolved -eq $protectedMedia -or $resolved.StartsWith($protectedMedia + [IO.Path]::DirectorySeparatorChar))) {
        throw 'Refusing to remove protected media storage.'
    }
    Write-Host ("Removing rebuildable candidate: {0}" -f $resolved)
    Remove-Item -LiteralPath $resolved -Recurse -Force -ErrorAction Stop
}
Write-Host 'Cleanup complete. Re-run without -Apply to verify the result.'
