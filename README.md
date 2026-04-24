# PT. CODEID HRMS

HRMS internship project for **PT. CODEID** built with:

- `frontend/`: React + Vite + TypeScript
- `backend/`: Go + Chi + GORM + PostgreSQL

This README is written for someone who wants to:

1. clone the project from GitHub
2. run it on localhost
3. log in with seeded accounts

So yes, **your teacher can clone this repo and run it locally**, as long as the prerequisites below are installed and the setup steps are followed.

## Features

- authentication with email/password
- JWT access token + refresh token flow
- Argon2id password hashing
- role-based access for `admin`, `manager`, and `employee`
- employee management
- department management
- job management
- attendance management
- leave request and approval flow
- holiday management
- audit logs
- PT. CODEID seeded organization structure

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

## Project Structure

```text
HRMS
|-- backend
|   |-- cmd
|   |-- internal
|   |-- migrations
|   `-- README.md
|-- frontend
|   |-- src
|   |-- tests
|   `-- README.md
|-- docs
|   |-- test-accounts.md
|   |-- github-actions-deploy.md
|   `-- README.md
`-- README.md
```

## Prerequisites

Install these first:

- `Git`
- `Go` `1.25+`
- `Node.js` `22+`
- `npm`
- `PostgreSQL` `15+` or `16+`

Recommended local ports:

- backend: `8080`
- frontend: `5173`
- PostgreSQL: `5432`

## 1. Clone the Repository

```powershell
git clone https://github.com/yesha-Sh/hrms-codeid.git
cd hrms-codeid
```

## 2. Prepare PostgreSQL

Create a local database named `hrms`.

Example with `psql`:

```powershell
createdb -U postgres hrms
```

If your PostgreSQL username or password is different, that is fine. You only need to match the values later in `backend/.env`.

## 3. Configure Backend Environment

Copy the example file:

```powershell
Copy-Item .\backend\.env.example .\backend\.env
```

Default local example already assumes:

- host: `127.0.0.1`
- port: `5432`
- database: `hrms`
- user: `postgres`
- password: `postgres`

If your PostgreSQL credentials are different, edit:

- [backend/.env.example](C:\laragon\www\HRMS\backend\.env.example)

or the copied local file:

- `backend/.env`

Important local backend values:

```env
APP_ENV=development
HTTP_PORT=8080
FRONTEND_ORIGIN=http://localhost:5173
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=hrms
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
JWT_ACCESS_SECRET=change-me-access-secret
JWT_REFRESH_SECRET=change-me-refresh-secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
ADMIN_SEED_EMAIL=admin@northstar.id
ADMIN_SEED_PASSWORD=ChangeMe123!
```

## 4. Run Backend Migrations and Seed Data

Open a terminal in the project root, then run:

```powershell
cd backend
go run ./cmd/migrate up
go run ./cmd/seed-admin
go run ./cmd/seed-demo
```

If these commands succeed, your local database is ready.

## 5. Start the Backend API

Still in `backend/`:

```powershell
go run ./cmd/api
```

Backend will run at:

- [http://localhost:8080](http://localhost:8080)

Health check:

- [http://localhost:8080/healthz](http://localhost:8080/healthz)

Expected response:

```json
{"status":"ok"}
```

## 6. Configure Frontend Environment

Open a second terminal from the project root.

Copy the example file:

```powershell
Copy-Item .\frontend\.env.example .\frontend\.env
```

Default frontend env:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 7. Install Frontend Dependencies

```powershell
cd frontend
npm install
```

## 8. Start the Frontend

```powershell
npm run dev
```

Frontend will run at:

- [http://localhost:5173](http://localhost:5173)

## 9. Login with Seeded Accounts

Primary demo accounts:

- Admin: `admin@northstar.id` / `ChangeMe123!`
- Team manager: `fajar.maulana@codeid.co.id` / `Manager123!`
- Subdepartment manager: `salma.nuraini@codeid.co.id` / `Manager123!`
- Division manager: `dimas.kusuma@codeid.co.id` / `Manager123!`
- Employee: `nabila.putri@codeid.co.id` / `Employee123!`

More accounts:

- [docs/test-accounts.md](C:\laragon\www\HRMS\docs\test-accounts.md)

## Local Run Summary

If you want the shortest working flow:

### Terminal 1

```powershell
cd backend
go run ./cmd/migrate up
go run ./cmd/seed-admin
go run ./cmd/seed-demo
go run ./cmd/api
```

### Terminal 2

```powershell
cd frontend
npm install
npm run dev
```

Then open:

- [http://localhost:5173](http://localhost:5173)

## Common Problems

### PostgreSQL connection fails

Check:

- PostgreSQL service is running
- database `hrms` exists
- `DB_USER` and `DB_PASSWORD` in `backend/.env` are correct

### `JWT_ACCESS_SECRET and JWT_REFRESH_SECRET are required`

This means:

- `backend/.env` is missing
- or the JWT variables are empty

Make sure these exist in `backend/.env`:

```env
JWT_ACCESS_SECRET=change-me-access-secret
JWT_REFRESH_SECRET=change-me-refresh-secret
```

### Frontend cannot call backend

Check:

- backend is running on port `8080`
- `frontend/.env` contains:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### Migration says `no change`

That usually means migrations already ran. It is not an error.

## Useful Commands

### Backend tests

```powershell
cd backend
go test ./...
```

### Frontend production build

```powershell
cd frontend
npm run build
```

### Frontend E2E tests

```powershell
cd frontend
npm run test:e2e
```

## For Deployment

Production deployment docs are separated from local installation docs:

- [docs/github-actions-deploy.md](C:\laragon\www\HRMS\docs\github-actions-deploy.md)
- [docs/github-secrets-checklist.md](C:\laragon\www\HRMS\docs\github-secrets-checklist.md)

## Additional Documentation

- [backend/README.md](C:\laragon\www\HRMS\backend\README.md)
- [frontend/README.md](C:\laragon\www\HRMS\frontend\README.md)
- [docs/README.md](C:\laragon\www\HRMS\docs\README.md)
