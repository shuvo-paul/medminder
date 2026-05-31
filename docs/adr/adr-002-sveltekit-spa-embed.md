# ADR-002: SvelteKit SPA Mode with Go Binary Embedding

## Status

Accepted

## Context

MedMinder is a self-contained application that should be simple to deploy. We needed to decide:

1. How to serve the frontend — SSR (server-side rendering) vs SPA (single-page application)
2. How to package the frontend for distribution — separate static hosting vs embedded in Go binary

## Decision

Use **SvelteKit in SPA mode** with **`adapter-static`**, embedded into the Go binary via `go:embed`.

- SvelteKit runs in SPA mode (`adapter-static` with `fallback: 'index.html'`)
- Vite builds the frontend to `cmd/server/web/dist/`
- Go binary embeds the entire `web/dist/` directory at compile time
- Chi serves the embedded files as a file server, falling back to `index.html` for client-side routing

## Consequences

**Positive:**
- Single-binary deployment — no separate static file server, no Node.js runtime in production
- No SSR complexity — simpler mental model, no server/client hydration bugs
- Embedded files are always in sync with the binary version
- Cold starts are fast — no network round-trips or CDN dependencies

**Negative:**
- No SSR means no server-rendered HTML for SEO or social previews (not a concern for a PWA)
- Binary is larger (~few MB for embedded assets)
- Every frontend change requires `make embed-frontend` + binary rebuild for production
- No incremental deployment of frontend changes independent of backend

## Alternatives Considered

| Alternative | Why Rejected |
|---|---|
| SvelteKit SSR | Adds deployment complexity — need Node.js runtime or separate container; SEO not needed |
| Separate static hosting (CDN/S3) | Two deployment targets; version skew risk between frontend and backend |
| Next.js / React | SvelteKit has a smaller bundle size and a simpler component model |
| Pure Go templates (no SPA framework) | Poor developer experience for interactive UIs; no component ecosystem |
