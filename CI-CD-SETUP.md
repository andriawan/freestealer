# CI/CD Setup Complete ✅

This document summarizes all the CI/CD infrastructure created for the FreeStealer project.

## 📋 What's Been Created

### GitHub Actions Workflows (5 workflows)

#### 1. **Test & Coverage** (`.github/workflows/test.yml`)
- ✅ Runs on every push/PR to main and develop
- ✅ Executes all tests with race detection
- ✅ Generates coverage reports
- ✅ Uploads to Codecov (optional)
- ✅ Enforces 30% minimum coverage threshold
- ✅ Caches Go modules for faster builds

#### 2. **Build** (`.github/workflows/build.yml`)
- ✅ Builds for 5 platforms: Linux (amd64, arm64), Windows (amd64), macOS (amd64, arm64)
- ✅ Runs golangci-lint for code quality
- ✅ Uploads artifacts with 7-day retention
- ✅ Matrix-based parallel builds

#### 3. **Deploy to Leapcell** (`.github/workflows/deploy-leapcell.yml`)
- ✅ Auto-deploys to Leapcell.io on push to main
- ✅ Manual trigger support
- ✅ Creates optimized production build
- ✅ Sets up proper Leapcell configuration
- ✅ Deployment verification

#### 4. **Release** (`.github/workflows/release.yml`)
- ✅ Triggered by version tags (v*.*.*)
- ✅ Creates GitHub releases automatically
- ✅ Builds binaries for all platforms
- ✅ Creates .tar.gz and .zip archives
- ✅ Builds multi-platform Docker images
- ✅ Pushes to GitHub Container Registry (ghcr.io)
- ✅ Manual workflow dispatch support

#### 5. **CodeQL Security** (`.github/workflows/codeql.yml`)
- ✅ Weekly security scans (Monday midnight)
- ✅ Runs on every push/PR
- ✅ Automated vulnerability detection
- ✅ GitHub Security integration

### Configuration Files

#### Docker Support
- ✅ `Dockerfile` - Multi-stage optimized build
- ✅ `.dockerignore` - Minimal image size
- ✅ `docker-compose.yml` - Local deployment
- ✅ Health check support
- ✅ Non-root user configuration

#### Development Tools
- ✅ `Makefile` - Build automation commands
- ✅ `.golangci.yml` - Linter configuration
- ✅ `.github/dependabot.yml` - Automatic dependency updates
- ✅ `.env.example` - Environment template

#### Scripts
- ✅ `scripts/coverage.ps1` - Coverage report generator (Windows)
- ✅ `scripts/coverage.sh` - Coverage report generator (Linux/Mac)

#### Documentation
- ✅ `README.md` - Updated with CI/CD badges and instructions
- ✅ `DEPLOYMENT.md` - Comprehensive deployment guide
- ✅ `.github/WORKFLOWS.md` - Workflow documentation

### Updated Files
- ✅ `.gitignore` - Added coverage and build artifacts
- ✅ `README.md` - Added badges, deployment info
- ✅ Coverage reports excluded from git

## 🚀 Quick Start Guide

### 1. Enable GitHub Actions

Actions are automatically enabled when you push these files to GitHub:

```bash
git add .
git commit -m "feat: add CI/CD infrastructure"
git push origin main
```

### 2. Set Up Secrets (Optional)

#### For Coverage Reporting (Optional)
1. Sign up at https://codecov.io
2. Add repository
3. Go to GitHub repo → Settings → Secrets → Actions
4. Add `CODECOV_TOKEN`

#### For Leapcell Deployment (Required for auto-deploy)
1. Create account at https://leapcell.io
2. Create new project
3. Get API token from dashboard
4. Add to GitHub secrets:
   - `LEAPCELL_TOKEN`
   - `LEAPCELL_PROJECT_ID`

### 3. Configure Environment Variables

On Leapcell dashboard, set:
```bash
PORT=8080
DATABASE_PATH=/data/kasir.db
GITHUB_CLIENT_ID=your_id
GITHUB_CLIENT_SECRET=your_secret
GITHUB_CALLBACK_URL=https://your-app.leapcell.app/auth/github/callback
SESSION_SECRET=your_secret
ENV=production
```

### 4. Test Locally

```bash
# Run tests
make test

# Generate coverage
./scripts/coverage.ps1  # Windows
./scripts/coverage.sh   # Linux/Mac

# Build
make build

# Test Docker
docker-compose up --build
```

## 📊 Workflow Triggers

| Workflow | Trigger Events |
|----------|---------------|
| Test & Coverage | Push/PR to main/develop |
| Build | Push/PR to main/develop |
| Deploy | Push to main (auto), Manual |
| Release | Tag push (v*), Manual |
| CodeQL | Push/PR, Weekly schedule |

## 🎯 Usage Examples

### Running Tests Locally
```bash
# All tests
go test ./... -v

# With coverage
make coverage

# Generate HTML report
make coverage-html
```

### Building Binaries
```bash
# Current platform
go build -o freestealer .

# All platforms (using Make)
make build

# Or manually for specific platform
GOOS=linux GOARCH=amd64 go build -o freestealer-linux-amd64 .
```

### Creating a Release
```bash
# 1. Commit all changes
git add .
git commit -m "chore: prepare v1.0.0"
git push

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. Watch the automation happen! 🎉
# Go to Actions tab to see progress
```

### Manual Deployment
```bash
# Deploy to Leapcell (if secrets configured)
# Go to Actions → Deploy to Leapcell → Run workflow

# Build Docker image
docker build -t freestealer:latest .

# Run locally
docker-compose up
```

## 📦 Release Assets

When you create a release (push a tag), the following assets are automatically created:

### Binary Archives
- `freestealer-linux-amd64.tar.gz`
- `freestealer-linux-arm64.tar.gz`
- `freestealer-windows-amd64.zip`
- `freestealer-darwin-amd64.tar.gz`
- `freestealer-darwin-arm64.tar.gz`

### Docker Images
- `ghcr.io/USERNAME/freestealer:latest`
- `ghcr.io/USERNAME/freestealer:v1.0.0`
- `ghcr.io/USERNAME/freestealer:1.0`
- `ghcr.io/USERNAME/freestealer:1`

Users can pull with:
```bash
docker pull ghcr.io/USERNAME/freestealer:latest
```

## 🔧 Customization

### Adjust Coverage Threshold
Edit `.github/workflows/test.yml`:
```yaml
if (( $(echo "$total < 30" | bc -l) )); then  # Change 30 to desired %
```

### Add New Build Platform
Edit `.github/workflows/build.yml`:
```yaml
matrix:
  goos: [linux, windows, darwin, freebsd]  # Add new OS
  goarch: [amd64, arm64, 386]  # Add new arch
```

### Change Auto-deploy Branch
Edit `.github/workflows/deploy-leapcell.yml`:
```yaml
on:
  push:
    branches: [ staging ]  # Deploy from staging instead
```

## 🛠️ Maintenance Commands

```bash
# Update dependencies
go get -u ./...
go mod tidy

# Run linter
golangci-lint run

# Check security issues
gosec ./...

# Update Swagger docs
swag init

# Clean build artifacts
make clean
```

## 📈 Monitoring

### CI/CD Status
- GitHub Actions tab shows workflow runs
- README badges show current status

### Coverage Tracking
- Codecov.io dashboard (if configured)
- Local reports: `coverage.html`

### Deployment Health
- Leapcell dashboard
- Health endpoint: `GET /health`

## ✅ Checklist for First Push

Before pushing to GitHub:

- [ ] Update `YOUR_USERNAME` in README.md badges
- [ ] Create GitHub repository if not exists
- [ ] Set up Leapcell account (if using auto-deploy)
- [ ] Add GitHub secrets (if using Codecov or Leapcell)
- [ ] Update GITHUB_CALLBACK_URL in .env
- [ ] Test locally: `make test`
- [ ] Test Docker build: `docker build -t test .`
- [ ] Review and customize workflows as needed
- [ ] Update LICENSE file with your name
- [ ] Add CHANGELOG.md (optional but recommended)

## 🎉 What Happens Next?

Once you push to GitHub:

1. **On Every Push/PR:**
   - ✅ Tests run automatically
   - ✅ Coverage is calculated
   - ✅ Binaries are built for all platforms
   - ✅ Code is linted
   - ✅ Security scan runs

2. **On Push to Main:**
   - ✅ All above, plus...
   - ✅ Automatic deployment to Leapcell (if configured)

3. **On Tag Push (v*):**
   - ✅ All above, plus...
   - ✅ GitHub release is created
   - ✅ Binaries are attached to release
   - ✅ Docker images are built and pushed
   - ✅ Changelog is auto-generated

4. **Weekly:**
   - ✅ Security scan runs
   - ✅ Dependabot checks for updates

## 🆘 Troubleshooting

### Tests Failing?
```bash
# Run locally first
go test ./... -v

# Check specific package
go test ./handlers -v
```

### Build Failing?
```bash
# Verify dependencies
go mod verify
go mod tidy

# Try clean build
go clean -cache
go build .
```

### Docker Issues?
```bash
# Test build
docker build -t test .

# Check logs
docker logs <container-id>

# Verify Docker is running
docker version
```

### Deployment Issues?
1. Check GitHub secrets are set
2. Verify Leapcell service status
3. Check workflow logs in Actions tab
4. Ensure environment variables are configured

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Leapcell Documentation](https://docs.leapcell.io)
- [Docker Documentation](https://docs.docker.com)
- [Go Release Documentation](https://golang.org/doc/devel/release)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)

## 🎯 Next Steps

1. **Push to GitHub** - Let CI/CD workflows run
2. **Monitor First Run** - Check Actions tab
3. **Fix Any Issues** - Review logs if anything fails
4. **Configure Secrets** - If using Codecov or Leapcell
5. **Create First Release** - Tag v1.0.0 when ready
6. **Update README** - Replace YOUR_USERNAME with actual username
7. **Celebrate** 🎉 - You have full CI/CD automation!

---

**Need Help?**
- Check `.github/WORKFLOWS.md` for detailed workflow docs
- See `DEPLOYMENT.md` for deployment instructions
- Review GitHub Actions logs for specific errors
- Open an issue on GitHub

**Made with** ❤️ **for automated deployments and releases!**
