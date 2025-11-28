param(
    [switch]$DryRun
)

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'

Write-Host "DryRun:" $DryRun

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error "GH token unavailable via 'gh auth token'"; exit 1 }

$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='post-comments-script' }

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $name = $_.Name
    Write-Host "\nChecking draft: $name"
    $content = Get-Content -Raw -Encoding UTF8 $_.FullName
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) { Write-Warning "no frontmatter in $name"; return }
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $title = $matches[1].Trim().Trim('"'); break } }
    if (-not $title) { Write-Warning "no title found in $name"; return }

    # find issue by exact title
    $list = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues?per_page=200&state=all" -Headers $headers -Method Get
    $match = $list | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) { Write-Warning "No exact match for title: $title"; return }
    $iss = $match[0]
    $issNum = $iss.number
    $issUrl = $iss.html_url
    Write-Host "Found issue #$issNum -> $issUrl"

    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    $body = $iss.body
    if (-not $body) { $body = "" }
    $has = $body -match [regex]::Escape($fileUrl)
    if ($has) { Write-Host "Issue #$issNum already contains link"; return }

    $commentBody = "Draft file: $fileUrl`n\n(Automated note: added by scripts/post_comments_on_missing_links.ps1)"

    if ($DryRun) { Write-Host "DryRun: would post comment to #$issNum -> $commentBody"; return }

    try {
        $payload = @{ body = $commentBody } | ConvertTo-Json
        Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$issNum/comments" -Headers $headers -Method Post -Body $payload -ContentType 'application/json'
        Write-Host "Posted comment to #$issNum"
    } catch {
        Write-Warning ("Failed to post comment to #{0}: {1}" -f $issNum, $_.Exception.Message)
    }
}

Write-Host "\nDone posting comments."
