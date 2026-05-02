<!-- Context: standards/code | Priority: critical | Version: 3.0 | Updated: 2026-05-01 -->
# Code Standards

## Quick Reference

**Core Philosophy**: Modular, Functional, Maintainable
**Golden Rule**: If you can't easily test it, refactor it

**Critical Patterns** (use these):
- ✅ Pure functions (same input = same output, no side effects)
- ✅ Immutability (create new data, don't modify)
- ✅ Composition (build complex from simple)
- ✅ Small functions (< 50 lines)
- ✅ Explicit dependencies (dependency injection)

**Anti-Patterns** (avoid these):
- ❌ Mutation, side effects, deep nesting
- ❌ God modules, global state, large functions

---

## Core Philosophy

**Modular**: Everything is a component - small, focused, reusable
**Functional**: Pure functions, immutability, composition over inheritance
**Maintainable**: Self-documenting, testable, predictable

## Principles

### Modular Design
- Single responsibility per module (package)
- Clear interfaces (explicit inputs/outputs via exported types)
- Independent and composable
- < 100 lines per file (ideally < 50)

### Functional Approach
- **Pure functions**: Same input = same output, no side effects
- **Immutability**: Pass by value, return new values, don't mutate receivers
- **Composition**: Build complex from small functions, compose interfaces
- **Declarative**: Describe what, not how

### Package Structure
```
internal/features/<feature>/
├── dto/              # Request/response types with Huma struct tags (Body wrappers, validation tags)
├── handler/         # Thin HTTP adapter: accepts dto.Input → calls service → returns dto.Output
├── service/         # Business logic: pure functions, no HTTP awareness
├── repository/      # Data access layer (sqlc queries)
└── routes.go        # Feature-owned route registration (pure wiring, no error translation)
```

## Feature Module Design Pattern (Mandatory)

**Every** feature module MUST live under `internal/features/<feature>/`. No exceptions. This is the only valid location for feature code.

```
internal/features/<feature>/
├── dto/              # Request/response types with Huma struct tags (Body wrappers, validation tags)
├── handler/         # Thin HTTP adapter: accepts dto.Input → calls service → returns dto.Output
├── service/         # Business logic: pure functions, no HTTP awareness
├── repository/      # Data access layer (sqlc queries)
└── routes.go        # Feature-owned route registration (pure wiring, no error translation)
```

**Rules (enforced, not optional):**
1. Feature code lives in `internal/features/<feature>/` — not scattered across `internal/` root
2. `dto/` owns all wire-format types with Huma struct tags — handlers do NOT define these
3. Handler functions accept `*dto.Input`, return `*dto.Output` — never bare types
4. Error translation (sentinel errors → `huma.Error400BadRequest`, etc.) lives **inside** the handler, not in `routes.go`
5. Routes are pure wiring: `huma.Register(api, huma.Operation{...}, handlers.MyHandler(svc))` — no inline closures, no error translation, no intermediate wrappers
6. All route registrations look identical: just `handlers.XxxHandler(svcA, svcB)` passed directly to `huma.Register`. If a handler needs external dependencies (e.g., `TokenService` for JWT extraction), pass them via the handler constructor, not the route closure
7. Shared types (e.g., `User`) live in `dto/` and are reused across input/output structs

**Why this pattern:**
- `dto/` = wire format + validation (what the client sends/receives)
- `handler/` = HTTP↔service adapter (where HTTP semantics belong)
- `service/` = pure business logic (testable without HTTP)
- `routes.go` = wiring only (reads like a manifest, not code)

Canonical reference: `internal/features/auth/`

## Patterns

### Pure Functions
```go
// ✅ Pure: same input = same output, no side effects
func Add(a, b int) int { return a + b }

func FormatName(firstName, lastName string) string {
    return firstName + " " + lastName
}

// ❌ Impure: mutates package-level state
var total int
func AddToTotal(value int) int {
    total += value
    return total
}
```

### Immutability
```go
// ✅ Immutable: return new slice, don't mutate the input
func AddItem(items []string, item string) []string {
    result := make([]string, len(items)+1)
    copy(result, items)
    result[len(items)] = item
    return result
}

// Or use append (creates new backing array when capacity exceeded)
func AddItem(items []string, item string) []string {
    return append(append([]string(nil), items...), item)
}

// ✅ Immutable: return new struct
func WithDiscount(price float64, pct float64) float64 {
    return price * (1 - pct/100)
}

// ❌ Mutable: modifies slice in place
func addItem(items []string, item string) {
    items = append(items, item) // local slice header change lost
}
```

### Composition
```go
// ✅ Compose small interfaces
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type ReadWriter interface {
    Reader
    Writer
}

// ✅ Compose small functions
func ProcessUser(ctx context.Context, u User) (User, error) {
    u, err := ValidateUser(u)
    if err != nil {
        return User{}, err
    }
    u, err := EnrichUserData(ctx, u)
    if err != nil {
        return User{}, err
    }
    return SaveUser(ctx, u)
}

// ❌ Deep inheritance (not applicable in Go — use composition)
```

### Declarative
```go
// ✅ Declarative: express intent
func ActiveUsers(users []User) []string {
    names := make([]string, 0, len(users))
    for _, u := range users {
        if u.IsActive {
            names = append(names, u.Name)
        }
    }
    return names
}

// ❌ Imperative with index tracking
func activeUsers(users []User) []string {
    var names []string
    for i := 0; i < len(users); i++ {
        if users[i].IsActive {
            names = append(names, users[i].Name)
        }
    }
    return names
}
```

## Naming

- **Files**: snake_case.go (`user_repository.go`, `medication_handler.go`)
- **Exported**: CamelCase (`NewUserService`, `CreateUser`)
- **Unexported**: camelCase (`userService`, `createUser`)
- **Interfaces**: "-er" suffix or descriptive names (`Reader`, `Storer`, `Validator`)
- **Predicates/booleans**: `IsValid`, `HasPermission`, `CanAccess`, `Enabled`
- **Variables**: descriptive (`userCount` not `uc`), short for small scopes
- **Constants**: CamelCase (Go idiom), e.g. `MaxRetries`, `DefaultTimeout`

## Error Handling

```go
// ✅ Explicit error handling with context
func ParseJSON(text string) (Value, error) {
    var data Value
    err := json.Unmarshal([]byte(text), &data)
    if err != nil {
        return Value{}, fmt.Errorf("parsing json: %w", err)
    }
    return data, nil
}

// ✅ Validate at boundaries
func (s *UserService) CreateUser(ctx context.Context, dto CreateUserDTO) (User, error) {
    if err := s.validator.Validate(dto); err != nil {
        return User{}, fmt.Errorf("create user: validation: %w", err)
    }
    user, err := s.repo.Insert(ctx, dto.ToUser())
    if err != nil {
        return User{}, fmt.Errorf("create user: %w", err)
    }
    return user, nil
}

// ❌ Ignoring errors or bare panics
func createUser(data string) User {
    var u User
    json.Unmarshal([]byte(data), &u) // error silently dropped
    return u
}
```

## Dependency Injection

```go
// ✅ Dependencies explicit — passed via constructor
type UserService struct {
    repo  UserRepository
    log   *slog.Logger
}

func NewUserService(repo UserRepository, log *slog.Logger) *UserService {
    return &UserService{repo: repo, log: log}
}

func (s *UserService) Create(ctx context.Context, user User) error {
    s.log.InfoContext(ctx, "creating user")
    return s.repo.Insert(ctx, user)
}

// ❌ Hidden dependencies — global/package-level access
var db *sql.DB

func CreateUser(user User) error {
    _, err := db.Exec("INSERT ...") // where did db come from?
    return err
}
```

## Anti-Patterns

❌ **Mutation**: Modifying input slices/structs in place
❌ **Side effects**: Logging inside pure functions, API calls in business logic
❌ **Deep nesting**: Use early returns / guard clauses instead
❌ **God packages**: Split into focused packages by responsibility
❌ **Global state**: `var` at package level, init() side effects
❌ **Silent errors**: `_` ignoring returned errors
❌ **Large functions**: Keep < 50 lines

## Best Practices

✅ Pure functions whenever possible
✅ Pass values, not pointers (only use pointers for mutation or nil)
✅ Small, focused functions (< 50 lines)
✅ Compose small interfaces into larger ones
✅ Explicit dependencies via constructor injection
✅ Validate at boundaries (handler/DTO level)
✅ Self-documenting code with doc comments on exported symbols
✅ Test in isolation (mock interfaces, table-driven tests)
✅ Use `fmt.Errorf("context: %w", err)` to wrap errors with context

**Golden Rule**: If you can't easily test it, refactor it.
