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

## Suggested commit message
```
feat(frontend): scaffold admin-fe & wx-fe examples, services, types, msw mocks, dev scripts

- add services templates and unwrapResponse helper
- add example pages (ProductList, OrderDetail)
- add MSW mock handlers and browser worker
- add GitHub issue & PR templates
```

## Notes
- This PR is intentionally additive and focused on scaffolding; further work includes full UI polish, Storybook, CI integration, and E2E.
