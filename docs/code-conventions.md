# Code Conventions

Applies to all Go code under `internal/`, `pkg/`, and `cmd/`.

## Imports

- Use goimports for formatting (handles grouping automatically)
- Alias imports only when necessary for clarity

## Naming Conventions

- Use CamelCase for exported names, camelCase for unexported
- **Test files MUST use `*_test.go` suffix** — never use other suffixes like `*_tests.go`
- Test functions: `TestFunctionName`, `TestStruct_MethodName`
- Interfaces with "-er" suffix (Reader, Writer) or descriptive names
- Avoid underscores in file names except for `_test.go` and `_mock.go`
- Repository files: `*_repository.go` (e.g., `user_repository.go`, `medication_repository.go`)
- Repository constructors: `NewUserRepository`, `NewMedicationRepository` (exported), implementation types private (e.g., `userRepository`)

## Types

- Define structs close to their usage
- Use interfaces to define behavior contracts
- Prefer composition over inheritance
- Return concrete types, accept interfaces

## Error Handling

- Use `fmt.Errorf` with context: `fmt.Errorf("doing X: %w", err)`
- Create sentinel errors for common cases: `var ErrNotFound = errors.New("not found")`
- Check errors immediately after function calls
- Never ignore errors with `_` without comment explaining why

## API Handlers

All API route handlers must use huma — not plain Chi handler functions. Handler functions take the signature `func(context.Context, *Input) (*Output, error)` where `Input` and `Output` are typed structs registered via `huma.Register`. Struct field tags (`doc`, `minLength`, `required`, etc.) drive both OpenAPI schema and request validation. Non-API routes (static files, SPA) remain as plain Chi handlers.

## Testing

### Conventions (TDD Required)

- All features must follow Test Driven Development
- Table-driven tests preferred
- Use testify/assert for assertions
- Mock external dependencies with hand-written mocks using `mock.Mock` from testify
- Tests should be in same package with `_test.go` suffix
- **Test packages**: external test packages (`handlers_test`, `service_test`, `repository_test`) for handlers/services/repos; same-package tests for utilities (dto, config, errors)

### Running Tests

```bash
make test                  # All tests
make test-cover            # With coverage
# Target specific tests:
go test -run TestName ./internal/features/auth/handlers/...
# Coverage HTML:
go tool cover -html=coverage.out
```

## Documentation

- All exported types and functions must have doc comments
- Comments start with the name being documented
- Use complete sentences with proper punctuation
- Comments explain **why** something is done, not **what** — only add comments for non-obvious logic

## Formatting

- Run `go fmt` after code implementation (or use goimports which includes formatting)
- Never commit unformatted code

## Code Generation

Regenerate after modifying sources:

| Trigger | Command |
|---|---|
| SQL queries or schema changed | `make sqlc-generate` |
| New migration needed | `make db-migrate-create NAME=<name>` |
| Frontend changes for prod build | `make embed-frontend` |

Generated files live in `internal/database/sqlc/` — **never edit these manually**.

## Anti-Patterns

- Never run `go build` or `go mod tidy` directly — use `make build` and `make tidy`
- Never create DB connections inside `router.New` — pass from `main.go`
- Never add new features without a branch
- Never edit generated files (`internal/database/sqlc/*.go`)
- Never import packages not in `go.mod`
- Never use emoji in code or comments
- Never commit secrets, credentials, or `.env` files
- Never force-push to `main`

## Git Conventions

- **Commit format**: [Conventional Commits](https://www.conventionalcommits.org) — `type(scope): description`
  - `feat(auth): add OAuth provider support`
  - `fix(reminders): correct timezone handling`
  - `refactor(router): extract infra routes`
- Branch naming: `feature/description` or `fix/description`
- Issue branches: `feature/issue-{number}-description` or `fix/issue-{number}-description`
- Commit messages: present tense, imperative mood
- Run `make test` before committing

## Before Committing

- [ ] Code formatted (`go fmt` / goimports)
- [ ] All tests pass (`make test`)
- [ ] Build succeeds (`make build`)
- [ ] No generated files edited manually
