# AGENTS.md - 1Channel Project Guide

## Project Overview

**1Channel** is a conversational multi-agent, multi-channel CRM system built with Go. It's platform-agnostic and designed as a modular monolith following Clean Architecture and Domain-Driven Design (DDD) principles.

## Architecture & Structure

### Directory Layout
```
├── cmd/api/                    # Application entrypoint (main.go)
├── internal/                   # Core business logic
│   └── module/                 # Each module follows this structure
│       ├── api.go              # Public interface (shared with other modules, low coupling)
│       ├── app/                # Use cases (application services)
│       ├── domain/             # Aggregates, entities, value objects, repository interfaces
│       └── infra/              # Implementations (DB, HTTP handlers, event handlers)
├── db/                         # Database layer
│   ├── migrations/             # SQL migration files (golang-migrate)
│   └── queries/                # SQLC query definitions
├── openspec/                   # Specification-driven development
│   ├── specs/                  # Domain specifications
│   └── changes/                # Change requests and archives
├── docker-compose.yml          # Docker services (app + postgres)
└── .github/workflows/          # CI/CD pipelines
```

### Module Design Principles
- **api.go**: Defines the public interface that other modules can depend on (prevents tight coupling)
- **Low coupling**: Modules communicate through interfaces, not direct dependencies
- **High cohesion**: Each module encapsulates its own domain logic

### Key Technical Decisions
- **Language**: Go 1.27
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Migrations**: golang-migrate/migrate
- **Architecture**: Modular monolith with Clean Architecture layers
- **API**: RESTful HTTP API (port 8080 by default)
- **Containerization**: Docker + Docker Compose

## Development Commands

### Docker Environment
```bash
# Start services (app + postgres)
docker compose up -d

# Stop services
docker compose down

# View logs
docker compose logs -f app

# Shell into container
docker compose exec app sh

# Rebuild after changes
docker compose up -d --build
```

### Essential Commands (inside container or local)
```bash
# Build and run
go build ./cmd/api              # Build the application
go run ./cmd/api                # Run directly

# Quality gates (MUST pass locally before push)
go mod verify                   # Verify dependencies
gofmt -l .                      # Check formatting (must be clean)
go vet ./...                    # Static analysis
go build ./...                  # Verify compilation
go test -race -count=1 ./...    # Run tests with race detection

# Database (SQLC)
sqlc generate                   # Generate Go code from SQL queries

# Migrations (golang-migrate)
migrate create -ext sql -dir db/migrations -seq migration_name
migrate -path db/migrations -database "$POSTGRES_URL" up
migrate -path db/migrations -database "$POSTGRES_URL" down
```

### Single Test Execution
```bash
go test -run TestFunctionName ./path/to/package
go test -v -count=1 ./...      # Verbose output
```

## Environment Setup

### Required Environment Variables
Copy `.env.example` to `.env` and configure:

```bash
# Application
HTTP_PORT=8080
LOG_LEVEL=debug

# Security
ONECHANNEL_SECRET=your-secret
META_SECRET=your-meta-secret

# WhatsApp Business API
WHATSAPP_PHONE_NUMBER_ID=your-phone-id
WHATSAPP_ACCESS_TOKEN=your-access-token

# Database (for Docker Compose)
POSTGRES_HOST=postgres          # Service name in docker-compose
POSTGRES_PORT=5432
POSTGRES_DATABASE=1channel_dev
POSTGRES_USERNAME=dev
POSTGRES_PASSWORD=dev
POSTGRES_URL=postgres://dev:dev@postgres:5432/1channel_dev?sslmode=disable
```

### Database Setup
1. Start services: `docker compose up -d`
2. Run migrations: `migrate -path db/migrations -database "$POSTGRES_URL" up`
3. Generate SQLC code: `sqlc generate`

## CI/CD Pipeline

### Quality Gates (GitHub Actions)
The CI runs these checks in order:
1. `go mod verify` - Dependency integrity
2. `gofmt -l .` - Code formatting (must be gofmt-clean)
3. `go vet ./...` - Static analysis
4. `go build ./...` - Compilation check
5. `go test -race -count=1 ./...` - Tests with race detection

### Docker Build
- Multi-stage build with Go 1.27-alpine
- Final image: Alpine 3.24 with minimal footprint
- Exposes port 8080
- Runs as non-root user `app`

## Git Workflow (Git Flow - Manual)

### Branch Strategy
```
main              # Production-ready code
develop           # Integration branch
feature/*         # New features (branch from develop)
release/*         # Release preparation (branch from develop)
hotfix/*          # Production fixes (branch from main)
```

### Common Workflows
```bash
# Start new feature
git checkout develop
git pull origin develop
git checkout -b feature/my-feature

# Finish feature
git checkout develop
git merge --no-ff feature/my-feature
git push origin develop
git branch -d feature/my-feature

# Start release
git checkout develop
git checkout -b release/v1.0.0

# Finish release
git checkout main
git merge --no-ff release/v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"
git checkout develop
git merge --no-ff release/v1.0.0
git branch -d release/v1.0.0

# Hotfix
git checkout main
git checkout -b hotfix/critical-fix
# ... fix ...
git checkout main
git merge --no-ff hotfix/critical-fix
git checkout develop
git merge --no-ff hotfix/critical-fix
git branch -d hotfix/critical-fix
```

## Testing Strategy

### Test Framework & Patterns
- **Library**: testify (`github.com/stretchr/testify`)
- **Pattern**: AAA (Arrange-Act-Assert)
- **Style**: Table-driven tests for parameterized testing

### Test Example
```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        want    *User
        wantErr bool
    }{
        {
            name: "valid input",
            input: CreateUserInput{
                Name:  "John Doe",
                Email: "john@example.com",
            },
            want:    &User{Name: "John Doe", Email: "john@example.com"},
            wantErr: false,
        },
        {
            name:    "empty name",
            input:   CreateUserInput{Name: "", Email: "john@example.com"},
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            svc := NewUserService(mockRepo)

            // Act
            got, err := svc.Create(context.Background(), tt.input)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Test Commands
```bash
# All tests
go test ./...

# Specific package
go test ./internal/module/...

# Single test
go test -run TestFunctionName ./internal/module/...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Development Conventions

### Code Style
- **Formatting**: Must be `gofmt`-clean (CI enforces this)
- **Imports**: Group standard library, external, internal packages
- **Error handling**: Return errors, don't panic
- **Naming**: Follow Go conventions (camelCase for private, PascalCase for public)

### Architecture Patterns
- **Domain Layer**: Pure business logic, no dependencies
- **Application Layer**: Use cases, orchestrates domain objects
- **Infrastructure Layer**: Implements interfaces defined in domain
- **Interface Layer**: HTTP handlers, request/response mapping

### Database Conventions
- **Migrations**: Sequential numbered files in `db/migrations/` (format: `000001_name.up.sql` / `000001_name.down.sql`)
- **Queries**: Type-safe SQL in `db/queries/` with SQLC
- **Generated Code**: Never edit generated SQLC files manually

## Common Pitfalls & Solutions

### 1. Formatting Errors
**Problem**: CI fails on `gofmt` check
**Solution**: Run `gofmt -w .` before committing

### 2. SQLC Regeneration
**Problem**: Database code out of sync
**Solution**: After changing SQL queries, run `sqlc generate`

### 3. Import Cycles
**Problem**: Go doesn't allow import cycles
**Solution**: Follow Clean Architecture layer dependencies (domain ← application ← infrastructure ← interfaces)

### 4. Environment Variables
**Problem**: Application fails to start
**Solution**: Ensure `.env` file exists with all required variables (see `.env.example`)

### 5. Docker Networking
**Problem**: App can't connect to database
**Solution**: Use Docker service name (`postgres`) as host, not `localhost`

## Working with OpenSpec

This project uses OpenSpec for specification-driven development:
- **Specifications**: Domain specs in `openspec/specs/`
- **Changes**: Change requests in `openspec/changes/`
- **Commands**: Use OpenSpec skills for managing changes

### OpenSpec Workflow
1. Create specification in `openspec/specs/`
2. Create change request in `openspec/changes/`
3. Implement changes following specs
4. Archive completed changes

## Quick Reference

### Build & Run (Docker)
```bash
docker compose up -d            # Start services
docker compose logs -f app      # View logs
docker compose down             # Stop services
```

### Quality Gates (run before push)
```bash
go mod verify && gofmt -l . && go vet ./... && go build ./... && go test -race -count=1 ./...
```

### Database
```bash
# Run migrations
migrate -path db/migrations -database "$POSTGRES_URL" up

# Create migration
migrate create -ext sql -dir db/migrations -seq migration_name

# Generate SQLC code
sqlc generate
```

### Git Flow
```bash
# Feature
git checkout -b feature/name develop

# Release
git checkout -b release/v1.0 develop

# Hotfix
git checkout -b hotfix/name main
```

## Important Notes

- **Never commit `.env` files** - they're in `.gitignore`
- **Always run quality gates locally** before pushing
- **SQLC generates code** - don't edit generated files manually
- **Follow Clean Architecture** - respect layer boundaries
- **Use OpenSpec** for feature specifications and change management
- **Run migrations** before starting the application
- **Use Docker service names** for inter-container communication

---

*This guide is maintained for AI agents working on the 1Channel project. Last updated: 2026-09-24*
