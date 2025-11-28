# PR Draft: Frontend examples, services, types, and dev setup

## Summary
This PR adds initial frontend scaffolding and example pages for `admin-fe` and `wx-fe`:

- TypeScript `services/*.ts` templates with `ApiResponse`/`unwrapResponse` helper.
- Example pages: ProductList and OrderDetail (admin & wx).
- Axios auth helpers and 401 handling in `services/api.ts`.
- Added minimal UI component stubs: `DataTable`, `ProductCard`.
- MSW mock handlers and instructions for local dev mocking.
- Dev scripts for local development (`vite` for admin-fe, `taro` scripts for wx-fe).
- GitHub Issue templates and PR template for repo.

## Files changed (high level)
- `admin-fe/src/services/*` (types & services)
- `wx-fe/src/services/*` (types & services)
- `.github/ISSUE_TEMPLATE/*`, `.github/PULL_REQUEST_TEMPLATE.md`
- `admin-fe/src/mocks/*`, `wx-fe/src/mocks/*`
- `admin-fe/src/components/*`, `wx-fe/src/components/*`
- `admin-fe/package.json`, `wx-fe/package.json`

## How to run & test locally
1. Install dependencies in each subproject:

```powershell
cd 'e:\project\tea\admin-fe'
npm.cmd install
cd '..\wx-fe'
npm.cmd install --legacy-peer-deps
```

2. Start dev server (admin):
```powershell
cd 'e:\project\tea\admin-fe'
npm.cmd run dev
```

3. Start Taro dev (wx):
```powershell
cd 'e:\project\tea\wx-fe'
npm.cmd run dev:weapp
```

4. Optional: enable MSW for browser/H5 mocking

```ts
// In your dev entry (e.g. src/main.tsx) add:
import { worker } from './mocks/browser';
worker.start();
```
## Testing checklist
- [ ] `npm install` in both `admin-fe` and `wx-fe` completes without error
- [ ] `npm run typecheck` in both projects returns no TypeScript errors
- [ ] `admin-fe` dev server starts and ProductList/OrderDetail pages render
- [ ] `wx-fe` H5 dev renders ProductList; Taro weapp dev starts when configured
- [ ] MSW mock responds to `/api/v1/products` in H5/dev mode

## Suggested commit message
```
feat(frontend): scaffold admin-fe & wx-fe examples, services, msw mocks, dev scripts

- add services templates and unwrapResponse helper
- add example pages (ProductList, OrderDetail)
- add MSW mock handlers and browser worker
- add GitHub issue & PR templates
```

## Files changed (detailed)
- `.github/ISSUE_TEMPLATE/*` — added templates for feature/bug/task
- `.github/PULL_REQUEST_TEMPLATE.md` — PR template
- `admin-fe/src/services/*`, `wx-fe/src/services/*` — API service templates and types
- `admin-fe/src/pages/*`, `wx-fe/src/pages/*` — example ProductList / OrderDetail pages
- `admin-fe/src/components/*`, `wx-fe/src/components/*` — minimal UI stubs
- `admin-fe/src/mocks/*`, `wx-fe/src/mocks/*` — MSW handlers and browser worker
- `scripts/create_frontend_commit.ps1` — helper to create branch and push
- `PR_DRAFT_FRONTEND.md`, `COMMIT_PR_FRONTEND.md` — PR and commit guidance

If you want I can also attach a short test matrix (browser list / node versions) to the PR body. Tell me which Node.js version you plan to use and whether Windows dev-only notes should be included.
```

## Notes
- This PR is intentionally additive and focused on scaffolding; further work includes full UI polish, Storybook, CI integration, and E2E.
