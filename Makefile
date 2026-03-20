.PHONY := build run test test-cover tidy clean web-install web-dev web-build web-preview embed-frontend

GO ?= go
PNPM ?= pnpm
BIN_DIR := bin
BIN := $(BIN_DIR)/medminder
CMD := cmd/server/main.go
WEB_DIR := web

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(CMD)

run: web-build
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
