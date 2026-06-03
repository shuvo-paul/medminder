.PHONY: start dev run build test test-cover tidy clean web-deps web-dev web-build web-preview embed-frontend db-migrate-up db-migrate-down db-migrate-steps db-migrate-force db-migrate-create sqlc-generate

GO ?= go
PNPM ?= pnpm
AIR ?= air
BIN_DIR := bin
BIN := $(BIN_DIR)/medminder
CMD := cmd/server/main.go
WEB_DIR := web

start:
	@trap 'kill 0' EXIT; \
	$(AIR) & \
	cd $(WEB_DIR) && $(PNPM) dev & \
	wait

dev:
	$(AIR)

run:
	./$(BIN)

build:
	@mkdir -p $(BIN_DIR)
	@mkdir -p cmd/server/web/dist
	$(GO) build -o $(BIN) $(CMD)

test:
	$(GO) test ./...

test-cover:
	$(GO) test -cover ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

web-deps:
	cd $(WEB_DIR) && $(PNPM) install

web-%: web-deps
	cd $(WEB_DIR) && $(PNPM) $*

embed-frontend: web-build build

db-migrate-up:
	$(GO) run -mod=mod cmd/migrate/main.go -direction up

db-migrate-down:
	$(GO) run -mod=mod cmd/migrate/main.go -direction down

db-migrate-steps:
	@echo "Usage: make db-migrate-steps STEPS=<n> (negative for down)"
	@if [ -z "$(STEPS)" ]; then echo "STEPS is required"; exit 1; fi
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
