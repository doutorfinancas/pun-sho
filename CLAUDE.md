# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**pun-sho** (PUNy-SHOrtener) is a URL shortener microservice in Go. It creates short links with QR code generation, visit tracking, TTL management, redirection limits, and label filtering.

## Common Commands

```bash
# Run all unit tests
make test/go

# Run unit tests with coverage report
make test/cover

# Run all tests (unit + HTTP integration via Docker)
make test

# Run a single test
go test -v -run TestFunctionName ./path/to/package/...

# Generate mocks (after changing generator/faker.go)
make generate/mocks

# Generate Swagger docs
make docs

# Start local PostgreSQL
make postgres

# Database migrations
make migration/create   # Create new migration
make migration/up       # Apply migrations
make migration/clean    # Rollback migrations

# Install pre-commit hooks (golangci-lint + goimports)
make hook/setup
```

## Architecture

Layered architecture with dependency injection wired in `main.go`:

```
api/          → HTTP handlers (Gin framework), middleware, request/response types
service/      → Business logic (ShortyService, QRCodeService)
entity/       → Domain models + repository implementations (GORM)
database/     → Generic GORM repository abstraction
```

**Request flow**: Gin handler → Service → Repository → GORM → PostgreSQL

## Development Principles

- **SOLID** — Single responsibility per package/struct. Depend on interfaces, not concrete types (see handler/service/repository boundaries).
- **DRY** — Extract shared logic into the existing utility packages (`str/`, `buf/`, `database/`) rather than duplicating across handlers or services.
- **YAGNI** — Don't add abstractions, config options, or endpoints until there's a concrete need. This is an MVP (v0.X).
- **KISS** — Prefer straightforward implementations. The codebase favors simple patterns (flat package structure, explicit wiring in `main.go`) over frameworks or code generation.

### Key patterns

- **Handler registration**: Handlers implement `HTTPHandler` interface with `Routes(*gin.RouterGroup)` and `Group() *string`
- **Authentication**: Simple bearer token via `Authorization` header, configured by `AUTH_TOKEN` env var
- **Soft deletes**: Links use `deleted_at` timestamp, never hard-deleted
- **Visit status tracking**: Each redirect attempt records a status (redirected, expired, limit_reached, blocked, deleted)
- **Bot filtering**: Social media bots are allowed via `ALLOWED_SOCIAL_BOTS` config; other bots are blocked with precise prefix matching

### Routes

- `GET /:slug` — Public redirect endpoint (api/url.go)
- `GET/POST/PATCH/DELETE /api/v1/short[/:id]` — CRUD for short links (api/shortener.go), auth required
- `POST /api/v1/preview` — QR code preview without persistence (api/preview.go)

## Testing

- Unit tests are colocated with implementation files (e.g., `service/shorty_test.go`)
- `test/` directory contains test utilities: mock DB setup (`NewMockDB`, `NewMockRepository`), mock row generators
- Uses `go-sqlmock` for database mocking and `testify/assert` for assertions
- Tests require `TEST_MODE=full` env var (set automatically by Makefile targets)
- Linting: golangci-lint with `.golangci.yml` config (line length 140, cyclomatic complexity 15)

## Configuration

All configuration via environment variables (see `.env.example`). Key vars: `API_PORT`, `AUTH_TOKEN`, `HOST_NAME`, `DB_ADAPTOR` (cockroach|postgres), `PUBLIC_ID_LENGTH`, `ALLOWED_SOCIAL_BOTS`, `QR_PNG_LOGO`.

## Database

- PostgreSQL (or CockroachDB) via GORM
- Migrations in `db_migrations/` using golang-migrate (sequential numbered SQL files)
- Docker Compose provides local PostgreSQL on port 5432
