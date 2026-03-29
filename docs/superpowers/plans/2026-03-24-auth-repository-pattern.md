# Auth Repository Pattern Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor auth handler to use separate repository package with interface in one file and postgres implementation in another.

**Architecture:** Move `UserRepository` interface from `handlers/register.go` to new `repository/user.go` file. Create `PostgresUserRepository` implementation in `repository/postgres.go`. Update imports in handlers and main.go.

**Tech Stack:** Go, sqlc, huma

---

## Task 1: Create repository package with interface

**Files:**
- Create: `internal/features/auth/repository/user.go`
- Modify: `internal/features/auth/handlers/register.go`

- [ ] **Step 1: Create user.go with UserRepository interface**

```go
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error)
}
```

- [ ] **Step 2: Update handlers/register.go to import from repository**

Remove the `UserRepository` interface definition and add import:

```go
import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/db"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"golang.org/x/crypto/bcrypt"
)
```

Change the handler parameter type:

```go
func RegisterHandler(repo repository.UserRepository, tokenSvc TokenServiceInterface) func(context.Context, *RegisterInput) (*RegisterOutput, error) {
```

- [ ] **Step 3: Run go build to verify**

```bash
go build ./...
```

Expected: Build passes

---

## Task 2: Create postgres implementation

**Files:**
- Create: `internal/features/auth/repository/postgres.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Create postgres.go**

```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/db"
)

type PostgresUserRepository struct {
	queries *db.Queries
}

func NewPostgresUserRepository(queries *db.Queries) *PostgresUserRepository {
	return &PostgresUserRepository{queries: queries}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error) {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash:  sql.NullString{String: passwordHash, Valid: true},
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	})
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *PostgresUserRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	return r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}
```

- [ ] **Step 2: Update main.go to use repository**

Remove the inline `userRepo` struct and its methods, then:

```go
import (
	// ... existing imports
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	// ... 
)

// In newRouter function, replace:
repo := &userRepo{queries: queries}
// With:
repo := repository.NewPostgresUserRepository(queries)
```

- [ ] **Step 3: Run go build to verify**

```bash
go build ./...
```

Expected: Build passes

---

## Task 3: Verify tests

**Files:**
- Test: `internal/features/auth/handlers/`

- [ ] **Step 1: Run handler tests**

```bash
go test ./internal/features/auth/handlers/... -v -count=1
```

Expected: All tests pass

- [ ] **Step 2: Run service tests**

```bash
go test ./internal/features/auth/service/... -v -count=1
```

Expected: All tests pass

---

## Task 4: Commit

- [ ] **Step 1: Add files**

```bash
git add internal/features/auth/repository/ internal/features/auth/handlers/register.go cmd/server/main.go
```

- [ ] **Step 2: Commit**

```bash
git commit -m "refactor: introduce repository pattern for auth

- Extract UserRepository interface to repository/user.go
- Add PostgresUserRepository implementation in repository/postgres.go
- Update handlers to use repository interface"
```

---

*Plan created from spec: docs/superpowers/specs/2026-03-24-auth-repository-design.md*
