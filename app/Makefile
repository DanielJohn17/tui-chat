# ----------------------------------------------------------------------------
# Environment Configuration
# ----------------------------------------------------------------------------
# Path to environment file
ENV_FILE ?= .env

# Load environment variables from .env file and export them to child processes
ifneq (,$(wildcard $(ENV_FILE)))
    include $(ENV_FILE)
    export $(shell sed -e 's/=.*//' -e 's/^export //' -e '/^#/d' -e '/^[[:space:]]*$$/d' $(ENV_FILE))
endif

# Default shell
SHELL := /bin/bash

# Database connection URL built from .env variables
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

BIN_DIR := bin
API_BIN := $(BIN_DIR)/api
TUI_BIN := $(BIN_DIR)/tui
SCHEMA_FILE := internal/api/db/schema.sql
GO_TUI_VERSION ?= v0.19.0

# Colors for terminal output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

.DEFAULT_GOAL := help

## ----------------------------------------------------------------------------
## Help
## ----------------------------------------------------------------------------
.PHONY: help
help: ## Display this help message
	@printf "$(CYAN)Usage:$(RESET) make [target]\n\n"
	@printf "$(CYAN)Available Targets:$(RESET)\n"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*?##/ { \
		category = $$1; \
		printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)

## ----------------------------------------------------------------------------
## Build Targets
## ----------------------------------------------------------------------------
.PHONY: build build-api build-tui
build: build-api build-tui ## Build both API and TUI binaries into bin/

build-api: ## Build the Go API binary
	@mkdir -p $(BIN_DIR)
	@printf "$(YELLOW)Building API binary -> $(API_BIN)...$(RESET)\n"
	@go build -o $(API_BIN) ./cmd/api
	@printf "$(GREEN)✓ API binary built successfully.$(RESET)\n"

build-tui: ## Build the Go TUI binary
	@mkdir -p $(BIN_DIR)
	@printf "$(YELLOW)Building TUI binary -> $(TUI_BIN)...$(RESET)\n"
	@go build -o $(TUI_BIN) ./cmd/tui
	@printf "$(GREEN)✓ TUI binary built successfully.$(RESET)\n"

## ----------------------------------------------------------------------------
## Execute / Run Targets
## ----------------------------------------------------------------------------
.PHONY: run-api run-tui dev-api
run-api: ## Run Go API server directly with go run
	@printf "$(YELLOW)Starting Go API server...$(RESET)\n"
	@go run ./cmd/api

run-tui: ## Run Go TUI application directly with go run
	@printf "$(YELLOW)Starting Go TUI client...$(RESET)\n"
	@go run ./cmd/tui

dev-api: ## Run API with live reloading (uses air if available, falls back to go run)
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml 2>/dev/null || air; \
	else \
		printf "$(YELLOW)air not found, falling back to go run ./cmd/api...$(RESET)\n"; \
		go run ./cmd/api; \
	fi

## ----------------------------------------------------------------------------
## Generation Targets
## ----------------------------------------------------------------------------
.PHONY: generate generate-api generate-tui generate-gsx sqlc generate-build-tui
generate: generate-api generate-tui ## Run all code generators (API sqlc + TUI gsx)

generate-api: ## Generate database queries and models with sqlc
	@printf "$(YELLOW)Running sqlc generate...$(RESET)\n"
	@sqlc generate
	@printf "$(GREEN)✓ sqlc code generation complete.$(RESET)\n"

sqlc: generate-api ## Alias for generate-api

generate-tui: ## Generate TUI GSX templates
	@printf "$(YELLOW)Generating go-tui GSX templates...$(RESET)\n"
	@go run github.com/grindlemire/go-tui/cmd/tui@$(GO_TUI_VERSION) generate ./...
	@printf "$(GREEN)✓ TUI GSX generation complete.$(RESET)\n"

generate-gsx: generate-tui ## Alias for generate-tui

generate-build-tui: generate-tui build-tui ## Generate GSX templates and build TUI binary

## ----------------------------------------------------------------------------
## Test Targets
## ----------------------------------------------------------------------------
.PHONY: test test-api test-tui test-cover
test: ## Run all tests across the project
	@printf "$(YELLOW)Running all project tests...$(RESET)\n"
	@go test -v ./...

test-api: ## Run API unit and integration tests
	@printf "$(YELLOW)Running API tests...$(RESET)\n"
	@go test -v ./internal/api/... ./test/...

test-tui: ## Run TUI tests
	@printf "$(YELLOW)Running TUI tests...$(RESET)\n"
	@go test -v ./internal/tui/...

test-cover: ## Run all tests and report coverage
	@printf "$(YELLOW)Running tests with coverage...$(RESET)\n"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@printf "$(GREEN)✓ Coverage report generated at coverage.out.$(RESET)\n"

## ----------------------------------------------------------------------------
## Database / Migration Targets (Environment from $(ENV_FILE))
## ----------------------------------------------------------------------------
.PHONY: check-db-env env-info migrate db-migrate migrate-status db-shell migrate-reset
check-db-env:
	@if [ ! -f "$(ENV_FILE)" ]; then \
		printf "$(YELLOW)Error: $(ENV_FILE) file not found. Please create $(ENV_FILE) with database credentials.$(RESET)\n" >&2; \
		exit 1; \
	fi
	@if [ -z "$(DB_NAME)" ] || [ -z "$(DB_USER)" ]; then \
		printf "$(YELLOW)Error: DB_NAME or DB_USER is not defined in $(ENV_FILE).$(RESET)\n" >&2; \
		exit 1; \
	fi

env-info: check-db-env ## Display database configuration loaded from .env
	@printf "$(CYAN)Database Configuration (loaded from $(ENV_FILE)):$(RESET)\n"
	@printf "  DB_HOST     = %s\n" "$(DB_HOST)"
	@printf "  DB_PORT     = %s\n" "$(DB_PORT)"
	@printf "  DB_USER     = %s\n" "$(DB_USER)"
	@printf "  DB_NAME     = %s\n" "$(DB_NAME)"
	@printf "  DB_URL      = postgres://$(DB_USER):******@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable\n"

migrate: check-db-env ## Apply database schema migration (schema.sql)
	@printf "$(YELLOW)Applying schema ($(SCHEMA_FILE)) to $(DB_NAME) on $(DB_HOST):$(DB_PORT) using $(ENV_FILE)...$(RESET)\n"
	@PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d "$(DB_NAME)" -f "$(SCHEMA_FILE)"
	@printf "$(GREEN)✓ Database migration applied successfully.$(RESET)\n"

db-migrate: migrate ## Alias for migrate

migrate-status: check-db-env ## Check database connection and list active tables
	@printf "$(YELLOW)Checking $(DB_NAME) tables on $(DB_HOST):$(DB_PORT) (from $(ENV_FILE))...$(RESET)\n"
	@PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d "$(DB_NAME)" -c "\dt"

db-shell: check-db-env ## Open interactive psql shell with credentials from .env
	@PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d "$(DB_NAME)"

migrate-reset: check-db-env ## Reset database schema (CAUTION: drops public schema and re-applies schema.sql)
	@printf "$(YELLOW)Resetting database public schema on $(DB_NAME) ($(DB_HOST):$(DB_PORT))...$(RESET)\n"
	@PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d "$(DB_NAME)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) migrate
	@printf "$(GREEN)✓ Database reset complete.$(RESET)\n"

## ----------------------------------------------------------------------------
## Quality & Clean Targets
## ----------------------------------------------------------------------------
.PHONY: fmt vet tidy clean
fmt: ## Format Go source code
	@printf "$(YELLOW)Formatting Go source files...$(RESET)\n"
	@go fmt ./...

vet: ## Run go vet code analysis
	@printf "$(YELLOW)Running go vet...$(RESET)\n"
	@go vet ./...

tidy: ## Tidy Go module dependencies
	@printf "$(YELLOW)Tidying Go modules...$(RESET)\n"
	@go mod tidy

clean: ## Clean build binaries, coverage files, and temporary artifacts
	@printf "$(YELLOW)Cleaning artifacts...$(RESET)\n"
	@rm -rf $(BIN_DIR) tmp/ coverage.out
	@printf "$(GREEN)✓ Clean complete.$(RESET)\n"
