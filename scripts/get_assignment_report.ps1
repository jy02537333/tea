$nums = 22..28
foreach ($n in $nums) {
    try {
        $b64 = gh issue view $n --json number,title,assignees,url --jq '. | @base64' 2>$null
        if (-not $b64) { Write-Warning "No data for #$n"; continue }
        $bytes = [System.Convert]::FromBase64String($b64)
        $text = [System.Text.Encoding]::UTF8.GetString($bytes)
        $obj = $text | ConvertFrom-Json
        $assignees = ''
        if ($obj.assignees) { $assignees = ($obj.assignees | ForEach-Object { $_.login }) -join ',' }
        Write-Host ("{0}`t{1}`tAssignees:{2}`t{3}" -f $obj.number, $obj.title, $assignees, $obj.url)
    } catch {
        $e = $_.Exception
        Write-Warning ("Failed to inspect #$n")
        Write-Host $e.Message
    }
}
