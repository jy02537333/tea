# Browser Login Tests — CI Integration

This folder contains Playwright-based smoke tests for the Admin-FE login flow.

Files:
- `login-test-pure-ui.js` — pure UI test (navigates to `login.html`, fills captcha shown on page or fetches it via page, submits via UI, and validates store→order flow). Writes a JSON report.
- `login-test-inject.js` — injection test (uses API to get token, injects into localStorage, verifies UI flow). Writes a JSON report.
- `login-test-ci.ts` — CI-friendly TypeScript script that logs in via API (Playwright request), injects token, validates flow, and writes a report to `reports/`.

Quick local run

1. Install dependencies and Playwright browsers:

```powershell
cd tools\browser-login
npm install
npx playwright install
```

2. Run CI script (headless) — writes report into `tools/browser-login/reports/`:

```powershell
#$env:ADMIN_FE_URL='http://localhost:8000'; $env:API_BASE='http://localhost:9292/api/v1'; $env:HEADLESS='1';
#npx ts-node login-test-ci.ts
npx ts-node login-test-ci.ts
```

3. Run pure UI locally (for debugging; may be flaky in headless):

```powershell
$env:ADMIN_FE_URL='http://localhost:8000'; $env:API_BASE='http://localhost:9292/api/v1'; $env:HEADLESS='1'; \
$env:REPORT_DIR='D:\\developTool\\work\\go\\tea\\tools\\browser-login\\reports'; node login-test-pure-ui.js
```

GitHub Actions example (workflow snippet)

Save this as `.github/workflows/browser-smoke.yml` in your repo to run CI test and upload the report as an artifact.

```yaml
name: Browser smoke tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  browser-smoke:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      - name: Install deps
        working-directory: tools/browser-login
        run: |
          npm ci
          npx playwright install --with-deps
      - name: Start API server (assumes binary available or start script)
        working-directory: API-Server
        run: |
          go build -o api_server_run.exe .
          Start-Process -FilePath .\api_server_run.exe -ArgumentList '' -NoNewWindow
      - name: Start static server
        working-directory: Admin-FE
        run: |
          python -m http.server 8000 &
      - name: Run CI test
        working-directory: tools/browser-login
        env:
          ADMIN_FE_URL: http://localhost:8000
          API_BASE: http://localhost:9292/api/v1
          HEADLESS: '1'
        run: |
          npx ts-node login-test-ci.ts || exit 1
      - name: Upload report
        uses: actions/upload-artifact@v4
        with:
          name: browser-smoke-report
          path: tools/browser-login/reports/*.json

```

Notes:
- The CI script `login-test-ci.ts` is deterministic (logs in via API) and preferable for automated runs. The pure UI script is useful for debugging UI issues.
- Reports are JSON files placed into `tools/browser-login/reports/`.

