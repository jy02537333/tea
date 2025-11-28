param(
    [int[]]$Issues = @(22,26),
    [switch]$DryRun
)

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'
$milestoneNumber = 1

Write-Host "DryRun:" $DryRun

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'GH token unavailable via "gh auth token"'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='sanitize-patch-script' }

foreach ($issNum in $Issues) {
    Write-Host "\nProcessing issue #$issNum"
    try {
        $iss = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$issNum" -Headers $headers -Method Get
    } catch {
        Write-Warning ("Failed to fetch #{0}: {1}" -f $issNum, $_.Exception.Message)
        continue
    }

    $body = $iss.body
    if (-not $body) { $body = "" }

    # sanitize control chars except CR/LF/Tab
    $san = $body -replace "[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]", ''
    # normalize line endings to LF
    $san = $san -replace "\r\n", "\n"
    $san = $san -replace "\r", "\n"

    $name = ($iss.title -replace '[^0-9a-zA-Z一-龥一二三四五六七八九十\s\-_,]','')
    # attempt to find corresponding draft filename by searching tasks dir for title (best-effort)
    $issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
    $draftFile = Get-ChildItem -Path $issueDir -Filter *.md | Where-Object { (Get-Content -Raw -Encoding UTF8 $_.FullName) -match [regex]::Escape($iss.title) } | Select-Object -First 1
    if ($draftFile) { $fileName = $draftFile.Name } else { $fileName = "$($issNum)_unknown.md" }
    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$fileName"

    $hasLink = ($san -match [regex]::Escape($fileUrl)) -or ($body -match [regex]::Escape($fileUrl))
    if ($hasLink) { Write-Host "#${issNum} already contains draft link"; continue }

    $newBody = $san + "`n`n---`nDraft file: $fileUrl"

    if ($DryRun) { Write-Host "DryRun: would PATCH #$issNum (sanitized body length: $($newBody.Length))"; continue }

    $payload = @{ body = $newBody; milestone = $milestoneNumber }
    try {
        $json = $payload | ConvertTo-Json -Depth 10
        Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$issNum" -Headers $headers -Method Patch -Body $json -ContentType 'application/json'
        Write-Host "Patched #$issNum successfully"
    } catch {
        Write-Warning ("Failed to PATCH #{0}: {1}" -f $issNum, $_.Exception.Message)
    }
}

Write-Host "\nDone."
