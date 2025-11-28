param(
    [int]$IssueNumber
)
if (-not $IssueNumber) { Write-Error 'Provide -IssueNumber'; exit 1 }
$owner = 'jy02537333'
$repo = 'tea'
try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'No GH token'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept='application/vnd.github+json'; 'User-Agent'='print-issue-body' }
try {
    $issue = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Get
    Write-Host "--- ISSUE #$IssueNumber BODY START ---"
    if ($issue.body) { Write-Host $issue.body } else { Write-Host '<empty>' }
    Write-Host "--- ISSUE #$IssueNumber BODY END ---"
} catch {
    Write-Warning ("Failed to fetch issue #{0}: {1}" -f $IssueNumber, $_.Exception.Message)
}
