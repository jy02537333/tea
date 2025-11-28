$nums = 22..28
foreach ($n in $nums) {
    $lines = gh issue view $n --json title --jq '.title' 2>$null
    if (-not $lines) { Write-Warning "no title for $n"; continue }
    # gh --jq '.title' prints a JSON string; remove quotes if present
    $raw = $lines.Trim()
    if ($raw.StartsWith('"') -and $raw.EndsWith('"')) { $raw = $raw.Trim('"') }
    Write-Host "=== $n original (as returned) ==="
    Write-Host $raw
    $utf8bytes = [System.Text.Encoding]::UTF8.GetBytes($raw)
    $gbk = [System.Text.Encoding]::GetEncoding(936).GetString($utf8bytes)
    Write-Host "=== $n re-decoded as GBK (codepage 936) ==="
    Write-Host $gbk
}
