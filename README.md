# MedMinder

A medication reminder application built with Go and SvelteKit (shadcn-svelte, Tailwind CSS v4, PWA).

## Development

### Prerequisites

- Go 1.25+
- Node.js LTS + pnpm
- [Air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`)
- Docker Engine 24+ and Docker Compose v2 (optional)

### Quick Start (local)

```bash
go mod download
make web-install
make start        # Go API on :8080, Vite on :5173
```

Open `http://localhost:5173`. The Go server hot-reloads on `.go` changes via Air; the frontend hot-reloads via Vite HMR.

### Quick Start (Docker)

```bash
docker compose up --build
```

Runs Go + Node.js in a single container alongside Postgres. Go API on `:8080`, Vite dev server on `:5173`.

To override any defaults, copy `.env.example` to `.env` and rerun `docker compose up --build`.

### Individual targets

| Command | Description |
|---|---|
| `make dev` | Go API only (Air hot-reload on `:8080`) |
| `make web-dev` | Frontend only (Vite HMR on `:5173`) |
| `make start` | Both together (Ctrl+C stops both) |
| `make embed-frontend` | Production build — frontend embedded in Go binary |

### API Docs

| URL | Description |
|---|---|
| `http://localhost:8080/api/openapi.json` | Live OpenAPI 3.1 spec (always in sync with code) |
| `http://localhost:8080/api/docs` | Swagger UI |

### Running Tests

```bash
# Local
make test

# Docker
docker compose run --rm app go test ./...
```

### Database Access

- Default credentials are defined in `.env.example`.
- Connect with `psql` using `docker compose exec db psql -U $DB_USER -d $DB_NAME`.

## Production

```bash
make embed-frontend   # pnpm build → go build → bin/medminder
./bin/medminder       # single binary on :8080
```

## License
MIT
