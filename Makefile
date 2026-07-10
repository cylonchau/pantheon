# Build configuration
.DEFAULT_GOAL := help
GOCMD=go
GOBUILD=$(GOCMD) build
GOBUILD_DIR=cmd
OUT_DIR ?= target
BIN_DIR := $(OUT_DIR)/bin
BUILDOPTS ?= -v

# Extract modules
modules := $(wildcard $(GOBUILD_DIR)/*)
SUBDIRS := $(notdir $(modules))

.PHONY: all build modules clean help test cover lint swagger

all: modules

# Run linter
lint:
	@echo "Running golangci-lint on pkg directory..."
	@golangci-lint run ./pkg/...

# Build all modules for the current platform
modules:
	@for dir in $(SUBDIRS); do \
		echo "Building module $$dir..."; \
		chmod +x scripts/build.sh && scripts/build.sh $$dir; \
	done

# Build a specific module
build:
	@if [ -z "$(module)" ]; then \
		echo "No module specified. Usage: make build module=<subdir>"; \
		exit 1; \
	fi
	@chmod +x scripts/build.sh && scripts/build.sh $(module)

# Run pkg tests with coverage summary
test:
	@echo "Running pkg unit tests..."
	@$(GOCMD) test -v -cover ./pkg/...

# Run pkg tests and generate HTML coverage report
cover:
	@echo "Generating coverage report for pkg..."
	@$(GOCMD) test -v -coverprofile=coverage.out ./pkg/...
	@$(GOCMD) tool cover -func=coverage.out
	@echo "HTML report generated at coverage.html"
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html

swagger:
	@echo "Generating swagger docs..."
	swag init -g cmd/pantheon-server/main.go -o docs --parseDependency --parseInternal

clean:
	@echo "Cleaning output directory..."
	@rm -rf $(OUT_DIR)
	@rm -f coverage.out coverage.html
	@echo "Done."

help:
	@echo "Available commands:"
	@echo "  make all             - Build all modules for current platform"
	@echo "  make modules         - Build all modules in $(GOBUILD_DIR)"
	@echo "  make build module=X  - Build a specific module (e.g., make build module=pantheon-server)"
	@echo "  make clean           - Remove $(OUT_DIR) directory and coverage files"
	@echo "  make swagger         - Generate swagger documentation"
	@echo "  make test            - Run all unit tests with coverage summary"
	@echo "  make cover           - Run tests and generate HTML coverage report"
	@echo "  make lint            - Run golangci-lint for code quality check"
	@echo "  make help            - Show this help message"
	@echo ""
	@echo "Example:"
	@for dir in $(SUBDIRS); do \
		echo "  make build module=$$dir"; \
	done
