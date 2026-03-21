# AGENTS.md - Coding Guidelines for MedMinder

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
make tidy          # go mod tidy
make build         # go build -o bin/medminder cmd/server/main.go
make run           # go run cmd/server/main.go (requires prior embed-frontend)
make test          # go test ./...
make test-cover    # go test -cover ./...
make clean         # remove bin/

# Frontend
make web-install   # pnpm install
make web-build     # pnpm build
make web-preview   # pnpm preview

# Production
make embed-frontend  # pnpm build + go build → bin/medminder (single binary)

# Docker (equivalent to make start + postgres)
docker compose up --build
```

If you need to run commands manually, stick to the equivalents shown above (e.g., `go test ./...`, `go test -run TestFunctionName ./...`).

## Project Structure

```
├── cmd/server/         # Application entrypoint
├── internal/           # Private code (see internal/AGENTS.md for Go style)
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
- Never commit secrets or credentials
- Configuration loaded from `configs/` directory

## Boundaries
- **Never commit directly to `main`** — all changes must go through a feature or fix branch and a PR
- Never commit secrets, credentials, or `.env` files
- Never force-push to `main`
