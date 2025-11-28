param()

$report = & (Join-Path $PSScriptRoot 'generate_detailed_report.ps1') | Out-String -Width 4096
$md = @"
# M4 Issue Sync Report
Generated: $(Get-Date -Format u)

```
$report
```
"@

$outPath = Join-Path $PSScriptRoot '..\tasks\m4_issue_sync_report.md'
$md | Out-File -FilePath $outPath -Encoding UTF8
Write-Host "Saved report to $outPath"
