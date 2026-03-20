# AGENTS.md - Coding Guidelines for MedMinder

## Project Overview
MedMinder is a medication reminder application built with Go 1.25.0 and SvelteKit.
Module: `github.com/shuvo-paul/medminder`

### Tech Stack

**Backend**
- HTTP router: [Chi](https://github.com/go-chi/chi)
- Database migrations: [golang-migrate](https://github.com/golang-migrate/migrate)
- Query generation: [sqlc](https://github.com/sqlc-dev/sqlc)
- Testing/assertions: [stretchr/testify](https://github.com/stretchr/testify)
- Live reload: [Air](https://github.com/air-verse/air)

**Frontend**
- Framework: [SvelteKit](https://kit.svelte.dev) (SPA mode)
- Build tool: [Vite](https://vitejs.dev)
- Styling: [Tailwind CSS](https://tailwindcss.com)
- Package manager: pnpm

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

### Dev workflow

**Local:**
```bash
make web-install   # first time only
make start         # browser: http://localhost:5173
```

**Docker:**
```bash
docker compose up --build   # browser: http://localhost:5173
```

## Project Structure

```
├── cmd/server/         # Application entrypoint
├── internal/           # Private code
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
├── web/              # SvelteKit frontend (src/, static/, vite.config.ts)
└── configs/          # Configuration files
```

## Code Style Guidelines

### Imports
- Group imports: stdlib, third-party, local
- Use goimports for formatting
- Alias imports only when necessary for clarity

```go
import (
    "context"
    "time"

    "github.com/stretchr/testify/assert"

    "github.com/shuvo-paul/medminder/internal/config"
)
```

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

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid case", "input", "output", false},
        {"error case", "bad", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

### Documentation
- All exported types and functions must have doc comments
- Comments start with the name being documented
- Use complete sentences with proper punctuation

```go
// UserService handles user-related business logic.
type UserService struct {
    repo UserRepository
}

// GetUser retrieves a user by their ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
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
