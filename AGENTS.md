# MedMinder — Agent Guidelines

## Project Overview
MedMinder is a medication reminder application built with Go 1.25.0 and SvelteKit.
Module: `github.com/shuvo-paul/medminder`

### Tech Stack

**Backend**
- HTTP router: Chi
- Database migrations: golang-migrate
- Query generation: sqlc
- Testing/assertions: testify
- Live reload: Air

**Frontend** — see `web/AGENTS.md` for full stack details.

## Build/Run/Test Commands

Prefer the Makefile targets so every environment uses the same workflow:

```bash
# Development
make start         # Go (Air :8080) + Vite (:5173) together — Ctrl+C stops both
make dev           # Go API only with Air hot-reload (:8080)
make web-dev       # Frontend only with Vite HMR (:5173)

# Go
make tidy
make build
make run           # requires prior embed-frontend
make test
make test-cover
make clean

# Frontend
make web-install   # first time only
make web-build
make web-preview

# Production
make embed-frontend  # pnpm build + go build → bin/medminder (single binary)

# Docker (equivalent to make start + postgres)
docker compose up --build
```

## Project Structure

```
├── cmd/server/         # Application entrypoint
├── internal/           # Private application code
│   ├── common/        # Shared utilities
│   ├── config/        # Configuration
│   ├── features/      # Feature modules
│   ├── middleware/    # HTTP middleware
│   ├── router/        # Route definitions
│   └── server/        # Server setup
├── pkg/               # Public packages
├── tests/             # Test suites
│   ├── e2e/          # End-to-end tests
│   ├── integration/  # Integration tests
│   └── testutil/     # Test helpers
├── migrations/        # Database migrations
├── web/              # SvelteKit frontend (see web/AGENTS.md)
└── configs/          # Configuration files
```

## Git Conventions
- Feature branches: `feature/description`
- Bug branches: `fix/description`
- Always checkout to a new branch before implementing a new feature or bug fix
- Commit messages: present tense, imperative mood
- Run `go test ./...` before committing

### Handling GitHub Issues
When mentioned in a GitHub issue to implement:
1. Create branch: `feature/issue-{number}-description` or `fix/issue-{number}-description`
2. Implement following TDD, run tests before committing
3. Push and create PR with `gh pr create`

## Environment
- Use `.env` files for local configuration (gitignored)
- Configuration loaded from `configs/` directory

## Go Code Style

Applies to all Go code under `internal/`, `pkg/`, and `cmd/`.

### Imports
- Use goimports for formatting (handles grouping automatically)
- Alias imports only when necessary for clarity

### Naming Conventions
- Use CamelCase for exported names, camelCase for unexported
- Test files: `*_test.go`
- Test functions: `TestFunctionName`, `TestStruct_MethodName`
- Interfaces with "-er" suffix (Reader, Writer) or descriptive names
- Avoid underscores in file names except for `_test.go` and `_mock.go`

### Types
- Define structs close to their usage
- Use interfaces to define behavior contracts
- Prefer composition over inheritance
- Return concrete types, accept interfaces

### Error Handling
- Use `fmt.Errorf` with context: `fmt.Errorf("doing X: %w", err)`
- Create sentinel errors for common cases: `var ErrNotFound = errors.New("not found")`
- Check errors immediately after function calls
- Never ignore errors with `_` without comment explaining why

### Testing (TDD Required)
- All features must follow Test Driven Development
- Table-driven tests preferred
- Use testify/assert for assertions
- Mock external dependencies
- Tests should be in same package with `_test.go` suffix

### Documentation
- All exported types and functions must have doc comments
- Comments start with the name being documented
- Use complete sentences with proper punctuation

## Boundaries
- **Never commit directly to `main`** — all changes must go through a feature or fix branch and a PR
- Never commit secrets, credentials, or `.env` files
- Never force-push to `main`
