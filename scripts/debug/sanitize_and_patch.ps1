param(
    [int[]]$IssueNumbers
)

if (-not $IssueNumbers) { Write-Error "Provide -IssueNumbers <22,26>"; exit 1 }

$owner = 'jy02537333'
$repo = 'tea'
$branch = 'feat/frontend-scaffold'

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'GH token unavailable via "gh auth token"'; exit 1 }

$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='sanitize-patch-script' }

$issueDir = Join-Path $PSScriptRoot "..\tasks\issues"

foreach ($num in $IssueNumbers) {
    Write-Host "\nProcessing issue #$num"
    try {
        $iss = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$num" -Headers $headers -Method Get
    } catch {
        $errMsg = $_.Exception.Message
        Write-Warning ("Failed to fetch #{0}: {1}" -f $num, $errMsg)
        continue
    }

    $body = $iss.body
    if (-not $body) { $body = "" }

    # sanitize: normalize Unicode and remove control characters except CR/LF/TAB
    try {
        $norm = [System.Text.NormalizationForm]::FormC
        if ([string]::IsNullOrEmpty($body)) { $bodyNorm = "" } else { $bodyNorm = [System.Text.Normalization]::Normalize($body, $norm) }
    } catch { $bodyNorm = $body }

    # remove disallowed control chars
    $san = $bodyNorm -replace '[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]', ''

    # find matching draft file by title
    $title = $iss.title
    $fileMatch = Get-ChildItem -Path $issueDir -Filter *.md | Where-Object {
        $c = Get-Content -Raw -Encoding UTF8 $_.FullName
        $fm = [regex]::Match($c, '(?s)^---\s*\r?\n(.*?)\r?\n---\s*\r?\n')
        if (-not $fm.Success) { return $false }
        $meta = $fm.Groups[1].Value
        foreach ($line in ($meta -split "`r?`n")) { if ($line -match '^[ \t]*title:[ \t]*(.*)$') { $t = $matches[1].Trim().Trim('"'); return $t -eq $title } }
        return $false
    }

    $fileUrl = $null
    if ($fileMatch) { $fileUrl = "https://github.com/$owner/$repo/blob/$branch/tasks/issues/$($fileMatch.Name)" }

    if ($fileUrl -and ($san -notmatch [regex]::Escape($fileUrl))) {
        $san = $san + "`n`n---`nDraft file: $fileUrl"
        $appendNote = $true
    } else { $appendNote = $false }

    # only patch if body changed or need to add link
    if ($san -ne $body -or $appendNote) {
        $payload = @{ body = $san }
        try {
            $json = $payload | ConvertTo-Json -Depth 10
            Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$num" -Headers $headers -Method Patch -Body $json -ContentType 'application/json'
            Write-Host "Patched #$num successfully"
        } catch {
            $errMsg2 = $_.Exception.Message
            Write-Warning ("Failed to patch #{0}: {1}" -f $num, $errMsg2)
        }
    } else {
        Write-Host "No changes needed for #$num"
    }
}

Write-Host "\nSanitize+patch run complete."
