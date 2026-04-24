# Backend README

Backend API for the PT. CODEID HRMS project.

For full localhost installation from clone to login, use the main guide first:

- [README.md](C:\laragon\www\HRMS\README.md)

## What This Backend Does

- auth and session flow
- role-based authorization
- employee, attendance, leave, team, assignment, holiday, and audit APIs
- PT. CODEID seed and demo data

## Important Directories

```text
backend
|-- cmd
|   |-- api
|   |-- migrate
|   |-- seed-admin
|   `-- seed-demo
|-- internal
|-- migrations
`-- README.md
```

## Local Environment

Example file:

- [backend/.env.example](C:\laragon\www\HRMS\backend\.env.example)

Create a local file:

```powershell
Copy-Item .env.example .env
```

Default local values are written for localhost development.

## Run Backend Only

```powershell
cd C:\laragon\www\HRMS\backend
go run ./cmd/migrate up
go run ./cmd/seed-admin
go run ./cmd/seed-demo
go run ./cmd/api
```

API:

- `http://localhost:8080`

Health check:

- `http://localhost:8080/healthz`

## Useful Commands

### Migrations

```powershell
go run ./cmd/migrate up
go run ./cmd/migrate down
```

### Seeds

```powershell
go run ./cmd/seed-admin
go run ./cmd/seed-demo
```

### Tests

```powershell
go test ./...
```

Main integration test:

- [backend/tests/integration/api_test.go](C:\laragon\www\HRMS\backend\tests\integration\api_test.go)

## API Areas

### Public auth
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

### Private
- `GET/PUT /api/v1/profile`
- `GET /api/v1/dashboard/admin`
- `GET /api/v1/dashboard/manager`
- `GET /api/v1/dashboard/employee`
- `GET/POST/PUT/DELETE /api/v1/employees`
- `GET/POST/PUT/DELETE /api/v1/attendances`
- `GET/POST/PUT/DELETE /api/v1/leave-requests`
- `GET/POST/PUT/DELETE /api/v1/employee-role-assignments`
- `GET /api/v1/teams`
- `GET /api/v1/teams/:id/members`
- `GET /api/v1/teams/:id/available-employees`
- `POST /api/v1/teams/:id/members`
- `DELETE /api/v1/teams/:id/members/:membershipId`

### Admin-only
- `GET/POST/PUT/DELETE /api/v1/departments`
- `GET/POST/PUT/DELETE /api/v1/jobs`
- `GET/POST/PUT/DELETE /api/v1/holidays`
- `GET /api/v1/audit-logs`
