#!/usr/bin/env pwsh
# Non-interactive script to get dev-login token and run authenticated endpoint checks
$ErrorActionPreference = 'Stop'

$base = 'http://localhost:9292/api/v1'
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$devRespPath = 'E:\project\tea\devlogin_resp.json'
$outFile = 'E:\project\tea\admin_api_test_results_auth.txt'

Write-Output "Starting auth test script: $(Get-Date -Format o)"

try {
    Write-Output "Requesting dev-login token..."
    $dev = Invoke-RestMethod -Uri "$base/user/dev-login" -Method Post -ContentType 'application/json' -Body '{"openid":"admin_openid"}' -TimeoutSec 10
    $dev | ConvertTo-Json -Depth 6 | Out-File -FilePath $devRespPath -Encoding utf8
    Write-Output "Saved dev-login response to $devRespPath"
} catch {
    Write-Output "ERROR: dev-login request failed: $($_.Exception.Message)"
    exit 2
}

# Try common token locations
$token = $null
if ($dev -ne $null) {
    if ($dev.data -ne $null -and $dev.data.token) { $token = $dev.data.token }
    if (-not $token -and $dev.token) { $token = $dev.token }
    if (-not $token -and $dev.data.access_token) { $token = $dev.data.access_token }
}

if (-not $token) {
    Write-Output "ERROR: could not locate token in dev-login response. Dumping response:" | Out-File -FilePath $outFile -Encoding utf8
    ($dev | ConvertTo-Json -Depth 6) | Out-File -FilePath $outFile -Append -Encoding utf8
    exit 3
}

Write-Output "Token found (length $($token.Length)); starting endpoint probes..."

$endpoints = @(
    '/health','/auth/captcha','/auth/login','/auth/me','/admin/menus','/admin/users',
    '/products','/products/1','/categories','/categories/1','/stores',
    '/admin/stores/1/orders','/admin/stores/1/orders/stats','/admin/stores/1/products','/admin/stores/1/products/1',
    '/admin/products','/admin/orders','/admin/orders/export','/admin/orders/1',
    '/orders/1/deliver','/orders/1/complete','/orders/1/cancel','/orders/1/admin-cancel','/orders/1/refund','/orders/1/refund/start','/orders/1/refund/confirm',
    '/admin/logs/operations?module=finance&order_id=1&limit=5','/admin/logs/operations','/admin/logs/operations/export','/admin/logs/access','/admin/logs/access/export',
    '/admin/refunds','/admin/refunds/export'
)

# endpoints that should be POST instead of GET
$postPatterns = @('/orders/1/deliver','/orders/1/complete','/orders/1/cancel','/orders/1/admin-cancel','/orders/1/refund','/orders/1/refund/start','/orders/1/refund/confirm')

# initialize/overwrite output file
"Authenticated API test results - $(Get-Date -Format o)`n`n" | Out-File -FilePath $outFile -Encoding utf8

foreach ($p in $endpoints) {
    $url = $base + $p
    $method = 'GET'
    if ($postPatterns -contains $p) { $method = 'POST' }
    Write-Output "Probing $method $url"
    try {
        if ($method -eq 'GET') {
            $resp = Invoke-WebRequest -Uri $url -Method Get -Headers @{ Authorization = "Bearer $token" } -UseBasicParsing -TimeoutSec 8
            $status = $resp.StatusCode
            $body = $resp.Content
        } else {
            # POST with empty JSON body
            $resp = Invoke-WebRequest -Uri $url -Method Post -Headers @{ Authorization = "Bearer $token" } -Body '{}' -ContentType 'application/json' -UseBasicParsing -TimeoutSec 8
            $status = $resp.StatusCode
            $body = $resp.Content
        }
    } catch {
        $status = 'ERROR'
        $body = $_.Exception.Message
    }
    if ([string]::IsNullOrEmpty($body)) { $preview = '(empty body)' } elseif ($body.Length -gt 800) { $preview = $body.Substring(0,800) + '...' } else { $preview = $body }
    $entry = "URL: $url`nMETHOD: $method`nSTATUS: $status`nBODY_PREVIEW:`n$preview`n----`n"
    $entry | Out-File -FilePath $outFile -Append -Encoding utf8
}

Write-Output "Finished probes; results saved to $outFile"
exit 0
