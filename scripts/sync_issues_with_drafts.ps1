param(
    [string]$MilestoneName = 'M4',
    [switch]$DryRun
)

$repoOwner = 'jy02537333'
$repoName = 'tea'
$branch = 'feat/frontend-scaffold'

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) { Write-Error "Issue drafts dir not found: $issueDir"; exit 1 }

Write-Host "Milestone:" $MilestoneName
Write-Host "DryRun:" $DryRun

# Ensure milestone exists (create if needed) and get milestone number via GH REST API
Write-Host "Ensuring milestone exists via gh api..."
$milestonesJson = gh api repos/$repoOwner/$repoName/milestones --paginate 2>$null
try {
    $milestones = $milestonesJson | ConvertFrom-Json
} catch {
    $milestones = @()
}
$ms = $milestones | Where-Object { $_.title -eq $MilestoneName }
if ($ms) {
    $milestoneNumber = $ms.number
    Write-Host "Found milestone '$MilestoneName' -> number $milestoneNumber"
} else {
    if ($DryRun) {
        Write-Host "DryRun: would create milestone '$MilestoneName'"
        $milestoneNumber = $null
    } else {
        Write-Host "Creating milestone '$MilestoneName' via API..."
        $create = gh api repos/$repoOwner/$repoName/milestones -f title="$MilestoneName" -f description="M4 checklist from tasks/m4_checklist.md" -X POST
        $created = $create | ConvertFrom-Json
        $milestoneNumber = $created.number
        Write-Host "Created milestone '$MilestoneName' -> number $milestoneNumber"
    }
}

$report = @()

Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $file = $_.FullName
    $name = $_.Name
    $content = Get-Content -Raw -Encoding UTF8 $file
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) { Write-Warning "no frontmatter in $name"; return }
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $title = $matches[1].Trim().Trim('"'); break } }
    if (-not $title) { Write-Warning "no title for $name"; return }

    # Find matching issue by exact title
    $found = @()
    $lines = gh issue list --limit 500 --json number,title,body,url --jq '.[] | @base64' 2>$null
    foreach ($ln in $lines) {
        try { $obj = ([System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($ln))) | ConvertFrom-Json; $found += $obj } catch { }
    }
    $match = $found | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}
    if (-not $match -or $match.Count -eq 0) { Write-Warning "No exact match for title: $title"; return }
    $issue = $match[0]

    $fileUrl = "https://github.com/$repoOwner/$repoName/blob/$branch/tasks/issues/$name"

    # Prepare new body: append draft link if not present
    $body = $issue.body
    if (-not $body) { $body = "" }
    if ($body -match [regex]::Escape($fileUrl)) {
        Write-Host "Issue #$($issue.number) already links to draft. Skipping body update."
        $bodyUpdated = $false
    } else {
        $append = "\n\n---\nDraft file: $fileUrl"
        $newBody = $body + $append
        if ($DryRun) { Write-Host "DryRun: would update body for #$($issue.number)"; $bodyUpdated = $true } else {
            gh issue edit $issue.number --body "$newBody" | Out-Null
            Write-Host "Updated body for #$($issue.number)"
            $bodyUpdated = $true
        }
    }

    # Set milestone (use milestone number via REST API if available)
    if (-not $milestoneNumber) {
        Write-Warning "No milestone number available for '$MilestoneName'; skipping milestone set for #$($issue.number)"
    } else {
        if ($DryRun) {
            Write-Host "DryRun: would set milestone #$milestoneNumber on #$($issue.number)"
        } else {
            # Use GH REST API to update issue body + milestone safely
            try {
                gh api repos/$repoOwner/$repoName/issues/$($issue.number) -X PATCH -f body="$newBody" -f milestone=$milestoneNumber | Out-Null
                Write-Host "Patched body+milestone for #$($issue.number) via API"
            } catch {
                Write-Warning "failed to update $($issue.url): $($_.Exception.Message)"
            }
        }
    }
    $report += [pscustomobject]@{
        file = $name
        title = $title
        issue = $issue.number
        url = $issue.url
        bodyUpdated = $bodyUpdated
    }
}

Write-Host "\nSummary:"
$report | ForEach-Object { Write-Host "#$_.[issue]`t$_.[title]`tBodyUpdated:$($_.bodyUpdated)`t$($_.url)" }
