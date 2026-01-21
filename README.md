# FreeStealer

[![Test & Coverage](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/test.yml/badge.svg)](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/test.yml)
[![Build](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/build.yml/badge.svg)](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/build.yml)
[![Deploy to Leapcell](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/deploy-leapcell.yml/badge.svg)](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/deploy-leapcell.yml)
[![CodeQL](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/codeql.yml/badge.svg)](https://github.com/YOUR_USERNAME/go-kasir-api/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_USERNAME/go-kasir-api)](https://goreportcard.com/report/github.com/YOUR_USERNAME/go-kasir-api)
[![codecov](https://codecov.io/gh/YOUR_USERNAME/go-kasir-api/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_USERNAME/go-kasir-api)
[![License](https://img.shields.io/github/license/YOUR_USERNAME/go-kasir-api)](LICENSE)
[![Release](https://img.shields.io/github/v/release/YOUR_USERNAME/go-kasir-api)](https://github.com/YOUR_USERNAME/go-kasir-api/releases)

A comprehensive REST API for tracking and sharing free tier information for hosting platforms like Railway, Koyeb, Vercel, and more. Built with Go, featuring GitHub OAuth authentication, voting system, and community comments.

## ✨ Features

- 🔐 **GitHub OAuth Authentication** - Secure login with GitHub
- 📊 **Free Tier Database** - Track hosting platforms' free tier offerings
- 👍 **Voting System** - Upvote/downvote tiers with toggle support
- 💬 **Comments** - Share experiences and tips (max 100 chars)
- 🔒 **Public/Private Tiers** - Control visibility of your submissions
- 📖 **Swagger Documentation** - Interactive API documentation
- 🚀 **Fast & Lightweight** - Pure Go SQLite driver, no CGO required
- 🔄 **Hot Reload** - Development with Air
- 🐳 **Docker Support** - Easy deployment with Docker
- 📦 **Multi-platform** - Binaries for Linux, Windows, macOS

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/go-kasir-api.git
cd go-kasir-api

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your GitHub OAuth credentials
# Get credentials from: https://github.com/settings/developers
```

### Running Locally

```bash
# Run the application
go run .

# Or with hot reload (development)
air

# Or using Make
make dev
```

The API will be available at `http://localhost:8080`

### Using Docker

```bash
# Build image
docker build -t freestealer .

# Run container
docker run -p 8080:8080 \
  -e GITHUB_CLIENT_ID=your_id \
  -e GITHUB_CLIENT_SECRET=your_secret \
  -e SESSION_SECRET=your_secret \
  freestealer
```

## 📚 API Documentation

Once running, access the Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

### Key Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/users` | Create a new user |
| GET | `/users` | Get all users |
| POST | `/tiers` | Create a new tier |
| GET | `/tiers` | Get tiers (with filters) |
| GET | `/tiers/:id` | Get specific tier |
| PUT | `/tiers/:id` | Update tier |
| DELETE | `/tiers/:id` | Delete tier |
| POST | `/votes` | Vote on a tier |
| POST | `/comments` | Add comment |
| GET | `/comments` | Get comments for tier |
| GET | `/auth/github` | Start GitHub OAuth |
| GET | `/auth/github/callback` | OAuth callback |
| GET | `/auth/me` | Get current user |
| GET | `/auth/logout` | Logout |

See [API_DOCS.md](API_DOCS.md) for detailed documentation.

## 🧪 Testing

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Generate coverage report
./scripts/coverage.ps1  # Windows
./scripts/coverage.sh   # Linux/Mac

# Or use Make
make test
make coverage
make coverage-html
```

**Current Coverage:**
- Database: 87.5%
- Handlers: 43.1%
- Overall: 31.4%

See [TESTING.md](TESTING.md) for more details.

## 🏗️ Project Structure

```
.
├── .github/
│   └── workflows/          # GitHub Actions CI/CD
│       ├── test.yml        # Test & coverage
│       ├── build.yml       # Multi-platform builds
│       ├── deploy-leapcell.yml  # Auto-deploy
│       ├── release.yml     # Release automation
│       └── codeql.yml      # Security analysis
├── auth/                   # Authentication logic
├── database/               # Database initialization
├── docs/                   # Swagger documentation
├── handlers/               # HTTP handlers
├── models/                 # Data models
├── scripts/                # Utility scripts
├── main.go                 # Application entry point
├── Dockerfile              # Docker configuration
├── Makefile                # Build automation
└── README.md
```

## 🚢 Deployment

### Automated Deployment (Recommended)

The project includes GitHub Actions workflows for automated deployment:

1. **Leapcell.io** - Automatic deployment on push to `main`
2. **Docker** - Automated image builds on releases
3. **GitHub Releases** - Automated binary releases on tags

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

### Quick Deploy to Leapcell

1. Set up GitHub secrets:
   - `LEAPCELL_TOKEN`
   - `LEAPCELL_PROJECT_ID`
2. Push to `main` branch
3. Deployment happens automatically! 🎉

### Manual Deployment

```bash
# Build for production
CGO_ENABLED=0 go build -ldflags="-s -w" -o freestealer .

# Run
./freestealer
```

## 🔧 Development

### Setup Development Environment

```bash
# Install Air for hot reload
go install github.com/air-verse/air@latest

# Run in development mode
air
```

### Generate Swagger Documentation

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Run security scan
gosec ./...
```

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- Tests pass (`make test`)
- Code is formatted (`go fmt ./...`)
- Coverage doesn't decrease significantly

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [GORM](https://gorm.io/) - ORM library
- [Chi](https://github.com/go-chi/chi) - HTTP router
- [Goth](https://github.com/markbates/goth) - OAuth library
- [Swaggo](https://github.com/swaggo/swag) - Swagger generation
- [Air](https://github.com/air-verse/air) - Hot reload
- [Logrus](https://github.com/sirupsen/logrus) - Logging

## 📧 Contact

For questions or support, please open an issue on GitHub.

## 🗺️ Roadmap

- [ ] Add more OAuth providers (Google, GitLab)
- [ ] Implement notification system
- [ ] Add tier comparison features
- [ ] Support tier history/versioning
- [ ] Add API rate limiting
- [ ] Implement full-text search
- [ ] Add export functionality (JSON, CSV)
- [ ] Create web frontend
- [ ] Add GraphQL API
- [ ] Implement caching layer

---

**Note:** Replace `YOUR_USERNAME` in the badges and links with your actual GitHub username.

Made with ❤️ using Go
