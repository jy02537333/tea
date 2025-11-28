param()

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'GH token unavailable via "gh auth token"'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='m4-report-simple' }

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) { Write-Error "Issue drafts dir not found: $issueDir"; exit 1 }

$rows = @()

$allIssues = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues?per_page=300&state=all" -Headers $headers -Method Get

Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $name = $_.Name
    $path = $_.FullName
    $content = Get-Content -Raw -Encoding UTF8 $path
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) { return }
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $title = $matches[1].Trim().Trim('"'); break } }
    if (-not $title) { return }

    $match = $allIssues | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) {
        $rows += [pscustomobject]@{ number='(none)'; title=$title; assignees=''; has_body_link=$false; comment_posted=$false; milestone=''; url=''; draft="https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name" }
        return
    }
    $iss = $match[0]
    $issNum = $iss.number
    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    $body = $iss.body
    if (-not $body) { $body = "" }
    $hasBodyLink = ($body -match [regex]::Escape($fileUrl))
    $comments = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$issNum/comments?per_page=200" -Headers $headers -Method Get
    $foundComment = $false
    foreach ($c in $comments) { if ($c.body -and ($c.body -match [regex]::Escape($fileUrl) -or $c.body -match 'Automated note: added by scripts/post_comments_on_missing_links.ps1')) { $foundComment = $true; break } }
    $assignees = ''
    if ($iss.assignees) { $assignees = ($iss.assignees | ForEach-Object { $_.login }) -join ',' }
    $milestoneTitle = ''
    if ($iss.milestone) { $milestoneTitle = $iss.milestone.title }
    $rows += [pscustomobject]@{ number=$issNum; title=$title; assignees=$assignees; has_body_link=$hasBodyLink; comment_posted=$foundComment; milestone=$milestoneTitle; url=$iss.html_url; draft=$fileUrl }
}

$out = "# M4 Issue Sync Report`nGenerated: $(Get-Date -Format u)`n`n"
$out += "|Number|Title|Assignees|HasBodyLink|CommentPosted|Milestone|IssueURL|DraftURL|`n"
$out += "|---|---|---|---:|---:|---|---|---|`n"
foreach ($r in $rows) { $out += "|$($r.number)|$([regex]::Escape($r.title))|$($r.assignees)|$($r.has_body_link)|$($r.comment_posted)|$($r.milestone)|$($r.url)|$($r.draft)`n" }

$outPath = Join-Path $PSScriptRoot "..\tasks\m4_issue_sync_report.md"
$out | Out-File -FilePath $outPath -Encoding UTF8
Write-Host "Saved report to $outPath"
