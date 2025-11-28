$owner = 'jy02537333'
$repo = 'tea'
$token = gh auth token
if (-not $token) { Write-Error 'No GH token'; exit 1 }
$headers = @{ Authorization = "token $token"; Accept = 'application/vnd.github+json'; 'User-Agent'='patch-script' }

foreach ($n in 22,26) {
    Write-Host "Patching milestone-only for #$n"
    $p = @{ milestone = 1 } | ConvertTo-Json
    try {
        $res = Invoke-RestMethod -Uri "https://api.github.com/repos/$owner/$repo/issues/$n" -Headers $headers -Method Patch -Body $p -ContentType 'application/json'
        Write-Host "Success #$n -> milestone: $($res.milestone.number)"
    } catch {
        Write-Warning ("Failed #{0}: {1}" -f $n, $_.Exception.Message)
    }
}
