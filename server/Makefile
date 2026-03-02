# Server Module Makefile

.PHONY: help
help: ## Display this help message
	@echo "Server Module Commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              Build server binary"
	@echo "  make run                Run server (builds first)"
	@echo "  make dev                Run with live reload (air)"
	@echo "  make linux              Build for Linux x64"
	@echo "  make production         Build production binary (obfuscated with garble)"
	@echo ""
	@echo "Test:"
	@echo "  make test               Run tests"
	@echo "  make itest              Run integration tests (requires PostgreSQL)"
	@echo ""
	@echo "Coverage:"
	@echo "  make cover              Generate coverage report (terminal)"
	@echo "  make cover-html         Generate HTML coverage report (opens in browser)"
	@echo "  make cover-func         Show coverage by function (sorted)"
	@echo ""
	@echo "Code Generation:"
	@echo "  make sqlc-gen           Regenerate database code from SQL"
	@echo ""
	@echo "Maintenance:"
	@echo "  make fmt                Format Go code"
	@echo "  make vet                Vet Go code"
	@echo "  make tidy               Run go mod tidy"
	@echo "  make deps               Download dependencies"
	@echo ""
	@echo "Quick commands from root:"
	@echo "  make server-build       Build from project root"
	@echo "  make server-dev         Run dev mode from root"

# Build & Run
.PHONY: build run dev linux
build:
	@mkdir -p bin
	@go build -o ./bin/switch-server .
	@echo "✓ Built: bin/switch-server"

run: build
	@./bin/switch-server

dev:
	@air

linux:
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o ./bin/switch-server-linux-amd64 .
	@echo "✓ Built: bin/switch-server-linux-amd64"

production:
	@mkdir -p bin
	@$(GOPATH)/bin/garble -literals -tiny build -o ./bin/switch-server .
	@echo "✓ Built (obfuscated): bin/switch-server"

# Test
.PHONY: test itest cover cover-html cover-func
test:
	@go test ./... -v

itest:
	@./scripts/test-integration.sh

# Coverage
cover:
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.out"

cover-html:
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report: coverage.html"
	@open coverage.html 2>/dev/null || true

cover-func:
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | sort -k3 -n

# Code Generation
.PHONY: sqlc-gen
sqlc-gen:
	@sqlc generate
	@echo "✓ Database code regenerated"
	@echo "  Generated: internal/db/*.go"
	@echo "  Source: internal/db/providers.sql"

# Maintenance
.PHONY: fmt vet tidy deps
fmt:
	@go fmt ./...

vet:
	@go vet ./...

tidy:
	@go mod tidy

deps:
	@go mod download
