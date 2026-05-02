<!-- Context: project-intelligence/technical | Priority: critical | Version: 1.1 | Updated: 2026-05-01 -->

# Technical Domain

**Purpose**: Tech stack, architecture, patterns, and standards for MedMinder — a medication reminder application.
**Last Updated**: 2026-05-01

## Quick Reference
- **Update When**: Tech stack changes | New feature modules | API/component patterns change | Architecture decisions
- **Audience**: Developers, AI agents

## Primary Stack

| Layer | Technology | Version | Rationale |
|-------|-----------|---------|-----------|
| Language | Go | 1.25.0 | Type safety, single binary, great concurrency |
| HTTP Router | Chi (chi/v5) | 5.2.5 | Lightweight, idiomatic Go, composable middleware |
| API Framework | huma v2 | 2.37.2 | Typed handlers → auto OpenAPI 3.1 + Scalar Docs |
| Database | PostgreSQL | latest | ACID compliance, JSON support, excellent reliability |
| DB Migrations | golang-migrate | 4.19.1 | Declarative up/down migrations, embeddable |
| Query Gen | sqlc | latest | Type-safe SQL → Go code, no runtime ORM overhead |
| Frontend | SvelteKit | latest (SPA) | Reactive, minimal boilerplate, compiled output |
| UI Components | shadcn-svelte | — | Accessible, customizable, Tailwind-native |
| Styling | Tailwind CSS | v4 | Utility-first, consistent design tokens |
| Icons | Lucide Svelte | — | Lightweight, tree-shakeable SVG icons |
| Auth | golang-jwt (jwt/v5) | 5.3.1 | JWT access + refresh token auth |
| Email | go-mail | 0.7.2 | SMTP email delivery for password resets |
| Logging | phuslu/log | 1.0.122 | Structured, zero-allocation logger |
| Testing | testify | 1.11.1 | Assertions, mocking, suite support |
| Infra | Docker Compose | — | Local PostgreSQL, dev app container |
| PWA | vite-plugin-pwa | — | Offline support, installable web app |

## Architecture Pattern

```
Type: Monolith (single binary w/ embedded SPA frontend)
Pattern: Feature-module Go backend + SvelteKit SPA frontend → embed.FS → single binary
```

### Key Architecture Decisions
- **Single binary deployment**: `go:embed` bundles compiled SvelteKit output into Go binary — zero runtime frontend dependencies
- **Feature-module layout**: Features live in `internal/features/<feature>/` with own `repository/`, `service/`, `handler/`, `dto/`, `routes.go`
- **Feature-owned routes**: Each feature registers its own routes via `RegisterRoutes(api huma.API, queries, ...)` — see `internal/features/auth/routes.go`
- **Router orchestrator**: `internal/router/router.go` wires features together — thin, no business logic
- **DB connection passed via DI**: `main.go` creates connection → passes to `router.New` → passes to features

### Layered Architecture (Handler → Service → Repository)
Each feature follows strict 3-layer separation for testability:

```
routes.go → wires handler to huma.Register()

handler/   → HTTP ONLY: parse input → call service → map errors → format response
service/   → Business logic: validation, hashing, orchestration. No HTTP awareness.
repository/→ Data access: sqlc queries, CRUD. No business logic.
```

**Rule**: Handlers never call repositories directly. Services never reference HTTP.

## Project Structure

```
.
├── cmd/server/              # Entrypoint: loads config, runs migrations, starts server
├── internal/
│   ├── common/              # Shared utilities: config, database, email, log
│   ├── database/sqlc/       # sqlc-generated Go code from SQL queries
│   ├── features/
│   │   └── auth/            # Feature module: routes.go, repository/, service/, handler/, dto/
│   ├── middleware/          # HTTP middleware
│   └── router/              # Thin orchestrator: infra.go, spa.go, router.go
├── migrations/              # Database migration SQL files
├── web/                     # SvelteKit frontend
│   ├── src/
│   │   ├── lib/components/  # Reusable UI components (shadcn-svelte + app components)
│   │   ├── lib/utils/       # Frontend utilities
│   │   └── routes/          # SvelteKit page routes
│   └── static/              # Static assets
├── configs/                 # Configuration files
└── Makefile                 # All build/run/test commands
```

## Code Patterns

### huma v2 API Handler + DTO Pattern
```go
// dto/auth.go — wire format + validation tags owned by dto package
// type RegisterInput struct { Body struct { Email string `json:"email" ...` } }
// type RegisterOutput struct { Body struct { User User `json:"user"` } }

// handlers/register.go — thin HTTP adapter (error translation lives here)
func RegisterHandler(svc service.AuthService) func(context.Context, *dto.RegisterInput) (*dto.RegisterOutput, error) {
    return func(ctx context.Context, input *dto.RegisterInput) (*dto.RegisterOutput, error) {
        if err := ValidateEmail(input.Body.Email); err != nil {
            return nil, huma.Error400BadRequest("Invalid email", err)
        }
        result, err := svc.Register(ctx, input.Body.Email, input.Body.DisplayName, input.Body.Password)
        if err != nil {
            if errors.Is(err, service.ErrEmailExists) {
                return nil, huma.Error409Conflict("Email already exists", err)
            }
            return nil, err
        }
        return &dto.RegisterOutput{Body: struct{ User dto.User }{User: result.User}}, nil
    }
}

// routes.go — pure wiring, one liner per route
func RegisterRoutes(api huma.API, queries *db.Queries, jwtSecret string, ...) {
    huma.Register(api, huma.Operation{
        OperationID: "register-user", Method: http.MethodPost,
        Path: "/api/auth/register", Summary: "Register a new user", Tags: []string{"auth"},
    }, handlers.RegisterHandler(authSvc))

    // All routes look identical — handler does JWT extraction internally when needed
    huma.Register(api, huma.Operation{
        OperationID: "logout-user", Method: http.MethodPost,
        Path: "/api/auth/logout", Summary: "Logout user", Tags: []string{"auth"},
    }, handlers.LogoutHandler(authSvc, tokenSvc))
}
```

**Handler with multiple dependencies** (e.g., JWT extraction): pass all dependencies via the handler constructor — not inline closures in `routes.go`. The handler signature reflects what it needs.

### SvelteKit Page Component
```svelte
<script lang="ts">
  import { Card, CardContent } from '$lib/components/ui/card';
  let items = $state<string[]>([]);
</script>
<div class="px-4 py-6">
  <h1 class="text-2xl font-semibold tracking-tight">Title</h1>
  <Card><CardContent>Content</CardContent></Card>
</div>
```

## Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Go files | `snake_case` | `user_repository.go`, `password_reset.go` |
| SvelteKit routes | kebab-case params | `+page.svelte`, `(protected)/` |
| Go types (exported) | PascalCase | `RegisterRoutes`, `NewUserRepository` |
| Go types (unexported) | camelCase | `registerInput`, `loginHandler` |
| Go packages | lowercase | `auth`, `repository`, `service` |
| DB columns / JSON | `snake_case` | `access_token`, `email_verified` |
| API paths | `/api/resource/action` | `/api/auth/register`, `/api/auth/login` |
| Git branches | `feature/description`, `fix/description` | `feature/issue-42-medication-crud` |

## Code Standards

### Layered Separation (Critical)
- **Handlers are thin adapters** — HTTP concerns ONLY: parse input from `dto.Input`, call service, map errors to HTTP codes (`huma.Error*`), return `dto.Output`. NO business logic, NO direct repository calls.
- **Services own business logic** — Validation, hashing, token generation, orchestration of repository calls. Testable without HTTP. Service structs accept repository interfaces via DI.
- **Repositories own data access** — Direct DB queries via sqlc. NO business logic. One repository per entity.
- **DTOs own wire format** — All request/response types with Huma struct tags live in `dto/`. Handlers accept `*dto.Input`, return `*dto.Output`.

### General Standards
- **TDD required** — Write tests first for all features; table-driven tests with testify
- **Feature-module layout** — Each feature is self-contained with `routes.go`, `repository/`, `service/`, `handler/`, `dto/`, `dto/`
- **Feature-owned routes** — `RegisterRoutes(api huma.API, ...)` per feature; router just calls it
- **huma typed handlers** — All API routes use `func(ctx, *Input) (*Output, error)`, never plain Chi handlers for API
- **Makefile targets** — Always use Makefile, never raw Go commands
- **goimports** — Standard formatting with import grouping
- **Error handling** — `fmt.Errorf("context: %w", err)`, sentinel errors (`var ErrNotFound = errors.New(...)`)
- **Doc comments** — All exported types and functions must have doc comments
- **git flow** — Feature branches, never commit to `main`, PRs required

## Security Requirements

| Requirement | Implementation |
|-------------|---------------|
| JWT auth | Access + refresh token flow via golang-jwt |
| Input validation | huma struct tags: `minLength`, `maxLength`, `pattern`, `required` |
| Password strength | Min 8 chars with uppercase, lowercase, number |
| Email validation | Regex pattern on email field |
| Password reset | Token-based reset flow via email |
| Parameterized queries | sqlc generates type-safe SQL (no injection risk) |
| Bearer token auth | Authorization header extraction + validation |

## 📂 Codebase References

| Pattern | File Location | Description |
|---------|--------------|-------------|
| Layered architecture (pattern) | `internal/features/auth/handlers/` + `service/` + `repository/` | 3-layer separation |
| huma API handler | `internal/features/auth/routes.go` | Register, login, logout endpoints |
| Feature structure | `internal/features/auth/` | Full feature module layout |
| Router orchestration | `internal/router/router.go` | DI wiring, feature registration |
| Handler example | `internal/features/auth/handlers/register.go` | Thin adapter: input → service → dto.Output |
| Service example | `internal/features/auth/service/token.go` | Token generation service |
| Repository example | `internal/features/auth/repository/user_repository.go` | Data access layer |
| DTO example | `internal/features/auth/dto/auth.go` | RegisterInput, LoginInput, etc. with Huma tags |
| SvelteKit page | `web/src/routes/login/+page.svelte` | Login page with validation |
| UI components | `web/src/lib/components/ui/` | shadcn-svelte component library |
| DB connection | `internal/common/database/db.go` | PostgreSQL connection setup |
| Migrations | `internal/common/database/migrate.go` | golang-migrate integration |
| Embedded frontend | `cmd/server/main.go` | `go:embed web/dist` pattern |
| Makefile | `Makefile` | All build/run/test targets |
| Config | `internal/common/config/` | Environment-based configuration |

## Related Files
- `business-domain.md` — Problem statement and user needs
- `business-tech-bridge.md` — Business → technology mapping
- `decisions-log.md` — Decision history with rationale
