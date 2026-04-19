# GitHub Secrets Copy-Paste Checklist

GitHub repository:
- `yesha-Sh/hrms-codeid`

Open:
- `Settings > Secrets and variables > Actions > New repository secret`

## Secrets to add

### Neon
- `NEON_DATABASE_URL`
  - value: your Neon production connection string

### Vercel
- `VERCEL_TOKEN`
- `VERCEL_ORG_ID`
- `VERCEL_PROJECT_ID`

### Railway
- `RAILWAY_TOKEN`
- `RAILWAY_PROJECT_ID`
- `RAILWAY_ENVIRONMENT`
- `RAILWAY_SERVICE`

## Helper script
If you already have GitHub CLI installed and authenticated, you can set all secrets from:

- `C:\laragon\www\HRMS\scripts\set-github-secrets.ps1`

Example:

```powershell
cd C:\laragon\www\HRMS
.\scripts\set-github-secrets.ps1 `
  -Repo 'yesha-Sh/hrms-codeid' `
  -NeonDatabaseUrl '<NEON_DATABASE_URL>' `
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
