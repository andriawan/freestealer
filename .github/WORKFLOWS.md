# GitHub Actions CI/CD Configuration

This document describes all GitHub Actions workflows configured for the FreeStealer project.

## Workflows Overview

| Workflow | Trigger | Purpose | Status |
|----------|---------|---------|--------|
| Test & Coverage | Push/PR to main, develop | Run tests and report coverage | Required |
| Build | Push/PR to main, develop | Build for multiple platforms | Required |
| Deploy to Leapcell | Push to main | Auto-deploy to production | Optional |
| Release | Tag push (v*) | Create releases with binaries | Manual |
| CodeQL | Push/PR/Schedule | Security analysis | Recommended |

## 1. Test & Coverage Workflow

**File:** `.github/workflows/test.yml`

### Triggers
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

### Jobs
- **test**: Runs all tests with race detection and coverage

### Steps
1. Checkout code
2. Set up Go 1.23
3. Cache Go modules
4. Download dependencies
5. Run tests with race detection
6. Generate coverage report
7. Upload to Codecov (optional)
8. Check coverage threshold (minimum 30%)

### Required Secrets
- `CODECOV_TOKEN` (optional): For uploading coverage to Codecov.io

### Configuration
```yaml
Coverage Threshold: 30%
Go Version: 1.23
Test Timeout: 5 minutes
```

## 2. Build Workflow

**File:** `.github/workflows/build.yml`

### Triggers
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

### Jobs
- **build**: Cross-platform binary compilation
- **lint**: Code quality checks

### Build Matrix
| OS | Architecture |
|--------|--------------|
| Linux | amd64, arm64 |
| Windows | amd64 |
| macOS | amd64, arm64 |

### Steps
1. Checkout code
2. Set up Go 1.23
3. Cache Go modules
4. Download dependencies
5. Build binary for platform
6. Upload artifact (7-day retention)

### Artifacts
- `freestealer-linux-amd64`
- `freestealer-linux-arm64`
- `freestealer-windows-amd64.exe`
- `freestealer-darwin-amd64`
- `freestealer-darwin-arm64`

## 3. Deploy to Leapcell Workflow

**File:** `.github/workflows/deploy-leapcell.yml`

### Triggers
- Push to `main` branch
- Manual workflow dispatch

### Jobs
- **deploy**: Deploy to Leapcell.io

### Steps
1. Checkout code
2. Set up Go 1.23
3. Build Linux binary
4. Create deployment package
5. Create Procfile and leapcell.json
6. Install Leapcell CLI
7. Deploy to production
8. Verify deployment

### Required Secrets
- `LEAPCELL_TOKEN`: Your Leapcell API token
- `LEAPCELL_PROJECT_ID`: Your Leapcell project ID

### Environment Variables
Set these in your Leapcell dashboard:
```bash
PORT=8080
DATABASE_PATH=/data/kasir.db
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret
GITHUB_CALLBACK_URL=https://your-app.leapcell.app/auth/github/callback
SESSION_SECRET=your_session_secret
ENV=production
```

## 4. Release Workflow

**File:** `.github/workflows/release.yml`

### Triggers
- Push tags matching `v*` pattern (e.g., `v1.0.0`, `v2.1.3`)
- Manual workflow dispatch

### Jobs
- **create-release**: Create GitHub release
- **build-and-upload**: Build binaries for all platforms
- **docker-release**: Build and push Docker image

### Steps

#### Create Release
1. Checkout code
2. Get version from tag
3. Generate changelog
4. Create GitHub release

#### Build and Upload
1. Build for each platform
2. Create archives (.tar.gz for Linux/Mac, .zip for Windows)
3. Upload to release

#### Docker Release
1. Set up Docker Buildx
2. Login to GitHub Container Registry
3. Build multi-platform image (amd64, arm64)
4. Push to `ghcr.io`

### Release Assets
- Source code (zip, tar.gz)
- Binary archives for all platforms
- Docker image: `ghcr.io/username/freestealer:version`

### Tags Format
```bash
v1.0.0      # Stable release
v1.0.0-rc1  # Release candidate
v1.0.0-beta # Beta release
```

## 5. CodeQL Security Analysis

**File:** `.github/workflows/codeql.yml`

### Triggers
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches
- Schedule: Every Monday at midnight (UTC)

### Jobs
- **analyze**: Security code analysis

### Steps
1. Checkout repository
2. Initialize CodeQL
3. Autobuild
4. Perform CodeQL analysis
5. Upload results to GitHub Security

### Languages Analyzed
- Go

## Setting Up Secrets

### GitHub Repository Secrets

Navigate to: `Settings → Secrets and variables → Actions`

#### Required for All Workflows
None (tests and builds work without secrets)

#### Optional for Enhanced Features
- `CODECOV_TOKEN`: Coverage reporting on Codecov.io
  - Get from: https://codecov.io

#### Required for Leapcell Deployment
- `LEAPCELL_TOKEN`: Your Leapcell API token
  - Get from: Leapcell dashboard → Settings → API Tokens
- `LEAPCELL_PROJECT_ID`: Your project identifier
  - Get from: Leapcell dashboard → Project → Settings

#### Auto-configured by GitHub
- `GITHUB_TOKEN`: Automatically provided for releases and Docker registry

## Workflow Dependencies

```
┌─────────────────┐
│  Push to main   │
└────────┬────────┘
         │
         ├─────────────────────────────────────┐
         │                                     │
         ▼                                     ▼
┌────────────────┐                  ┌──────────────────┐
│  Test & Coverage│                 │      Build       │
└────────┬────────┘                 └──────────────────┘
         │
         │ (if tests pass)
         │
         ▼
┌────────────────┐
│ Deploy Leapcell│
└────────────────┘

On Tag Push (v*)
         │
         ▼
┌────────────────┐
│    Release     │
│  ├─ Binaries   │
│  └─ Docker     │
└────────────────┘
```

## Best Practices

### 1. Branch Protection
Enable branch protection for `main`:
- Require status checks (test, build)
- Require pull request reviews
- Require branches to be up to date

### 2. Release Process
```bash
# 1. Update version
git checkout main
git pull

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. Wait for automated release
# Check: Actions → Release workflow
```

### 3. Monitoring Deployments
- Check workflow status in Actions tab
- Monitor Leapcell dashboard for deployment status
- Review logs for any errors

### 4. Rollback Process
If deployment fails:
```bash
# Revert to previous tag
git tag -d v1.0.1  # Delete bad tag locally
git push origin :refs/tags/v1.0.1  # Delete from remote

# Deploy previous version
git push origin main  # Redeploy last good commit
```

## Troubleshooting

### Tests Failing
1. Check test logs in Actions tab
2. Run tests locally: `go test ./... -v`
3. Check for environment-specific issues

### Build Failing
1. Verify Go version compatibility
2. Check for missing dependencies
3. Run `go mod verify` locally

### Deployment Failing
1. Verify Leapcell secrets are set
2. Check Leapcell service status
3. Review deployment logs in workflow

### Docker Build Failing
1. Check Dockerfile syntax
2. Verify base image availability
3. Test build locally: `docker build -t test .`

## Extending Workflows

### Add New Platform Build
Edit `.github/workflows/build.yml`:
```yaml
matrix:
  goos: [linux, windows, darwin, freebsd]  # Add freebsd
  goarch: [amd64, arm64, 386]  # Add 386
```

### Add Staging Environment
Create `.github/workflows/deploy-staging.yml`:
```yaml
on:
  push:
    branches: [ develop ]

jobs:
  deploy-staging:
    # Similar to deploy-leapcell.yml
    # Use different LEAPCELL_PROJECT_ID
```

### Add Performance Tests
Add to `.github/workflows/test.yml`:
```yaml
- name: Run benchmarks
  run: go test -bench=. -benchmem ./...
```

## Maintenance

### Weekly Tasks
- Review CodeQL security alerts
- Check Dependabot PRs
- Update dependencies if needed

### Monthly Tasks
- Review and update Go version
- Check for workflow improvements
- Review coverage trends

### Per Release
- Update CHANGELOG.md
- Test release locally
- Verify all platforms build
- Check deployment health

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Leapcell Documentation](https://docs.leapcell.io)
- [Go Release Best Practices](https://golang.org/doc/install/source#release)
- [Docker Multi-platform Builds](https://docs.docker.com/build/building/multi-platform/)
