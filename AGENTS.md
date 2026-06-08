# MedMinder — Agent Guidelines

## Project Overview
MedMinder is a medication reminder application built with Go 1.25.0 and SvelteKit.
Module: `github.com/shuvo-paul/medminder`

### Tech Stack

**Backend**
- HTTP router: Chi (wrapped by huma for API routes)
- API framework: huma v2 — typed handlers, auto-generated OpenAPI 3.1 spec at `/api/openapi.json`, Swagger UI at `/api/docs`
- Database migrations: golang-migrate
- Query generation: sqlc
- Testing/assertions: testify
- Live reload: Air

**Frontend** — SvelteKit (SPA mode) + Tailwind CSS v4 + shadcn-svelte. See [`docs/frontend.md`](docs/frontend.md).

## Commands

| Task | Command |
|---|---|
| Dev (Go + Vite) | `make start` |
| Go API only | `make dev` |
| Frontend only | `make web-dev` |
| Run tests | `make test` |
| Test with coverage | `make test-cover` |
| Build | `make build` |
| Tidy modules | `make tidy` |
| Regenerate sqlc code | `make sqlc-generate` |
| Create migration | `make db-migrate-create NAME=<name>` |
| Apply migrations | `make db-migrate-up` |
| Production build | `make embed-frontend` |
| Start Postgres | `docker compose up -d` |

## Critical Rules

1. **Use Makefile for Go builds** — never run `go build` or `go mod tidy` directly.
2. **Never push without approval** — get explicit human approval before `git push`. "Open a PR" is intent, not permission.
3. **TDD required** — write tests before implementation. Run `make test` before committing.
4. **Branch before work** — always create a feature or fix branch; never commit to `main`.
5. **Lazy-load docs/** — read files from `docs/` only when relevant to the current task, not upfront in every session.
6. **Never commit secrets** — no credentials, `.env` files, or sensitive keys.
7. **Never force-push** — rebase safely or merge instead.

For detailed code conventions, see [`docs/code-conventions.md`](docs/code-conventions.md).

## Architecture Decision Records

Before proposing architectural changes, check `docs/adr/`:
- **ADR-001**: Chi + huma API routing
- **ADR-002**: SvelteKit SPA embed

Existing ADRs document why current patterns exist and what alternatives were rejected. Creating a new ADR is required when adopting a significant new dependency, framework, deployment model, or architectural pattern.

## Project Structure

```
├── cmd/server/         # Application entrypoint
├── internal/           # Private application code
│   ├── common/         # Shared utilities
│   ├── config/         # Configuration
│   ├── features/      # Feature modules (auth, medications, reminders)
│   │   └── <feature>/  # Feature package
│   │       ├── repository/  # Data access layer (e.g., user_repository.go)
│   │       ├── service/      # Business logic
│   │       ├── handler/      # HTTP handlers
│   │       ├── routes.go    # Feature-owned route registration
│   │       └── dto/          # Data transfer objects
│   ├── middleware/     # HTTP middleware
│   ├── router/        # Route orchestrator (infra, SPA, huma setup)
│   └── server/        # Server setup
├── pkg/               # Public packages
├── tests/             # Test suites
│   ├── e2e/          # End-to-end tests
│   ├── integration/  # Integration tests
│   └── testutil/     # Test helpers
├── migrations/        # Database migrations
├── web/              # SvelteKit frontend (see docs/frontend.md)
├── docs/             # Documentation & ADRs
├── configs/          # Configuration files
```

## Feature-Owned Route Registration

Each feature package owns its route registration via a `RegisterRoutes` function in `routes.go`:

```go
// internal/features/auth/routes.go
func RegisterRoutes(api huma.API, queries *db.Queries, jwtSecret string)
```

The router orchestrates by calling feature `RegisterRoutes` functions. This keeps feature wiring inside the feature package.

## Environment

- Use `.env` files for local configuration (gitignored)
- Configuration loaded from `configs/` directory
- Reference `.env.example` for dev credentials and required variables

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues on `shuvo-paul/medminder`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default label names used for triage: needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context repo — one CONTEXT.md + docs/adr/ at root. See `docs/agents/domain.md`.
