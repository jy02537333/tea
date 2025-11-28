$nums = 22..28
foreach ($n in $nums) {
    Write-Host "Assigning #$n"
    try {
        gh issue edit $n --add-assignee jy02537333 | Out-Null
        Write-Host "Assigned #$n"
    } catch {
        Write-Warning ("Failed to assign #{0}: {1}" -f $n, $_.Exception.Message)
    }
}
