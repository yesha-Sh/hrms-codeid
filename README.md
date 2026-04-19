# PT. CODEID HRMS

HRMS internship project for **PT. CODEID** with:

- `frontend/`: React + Vite + TypeScript
- `backend/`: Go + Chi + GORM + PostgreSQL

The system is built around PT. CODEID's organization structure, seeded demo data, role-based access, and realistic HR flows for:

- `admin`
- `manager`
- `employee`

## Features

### Core HRMS
- authentication with email/password
- JWT access token + refresh token flow
- Argon2id password hashing
- role-based routing and authorization
- employee management
- department management
- job management
- attendance management
- leave request and approval flow
- holiday management
- audit logs
- profile page

### PT. CODEID Structure
- seeded PT. CODEID organization
- division, subdepartment, and team manager structure
- cross-department team support
- remote and client-based worker handling
- seeded realistic employees, teams, attendance, leave, and assignments

### Testing
- backend automated auth + integration tests
- frontend Playwright E2E regression suite

## Monorepo Structure

```text
C:\laragon\www\HRMS
|-- backend
|   |-- cmd
|   |-- internal
|   |-- migrations
|   `-- README.md
|-- docs
|   |-- github-actions-deploy.md
|   |-- test-accounts.md
|   `-- README.md
|-- frontend
|   |-- src
|   |-- tests
|   `-- README.md
|-- .github
|   `-- workflows
`-- README.md
```

## Tech Stack

### Backend
- Go
- Chi
- GORM
- PostgreSQL
- golang-migrate
- JWT
- Argon2id

### Frontend
- React
- Vite
- TypeScript
- React Router
- Playwright

## Quick Start

### 1. Backend setup

Copy or update backend environment values:

File: [C:\laragon\www\HRMS\backend\.env.example](C:\laragon\www\HRMS\backend\.env.example)

Then run:

```powershell
cd C:\laragon\www\HRMS\backend
go run ./cmd/migrate up
go run ./cmd/seed-admin
go run ./cmd/seed-demo
go run ./cmd/api
```

API default:

- `http://localhost:8080`

Health check:

- `http://localhost:8080/healthz`

### 2. Frontend setup

File: [C:\laragon\www\HRMS\frontend\.env.example](C:\laragon\www\HRMS\frontend\.env.example)

Then run:

```powershell
cd C:\laragon\www\HRMS\frontend
npm install
npm run dev
```

Frontend default:

- `http://localhost:5173`

## Demo Accounts

Primary accounts:

- Admin: `admin@northstar.id` / `ChangeMe123!`
- Team manager: `fajar.maulana@codeid.co.id` / `Manager123!`
- Subdepartment manager: `salma.nuraini@codeid.co.id` / `Manager123!`
- Division manager: `dimas.kusuma@codeid.co.id` / `Manager123!`
- Employee: `nabila.putri@codeid.co.id` / `Employee123!`

Full seeded account matrix:

- [C:\laragon\www\HRMS\docs\test-accounts.md](C:\laragon\www\HRMS\docs\test-accounts.md)

## Recommended Demo Flow

### Admin
- log in as admin
- open dashboard
- open employees
- test management scope filter
- create or edit employees
- export CSV

### Manager
- log in as `fajar.maulana@codeid.co.id`
- open team page
- add available employee to team
- open attendance
- add self attendance
- open leave
- request leave

### Approval Flow
- log in as `yoga.prasetyo@codeid.co.id`
- create a leave request
- log out
- log in as `dimas.kusuma@codeid.co.id`
- approve the request from approvals page

### Employee
- log in as `nabila.putri@codeid.co.id`
- create attendance
- create leave request
- update pending leave
- open profile page

## Test Commands

### Backend tests

```powershell
cd C:\laragon\www\HRMS\backend
go test ./...
```

### Frontend build

```powershell
cd C:\laragon\www\HRMS\frontend
npm run build
```

### Frontend E2E tests

```powershell
cd C:\laragon\www\HRMS\frontend
npm run test:e2e
```

Extra Playwright modes:

```powershell
npm run test:e2e:headed
npm run test:e2e:debug
```

## GitHub and Auto Deploy

This repo is prepared for GitHub CI/CD with:

- [C:\laragon\www\HRMS\.github\workflows\ci.yml](C:\laragon\www\HRMS\.github\workflows\ci.yml)
- [C:\laragon\www\HRMS\.github\workflows\deploy-frontend-vercel.yml](C:\laragon\www\HRMS\.github\workflows\deploy-frontend-vercel.yml)
- [C:\laragon\www\HRMS\.github\workflows\deploy-backend-railway.yml](C:\laragon\www\HRMS\.github\workflows\deploy-backend-railway.yml)
- [C:\laragon\www\HRMS\.github\workflows\migrate-neon.yml](C:\laragon\www\HRMS\.github\workflows\migrate-neon.yml)

Deployment and secret setup guide:

- [C:\laragon\www\HRMS\docs\github-actions-deploy.md](C:\laragon\www\HRMS\docs\github-actions-deploy.md)

## What The Automated Tests Cover

### Backend
- Argon2 hashing and verification
- JWT generation and parsing
- login, refresh, logout
- admin employee CRUD
- management scope filtering
- employee finalized leave restrictions
- manager approval rules
- team manager attendance and team membership flow
- overlapping secondary assignment rejection

### Frontend E2E
- auth flows for all roles
- admin employee filtering, create/delete, export
- team manager team flow
- manager self attendance and self leave
- subdepartment manager approval flow
- employee attendance and leave flows
- finalized leave lock behavior
- profile database-backed selections

## Notes

- The backend integration suite creates a temporary PostgreSQL test database automatically.
- The Playwright suite seeds backend data before running.
- The E2E setup uses an isolated API port for browser tests.

## Additional Docs

- Backend docs: [C:\laragon\www\HRMS\backend\README.md](C:\laragon\www\HRMS\backend\README.md)
- Frontend docs: [C:\laragon\www\HRMS\frontend\README.md](C:\laragon\www\HRMS\frontend\README.md)
- Documentation index: [C:\laragon\www\HRMS\docs\README.md](C:\laragon\www\HRMS\docs\README.md)
- Test accounts: [C:\laragon\www\HRMS\docs\test-accounts.md](C:\laragon\www\HRMS\docs\test-accounts.md)
- GitHub deploy guide: [C:\laragon\www\HRMS\docs\github-actions-deploy.md](C:\laragon\www\HRMS\docs\github-actions-deploy.md)
