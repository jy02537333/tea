param(
    [int]$Issue = 22
)

$owner = 'jy02537333'
$repo = 'tea'

try { $token = gh auth token 2>$null } catch { $token = $null }
if (-not $token) { Write-Error 'GH token unavailable via "gh auth token"'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json' }

try {
    $iss = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$Issue" -Headers $headers -Method Get
} catch {
    Write-Error "Failed to fetch issue #$Issue: $($_.Exception.Message)"
    exit 1
}

$b = $iss.body
if (-not $b) { Write-Host 'body empty'; exit }
Write-Host "Length: $($b.Length)"

$non = ($b -replace '[\x20-\x7E\r\n\t]','')
Write-Host "Non-printable count: $($non.Length)"
if ($non.Length -gt 0) {
    $chars = $non.ToCharArray() | Select-Object -First 200
    foreach ($c in $chars) {
        $code = [int][char]$c
        Write-Host ("{0} 0x{1} {2}" -f $code, [Convert]::ToString($code,16).ToUpper(), $c)
    }
}

Write-Host '--- Preview (first 800 chars) ---'
Write-Host ($b.Substring(0,[Math]::Min(800,$b.Length)))
