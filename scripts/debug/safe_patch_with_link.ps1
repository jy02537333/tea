param(
    [int]$IssueNumber,
    [string]$DraftUrl
)
if (-not $IssueNumber -or -not $DraftUrl) { Write-Error 'Usage: -IssueNumber <n> -DraftUrl <url>'; exit 1 }
$owner = 'jy02537333'
$repo = 'tea'
try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'No GH token'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept='application/vnd.github+json'; 'User-Agent'='safe-patcher' }

function Sanitize([string]$s) {
    if ($null -eq $s) { return '' }
    $norm = [string]$s
    try {
        $norm = $norm.Normalize([System.Text.NormalizationForm]::FormC)
    } catch { }
    # remove C0 control chars except \r \n \t
    $san = -join ($norm.ToCharArray() | Where-Object { $c = [int]$_; ($c -ge 32) -or ($c -eq 9) -or ($c -eq 10) -or ($c -eq 13) })
    return $san
}

try {
    # Step 1: minimal body to ensure patch acceptance
    $payload1 = @{ body = "[sync] temporary body" } | ConvertTo-Json
    Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Patch -Body $payload1 -ContentType 'application/json'
    Start-Sleep -Seconds 1
    # Step 2: fetch current body
    $issue = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Get
    $current = $issue.body
    $linkLine = "\n\n草稿链接: $DraftUrl"
    $newBody = Sanitize($current) + $linkLine
    $payload2 = @{ body = $newBody } | ConvertTo-Json
    Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Patch -Body $payload2 -ContentType 'application/json'
    Write-Host ("Patched #{0} with draft link" -f $IssueNumber)
} catch {
    Write-Warning ("Failed to safe-patch #{0}: {1}" -f $IssueNumber, $_.Exception.Message)
}
