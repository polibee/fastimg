[CmdletBinding()]
param(
    [double]$WarnGB = 8
)

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path

function Get-DirectoryGB([string]$Path) {
    if (-not (Test-Path -LiteralPath $Path)) { return 0.0 }
    $sum = (Get-ChildItem -LiteralPath $Path -Force -File -Recurse -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum).Sum
    return [math]::Round(([double]$sum / 1GB), 2)
}

$rows = @()
function Add-Row([string]$Path, [string]$Class) {
    if (Test-Path -LiteralPath $Path) {
        $script:rows += [pscustomobject]@{
            Class = $Class
            GB = Get-DirectoryGB $Path
            Path = $Path
        }
    }
}

Get-ChildItem -LiteralPath $repoRoot -Force -Directory -Filter '.gocache*' | ForEach-Object { Add-Row $_.FullName '可重建 Go 缓存' }
$backendRoot = Join-Path $repoRoot 'backend'
Get-ChildItem -LiteralPath $backendRoot -Force -Directory -Filter '.gocache*' | ForEach-Object { Add-Row $_.FullName '可重建 Go 缓存' }
Add-Row (Join-Path $backendRoot 'storage/.cache-go-build') '可重建 Go 缓存'
Add-Row (Join-Path $backendRoot '.cache') '可重建 Go 缓存'
Add-Row (Join-Path $repoRoot 'admin/node_modules') '可重建前端依赖'
Add-Row (Join-Path $backendRoot '.runtime') '运行产物，先检查当前进程'
Add-Row (Join-Path $backendRoot 'storage/bin') '构建产物，人工确认'
Add-Row (Join-Path $backendRoot 'storage/fastimg') '受保护：媒体数据'
Add-Row (Join-Path $backendRoot 'storage/logs') '受保护：运行日志'

$rows | Sort-Object GB -Descending | Format-Table Class, GB, Path -AutoSize
$rebuildable = ($rows | Where-Object { $_.Class -in @('可重建 Go 缓存', '可重建前端依赖') } | Measure-Object -Property GB -Sum).Sum
Write-Host ("可重建候选合计：{0:N2} GB" -f [double]$rebuildable)
if ([double]$rebuildable -gt $WarnGB) {
    Write-Warning ("可重建候选超过 {0:N2} GB，请先执行清理预览。" -f $WarnGB)
}
