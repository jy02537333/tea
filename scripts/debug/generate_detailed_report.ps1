param()

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'GH token unavailable via "gh auth token"'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='detailed-report-script' }

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) { Write-Error "Issue drafts dir not found: $issueDir"; exit 1 }

$report = @()

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

    # find issue by title via API
    $allIssues = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues?per_page=200&state=all" -Headers $headers -Method Get
    $match = $allIssues | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) {
        $report += [pscustomobject]@{ number = $null; title = $title; assignees = ''; has_body_link = $false; comment_posted = $false; milestone = $null; url = $null }
        return
    }
    $iss = $match[0]
    $issNum = $iss.number
    $issUrl = $iss.html_url

    # check body for exact draft link
    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    $body = $iss.body
    if (-not $body) { $body = "" }
    $hasBodyLink = ($body -match [regex]::Escape($fileUrl))

    # fetch comments and check whether a comment contains the draft link or the automated note marker
    $comments = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$issNum/comments?per_page=200" -Headers $headers -Method Get
    $foundComment = $false
    foreach ($c in $comments) {
        if ($c.body -and ($c.body -match [regex]::Escape($fileUrl) -or $c.body -match 'Automated note: added by scripts/post_comments_on_missing_links.ps1')) { $foundComment = $true; break }
    }

    $assignees = ''
    if ($iss.assignees) { $assignees = ($iss.assignees | ForEach-Object { $_.login }) -join ',' }
    $milestoneTitle = $null
    if ($iss.milestone) { $milestoneTitle = $iss.milestone.title }

    $report += [pscustomobject]@{ number = $issNum; title = $title; assignees = $assignees; has_body_link = $hasBodyLink; comment_posted = $foundComment; milestone = $milestoneTitle; url = $issUrl }
}

Write-Host "IssueNumber`tTitle`tAssignees`tHasBodyLink`tCommentPosted`tMilestone`tURL"
$report | ForEach-Object { Write-Host ("{0}`t{1}`t{2}`t{3}`t{4}`t{5}`t{6}" -f $_.number, $_.title, $_.assignees, ($_.has_body_link -as [string]), ($_.comment_posted -as [string]), $_.milestone, $_.url) }
