foreach ($n in 22..28) {
    Write-Host "=== Issue $n ==="
    try {
        gh issue view $n --json number,title,state,url 2>$null | Out-String | Write-Host
    } catch {
        Write-Warning ("gh failed for issue {0}: {1}" -f $n, $_.Exception.Message)
    }
}
