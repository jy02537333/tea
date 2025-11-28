param(
    [switch]$DryRun
)

# Close later-created duplicate issues whose titles match drafts in tasks/issues.
# Keeps the earliest-created (lowest number) issue open and closes later ones with
# a comment linking to the primary issue. Use -DryRun to preview actions.

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Definition
$issueDir = Join-Path $scriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) {
    Write-Error "Issue drafts directory not found: $issueDir"
    exit 1
}

Write-Host "DryRun:" $DryRun

# Read draft titles
$titles = @{}
Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $content = Get-Content -Raw -Encoding UTF8 $_.FullName
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if ($front.Success) {
        $meta = $front.Groups[1].Value
        foreach ($line in ($meta -split "`r?`n")) {
            if ($line -match '^[ \t]*title:[ \t]*(.*)$') {
                $title = $matches[1].Trim().Trim('"')
                if ($title) { $titles[$title] = $true }
                break
            }
        }
    }
}

if (-not $titles.Keys) {
    Write-Warning "No draft titles found in $issueDir"
    exit 0
}

Write-Host "Found" ($titles.Keys.Count) "draft titles to check for duplicates."

# Get all issues as base64 lines and decode safely
$items = @()
try {
    $lines = gh issue list --limit 500 --json number,title,url,createdAt,state --jq '.[] | @base64' 2>$null
} catch {
    Write-Error "Failed to run 'gh issue list': $($_.Exception.Message)"
    exit 1
}

foreach ($ln in $lines) {
    try {
        $bytes = [System.Convert]::FromBase64String($ln)
        $text = [System.Text.Encoding]::UTF8.GetString($bytes)
        $obj = $text | ConvertFrom-Json -ErrorAction Stop
        $items += $obj
    } catch {
        Write-Warning "Skipping malformed GH item: $($_.Exception.Message)"
    }
}

# Group by title for only the titles we care about
$groups = $items | Where-Object { $titles.ContainsKey($_.title) } | Group-Object -Property title

$actions = @()
foreach ($g in $groups) {
    if ($g.Count -le 1) { continue }
    $sorted = $g.Group | Sort-Object {[int]$_.number}
    $primary = $sorted[0]
    $dups = $sorted | Select-Object -Skip 1
    foreach ($d in $dups) {
        if ($d.state -eq 'closed') { continue }
        $actions += [pscustomobject]@{
            title = $g.Name
            primary_number = $primary.number
            primary_url = $primary.url
            dup_number = $d.number
            dup_url = $d.url
        }
    }
}

if (-not $actions) {
    Write-Host "No open duplicates found for draft titles."
    exit 0
}

Write-Host "Planned actions:" ($actions.Count) "duplicate(s) to close."
foreach ($a in $actions) {
    if ($DryRun) {
        Write-Host "DryRun: would close issue #$($a.dup_number) ($($a.title)) -> link to #$($a.primary_number)"
    } else {
        Write-Host "Closing #$($a.dup_number) and adding comment linking to #$($a.primary_number)"
        $body = "Closing duplicate of #$($a.primary_number). See primary: $($a.primary_url)"
        try {
            gh issue comment $a.dup_number --body "$body" | Out-Null
            gh issue close $a.dup_number | Out-Null
            Write-Host "Closed #$($a.dup_number) -> $($a.dup_url)"
        } catch {
            Write-Warning "Failed to close #$($a.dup_number): $($_.Exception.Message)"
        }
    }
}

Write-Host "Done."
