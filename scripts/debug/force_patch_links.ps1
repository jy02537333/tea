param()

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'
$milestoneTitle = 'M4'

Write-Host "Loading milestone..."
$msJson = gh api repos/$owner/$repo/milestones --paginate 2>$null
try { $msArr = $msJson | ConvertFrom-Json } catch { $msArr = @() }
$ms = $msArr | Where-Object { $_.title -eq $milestoneTitle }
$milestoneNumber = $null
if ($ms) { $milestoneNumber = $ms.number; Write-Host "Milestone $milestoneTitle -> #$milestoneNumber" } else { Write-Warning "Milestone not found" }

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $name = $_.Name
    $content = Get-Content -Raw -Encoding UTF8 $_.FullName
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) { Write-Warning "no frontmatter in $name"; return }
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $title = $matches[1].Trim().Trim('"'); break } }
    if (-not $title) { Write-Warning "no title found in $name"; return }

    $found = @()
    $lines = gh issue list --limit 500 --json number,title --jq '.[] | @base64' 2>$null
    foreach ($ln in $lines) {
        try { $txt = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($ln)); $obj = $txt | ConvertFrom-Json; $found += $obj } catch { }
    }
    $match = $found | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) { Write-Warning "No issue for title: $title"; return }
    $issNum = $match[0].number
    Write-Host "Checking issue #$issNum for $name"
    $issJson = gh api repos/$owner/$repo/issues/$issNum 2>$null
    try { $issObj = $issJson | ConvertFrom-Json } catch { $issObj = $null }
    $body = $issObj.body
    if (-not $body) { $body = "" }
    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    if ($body -match [regex]::Escape($fileUrl)) { Write-Host "#${issNum} already has link"; continue }

    $newBody = $body + "`n`n---`nDraft file: $fileUrl"
    $payload = @{ body = $newBody }
    if ($milestoneNumber) { $payload.milestone = $milestoneNumber }
    $tmp = [System.IO.Path]::GetTempFileName()
    $json = $payload | ConvertTo-Json -Depth 5
    Set-Content -Path $tmp -Value $json -Encoding UTF8
    Write-Host "Patching #$issNum with payload file $tmp"
    gh api repos/$owner/$repo/issues/$issNum -X PATCH --input $tmp | Out-Null
    Remove-Item $tmp -ErrorAction SilentlyContinue
    Write-Host "Patched #$issNum"
}
