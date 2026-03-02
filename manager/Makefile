# Manager Module Makefile

.PHONY: help
help: ## Display this help message
	@echo "Manager Module Commands (React UI)"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              Build React UI for production"
	@echo "  make dev                Run dev server with hot reload (Vite)"
	@echo "  make preview            Preview production build"
	@echo ""
	@echo "Dependencies:"
	@echo "  make install            Install dependencies"
	@echo ""
	@echo "Lint:"
	@echo "  make lint               Lint TypeScript/React code"
	@echo "  make lint-fix           Fix linting issues"
	@echo ""
	@echo "Direct pnpm commands:"
	@echo "  pnpm dev                Run dev server (port 3000)"
	@echo "  pnpm build              Build for production"
	@echo "  pnpm lint               Run ESLint"
	@echo ""
	@echo "Quick commands from root:"
	@echo "  make manager-build      Build from project root"
	@echo "  make manager-dev        Run dev mode from root"

PNPM=pnpm

# Build & Run
.PHONY: build dev preview
build:
	@$(PNPM) build
	@echo "✓ Built: dist/"

dev:
	@$(PNPM) dev

preview:
	@$(PNPM) preview

# Dependencies
.PHONY: install
install:
	@$(PNPM) install

# Lint
.PHONY: lint lint-fix
lint:
	@$(PNPM) lint

lint-fix:
	@$(PNPM) lint --fix
