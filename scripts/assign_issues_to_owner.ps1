param(
    [string]$Owner = 'jy02537333',
    [switch]$DryRun
)

# This script reads issue draft files from tasks/issues, extracts the `title` from each
# file's frontmatter, then finds the earliest-created issue with that exact title
# and assigns it to $Owner. This avoids embedding non-ASCII literals in the script.

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) {
    Write-Error "Issue drafts directory not found: $issueDir"
    exit 1
}

Write-Host "Owner:" $Owner
Write-Host "DryRun:" $DryRun

Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $file = $_.FullName
    Write-Host "\nProcessing draft file: $($_.Name)"
    $content = Get-Content -Raw -Encoding UTF8 $file
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
    if (-not $front.Success) {
        Write-Warning "No frontmatter found in $file"
        return
    }

    # Extract title from frontmatter
    $meta = $front.Groups[1].Value
    $title = ''
    foreach ($line in ($meta -split "`r?`n")) {
        if ($line -match '^[ \t]*title:[ \t]*(.*)$') {
            $title = $matches[1].Trim().Trim('"')
            break
        }
    }

    if (-not $title) {
        Write-Warning "No title in frontmatter for $file"
        return
    }

    Write-Host "Title: $title"

    # Query issues and parse JSON safely (write gh output to a UTF-8 temp file first)
    # Use gh's --jq to output one base64-encoded JSON object per line to avoid
    # problems with embedded newlines or encoding in titles. Decode each line
    # and convert to objects safely.
    $jsonList = @()
    try {
        $lines = gh issue list --limit 200 --json number,title,url,createdAt --jq '.[] | @base64' 2>$null
    } catch {
        Write-Warning "Failed to run gh issue list: $($_.Exception.Message)"
        return
    }

    if ($lines) {
        foreach ($ln in $lines) {
            try {
                $bytes = [System.Convert]::FromBase64String($ln)
                $text = [System.Text.Encoding]::UTF8.GetString($bytes)
                $obj = $text | ConvertFrom-Json -ErrorAction Stop
                $jsonList += $obj
            } catch {
                Write-Warning "Failed to decode/parse one GH item: $($_.Exception.Message)"
            }
        }
    }

    $found = $jsonList | Where-Object { $_.title -eq $title } | Sort-Object {[int]$_.number}

    if ($found -and $found.Count -gt 0) {
        $issue = $found[0]
        Write-Host "Found issue: $($issue.number) $($issue.title) $($issue.url)"
        if ($DryRun) {
            Write-Host "DryRun: would assign $($issue.number) -> $Owner"
        } else {
            try {
                gh issue edit $issue.number --add-assignee $Owner | Out-Null
                Write-Host "Assigned: https://github.com/jy02537333/tea/issues/$($issue.number) -> $Owner"
            } catch {
                Write-Warning "Failed to assign issue $($issue.number): $($_.Exception.Message)"
            }
        }
    } else {
        Write-Warning "No exact match found for title: $title"
    }
}
