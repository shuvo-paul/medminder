# OpenAPI Documentation via huma v2

**Date:** 2026-03-22

## Context

MedMinder had no API documentation and no consistent handler pattern. The SRS defines 82 planned endpoints across 10 modules. Adding OpenAPI documentation early establishes a machine-readable contract that stays in sync with the code as features are built.

## Decision

Integrate **huma v2** with the existing Chi router using the `humachi` adapter. huma wraps Chi — Chi stays as the base router, API routes register through huma, and static/SPA routes remain as plain Chi handlers.

**Why huma over alternatives:**
- Typed `Input`/`Output` structs auto-generate the OpenAPI 3.1 spec at runtime — no annotation drift, no generation step
- Built-in request validation from struct field tags
- Swagger UI bundled at `/docs` at no extra cost
- Clean handler signature (`func(context.Context, *Input) (*Output, error)`) is more readable and testable than plain `http.HandlerFunc`

## What Changed

- `cmd/server/main.go` — extracted `newRouter(distFS fs.FS) http.Handler`, integrated huma, migrated health check to a typed huma operation
- `cmd/server/main_test.go` — tests for health check response shape and `/openapi.json` availability
- `go.mod` / `go.sum` — added `github.com/danielgtaylor/huma/v2 v2.37.2`
- `AGENTS.md` — updated tech stack, added API Handlers coding guideline
- `README.md` — added API Docs section

## Endpoints Added

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/healthz` | Health check (now in OpenAPI spec) |
| `GET` | `/openapi.json` | Live OpenAPI 3.1 spec |
| `GET` | `/docs` | Swagger UI |

## Handler Pattern (all future API features)

```go
type CreateXInput struct {
    Body struct {
        Name string `json:"name" doc:"..." minLength:"1"`
    }
}

type CreateXOutput struct {
    Body *X
}

huma.Register(api, huma.Operation{
    OperationID: "create-x",
    Method:      http.MethodPost,
    Path:        "/api/x",
    Summary:     "Create X",
    Tags:        []string{"x"},
}, handler.Create)
```

Struct field tags (`doc`, `minLength`, `maxLength`, `minimum`, `required`, `pattern`) drive both the OpenAPI schema and automatic request validation.
