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
MIGRATIONS_DIR := internal/api/db/migrations
GOOSE ?= goose
# Bubble Tea does not require code generation

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

build-tui: ## Build the Go TUI binary (optionally: make build-tui API_URL=https://...)
	@mkdir -p $(BIN_DIR)
	@printf "$(YELLOW)Building TUI binary -> $(TUI_BIN)...$(RESET)\n"
	@if [ -n "$(API_URL)" ]; then \
		go build -ldflags "-X main.DefaultAPIURL=$(API_URL)" -o $(TUI_BIN) ./cmd/tui; \
	else \
		go build -o $(TUI_BIN) ./cmd/tui; \
	fi
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
	@printf "$(GREEN)✓ Bubble Tea does not require code generation.$(RESET)\n"
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
.PHONY: check-db-env env-info migrate db-migrate migrate-down migrate-status migrate-version migrate-reset migrate-create db-shell
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

migrate: check-db-env ## Apply database migrations using goose
	@printf "$(YELLOW)Applying goose migrations up on $(DB_NAME)...$(RESET)\n"
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up
	@printf "$(GREEN)✓ Database migrations applied successfully.$(RESET)\n"

db-migrate: migrate ## Alias for migrate

migrate-down: check-db-env ## Roll back the most recent goose migration
	@printf "$(YELLOW)Rolling back latest goose migration on $(DB_NAME)...$(RESET)\n"
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down
	@printf "$(GREEN)✓ Migration rolled back successfully.$(RESET)\n"

migrate-status: check-db-env ## Check goose migration status
	@printf "$(YELLOW)Checking migration status for $(DB_NAME)...$(RESET)\n"
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migrate-version: check-db-env ## Print current goose migration version
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" version

migrate-reset: check-db-env ## Roll back all migrations and re-apply them
	@printf "$(YELLOW)Resetting and re-applying all migrations on $(DB_NAME)...$(RESET)\n"
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset
	@$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up
	@printf "$(GREEN)✓ Database reset and re-migrated successfully.$(RESET)\n"

migrate-create: ## Create a new SQL migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then \
		printf "$(YELLOW)Error: Migration name required. Usage: make migrate-create name=add_column_name$(RESET)\n" >&2; \
		exit 1; \
	fi
	@$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

db-shell: check-db-env ## Open interactive psql shell with credentials from .env
	@PGPASSWORD="$(DB_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d "$(DB_NAME)"

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
