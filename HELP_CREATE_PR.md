How to create the PR locally using GitHub CLI (`gh`) or git

Option A — Using `gh` (recommended if installed):

1. Ensure you are on the new branch (created by the script):

```powershell
cd 'e:\project\tea'
git checkout feat/frontend-scaffold
```

2. Create PR with the drafted body:

```powershell
gh pr create --title "feat(frontend): scaffold admin-fe & wx-fe" --body-file PR_DRAFT_FRONTEND.md --base main
```

Option B — Using git + browser (no `gh`):

1. Push branch (if not already pushed):

```powershell
cd 'e:\project\tea'
git push -u origin feat/frontend-scaffold
```

2. Open the GitHub compare URL printed by the script or:

https://github.com/jy02537333/tea/pull/new/feat/frontend-scaffold

Notes:
- Replace `main` with your repository default branch if different.
- `gh` will prompt for login if required and uses your local credentials; it's the easiest way to create a PR from CLI.
