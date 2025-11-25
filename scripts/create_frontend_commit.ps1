# Create branch, commit frontend scaffold changes and (optionally) open a PR
# Usage: run in PowerShell from any location; script will cd into repo root path below.

$repoRoot = 'e:\project\tea'
Set-Location $repoRoot

$branch = 'feat/frontend-scaffold'
$commitMsg = 'feat(frontend): scaffold admin-fe & wx-fe examples, services, msw mocks, dev scripts'

Write-Host "Creating branch $branch and committing changes..."
git checkout -b $branch

# Stage the main frontend changes; use a safe all-files add for PowerShell
git add -A

git commit -m "$commitMsg"

Write-Host 'Pushing branch to origin...'
git push -u origin $branch

# If GitHub CLI is available, create a PR directly
if (Get-Command gh -ErrorAction SilentlyContinue) {
  gh pr create --title "$commitMsg" --body-file PR_DRAFT_FRONTEND.md --base main
} else {
  Write-Host "GitHub CLI not found. Open the URL to create a PR manually:"
  Write-Host "https://github.com/jy02537333/tea/compare/$branch?expand=1"
}

Write-Host 'Done.'
