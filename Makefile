.PHONY := build run test test-cover tidy clean

GO ?= go
BIN_DIR := bin
BIN := $(BIN_DIR)/medminder
CMD := cmd/server/main.go

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) $(CMD)

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
