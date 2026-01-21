# Deployment Guide

This guide covers deploying FreeStealer to various platforms and using the automated CI/CD workflows.

## Table of Contents

- [GitHub Actions Workflows](#github-actions-workflows)
- [Deploying to Leapcell.io](#deploying-to-leapcellIO)
- [Docker Deployment](#docker-deployment)
- [Manual Deployment](#manual-deployment)
- [Release Process](#release-process)

## GitHub Actions Workflows

The project includes several automated workflows:

### 1. Test & Coverage (`test.yml`)

Runs on every push and pull request to `main` and `develop` branches.

**Features:**
- Runs all tests with race detection
- Generates coverage reports
- Uploads to Codecov
- Enforces minimum 30% coverage threshold

**Secrets Required:**
- `CODECOV_TOKEN` (optional): For uploading to Codecov

### 2. Build (`build.yml`)

Builds the application for multiple platforms.

**Platforms:**
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64)

**Features:**
- Cross-platform compilation
- Build artifact retention (7 days)
- Code linting with golangci-lint

### 3. Deploy to Leapcell (`deploy-leapcell.yml`)

Automatically deploys to Leapcell.io when pushing to `main` branch.

**Secrets Required:**
- `LEAPCELL_TOKEN`: Your Leapcell API token
- `LEAPCELL_PROJECT_ID`: Your Leapcell project ID

### 4. Release (`release.yml`)

Creates GitHub releases with pre-built binaries.

**Triggers:**
- Push tags matching `v*` (e.g., `v1.0.0`)
- Manual workflow dispatch

**Features:**
- Multi-platform binary builds
- Automatic changelog generation
- Docker image publication to GitHub Container Registry
- Release artifact upload

### 5. CodeQL Security Analysis (`codeql.yml`)

Runs security analysis on the codebase.

**Schedule:**
- Every Monday at midnight
- On pull requests

## Deploying to Leapcell.io

### Prerequisites

1. Create a Leapcell account at https://leapcell.io
2. Create a new project
3. Get your API token and project ID

### Setup GitHub Secrets

1. Go to your GitHub repository settings
2. Navigate to Secrets and Variables → Actions
3. Add the following secrets:
   - `LEAPCELL_TOKEN`: Your Leapcell API token
   - `LEAPCELL_PROJECT_ID`: Your project ID

### Automatic Deployment

Push to the `main` branch:

```bash
git push origin main
```

The deployment workflow will automatically:
1. Build the application
2. Create deployment package
3. Deploy to Leapcell
4. Verify deployment

### Manual Deployment

Trigger the workflow manually:

1. Go to Actions → Deploy to Leapcell
2. Click "Run workflow"
3. Select branch and run

### Environment Variables on Leapcell

Set these environment variables in your Leapcell project dashboard:

```bash
PORT=8080
DATABASE_PATH=/data/kasir.db
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret
GITHUB_CALLBACK_URL=https://your-app.leapcell.app/auth/github/callback
SESSION_SECRET=your_session_secret
```

## Docker Deployment

### Build Docker Image

```bash
docker build -t freestealer:latest .
```

### Run Locally

```bash
docker run -p 8080:8080 \
  -e GITHUB_CLIENT_ID=your_id \
  -e GITHUB_CLIENT_SECRET=your_secret \
  -e SESSION_SECRET=your_secret \
  freestealer:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  freestealer:
    image: freestealer:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_PATH=/data/kasir.db
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - SESSION_SECRET=${SESSION_SECRET}
    volumes:
      - ./data:/data
    restart: unless-stopped
```

Run:

```bash
docker-compose up -d
```

### Pull from GitHub Container Registry

After a release:

```bash
docker pull ghcr.io/your-username/freestealer:latest
docker run -p 8080:8080 ghcr.io/your-username/freestealer:latest
```

## Manual Deployment

### Build from Source

```bash
# Clone repository
git clone https://github.com/your-username/freestealer.git
cd freestealer

# Build
go build -ldflags="-s -w" -o freestealer .

# Run
./freestealer
```

### Linux Server (systemd)

Create `/etc/systemd/system/freestealer.service`:

```ini
[Unit]
Description=FreeStealer API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/freestealer
ExecStart=/opt/freestealer/freestealer
Restart=always
RestartSec=3
Environment="PORT=8080"
Environment="DATABASE_PATH=/var/lib/freestealer/kasir.db"
EnvironmentFile=/etc/freestealer/.env

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable freestealer
sudo systemctl start freestealer
sudo systemctl status freestealer
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Release Process

### Create a Release

1. **Update Version**: Update version in relevant files
2. **Update Changelog**: Document changes in CHANGELOG.md
3. **Commit Changes**:
   ```bash
   git add .
   git commit -m "chore: bump version to v1.0.0"
   ```

4. **Create and Push Tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

5. **Automatic Process**:
   - GitHub Actions builds binaries for all platforms
   - Creates GitHub release with binaries
   - Builds and publishes Docker image
   - Updates release notes

### Manual Release

Use workflow dispatch:

1. Go to Actions → Release
2. Click "Run workflow"
3. Enter version (e.g., `v1.0.0`)
4. Run workflow

### Download Release Assets

Users can download pre-built binaries from:
```
https://github.com/your-username/freestealer/releases/latest
```

## Platform-Specific Notes

### Leapcell.io

- Automatically scales based on traffic
- Provides HTTPS by default
- Handles database persistence in `/data` directory
- Zero-downtime deployments

### Heroku

Create `Procfile`:
```
web: ./freestealer
```

Deploy:
```bash
heroku create your-app-name
git push heroku main
```

### Railway

Railway auto-detects Go applications. Just connect your repository.

### Vercel

Not recommended for this backend API. Use Leapcell, Railway, or Heroku instead.

## Monitoring

### Health Check Endpoint

```bash
curl http://localhost:8080/health
```

### Logs

View logs in your deployment platform:

- **Leapcell**: Dashboard → Logs
- **Docker**: `docker logs freestealer`
- **Systemd**: `journalctl -u freestealer -f`

## Troubleshooting

### Build Failures

1. Check Go version (requires 1.23+)
2. Verify all dependencies: `go mod verify`
3. Clear build cache: `go clean -cache`

### Deployment Failures

1. Check secrets are set correctly
2. Verify environment variables
3. Check platform-specific logs
4. Ensure port 8080 is available

### Database Issues

1. Ensure write permissions for database directory
2. Check DATABASE_PATH environment variable
3. Verify SQLite is properly initialized

## Support

For deployment issues:
- Check GitHub Actions logs
- Review platform-specific documentation
- Open an issue on GitHub
