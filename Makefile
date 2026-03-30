.PHONY := build dev start run test test-cover tidy clean web-install web-dev web-build web-preview embed-frontend db-migrate-up db-migrate-down db-migrate-steps db-migrate-force db-migrate-create sqlc-generate

GO ?= go
PNPM ?= pnpm
AIR ?= air
BIN_DIR := bin
BIN := $(BIN_DIR)/medminder
CMD := cmd/server/main.go
WEB_DIR := web
MIGRATE := migrate
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= medminder
DB_PASSWORD ?= medminder
DB_NAME ?= medminder
DB_SSLMODE ?= disable
MIGRATION_SOURCE ?= file://internal/common/database/migrations

build:
	@mkdir -p $(BIN_DIR)
	@mkdir -p cmd/server/web/dist
	$(GO) build -o $(BIN) $(CMD)

# Development: Go server with Air hot-reload. No frontend build.
# Use alongside `make web-dev` in a second terminal.
dev:
	$(AIR)

# Start both Air (Go) and Vite (frontend) together. Ctrl+C stops both.
start:
	trap 'kill 0' EXIT; $(AIR) & cd $(WEB_DIR) && $(PNPM) dev

# Run Go server directly (use after make embed-frontend)
run:
	$(GO) run $(CMD)

test:
	$(GO) test ./...

test-cover:
	$(GO) test -cover ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

web-install:
	cd $(WEB_DIR) && $(PNPM) install

web-dev:
	cd $(WEB_DIR) && $(PNPM) dev

web-build:
	cd $(WEB_DIR) && $(PNPM) build

web-preview:
	cd $(WEB_DIR) && $(PNPM) preview

embed-frontend: web-build build

db-migrate-up:
	$(GO) run -mod=mod cmd/migrate/main.go -direction up

db-migrate-down:
	$(GO) run -mod=mod cmd/migrate/main.go -direction down

db-migrate-steps:
	@echo "Usage: make db-migrate-steps STEPS=<n> (negative for down)"
	@if [ -z "$(STEPS)" ]; then exit 1; fi
	$(GO) run -mod=mod cmd/migrate/main.go -direction steps -steps $(STEPS)

db-migrate-force:
	@echo "Usage: make db-migrate-force VERSION=<version>"
	@if [ -z "$(VERSION)" ]; then exit 1; fi
	$(GO) run -mod=mod cmd/migrate/main.go -direction force -version $(VERSION)

db-migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make db-migrate-create NAME=<name>"; exit 1; fi
	$(GO) run -mod=mod cmd/migrate/main.go -direction create -name $(NAME)

sqlc-generate:
	sqlc generate
