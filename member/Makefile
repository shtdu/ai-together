# Member Module Makefile

.PHONY: help
help: ## Display this help message
	@echo "Member Module Commands (Desktop App)"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              Build desktop app"
	@echo "  make run                Run app (builds first)"
	@echo "  make dev                Run with hot reload (wails3)"
	@echo "  make package            Package release bundle"
	@echo ""
	@echo "Cross-compile (requires mingw-w64):"
	@echo "  make win                Build Windows package"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend-dev       Run Vite dev server (port 9245)"
	@echo "  make frontend-build     Build frontend for production"
	@echo "  make frontend-install   Install frontend dependencies"
	@echo "  make frontend-lint      Run ESLint on frontend code"
	@echo ""
	@echo "Code Generation:"
	@echo "  make api-gen            Regenerate API client from OpenAPI spec"
	@echo "  make update-assets      Update build metadata"
	@echo ""
	@echo "Test:"
	@echo "  make test               Run tests"
	@echo "  make cli-test           Run CLI tests"
	@echo ""
	@echo "CLI Tool:"
	@echo "  make cli-build          Build CLI tool (member-cli)"
	@echo "  make cli-clean          Clean CLI binary"
	@echo ""
	@echo "Maintenance:"
	@echo "  make fmt                Format Go code"
	@echo "  make vet                Vet Go code"
	@echo "  make tidy               Run go mod tidy"
	@echo "  make deps               Download dependencies"
	@echo ""
	@echo "Wails3 Tasks (use directly):"
	@echo "  wails3 task build       Build for current platform"
	@echo "  wails3 task package     Package for current platform"
	@echo "  wails3 task darwin:build/package  macOS specific"
	@echo "  wails3 task windows:build/package Windows specific"
	@echo ""
	@echo "Quick commands from root:"
	@echo "  make member-build       Build from project root"
	@echo "  make member-dev         Run dev mode from root"

# Build & Run (via wails3)
WAILS_TASK=wails3 task
.PHONY: build run dev package win
build:
	@$(WAILS_TASK) build

run:
	@$(WAILS_TASK) run

dev:
	@$(WAILS_TASK) dev

package:
	@$(WAILS_TASK) package

win:
	@env ARCH=amd64 $(WAILS_TASK) windows:package

# Frontend
.PHONY: frontend-dev frontend-build frontend-install frontend-lint
frontend-dev:
	@cd frontend && npm run dev

frontend-build:
	@cd frontend && npm run build

frontend-install:
	@cd frontend && npm install

frontend-lint:
	@cd frontend && npm run lint

# Code Generation
.PHONY: api-gen update-assets
api-gen:
	@cd ../../shared/integration && go generate
	@echo "✓ API client regenerated"
	@echo "  Generated: shared/integration/generated.go"
	@echo "  Source: docs/client_api/server_api.yaml"

update-assets:
	@wails3 update build-assets -name "CodeTogether" -binaryname "CodeTogether" -config build/config.yml -dir build
	@echo "✓ Build assets updated"

# Test
.PHONY: test cli-test
test:
	@go test ./... -v

cli-test:
	@go test ./cmd/cli/... -v

# CLI Tool
.PHONY: cli-build cli-clean
cli-build:
	@echo "Building CLI tool..."
	@go build -o bin/member-cli ./cmd/cli
	@echo "✓ CLI binary built: bin/member-cli"

cli-clean:
	@echo "Cleaning CLI binary..."
	@rm -f bin/member-cli
	@echo "✓ CLI binary cleaned"

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
