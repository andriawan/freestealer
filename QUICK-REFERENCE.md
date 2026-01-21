# FreeStealer CI/CD Quick Reference

## 🚀 Common Commands

### Testing
```bash
# Run all tests
make test
go test ./... -v

# With coverage
make coverage
go test ./... -cover

# Coverage report (HTML)
make coverage-html
.\scripts\coverage.ps1    # Windows
./scripts/coverage.sh     # Linux/Mac
```

### Building
```bash
# Current platform
make build
go build -o freestealer .

# All platforms (manual)
GOOS=linux GOARCH=amd64 go build -o freestealer-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o freestealer-windows-amd64.exe .
GOOS=darwin GOARCH=amd64 go build -o freestealer-darwin-amd64 .
```

### Docker
```bash
# Build image
docker build -t freestealer:latest .

# Run container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Development
```bash
# Hot reload
make dev
air

# Format code
go fmt ./...

# Lint
golangci-lint run

# Update dependencies
go get -u ./...
go mod tidy
```

## 📋 GitHub Actions Workflows

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| Test & Coverage | `test.yml` | Push/PR | Run tests, coverage |
| Build | `build.yml` | Push/PR | Multi-platform builds |
| Deploy | `deploy-leapcell.yml` | Push to main | Auto-deploy |
| Release | `release.yml` | Tag (v*) | Create releases |
| CodeQL | `codeql.yml` | Weekly/Push | Security scan |

## 🔑 GitHub Secrets (Optional)

| Secret | Required For | Get From |
|--------|-------------|----------|
| `CODECOV_TOKEN` | Coverage reporting | codecov.io |
| `LEAPCELL_TOKEN` | Auto-deployment | Leapcell dashboard |
| `LEAPCELL_PROJECT_ID` | Auto-deployment | Leapcell project settings |

## 🏷️ Creating a Release

```bash
# 1. Ensure all tests pass
make test

# 2. Commit all changes
git add .
git commit -m "chore: prepare release v1.0.0"
git push

# 3. Create tag
git tag -a v1.0.0 -m "Release v1.0.0"

# 4. Push tag (triggers release workflow)
git push origin v1.0.0
```

**Automated Actions:**
- ✅ Create GitHub release
- ✅ Build binaries (5 platforms)
- ✅ Create archives (.tar.gz, .zip)
- ✅ Build Docker images
- ✅ Push to ghcr.io

## 🌐 Deployment

### Automatic (Leapcell)
Push to main → Auto-deploys (if secrets configured)

### Manual Trigger
1. Go to Actions → Deploy to Leapcell
2. Click "Run workflow"
3. Select branch → Run

### Docker Deployment
```bash
# Using Docker Compose
docker-compose up -d

# Or pull from registry (after release)
docker pull ghcr.io/USERNAME/freestealer:latest
docker run -p 8080:8080 ghcr.io/USERNAME/freestealer:latest
```

## 🔍 Monitoring

### Check Workflow Status
```
GitHub Repository → Actions tab
```

### View Coverage
```
codecov.io dashboard (if configured)
Local: coverage.html
```

### Check Deployment
```
Leapcell dashboard
Health check: GET http://your-app/health
```

## 📊 Project Status Badges

Add to README.md:
```markdown
[![Test](https://github.com/USER/REPO/workflows/Test%20&%20Coverage/badge.svg)](...)
[![Build](https://github.com/USER/REPO/workflows/Build/badge.svg)](...)
[![Deploy](https://github.com/USER/REPO/workflows/Deploy%20to%20Leapcell/badge.svg)](...)
[![Coverage](https://codecov.io/gh/USER/REPO/branch/main/graph/badge.svg)](...)
```

## 🛠️ Troubleshooting

### Tests failing locally?
```bash
go test ./... -v
go mod verify
go mod tidy
```

### Docker build fails?
```bash
docker system prune -a
docker build --no-cache -t freestealer .
```

### Workflow failing?
1. Check Actions tab logs
2. Verify secrets are set
3. Check syntax of YAML files
4. Review error messages

## 📁 Project Structure

```
.
├── .github/
│   ├── workflows/         # CI/CD workflows
│   ├── dependabot.yml     # Dependency updates
│   └── WORKFLOWS.md       # Workflow docs
├── scripts/
│   ├── coverage.ps1       # Coverage script (Windows)
│   └── coverage.sh        # Coverage script (Linux/Mac)
├── Dockerfile             # Docker image
├── docker-compose.yml     # Local deployment
├── Makefile              # Build commands
├── .golangci.yml         # Linter config
├── .env.example          # Environment template
├── DEPLOYMENT.md         # Deployment guide
└── CI-CD-SETUP.md        # This setup guide
```

## 🎯 Next Steps After Setup

1. ✅ Update USERNAME in README badges
2. ✅ Push to GitHub
3. ✅ Watch Actions run
4. ✅ Configure secrets (optional)
5. ✅ Create first release
6. ✅ Deploy to production

## 📚 Documentation

- **Workflows**: `.github/WORKFLOWS.md`
- **Deployment**: `DEPLOYMENT.md`
- **Testing**: `TESTING.md`
- **API Docs**: `API_DOCS.md`
- **Setup**: `CI-CD-SETUP.md`

---

**Quick Links:**
- 📊 [Actions](../../actions)
- 🏷️ [Releases](../../releases)
- 📦 [Packages](../../packages)
- 🔒 [Security](../../security)
