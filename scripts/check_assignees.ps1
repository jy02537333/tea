$nums = 22..28
foreach ($n in $nums) {
    Write-Host "Issue #$n";
    $o = gh issue view $n --json number,title,assignees,url 2>$null | ConvertFrom-Json
    $assignees = ''
    if ($o.assignees) { $assignees = ($o.assignees | ForEach-Object { $_.login } -join ',') }
    Write-Host ("{0}`t{1}`tAssignees:{2}`t{3}" -f $o.number, $o.title, $assignees, $o.url)
}
