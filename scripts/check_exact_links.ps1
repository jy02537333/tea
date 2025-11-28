param()

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) { Write-Error "Issue drafts dir not found: $issueDir"; exit 1 }

Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $name = $_.Name
    $content = Get-Content -Raw -Encoding UTF8 $_.FullName
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) { Write-Warning "no frontmatter in $name"; return }
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $title = $matches[1].Trim().Trim('"'); break } }
    if (-not $title) { Write-Warning "no title found in $name"; return }

    # find issue by exact title
    $found = @()
    $lines = gh issue list --limit 500 --json number,title --jq '.[] | @base64' 2>$null
    foreach ($ln in $lines) {
        try { $txt = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($ln)); $obj = $txt | ConvertFrom-Json; $found += $obj } catch { }
    }
    $match = $found | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) { Write-Host "[NO MATCH] $name -> title: $title"; return }
    $issue = $match[0]
    $issNum = $issue.number
    $issUrl = "https://github.com/$owner/$repo/issues/$issNum"

    $issJson = gh api repos/$owner/$repo/issues/$issNum 2>$null | Out-String
    try { $issObj = $issJson | ConvertFrom-Json } catch { $issObj = $null }
    $body = $issObj.body
    if (-not $body) { $body = "" }
    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    $has = $body -match [regex]::Escape($fileUrl)

    Write-Host "#${issNum}`t$name`tHasLink:${has}`tMilestone:${([bool]$issObj.milestone)}
Excerpt:`n$([string]$body).Substring(0,[Math]::Min(200, ($body.Length)))`n"
}
