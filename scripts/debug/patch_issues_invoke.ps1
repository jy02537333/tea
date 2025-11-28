param(
    [switch]$DryRun
)

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'
$milestoneTitle = 'M4'

Write-Host "DryRun:" $DryRun

# get token from gh
try {
    $token = gh auth token 2>$null
} catch {
    Write-Error "Failed to get GH token via 'gh auth token'"
    exit 1
}
if (-not $token) { Write-Error "No GH token available"; exit 1 }

$headers = @{
    Authorization = "token $token"
    Accept = 'application/vnd.github+json'
    'User-Agent' = 'patch-script'
}

Write-Host "Fetching milestone list..."
$msArr = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/milestones?per_page=100" -Headers $headers -Method Get
$ms = $msArr | Where-Object { $_.title -eq $milestoneTitle }
if ($ms) { $milestoneNumber = $ms.number; Write-Host "Found milestone '$milestoneTitle' -> #$milestoneNumber" } else {
    if ($DryRun) { Write-Host "DryRun: would create milestone '$milestoneTitle'"; $milestoneNumber = $null }
    else {
        Write-Host "Creating milestone '$milestoneTitle' via API..."
        $body = @{ title = $milestoneTitle; description = 'M4 checklist from tasks/m4_checklist.md' } | ConvertTo-Json
        $created = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/milestones" -Headers $headers -Method Post -Body $body -ContentType 'application/json'
        $milestoneNumber = $created.number
        Write-Host "Created milestone -> #$milestoneNumber"
    }
}

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $name = $_.Name
    Write-Host "\nProcessing draft: $name"
    $content = Get-Content -Raw -Encoding UTF8 $_.FullName
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) { Write-Warning "no frontmatter in $name"; return }
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $title = $matches[1].Trim().Trim('"'); break } }
    if (-not $title) { Write-Warning "no title found in $name"; return }

    # find issue
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
    $bodyNeeds = -not ($body -match [regex]::Escape($fileUrl))
    $milestoneSet = $false
    if ($milestoneNumber) { if ($iss.milestone -and $iss.milestone.number -eq $milestoneNumber) { $milestoneSet = $true } }

    if ($bodyNeeds -or -not $milestoneSet) {
        $newBody = $body
        if ($bodyNeeds) { $newBody = $newBody + "`n`n---`nDraft file: $fileUrl" }

        $payload = @{ body = $newBody }
        if ($milestoneNumber) { $payload.milestone = $milestoneNumber }

        if ($DryRun) {
            Write-Host "DryRun: would PATCH #$issNum (bodyNeeds=$bodyNeeds, milestoneNeeds=$(-not $milestoneSet))"
        } else {
            try {
                $json = $payload | ConvertTo-Json -Depth 10
                Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$issNum" -Headers $headers -Method Patch -Body $json -ContentType 'application/json'
                Write-Host "Patched #$issNum"
            } catch {
                Write-Warning ("Failed to patch #{0}: {1}" -f $issNum, $_.Exception.Message)
            }
        }
    } else {
        Write-Host "No changes needed for #$issNum"
    }
}

Write-Host "\nDone."
