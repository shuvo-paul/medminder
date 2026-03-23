# Auth Repository Pattern - Design Specification

**Date:** 2026-03-24  
**Feature:** Auth User Repository Pattern Implementation  
**Issue:** #21 (part of refactoring)

## Overview

Refactor the current auth handler to use a proper repository pattern with separated interface and implementation.

## Current State

- `UserRepository` interface defined in `handlers/register.go`
- Concrete implementation (`userRepo`) in `cmd/server/main.go`
- Direct dependency on `*db.Queries` in main

## Proposed Changes

### 1. Create repository package

```
internal/features/auth/repository/
  user.go       # UserRepository interface
  postgres.go   # PostgresUserRepository implementation
```

### 2. user.go (Interface)

```go
type UserRepository interface {
    CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error)
    GetUserByEmail(ctx context.Context, email string) (db.User, error)
    CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error)
}
```

### 3. postgres.go (Implementation)

- Wrap `*db.Queries`
- Implement all `UserRepository` methods
- Use sqlc-generated query methods internally

### 4. Update handlers

- Import `UserRepository` from `repository` package
- Keep `TokenServiceInterface` in handlers (or move to service)

### 5. Update main.go

- Replace inline `userRepo` with `repository.NewPostgresUserRepository(queries)`

## Files to Modify

- `internal/features/auth/handlers/register.go` — import repository interface
- `cmd/server/main.go` — use repository implementation

## Files to Create

- `internal/features/auth/repository/user.go` — interface definition
- `internal/features/auth/repository/postgres.go` — postgres implementation

## Testing Approach

- Unit tests for repository implementation (using test DB or testcontainers)
- Handler tests can inject mock repository

## Acceptance Criteria

- [ ] `UserRepository` interface in separate package
- [ ] `PostgresUserRepository` implements interface
- [ ] Handlers use repository interface
- [ ] Server builds and works with new structure
- [ ] All existing tests pass

---

*Approved for implementation*
