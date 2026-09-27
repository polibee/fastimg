[CmdletBinding()]
param(
    [string]$BaseUrl = 'http://127.0.0.1:53085',
    [string]$Path = '/api/v1/discovery/status',
    [int]$Requests = 100,
    [int]$Concurrency = 10,
    [int]$TimeoutSeconds = 10
)

$ErrorActionPreference = 'Stop'
if ($Requests -lt 1 -or $Concurrency -lt 1) { throw 'Requests and Concurrency must be greater than zero.' }

Add-Type -AssemblyName System.Net.Http
$uri = ([Uri]($BaseUrl.TrimEnd('/') + '/' + $Path.TrimStart('/'))).AbsoluteUri
$handler = [System.Net.Http.HttpClientHandler]::new()
$client = [System.Net.Http.HttpClient]::new($handler)
$client.Timeout = [TimeSpan]::FromSeconds($TimeoutSeconds)
$success = 0
$failed = 0
$statusCounts = @{}
$watch = [Diagnostics.Stopwatch]::StartNew()

try {
    for ($offset = 0; $offset -lt $Requests; $offset += $Concurrency) {
        $count = [math]::Min($Concurrency, $Requests - $offset)
        $tasks = @()
        for ($i = 0; $i -lt $count; $i++) { $tasks += $client.GetAsync($uri) }
        foreach ($task in $tasks) {
            try {
                $response = $task.GetAwaiter().GetResult()
                $code = [int]$response.StatusCode
                if (-not $statusCounts.ContainsKey($code)) { $statusCounts[$code] = 0 }
                $statusCounts[$code]++
                if ($response.IsSuccessStatusCode) { $success++ } else { $failed++ }
                $response.Dispose()
            } catch {
                $failed++
                Write-Warning "Request failed: $($_.Exception.Message)"
            }
        }
    }
} finally {
    $watch.Stop()
    $client.Dispose()
    $handler.Dispose()
}

$rps = [math]::Round(($Requests / [math]::Max($watch.Elapsed.TotalSeconds, 0.001)), 2)
Write-Host "Target: $uri"
Write-Host "Requests: $Requests; concurrency: $Concurrency; success: $success; failed: $failed; elapsed: $([math]::Round($watch.Elapsed.TotalSeconds, 2))s; about $rps req/s"
Write-Host ('Status codes: ' + (($statusCounts.GetEnumerator() | Sort-Object Name | ForEach-Object { "$($_.Name)=$($_.Value)" }) -join ', '))
if ($failed -gt 0) { exit 1 }
