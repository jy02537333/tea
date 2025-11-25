Commit message
----------------

feat(frontend): scaffold admin-fe & wx-fe examples, services, msw mocks, dev scripts


PR body (use as PR description)
--------------------------------

This PR adds initial frontend scaffolding and example pages for `admin-fe` and `wx-fe`:

- TypeScript `services/*.ts` templates with `ApiResponse`/`unwrapResponse` helper.
- Example pages: ProductList and OrderDetail (admin & wx).
- Axios auth helpers and 401 handling in `services/api.ts`.
- Added minimal UI component stubs: `DataTable`, `ProductCard`.
- MSW mock handlers and instructions for local dev mocking.
- Dev scripts for local development (`vite` for admin-fe, `taro` scripts for wx-fe).
- GitHub Issue templates and PR template for repo.

How to run locally
-------------------
1. Install dependencies:

```powershell
cd 'e:\project\tea\admin-fe'
npm.cmd install

cd 'e:\project\tea\wx-fe'
npm.cmd install --legacy-peer-deps
```

2. Start dev servers:

```powershell
cd 'e:\project\tea\admin-fe'
npm.cmd run dev

cd 'e:\project\tea\wx-fe'
npm.cmd run dev:weapp
```

Optional: enable MSW for H5/browser development by calling `worker.start()` in your dev entry.
