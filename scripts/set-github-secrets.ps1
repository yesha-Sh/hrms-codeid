param(
  [Parameter(Mandatory = $true)] [string]$Repo,
  [Parameter(Mandatory = $true)] [string]$NeonDatabaseUrl,
  [Parameter(Mandatory = $true)] [string]$JwtAccessSecret,
  [Parameter(Mandatory = $true)] [string]$JwtRefreshSecret,
  [Parameter(Mandatory = $true)] [string]$VercelToken,
  [Parameter(Mandatory = $true)] [string]$VercelOrgId,
  [Parameter(Mandatory = $true)] [string]$VercelProjectId,
  [Parameter(Mandatory = $true)] [string]$RailwayToken,
  [Parameter(Mandatory = $true)] [string]$RailwayProjectId,
  [Parameter(Mandatory = $true)] [string]$RailwayEnvironment,
  [Parameter(Mandatory = $true)] [string]$RailwayService
)

$ErrorActionPreference = 'Stop'

function Set-GitHubSecret {
  param(
    [string]$Name,
    [string]$Value
  )

  Write-Host "Setting GitHub secret $Name for $Repo..."
  $Value | gh secret set $Name --repo $Repo
}

Set-GitHubSecret 'NEON_DATABASE_URL' $NeonDatabaseUrl
Set-GitHubSecret 'JWT_ACCESS_SECRET' $JwtAccessSecret
Set-GitHubSecret 'JWT_REFRESH_SECRET' $JwtRefreshSecret
Set-GitHubSecret 'VERCEL_TOKEN' $VercelToken
Set-GitHubSecret 'VERCEL_ORG_ID' $VercelOrgId
Set-GitHubSecret 'VERCEL_PROJECT_ID' $VercelProjectId
Set-GitHubSecret 'RAILWAY_TOKEN' $RailwayToken
Set-GitHubSecret 'RAILWAY_PROJECT_ID' $RailwayProjectId
Set-GitHubSecret 'RAILWAY_ENVIRONMENT' $RailwayEnvironment
Set-GitHubSecret 'RAILWAY_SERVICE' $RailwayService

Write-Host 'All GitHub Actions secrets have been set.'
