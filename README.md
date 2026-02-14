# MedMinder

A medication reminder application built with Go.

## Development Environment

MedMinder ships with a Docker-based developer experience featuring hot reload via [Air](https://github.com/air-verse/air) and a Postgres database.

### Prerequisites

- Docker Engine 24+
- Docker Compose v2

### Quick Start

```bash
docker compose up --build
```

The API listens on `http://localhost:8080` (configurable via `APP_PORT`). Source files are bind-mounted, so edits automatically trigger a rebuild inside the container.

To override any defaults (for example when preparing staging/production), copy `.env.example` to `.env`, update the values, and rerun `docker compose up --build`. Compose automatically picks up `.env` when it exists.

### Running Tests

```bash
docker compose run --rm app go test ./...
```

### Database Access

- Default credentials are defined in `.env.example`.
- Connect with `psql` using `docker compose exec db psql -U $DB_USER -d $DB_NAME`.
- The Postgres container stores data under `/var/lib/postgresql` (per the 18.x images). If you ran an older setup that mounted `/var/lib/postgresql/data`, remove the stale volume with `docker volume rm medminder_postgres_data` before starting the upgraded stack.

## License
MIT
