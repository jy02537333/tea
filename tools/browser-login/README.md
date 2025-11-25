Playwright Browser Login Test (Admin-FE)

Purpose
- Automate opening the Admin-FE, perform login with test credentials, and verify that `localStorage.tea_admin_token` is saved.

Prerequisites
- Node.js (>= 16)
- From repo root, run:

```powershell
cd tools\browser-login
npm install
npx playwright install
```

Usage
- Default (headless):
```powershell
# assumes Admin-FE is served at http://localhost:8000
ADMIN_FE_URL=http://localhost:8000 TEST_USER=admin TEST_PASS=pass node login-test.js
```

- Debug (show browser):
```powershell
set HEADLESS=0
ADMIN_FE_URL=http://localhost:8000 TEST_USER=admin TEST_PASS=pass node login-test.js
```

Notes
- The script attempts several selectors to open the login modal and click the login button. If your Admin-FE uses a different flow, adjust selectors in `login-test.js`.
- The script only prints the token and a snippet of the sidebar menu text (if found). It does not perform assertions or modify files.
- This folder and instructions do not perform any git operations.
