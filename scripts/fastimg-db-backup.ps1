[CmdletBinding()]
param(
    [string]$EnvFile,
    [string]$OutputPath,
    [string]$RestoreDatabase,
    [switch]$VerifyDump,
    [switch]$ConfirmRestore
)

$ErrorActionPreference = 'Stop'
if ([string]::IsNullOrWhiteSpace($EnvFile)) { $EnvFile = Join-Path $PSScriptRoot '..\backend\.env' }

function Read-DotEnv([string]$Path) {
    $values = @{}
    if (-not (Test-Path -LiteralPath $Path)) { return $values }
    foreach ($line in Get-Content -LiteralPath $Path) {
        if ($line -match '^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)\s*$') {
            $value = $Matches[2].Trim()
            if ($value.Length -ge 2 -and (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'")))) {
                $value = $value.Substring(1, $value.Length - 2)
            }
            $values[$Matches[1]] = $value
        }
    }
    return $values
}

function Get-Setting($values, [string]$Name, [string]$Default = '') {
    if ($values.ContainsKey($Name) -and $values[$Name] -ne '') { return [string]$values[$Name] }
    return $Default
}

function Find-PgTool([string]$Name) {
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($command) { return $command.Source }
    $candidates = @(
        (Join-Path 'D:\laragon\bin\postgresql\postgresql\bin' $Name),
        (Join-Path 'C:\Program Files\PostgreSQL\16\bin' $Name),
        (Join-Path 'C:\Program Files\PostgreSQL\15\bin' $Name)
    )
    foreach ($candidate in $candidates) { if (Test-Path -LiteralPath $candidate) { return $candidate } }
    throw "PostgreSQL tool not found: $Name. Install the client tools or add them to PATH."
}

function Invoke-Pg([string]$Tool, [string[]]$Arguments, [string]$Password) {
    $previousPassword = $env:PGPASSWORD
    try {
        if ($Password) { $env:PGPASSWORD = $Password } else { Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue }
        & $Tool @Arguments | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "$([IO.Path]::GetFileName($Tool)) exited with code $LASTEXITCODE" }
    } finally {
        if ($null -eq $previousPassword) { Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue } else { $env:PGPASSWORD = $previousPassword }
    }
}

$settingsPath = (Resolve-Path $EnvFile).Path
$settings = Read-DotEnv $settingsPath
$hostName = Get-Setting $settings 'DB_HOST' '127.0.0.1'
$port = Get-Setting $settings 'DB_PORT' '5432'
$database = Get-Setting $settings 'DB_DATABASE' 'fastimg_dev'
$username = Get-Setting $settings 'DB_USERNAME' 'postgres'
$password = Get-Setting $settings 'DB_PASSWORD'

if ([string]::IsNullOrWhiteSpace($OutputPath)) {
    $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
    $OutputPath = Join-Path ([IO.Path]::GetTempPath()) "fastimg-backups\fastimg-$stamp.dump"
}
$outputFullPath = [IO.Path]::GetFullPath($OutputPath)
$outputParent = Split-Path -Parent $outputFullPath
New-Item -ItemType Directory -Path $outputParent -Force | Out-Null

$pgDump = Find-PgTool 'pg_dump.exe'
$pgRestore = Find-PgTool 'pg_restore.exe'

Write-Host "Backing up database: $database@$hostName`:$port (password is not printed)"
Invoke-Pg $pgDump @('-h', $hostName, '-p', $port, '-U', $username, '-d', $database, '-Fc', '--no-owner', '--no-privileges', '-Z', '6', '-f', $outputFullPath) $password
if (-not (Test-Path -LiteralPath $outputFullPath)) { throw "Backup file was not created: $outputFullPath" }
$sizeMB = [math]::Round(((Get-Item -LiteralPath $outputFullPath).Length / 1MB), 2)
Write-Host "Backup complete: $outputFullPath ($sizeMB MB)"

if ($VerifyDump) {
    Write-Host 'Verifying the backup catalog.'
    Invoke-Pg $pgRestore @('--list', $outputFullPath) ''
    Write-Host 'pg_restore can read the backup catalog.'
}

if ($RestoreDatabase) {
    if (-not $ConfirmRestore) { throw 'Restore can overwrite target objects; pass -ConfirmRestore explicitly.' }
    if ($RestoreDatabase -eq $database) { throw 'Refusing to restore into the source database. Use an isolated temporary database.' }
    Write-Host "Restoring into isolated target: $RestoreDatabase@$hostName`:$port"
    Invoke-Pg $pgRestore @('-h', $hostName, '-p', $port, '-U', $username, '-d', $RestoreDatabase, '--clean', '--if-exists', '--no-owner', '--no-privileges', '--exit-on-error', $outputFullPath) $password
    Write-Host 'Restore complete. Use psql on the target database to check migration version and key table counts.'
}

Write-Host 'Note: this script never creates or drops databases; a DBA must create an isolated target database before restore.'
