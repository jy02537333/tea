param(
    [int]$MilestoneNumber = 1
)

$owner = 'jy02537333'
$repo = 'tea'
Write-Host "Listing issues in milestone #$MilestoneNumber"
try {
    $json = gh api repos/$owner/$repo/issues?milestone=$MilestoneNumber --paginate 2>$null
    $arr = $json | ConvertFrom-Json
    foreach ($it in $arr) {
        Write-Host ("#{0}`t{1}`t{2}" -f $it.number, $it.title, $it.html_url)
    }
} catch {
    Write-Warning "Failed to list milestone issues: $($_.Exception.Message)"
}
