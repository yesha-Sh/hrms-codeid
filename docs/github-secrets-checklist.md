# GitHub Secrets Copy-Paste Checklist

GitHub repository:
- `yesha-Sh/hrms-codeid`

Open:
- `Settings > Secrets and variables > Actions > New repository secret`

## Secrets to add

### Neon
- `NEON_DATABASE_URL`
  - value: your Neon production connection string

### Backend app secrets
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`

### Vercel
- `VERCEL_TOKEN`
- `VERCEL_ORG_ID`
- `VERCEL_PROJECT_ID`

### Railway
- `RAILWAY_TOKEN`
- optional: `RAILWAY_API_TOKEN`
- `RAILWAY_PROJECT_ID`
- `RAILWAY_ENVIRONMENT`
- `RAILWAY_SERVICE`

Notes:
- `RAILWAY_TOKEN` is enough for deploy-focused workflow usage.
- `RAILWAY_API_TOKEN` is only needed if you want GitHub Actions to run `railway link` and `railway variables set` automatically.

## Helper script
If you already have GitHub CLI installed and authenticated, you can set all secrets from:

- `C:\laragon\www\HRMS\scripts\set-github-secrets.ps1`

Example:

```powershell
cd C:\laragon\www\HRMS
.\scripts\set-github-secrets.ps1 `
  -Repo 'yesha-Sh/hrms-codeid' `
  -NeonDatabaseUrl '<NEON_DATABASE_URL>' `
  -JwtAccessSecret '<JWT_ACCESS_SECRET>' `
  -JwtRefreshSecret '<JWT_REFRESH_SECRET>' `
  -VercelToken '<VERCEL_TOKEN>' `
  -VercelOrgId '<VERCEL_ORG_ID>' `
  -VercelProjectId '<VERCEL_PROJECT_ID>' `
  -RailwayToken '<RAILWAY_TOKEN>' `
  -RailwayProjectId '<RAILWAY_PROJECT_ID>' `
  -RailwayEnvironment '<RAILWAY_ENVIRONMENT>' `
  -RailwayService '<RAILWAY_SERVICE>'
```

## Important note
Do not commit real secrets into the repo.
GitHub Actions secrets must be stored in the GitHub repository settings, not in tracked files.
