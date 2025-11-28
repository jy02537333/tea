param(
    [int]$IssueNumber
)
if (-not $IssueNumber) { Write-Error 'Provide -IssueNumber'; exit 1 }
$owner = 'jy02537333'
$repo = 'tea'
try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'No GH token'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept='application/vnd.github+json'; 'User-Agent'='test-patch' }
$payload = @{ body = "Draft file test minimal" } | ConvertTo-Json
try {
    $res = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Patch -Body $payload -ContentType 'application/json'
    Write-Host ("Patched #{0}: OK" -f $IssueNumber)
    $resJson = $res | ConvertTo-Json -Depth 3
    Write-Host $resJson
} catch {
    Write-Warning ("Failed minimal patch #{0}: {1}" -f $IssueNumber, $_.Exception.Message)
}
