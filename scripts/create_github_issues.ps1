param(
    [switch]$DryRun
)

# Batch create issues from markdown drafts in tasks/issues
# Requires GitHub CLI `gh` and that user is authenticated (`gh auth login`).

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"
if (-not (Test-Path $issueDir)) {
    Write-Error "Issue directory not found: $issueDir"
    exit 1
}

Get-ChildItem -Path $issueDir -Filter *.md | ForEach-Object {
    $file = $_.FullName
    $content = Get-Content -Raw -Path $file -Encoding UTF8

    # parse frontmatter using .NET regex with Singleline mode
    $front = [regex]::Match($content, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n(.*)$')
    if ($front.Success) {
        $meta = $front.Groups[1].Value
        $body = $front.Groups[2].Value

        $title = ""
        $labels = ""
        $assignees = ""

        foreach ($line in ($meta -split "`r?`n")) {
            if ($line -match '^[ \t]*title:[ \t]*(.*)$') {
                $title = $matches[1].Trim()
                $title = $title.Trim('"')
            } elseif ($line -match '^[ \t]*labels:[ \t]*(.*)$') {
                $labels = $matches[1].Trim().Trim('"')
            } elseif ($line -match '^[ \t]*assignees:[ \t]*(.*)$') {
                $assignees = $matches[1].Trim().Trim('"')
            }
        }

        if (-not $title) { $title = "Untitled" }

        Write-Host "Issue file: $file`nTitle: $title`nLabels: $labels`nAssignees: $assignees`n"

        if ($DryRun) { continue }

        $args = @("issue","create","--title", $title, "--body-file", $file)
        if ($labels) {
            $labels.Split(',') | ForEach-Object { $args += @("--label", $_.Trim()) }
        }
        if ($assignees -and $assignees -ne "") {
            $assignees.Split(',') | ForEach-Object { $args += @("--assignee", $_.Trim()) }
        }

        $result = gh @args
        Write-Host $result
    } else {
        Write-Warning "Skipping file without frontmatter: $file"
    }
}
