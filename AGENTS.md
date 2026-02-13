# AGENTS.md - Coding Guidelines for MedMinder

## Project Overview
MedMinder is a medication reminder application built with Go 1.25.0.
Module: `github.com/shuvo-paul/medminder`

## Build/Run/Test Commands

```bash
# Install dependencies
go mod tidy

# Build the application
go build -o bin/medminder cmd/server/main.go

# Run the application
go run cmd/server/main.go

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a single test by name
go test -run TestFunctionName ./...

# Run tests for a specific package
go test ./internal/features/...

# Run tests with verbose output
go test -v ./...
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
├── web/              # Frontend assets
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
- Commit messages: present tense, imperative mood
- Run `go test ./...` before committing

## Environment
- Use `.env` files for local configuration (gitignored)
- Never commit secrets or credentials
- Configuration loaded from `configs/` directory
