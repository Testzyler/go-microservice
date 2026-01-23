GO ?= go
BUF ?= buf
K6 ?= k6
ENV_FILES ?= .env
E2E_BASE_URL ?= http://localhost:8080

.PHONY: deps fmt buf-gen unit test e2e perf gateway auth auth-workers auth-scheduler serve-all docker-up docker-down

deps:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

buf-gen:
	cd global/proto && $(BUF) generate

unit:
	ENV_FILES=$(ENV_FILES) $(GO) test ./...

test: unit

e2e:
	cd tests/e2e && E2E_BASE_URL=$(E2E_BASE_URL) bun run test:e2e

perf:
	BASE_URL?=http://localhost:8080 $(K6) run tests/k6/auth.js

gateway:
	ENV_FILES=$(ENV_FILES) $(GO) run . serve-internal-gateway

auth:
	ENV_FILES=$(ENV_FILES) $(GO) run . serve-auth

auth-workers:
	ENV_FILES=$(ENV_FILES) $(GO) run . serve-auth-worker mailer

auth-scheduler:
	ENV_FILES=$(ENV_FILES) $(GO) run . serve-auth-scheduler

serve-all:
	ENV_FILES=$(ENV_FILES) $(GO) run . serve-all

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v
