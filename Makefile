.DEFAULT_GOAL := help

GO ?= go
APP ?= go-ddd
COMPOSE ?= docker compose

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "%-20s %s\n", $$1, $$2}'

.PHONY: run serve build
run: ## Run app (default command -> serve)
	$(GO) run .

serve: ## Run explicit serve command
	$(GO) run . serve

build: ## Build binary to bin/go-ddd
	@mkdir -p bin
	$(GO) build -o bin/$(APP) .

.PHONY: test fmt vet tidy vendor
test: ## Run all tests
	$(GO) test ./...

fmt: ## Run go fmt
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

tidy: ## Run go mod tidy
	$(GO) mod tidy

vendor: ## Sync vendor directory
	$(GO) mod vendor

.PHONY: db-up db-down db-logs db-ps
db-up: ## Start postgres via docker compose
	$(COMPOSE) up -d postgres

db-down: ## Stop compose services
	$(COMPOSE) down

db-logs: ## Tail postgres logs
	$(COMPOSE) logs -f postgres

db-ps: ## Show compose service status
	$(COMPOSE) ps

.PHONY: migrate-up migrate-down migrate-version migrate-force
migrate-up: ## Run migration up (optional: STEPS=1)
	$(GO) run . migrate --command up $(if $(STEPS),--steps $(STEPS),)

migrate-down: ## Run migration down (optional: STEPS=1)
	$(GO) run . migrate --command down $(if $(STEPS),--steps $(STEPS),)

migrate-version: ## Show current migration version
	$(GO) run . migrate --command version

migrate-force: ## Force migration version (required: VERSION=<n>)
	@test -n "$(VERSION)" || (echo "VERSION is required. Example: make migrate-force VERSION=3" && exit 1)
	$(GO) run . migrate --command force --version $(VERSION)
