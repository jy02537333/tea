param(
    [int]$IssueNumber,
    [string]$DraftUrl
)
if (-not $IssueNumber -or -not $DraftUrl) { Write-Error 'Usage: -IssueNumber <n> -DraftUrl <url>'; exit 1 }
$owner = 'jy02537333'
$repo = 'tea'
try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'No GH token'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept='application/vnd.github+json'; 'User-Agent'='safe-patcher-verbose' }

function Sanitize([string]$s) {
    if ($null -eq $s) { return '' }
    $norm = [string]$s
    try { $norm = $norm.Normalize([System.Text.NormalizationForm]::FormC) } catch { }
    $san = -join ($norm.ToCharArray() | Where-Object { $c = [int]$_; ($c -ge 32) -or ($c -eq 9) -or ($c -eq 10) -or ($c -eq 13) })
    return $san
}

Write-Host "Safe verbose patch for #$IssueNumber -> $DraftUrl"
try {
    Write-Host "STEP 1: PATCH minimal body"
    $payload1 = @{ body = "[sync] temporary body" } | ConvertTo-Json
    $res1 = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Patch -Body $payload1 -ContentType 'application/json'
    Write-Host "Response from step1 body:"; Write-Host $res1.body

    Write-Host "STEP 2: GET current issue"
    $issue = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Get
    Write-Host "Current body:"; Write-Host $issue.body

    Write-Host "STEP 3: Build new body"
    $current = $issue.body
    $linkLine = "`n`n草稿链接: $DraftUrl"
    Write-Host "LinkLine (raw):"; Write-Host '>>>START_LINKLINE>>>'
    Write-Host $linkLine
    Write-Host '<<<END_LINKLINE<<<'
    $san = Sanitize($current)
    Write-Host "Sanitized(current) type: $($san.GetType().FullName) length: $($san.Length)"
    Write-Host "LinkLine type: $($linkLine.GetType().FullName) length: $($linkLine.Length)"
    Write-Host "Is linkLine empty? $([string]::IsNullOrEmpty($linkLine))"
    $newBody = $san + $linkLine
    Write-Host "Concat test (san + '###' + linkLine):"; Write-Host ($san + '###' + $linkLine)
    Write-Host "New body preview:"; Write-Host '>>>START_NEWBODY>>>'
    Write-Host $newBody
    Write-Host '<<<END_NEWBODY<<<'

    Write-Host "STEP 4: PATCH new body"
    $payload2 = @{ body = $newBody } | ConvertTo-Json
    $res2 = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$IssueNumber" -Headers $headers -Method Patch -Body $payload2 -ContentType 'application/json'
    Write-Host "Response from step4 body:"; Write-Host $res2.body
    Write-Host "Done.";
} catch {
    Write-Warning ("Verbose safe-patch failed for #{0}: {1}" -f $IssueNumber, $_.Exception.Message)
}
