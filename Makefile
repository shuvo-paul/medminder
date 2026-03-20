.PHONY := build dev start run test test-cover tidy clean web-install web-dev web-build web-preview embed-frontend

GO ?= go
PNPM ?= pnpm
AIR ?= air
BIN_DIR := bin
BIN := $(BIN_DIR)/medminder
CMD := cmd/server/main.go
WEB_DIR := web

build:
	@mkdir -p $(BIN_DIR)
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
