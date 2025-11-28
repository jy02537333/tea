param()

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'GH token unavailable via "gh auth token"'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='find-missing-links' }

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) { Write-Error "Issue drafts dir not found: $issueDir"; exit 1 }

$allIssues = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues?per_page=200&state=all" -Headers $headers -Method Get

Write-Host "Missing-body-link issues (number - title):"

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
    if (-not $match -or $match.Count -eq 0) { return }
    $iss = $match[0]
    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    $body = $iss.body
    if (-not $body) { $body = "" }
    $hasBodyLink = ($body -match [regex]::Escape($fileUrl))
    if (-not $hasBodyLink) {
        Write-Host ("{0} - {1}" -f $iss.number, $iss.title)
    }
}
