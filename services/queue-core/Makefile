# --- Makefile for Queue-Core ---

# Go command
GO := go

# Directories
CMD_DIR := ./cmd
INTERNAL_DIR := ./internal
TEST_DIR := ./tests

# Targets
.PHONY: all build run clean fmt lint test unit integration load

# Default: build
all: build

# Build binary
build:
	$(GO) build -o queue-core $(CMD_DIR)/main.go

# Run binary
run: build
	./queue-core

# Clean binary
clean:
	rm -f queue-core

# Format code
fmt:
	$(GO) fmt ./...

# Lint code (requires golangci-lint installed)
lint:
	golangci-lint run

# Run all tests
test: unit integration load

# Unit tests
unit:
	$(GO) test -v $(TEST_DIR)/unit/...

# Integration tests
integration:
	$(GO) test -v $(TEST_DIR)/integration/...

# Load tests
load:
	$(GO) test -v $(TEST_DIR)/load/...
