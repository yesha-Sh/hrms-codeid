# Frontend README

Frontend application for the PT. CODEID HRMS project.

For full localhost installation from clone to login, use the main guide first:

- [README.md](C:\laragon\www\HRMS\README.md)

## Stack

- React
- Vite
- TypeScript
- React Router

## Local Environment

Example file:

- [frontend/.env.example](C:\laragon\www\HRMS\frontend\.env.example)

Create a local file:

```powershell
Copy-Item .env.example .env
```

Default local API target:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Run Frontend Only

```powershell
cd C:\laragon\www\HRMS\frontend
npm install
npm run dev
```

App:

- `http://localhost:5173`

## Build

```powershell
npm run build
```

## E2E Tests

```powershell
npm run test:e2e
```

Extra modes:

```powershell
npm run test:e2e:headed
npm run test:e2e:debug
```

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
