# Backend README

Backend API for the PT. CODEID HRMS project.

## Stack
- Go
- Chi
- GORM
- PostgreSQL
- golang-migrate
- JWT
- Argon2id

## Main Responsibilities
- auth and session flow
- role-based authorization
- employee, attendance, leave, team, assignment, holiday, and audit APIs
- PT. CODEID seed/reference data

## Important Directories

```text
backend
|-- cmd
|   |-- api
|   |-- migrate
|   |-- seed-admin
|   `-- seed-demo
|-- internal
|   |-- auth
|   |-- config
|   |-- db
|   |-- middleware
|   |-- models
|   |-- modules
|   `-- server
|-- migrations
`-- README.md
```

## Environment

Example file:

- [C:\laragon\www\HRMS\backend\.env.example](C:\laragon\www\HRMS\backend\.env.example)

Key values:
- `APP_ENV`
- `HTTP_PORT`
- `DATABASE_URL`
- `FRONTEND_ORIGIN`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SSLMODE`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `JWT_ACCESS_TTL`
- `JWT_REFRESH_TTL`
- `ADMIN_SEED_EMAIL`
- `ADMIN_SEED_PASSWORD`

## Running Locally

```powershell
cd C:\laragon\www\HRMS\backend
go run ./cmd/migrate up
go run ./cmd/seed-admin
go run ./cmd/seed-demo
go run ./cmd/api
```

## Migrations

Run up:

```powershell
go run ./cmd/migrate up
```

Run down:

```powershell
go run ./cmd/migrate down
```

## Seed Commands

Admin user:

```powershell
go run ./cmd/seed-admin
```

Full PT. CODEID demo/reference data:

```powershell
go run ./cmd/seed-demo
```

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

## Automated Tests

Run all tests:

```powershell
cd C:\laragon\www\HRMS\backend
go test ./...
```

Included:
- auth unit tests
- Postgres integration tests with temporary DB setup

Main test file:

- [C:\laragon\www\HRMS\backend\tests\integration\api_test.go](C:\laragon\www\HRMS\backend\tests\integration\api_test.go)

## Seeded Accounts

See root project README:

- [C:\laragon\www\HRMS\README.md](C:\laragon\www\HRMS\README.md)

Extended documentation:

- [C:\laragon\www\HRMS\docs\github-actions-deploy.md](C:\laragon\www\HRMS\docs\github-actions-deploy.md)
- [C:\laragon\www\HRMS\docs\test-accounts.md](C:\laragon\www\HRMS\docs\test-accounts.md)
