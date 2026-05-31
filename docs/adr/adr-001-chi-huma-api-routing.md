# ADR-001: Chi + huma for API Routing

## Status

Accepted

## Context

We needed an HTTP router and API framework for the MedMinder backend. The key requirements were:

1. Type-safe API handlers that auto-generate OpenAPI 3.1 specs
2. Automatic request validation from struct tags
3. Interactive API documentation (Swagger UI)
4. Lightweight, idiomatic Go — no heavy frameworks
5. Support for non-API routes (SPA file serving, service workers)

## Decision

Use **Chi** as the HTTP router, wrapped by **huma v2** for API routes.

- Chi handles all routing, middleware, and non-API endpoints
- huma v2 wraps Chi for API routes, providing typed handlers, auto-generated OpenAPI 3.1 at `/api/openapi.json`, and Swagger UI at `/api/docs`
- huma handler signature: `func(context.Context, *Input) (*Output, error)`
- Struct field tags (`doc`, `minLength`, `required`, etc.) drive both OpenAPI schema and request validation

## Consequences

**Positive:**
- OpenAPI 3.1 spec is always in sync with code — no drift
- Request validation is declarative, not manual
- Swagger UI provides interactive API documentation at `/api/docs`
- Chi remains available for non-API routes (SPA, health checks)

**Negative:**
- huma v2 has a learning curve for its typed handler patterns
- Two-layer abstraction (Chi + huma) adds some indirection
- huma's opinionated Input/Output struct model limits flexibility for unusual endpoints

## Alternatives Considered

| Alternative | Why Rejected |
|---|---|
| Raw Chi only | No OpenAPI generation, no request validation from struct tags |
| Gin | Adds a framework dependency; Chi is lighter and more idiomatic |
| Echo | Similar to Gin — framework weight for features we don't need |
| net/http (stdlib) | Too low-level; would need to build middleware, routing, validation manually |
| huma v2 with stdlib adapter | huma's Chi adapter has better middleware compatibility |
