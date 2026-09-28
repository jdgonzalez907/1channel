# AGENTS.md - 1Channel Project Guide

## Project Overview

**1Channel** is a conversational multi-agent, multi-channel CRM system in one monorepo: a Go REST API (`backend/`) and a Vue single-page app (`frontend/`). The backend is platform-agnostic and designed as a modular monolith following Clean Architecture and Domain-Driven Design (DDD) principles.

A contact (customer) talks to the company through a channel; an agent (system user) answers. Conversations are short support/sales threads. The current REST API is internal and first-party (the same person builds front and back), so pragmatic decisions beat over-engineering. Front and back live in the same repo but build, release, and deploy independently.

## Architecture & Structure

### Directory Layout
```
├── backend/                    # REST API (Go)
│   ├── cmd/api/                # Application entrypoint (main.go, router, config)
│   ├── internal/
│   │   ├── modules/            # Business modules (one per bounded context)
│   │   │   └── <module>/       # e.g. contacts, conversations, users
│   │   │       ├── api.go      # Public write API consumed by other modules
│   │   │       ├── app/        # Use cases (application services)
│   │   │       ├── domain/     # Aggregates, entities, value objects, repository ports
│   │   │       └── infra/      # Adapters: pg repositories, http handlers
│   │   └── shared/infra/       # Cross-cutting infra
│   │       ├── http/httperror/ # problem+json helpers
│   │       ├── http/httputil/  # JSON, path params, lookup errors
│   │       ├── http/middleware/# Auth, logging, timeout
│   │       └── pgdb/           # pgx pool, sqlc wrapper, UUID/time helpers
│   ├── db/
│   │   ├── migrations/         # golang-migrate SQL files
│   │   └── queries/            # SQLC query definitions
│   ├── docs/endpoints.md       # API reference
│   ├── docker-compose.yml      # dev: postgres 18 + migrate
│   ├── Dockerfile              # image: ghcr.io/<owner>/<repo>/api
│   └── .env.example            # dev config for the backend
├── frontend/                   # SPA (Vue 3 + TypeScript + Vite + pnpm)
│   ├── src/                    # Vue shell
│   ├── nginx.conf              # serves the SPA and proxies /v1 -> api:8080
│   ├── Dockerfile              # image: ghcr.io/<owner>/<repo>/web
│   └── .env.example            # future VITE_* variables
├── docker-compose.prod.yml     # web + api only (no DB; configured by envs)
├── Makefile                    # single interface (backend + frontend)
├── openspec/                   # Spec-driven development (source of truth)
│   ├── specs/                  # Durable capability specs
│   └── changes/                # In-flight changes + archive
└── .github/workflows/          # CI/CD per component (backend.yml, frontend.yml)
```

### Module Design Principles
- **api.go**: defines the public interface other modules may depend on (write use cases only). Keeps modules decoupled.
- **Low coupling / high cohesion**: modules talk to each other only through their `api.go`, never through each other's `app`/`domain`/`infra`.
- **Reads do not cross modules**: read endpoints resolve their own data straight from SQLC (see below).

### Layer Dependency Rule
```
  domain / app   SHALL NOT import infra
  infra          MAY import domain / app   (infra -> domain is valid)
  shared/infra MAY import a module's domain (infra -> domain)
```
- The forbidden direction is **domain -> infra**. Never the other way around.
- Crossing a *module* boundary still goes through `api.go`, even if the direction (infra -> domain) is allowed.

### Key Technical Decisions
- **Language**: Go 1.27
- **Database**: PostgreSQL 18 with SQLC
- **Migrations**: golang-migrate; in dev they are edited in place and the DB is recreated (nothing is in production yet)
- **Timestamps**: UTC end to end (Postgres `timezone=UTC`, pgx `ScanLocation=time.UTC`, `pgdb.FromTimestamp` normalizes)
- **Architecture**: modular monolith with Clean Architecture layers
- **API**: RESTful HTTP (port 8080)
- **SQLC**: `emit_json_tags: false` — generated structs are row types and are never serialized
- **Reads**: HTTP handler -> `*sqlc.Queries` directly -> Response DTO (no app/domain/repository)
- **Pagination**: explicit `before_sent_at` / `before_id` (+ `next_before_sent_at` / `next_before_id`), not an opaque cursor
- **Containerization**: Docker + Docker Compose
- **Frontend**: Vue 3 + TypeScript + Vite + pnpm (Composition API, `<script setup>`), served by nginx in the `web` image
- **Monorepo**: `backend/` (Go) and `frontend/` (SPA) build and deploy independently (CI filters by `paths`)
- **Deployment**: production compose with `web` + `api` only; the database is external and configured by envs; TLS is terminated by Cloudflare Tunnel (containers are plain HTTP); deploy by commit sha per component
- **API boundary**: `/v1` is the compatibility contract; the front consumes the API same-origin (no CORS)

## Domain Model & Business Rules

### Entities
```
  User (agent)    id, created_at
  Contact         id, external_contact_id (channel id), created_at
  Conversation    id, status, user_id?, contact_id?, created_at, updated_at?,
                  finished_at?, last_message_at, last_message_id, unread_count
  Message         id, conversation_id, status, type, text?, user_id?,
                  contact_id?, external_id?, sent_at, read_at?, edited_at?, deleted_at?
```

### States
```
  conversation: pending | assigned | expired | resolved
  message:      sent | read | deleted | failed
  message type: text   (MVP)
```

### Conversation lifecycle
```
  pending --(agent replies)--> assigned
  pending/assigned --(expires)--> expired
  assigned --(agent resolves)--> resolved
  expired/resolved = finished (finished_at required)
```

### Business rules
- Each conversation has **at most one assigned agent**.
- Every conversation is born with **at least one message** (no empty conversation).
- A contact has **at most one open conversation** (`pending`/`assigned`).
- The contact is resolved or created by its `external_contact_id`.
- A message has **exactly one owner**: agent XOR contact. Text is 1..1000 graphemes.
- Replying to a `pending` conversation with no agent assigns that agent and moves it to `assigned`.
- If the external channel send fails, the agent message becomes `failed` (terminal, blocks edits) and the operation returns an error.
- Edit/delete: only the owner; edits with a timestamp `<=` the last applied edit are discarded; delete is idempotent; the contact may edit/delete even on a finished conversation.
- Marking read: only the assigned agent, only the contact's messages, and it works on finished conversations too.
- Expire and resolve are idempotent. Resolve is only allowed by the assigned agent.
- `updated_at` only moves forward.
- `unread_count` counts the contact's messages with `read_at` null, **including deleted ones** (a deletion is still an interaction by the contact).
- `deleted`: the text is **not** erased in the DB; the API exposes it as `null`.
- Timestamps are always UTC.

### Inbox read model
- Ordered by the last message `(sent_at, id)` descending.
- Visible = `pending` OR assigned to the requester.
- Row: contact (`id`, `external_id`), last-message preview (`text`, `sent_at`, `owner`), `unread_count`.
- Maintained denormalized on `conversations` (`last_message_at`, `last_message_id`, `unread_count`) by `RefreshConversationLastMessage` inside `Save`'s transaction, only when there are modified messages.

## REST API

### Write (`http-api`)
```
  POST   /v1/users                              -> 201 {id}          (public, bootstrap)
  POST   /v1/conversations/{id}/messages        -> 201 {id}
  PATCH  /v1/conversations/{id}/messages        -> 204   ({"status":"read"})
  PATCH  /v1/conversations/{id}                 -> 204   ({"status":"resolved"})
```
An agent cannot `expired` via HTTP (422).

### Read (`http-read-api`)
```
  GET /v1/conversations[?status=open|finished&external_contact_id=&before_*]
  GET /v1/conversations/{id}[?before_*]
  GET /v1/contacts/{id}
  GET /v1/users/{id}
```

### Auth & errors
- `Authorization: Bearer <user id>`; missing/malformed/nonexistent user -> 401. Existence is validated once in `Auth`.
- Errors use `Content-Type: application/problem+json` with `title`, `status`, `detail`, `instance`.
- Code mapping: 400 malformed/position, 401 auth, 403 ownership, 404 missing, 409 state conflict, 422 validation, 500 internal, 504 timeout.

## Development Commands

### Docker Environment
Compose raises data only; the app runs on the host with `go run` for fast iteration. The dev compose lives in `backend/` and is driven from the root `Makefile`.

```bash
make up                    # Start postgres 18 (UTC) in background
make down                  # Stop services
docker compose -f backend/docker-compose.yml --project-directory backend ps
docker compose -f backend/docker-compose.yml --project-directory backend logs -f postgres
make migrate-up            # Apply migrations on demand
make reset                 # Recreate DB from scratch
make seed                  # Load dev test data (SOLO dev; DESTRUCTIVE)
```

The frontend runs separately with Vite: `make run-web` (proxies `/v1` to `localhost:8080`). `make run` launches both backend and frontend.

If `docker` asks for permissions: `make DOCKER="sudo docker" up` (or add your user to the `docker` group). Note: with passworded sudo, `make` may not work non-interactively; the `migrate` CLI against `localhost:5432` is the fallback (see below).

### Dev Seed Data

`make seed` runs `backend/db/seed/dev_seed.sql` inside the `postgres` container. It is a plain
SQL script (not a migration and never for production) that **truncates** `users`,
`contacts`, `conversations` and `messages`, then reloads realistic test data: 3 agents,
50 contacts, 90 conversations (30 open / 60 finished) and 3000 messages. It aborts if the
target database name does not contain `dev`. Requires the database to be migrated and the
`postgres` service running. Re-run it freely; it always rebuilds from scratch.

### Essential Commands (host or container)
```bash
# Backend (module lives in backend/)
go -C backend build ./...       # Compile everything
go -C backend run ./cmd/api     # Run the app (needs .env sourced; or: make run-api)
go -C backend build ./cmd/api   # Build only the binary

# Quality gates (MUST pass locally before push)
go -C backend mod verify
gofmt -l backend                # Must be clean
go -C backend vet ./...
go -C backend build ./...
go -C backend test -race -count=1 ./...

# Frontend
pnpm --dir frontend install
pnpm --dir frontend type-check
pnpm --dir frontend build

# SQLC (regenerate after changing backend/db/queries/*.sql; run from backend/)
sqlc generate

# Migrations
make migrate-up
make migrate-down
make migrate-create NAME=migration_name
```

`sqlc` and `migrate` may not be on `PATH`: they live in `$(go env GOPATH)/bin`.

### Migrations without Docker
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
DSN="postgres://dev:dev@localhost:5432/1channel_dev?sslmode=disable"
migrate -path backend/db/migrations -database "$DSN" up
migrate -path backend/db/migrations -database "$DSN" down 2   # then up to re-apply edited files
```

## Environment Setup

### Required Environment Variables
Copy `backend/.env.example` to `backend/.env`. The app does not load `.env` itself: `make run-api` sources it, and Docker Compose autoloads it (from `backend/`) for interpolation. The frontend has its own `frontend/.env` for future `VITE_*` variables.

```bash
HTTP_PORT=8080
LOG_LEVEL=debug
TZ=UTC

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DATABASE=1channel_dev
POSTGRES_USERNAME=dev
POSTGRES_PASSWORD=dev
```

`POSTGRES_URL` is an optional DSN override; when set it takes precedence over the `POSTGRES_*` parts.

Integration variables (`META_*`, `WHATSAPP_*`, `ONECHANNEL_SECRET`) are intentionally absent until the webhook and the real sender exist.

### Database Setup
1. Start PostgreSQL: `make up`
2. Run migrations: `make migrate-up`
3. Generate SQLC code: `sqlc generate`

## CI/CD Pipeline

### Quality Gates (GitHub Actions)
Two workflows run independently, filtered by `paths`:

- **`backend.yml`** (on `backend/**`): `go mod verify`, `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test -race -count=1 ./...` (run with `working-directory: backend`).
- **`frontend.yml`** (on `frontend/**`): `pnpm install --frozen-lockfile`, `pnpm type-check`, `pnpm build`.

CI has **no database**: tests must not require one. A commit that only touches one component does not rebuild the other.

### Docker Build
- **`api`** (`backend/Dockerfile`): multi-stage with Go 1.27-alpine; final image Alpine 3.24, non-root user `app`, port 8080.
- **`web`** (`frontend/Dockerfile`): multi-stage with node:24-alpine (pnpm build) → nginx:1.27-alpine serving `dist/` and proxying `/v1` to `api:8080`.

### Production Deploy
`docker-compose.prod.yml` runs only `web` + `api`. Images are versioned by sha per component; the database is external and configured by envs injected at compose time (a `.env` on the host, next to the compose file). TLS is terminated by Cloudflare Tunnel, so containers speak plain HTTP (`web` listens on `:80`, `api` stays internal on `:8080`).

```bash
# 1) Migrate FIRST (manual; the host has no `migrate` binary)
docker run --rm -v "$PWD/backend/db/migrations:/migrations" \
  migrate/migrate -path=/migrations -database "$POSTGRES_URL" up

# 2) Deploy one component independently (change the sha, recreate only that service)
API_TAG=<sha> docker compose -f docker-compose.prod.yml up -d --no-deps api
WEB_TAG=<sha> docker compose -f docker-compose.prod.yml up -d --no-deps web
```

## Git Workflow (Git Flow - Manual)

### Branch Strategy
```
main              # Production-ready
develop           # Integration
feature/*         # from develop
release/*         # from develop
hotfix/*          # from main
```

### Common Workflows
```bash
# Feature
git checkout develop && git pull origin develop
git checkout -b feature/my-feature
# ... finish ...
git checkout develop && git merge --no-ff feature/my-feature
git push origin develop && git branch -d feature/my-feature

# Release
git checkout -b release/v1.0.0 develop
git checkout main && git merge --no-ff release/v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"
git checkout develop && git merge --no-ff release/v1.0.0

# Hotfix
git checkout -b hotfix/critical-fix main
git checkout main && git merge --no-ff hotfix/critical-fix
git checkout develop && git merge --no-ff hotfix/critical-fix
```

## Testing Strategy

### Principles
- **Unit tests only.** No integration tests, no tests that need a database. CI has no Postgres.
- **Library**: testify; **pattern**: AAA; **style**: table-driven.
- **Writes**: the handler depends on `app` use-case interfaces; those are mocked with testify.
- **Reads**: no interfaces and no mocks (infra -> infra). Only the pure logic is unit-tested: tab/status parsing, visibility, pagination position parsing, and row->Response mapping.
- The read-model refresh (`RefreshConversationLastMessage`) and migrations are not covered by automated tests; verify them manually.

### Test Example
```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        want    *User
        wantErr bool
    }{
        {name: "valid input", input: CreateUserInput{Name: "John Doe", Email: "john@example.com"}, wantErr: false},
        {name: "empty name", input: CreateUserInput{Name: "", Email: "john@example.com"}, wantErr: true},
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
go -C backend test ./...
go -C backend test ./internal/modules/...
go -C backend test -run TestFunctionName ./internal/modules/...
go -C backend test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Development Conventions

### Code Style
- **Formatting**: `gofmt`-clean (CI enforces it)
- **Imports**: standard library, external, internal groups
- **Errors**: return, don't panic
- **Naming**: Go conventions

### Handler & DTO Conventions
```
  Files:   <x>_write_handler.go / <x>_read_handler.go
           write_dto.go / read_dto.go
  Types:   XWriteHandler / XReadHandler
           Request (input) / Response (output)   -- never the "DTO" suffix
  Handlers are symmetric write/read; reads receive `*sqlc.Queries` directly.
```

### Shared Infra Helpers (reuse, don't duplicate)
```go
httputil.RequirePathUUID(w, r, "id", "invalid X id")  // parse {id} else 400
httputil.LookupError(w, r, err, "X not found")        // IsNoRows -> 404, else 500
middleware.RequireUserID(w, r)                         // id from context else 401
middleware.Auth(lookup)                                // bearer + existence -> 401
pgdb.IsNoRows(err)                                     // errors.Is(err, pgx.ErrNoRows)
pgdb.UUIDString / UUIDStringPtr / FromTimestamp / FromTimestampPtr
```
Rule of thumb: if a block repeats across two or more handlers, move it to `shared/infra`.

### Query Naming
- SQLC query names describe the data operation; **no `read`/`write` words** (Postgres doesn't know about that).

### Timestamps in Responses
- Response DTOs use `time.Time` / `*time.Time`; `encoding/json` serializes RFC3339, and `pgdb` normalizes to UTC. **Do not** call `.Format(...)` manually.

### Architecture Patterns
- **Domain**: pure business logic; no infra imports
- **Application**: use cases orchestrating the domain
- **Infrastructure**: adapters (pg repos, http handlers)
- **Reads**: handlers query SQLC directly and map to Response

## Working Style (read this before changing anything)

- **MVP**: prefer the direct solution. Don't add layers, abstractions, or "future-proofing" that isn't needed.
- **infra -> infra = no interfaces**: read handlers depend on `*sqlc.Queries`. No interfaces/mocks for reads.
- **Unit tests only**: no integration/DB tests. Don't reintroduce them without being asked.
- **First-party API**: one person controls front and back, so ergonomics favor explicit parameters (e.g., `before_sent_at`/`before_id`) over opaque tokens.
- **Ask before expanding scope**: don't silently narrow, defer, or simplify specified behavior. If a task needs more than the spec describes, surface it.
- **Clean Architecture direction**: `infra -> domain` is fine; `domain -> infra` is not.

## Common Pitfalls & Solutions

### 1. Formatting Errors
**Problem**: CI fails on `gofmt`.
**Solution**: `gofmt -w .` before committing.

### 2. SQLC Regeneration
**Problem**: DB code out of sync.
**Solution**: after changing `backend/db/queries/*.sql`, run `sqlc generate` from `backend/`. Never edit generated files.

### 3. Import Cycles / Layer Violations
**Problem**: domain importing infra.
**Solution**: `domain`/`app` must not import `infra`. Cross-module only via `api.go`.

### 4. Environment Variables
**Problem**: app fails to start.
**Solution**: `backend/.env` must exist (see `backend/.env.example`); `make run-api` sources it.

### 5. Database Connectivity
**Problem**: app can't connect.
**Solution**: host uses `POSTGRES_HOST=localhost`; `postgres` is only the Compose-internal name.

### 6. Edited Migrations Not Applied
**Problem**: editing an already-applied migration has no effect.
**Solution**: recreate the DB (`make reset`, or `migrate down <n>` then `up`). Dev only.

## Working with OpenSpec

`openspec/specs/*` is the **source of truth** for behavior. Read the relevant spec before changing behavior; `AGENTS.md` is a summary, not a replacement.

- **Specs**: durable capabilities in `openspec/specs/`
- **Changes**: proposals/design/tasks in `openspec/changes/` (+ archive)
- **Workflow**: propose -> apply -> archive; sync delta specs into main specs on archive.

## Quick Reference

### Build & Run
```bash
make up                         # Start postgres 18 (UTC)
make migrate-up                 # Apply migrations
make run                        # Backend + frontend at once
make run-api                    # Backend only (go run ./cmd/api, sourcing backend/.env)
make run-web                    # Frontend only (Vite, proxies /v1 to :8080)
make down                       # Stop services
```

### Quality Gates
```bash
# Backend (from repo root)
go -C backend mod verify && gofmt -l backend && go -C backend vet ./... && go -C backend build ./... && go -C backend test -race -count=1 ./...
# Frontend
pnpm --dir frontend type-check && pnpm --dir frontend build
```

### Database
```bash
make migrate-up
make migrate-down
make migrate-create NAME=name
sqlc generate
```

## Important Notes

- **Never commit `.env`** — `.env` / `.env.*` are gitignored (only `.env.example` is tracked). There is one `.env` per project (`backend/`, `frontend/`).
- **Run quality gates locally** before pushing.
- **SQLC generates code** — don't edit generated files.
- **Infra -> domain is allowed; domain -> infra is not.**
- **Reads bypass app/domain**; writes go through use cases.
- **Unit tests only**; CI has no database.
- **Timestamps are UTC**; Response DTOs use `time.Time`.
- **App runs on the host** against `localhost`; `postgres` only exists inside Compose.
- **The Go module lives in `backend/`**; use `go -C backend ...` or the root `Makefile`.
- **Production** is `docker-compose.prod.yml` with only `web` + `api`; the database is external via envs; TLS is terminated by Cloudflare Tunnel. Migrations in production are run manually, before deploying.
- **OpenSpec specs are authoritative** for behavior.

---

*This guide is maintained for AI agents working on the 1Channel project. Last updated: 2026-09-27*
