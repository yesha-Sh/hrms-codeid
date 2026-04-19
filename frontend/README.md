# Frontend README

Frontend application for the PT. CODEID HRMS project.

## Stack
- React
- Vite
- TypeScript
- React Router
- Playwright

## Main Responsibilities
- login and role-based routing
- admin, manager, and employee UI
- CRUD and table workflows
- profile flow
- CSV export
- E2E regression coverage

## Environment

Example file:

- [C:\laragon\www\HRMS\frontend\.env.example](C:\laragon\www\HRMS\frontend\.env.example)

Required:
- `VITE_API_BASE_URL`

Default:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Run Locally

```powershell
cd C:\laragon\www\HRMS\frontend
npm install
npm run dev
```

## Production Build

```powershell
cd C:\laragon\www\HRMS\frontend
npm run build
```

## E2E Regression Suite

Run:

```powershell
cd C:\laragon\www\HRMS\frontend
npm run test:e2e
```

Extra modes:

```powershell
npm run test:e2e:headed
npm run test:e2e:debug
```

## E2E Notes

- Playwright uses:
  - frontend on `http://127.0.0.1:4173`
  - isolated backend on `http://127.0.0.1:18080`
- the suite runs migrations and seeds automatically before execution
- tests use seeded PT. CODEID accounts and role flows

Important files:
- [C:\laragon\www\HRMS\frontend\playwright.config.ts](C:\laragon\www\HRMS\frontend\playwright.config.ts)
- [C:\laragon\www\HRMS\frontend\tests\e2e\global-setup.ts](C:\laragon\www\HRMS\frontend\tests\e2e\global-setup.ts)
- [C:\laragon\www\HRMS\frontend\tests\e2e\helpers.ts](C:\laragon\www\HRMS\frontend\tests\e2e\helpers.ts)

## Main Role Screens

### Admin
- dashboard
- employees
- departments
- attendance
- leave
- jobs
- holidays
- audit logs
- profile

### Manager
- dashboard
- team
- attendance
- leave
- approvals
- profile

### Employee
- dashboard
- attendance
- leave
- profile

## Demo Accounts

See root project README:

- [C:\laragon\www\HRMS\README.md](C:\laragon\www\HRMS\README.md)

Supporting docs:

- [C:\laragon\www\HRMS\docs\test-accounts.md](C:\laragon\www\HRMS\docs\test-accounts.md)
- [C:\laragon\www\HRMS\docs\github-actions-deploy.md](C:\laragon\www\HRMS\docs\github-actions-deploy.md)
