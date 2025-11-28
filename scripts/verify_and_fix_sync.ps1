param(
    [switch]$DryRun
)

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'
$milestoneTitle = 'M4'

Write-Host "DryRun:" $DryRun

# Ensure milestone exists
Write-Host "Fetching milestones..."
$msJson = gh api repos/$owner/$repo/milestones --paginate 2>$null
try { $msArr = $msJson | ConvertFrom-Json } catch { $msArr = @() }
$ms = $msArr | Where-Object { $_.title -eq $milestoneTitle }
if ($ms) { $milestoneNumber = $ms.number; Write-Host "Found milestone '$milestoneTitle' -> #$milestoneNumber" } else {
    if ($DryRun) { Write-Host "DryRun: would create milestone '$milestoneTitle'"; $milestoneNumber = $null }
    else {
        Write-Host "Creating milestone '$milestoneTitle'..."
        $created = gh api repos/$owner/$repo/milestones -f title="$milestoneTitle" -f description="M4 checklist from tasks/m4_checklist.md" -X POST
        $createdObj = $created | ConvertFrom-Json
        $milestoneNumber = $createdObj.number
        Write-Host "Created milestone -> #$milestoneNumber"
    }
}

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) { Write-Error "Issue drafts dir not found: $issueDir"; exit 1 }

$report = @()

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

    # find issue by exact title
    $found = @()
    $lines = gh issue list --limit 500 --json number,title,body,url --jq '.[] | @base64' 2>$null
    foreach ($ln in $lines) {
        try { $txt = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($ln)); $obj = $txt | ConvertFrom-Json; $found += $obj } catch { }
    }
    $match = $found | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) { Write-Warning "No exact match for title: $title"; $report += [pscustomobject]@{ file=$name; title=$title; issue=$null; url=$null; bodyUpdated=$false; milestoneSet=$false; note='no match' }; return }
    $issue = $match[0]
    $issNum = $issue.number
    $issUrl = $issue.url
    Write-Host "Found issue #$issNum -> $issUrl"

    # fetch issue full info via API to inspect milestone
    $issJson = gh api repos/$owner/$repo/issues/$issNum 2>$null
    try { $issObj = $issJson | ConvertFrom-Json } catch { $issObj = $null }

    $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$name"
    $body = $issObj.body
    if (-not $body) { $body = "" }
    $bodyNeeds = -not ($body -match [regex]::Escape($fileUrl))

    $milestoneSet = $false
    if ($milestoneNumber) {
        if ($issObj.milestone -and $issObj.milestone.number -eq $milestoneNumber) { $milestoneSet = $true } else { $milestoneSet = $false }
    }

    $bodyUpdated = $false
    if ($bodyNeeds -or -not $milestoneSet) {
        $append = "\n\n---\nDraft file: $fileUrl"
        $newBody = $body
        if ($bodyNeeds) { $newBody = $newBody + $append }

        if ($DryRun) {
            Write-Host "DryRun: would patch #$issNum (bodyNeeds=$bodyNeeds, milestoneNeeds=$(-not $milestoneSet))"
            $bodyUpdated = $bodyNeeds
        } else {
            # patch via API
            try {
                if ($milestoneNumber) {
                    gh api repos/$owner/$repo/issues/$issNum -X PATCH -f body="$newBody" -f milestone=$milestoneNumber | Out-Null
                } else {
                    gh api repos/$owner/$repo/issues/$issNum -X PATCH -f body="$newBody" | Out-Null
                }
                Write-Host "Patched #$issNum (body patched if needed, milestone set if available)"
                $bodyUpdated = $bodyNeeds
                $milestoneSet = $true
            } catch {
                Write-Warning ("Failed to patch #{0}: {1}" -f $issNum, $_.Exception.Message)
            }
        }
    } else {
        Write-Host "No changes needed for #$issNum (body already links, milestone set)"
    }

    $report += [pscustomobject]@{ file=$name; title=$title; issue=$issNum; url=$issUrl; bodyUpdated=$bodyUpdated; milestoneSet=$milestoneSet }
}

Write-Host "\nVerification summary:"
$report | ForEach-Object { Write-Host ("#{0}`t{1}`tBodyUpdated:{2}`tMilestoneSet:{3}`t{4}" -f $_.issue, $_.title, $_.bodyUpdated, $_.milestoneSet, $_.url) }
