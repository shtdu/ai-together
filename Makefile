# Makefile for Code Together

SERVER_DIR=server
MEMBER_DIR=member
MANAGER_DIR=manager
SERVER_BINARY_NAME=codetogether
SERVER_BUILD_DIR=$(SERVER_DIR)/bin
MEMBER_TASK=wails3 task
PNPM=pnpm

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: help
help: ## Display this help message
	@echo "Code Together - Build and Development"
	@echo ""
	@echo "Quick Start:"
	@echo "  make build              Build all components"
	@echo "  make dev                Run all in dev mode (server, member, manager)"
	@echo "  make test               Run all tests"
	@echo "  make clean              Remove build artifacts"
	@echo ""
	@echo "Release:"
	@echo "  make release            Build release binaries for current platform"
	@echo "  make release-all        Build release binaries for all platforms"
	@echo ""
	@echo "Docker:"
	@echo "  make docker             Build all Docker images (current arch)"
	@echo "  make docker-multi       Build and push multi-arch Docker images"
	@echo "  make docker-server      Build server Docker image"
	@echo "  make docker-manager     Build manager Docker image"
	@echo "  make docker-server-multi Build and push multi-arch server image"
	@echo "  make docker-manager-multi Build and push multi-arch manager image"
	@echo ""
	@echo "Module Commands:"
	@echo "  make server             Server commands"
	@echo "  make member             Member client commands"
	@echo "  make manager            Manager UI commands"
	@echo "  make tools              Build hook tools"
	@echo "  make docs               Compile manuals to HTML"

.PHONY: server
server: ## Show server-specific commands
	@echo "Server Commands:"
	@echo "  make server-build       Build server binary"
	@echo "  make server-production  Build production server binary (obfuscated with garble)"
	@echo "  make server-dev         Run with live reload (air)"
	@echo "  make server-test        Run tests"
	@echo "  make server-cover       Generate coverage report"
	@echo "  make server-cover-html  Generate HTML coverage report"
	@echo "  make server-linux       Build for Linux x64"
	@echo ""
	@echo "For more commands, run: cd server && make help"

.PHONY: member
member: ## Show member-specific commands
	@echo "Member Client Commands:"
	@echo "  make member-build       Build desktop app"
	@echo "  make member-dev         Run with hot reload (wails3)"
	@echo "  make member-test        Run tests"
	@echo "  make member-release     Build release package"
	@echo "  make member-win         Build Windows package (cross-compile)"
	@echo ""
	@echo "For more commands, run: cd member && make help"

.PHONY: manager
manager: ## Show manager-specific commands
	@echo "Manager UI Commands:"
	@echo "  make manager-build      Build React UI"
	@echo "  make manager-dev        Run dev server (Vite)"
	@echo "  make manager-lint       Lint TypeScript"
	@echo ""
	@echo "For more commands, run: cd manager && make help"

.PHONY: tools
tools: ## Show tools-specific commands
	@echo "Hook Tools Commands:"
	@echo "  make tools-build        Build hook-collector and hook-browser"
	@echo "  make hook-collector      Build hook-collector"
	@echo "  make hook-browser       Build hook-browser"
	@echo "  make shared-api-gen     Generate shared API"
	@echo ""

# Common tasks
.PHONY: build dev test clean shared-api-gen
build: server-build member-build manager-build

shared-api-gen: ## Regenerate shared API client from OpenAPI spec
	@cd shared/integration && go generate
	@echo "✓ Shared API client regenerated"
dev: server-dev member-dev ## Run server and member in dev mode
test: server-test member-test
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(SERVER_BUILD_DIR) $(MEMBER_DIR)/bin $(MANAGER_DIR)/dist hook-collector hook-browser
	@echo "Clean completed"

# Release tasks
.PHONY: release release-all
release: ## Build release for current platform
	@echo "Building release for $(GOOS)/$(GOARCH)..."
	@$(MAKE) server-production
	@$(MAKE) member-build
	@$(MAKE) manager-build
	@echo "Release build completed"

release-all: ## Build release for all platforms
	@echo "Building release for all platforms..."
	@$(MAKE) server-linux
	@cd $(MEMBER_DIR) && wails3 task darwin:package
	@cd $(MEMBER_DIR) && env ARCH=amd64 wails3 task windows:package
	@echo "All platform releases completed"

# Server targets
.PHONY: server-build server-production server-dev server-test server-cover server-cover-html server-linux
server-build:
	@mkdir -p $(SERVER_BUILD_DIR)
	@cd $(SERVER_DIR) && go build -o ./bin/$(SERVER_BINARY_NAME) .
	@echo "✓ Server built: $(SERVER_BUILD_DIR)/$(SERVER_BINARY_NAME)"

server-production:
	@mkdir -p $(SERVER_BUILD_DIR)
	@cd $(SERVER_DIR) && $(shell go env GOPATH)/bin/garble -literals -tiny build -o ./bin/$(SERVER_BINARY_NAME) .
	@echo "✓ Server built (obfuscated): $(SERVER_BUILD_DIR)/$(SERVER_BINARY_NAME)"

server-dev:
	@cd $(SERVER_DIR) && air

server-test:
	@cd $(SERVER_DIR) && go test ./... -v

server-cover:
	@cd $(SERVER_DIR) && go test $$(go list ./... | grep -v -E 'scripts|internal/db|repository|testutil') -coverprofile=coverage.out -covermode=atomic
	@cd $(SERVER_DIR) && go tool cover -func=coverage.out | tail -1

server-cover-html:
	@cd $(SERVER_DIR) && go test $$(go list ./... | grep -v -E 'scripts|internal/db|repository|testutil') -coverprofile=coverage.out -covermode=atomic
	@cd $(SERVER_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "✓ HTML coverage report: $(SERVER_DIR)/coverage.html"
	@open $(SERVER_DIR)/coverage.html 2>/dev/null || true

server-linux:
	@mkdir -p $(SERVER_BUILD_DIR)
	@cd $(SERVER_DIR) && GOOS=linux GOARCH=amd64 go build -o ./bin/$(SERVER_BINARY_NAME)-linux-amd64 .
	@echo "✓ Server built: $(SERVER_BUILD_DIR)/$(SERVER_BINARY_NAME)-linux-amd64"

# Member targets
.PHONY: member-build member-dev member-test member-release member-win
member-build:
	@cd $(MEMBER_DIR) && $(MEMBER_TASK) build

member-dev:
	@cd $(MEMBER_DIR) && $(MEMBER_TASK) dev

member-test:
	@cd $(MEMBER_DIR) && go test ./... -v

member-release:
	@cd $(MEMBER_DIR) && $(MEMBER_TASK) package

member-win:
	@cd $(MEMBER_DIR) && env ARCH=amd64 $(MEMBER_TASK) windows:package

# Manager targets
.PHONY: manager-build manager-dev manager-lint
manager-build:
	@cd $(MANAGER_DIR) && $(PNPM) build
	@echo "✓ Manager UI built: $(MANAGER_DIR)/dist"

manager-dev:
	@cd $(MANAGER_DIR) && $(PNPM) dev

 manager-lint:
	@cd $(MANAGER_DIR) && $(PNPM) lint

# Tools targets
.PHONY: tools-build hook-collector hook-browser
tools-build: hook-collector hook-browser

hook-collector:
	@echo "Building hook-collector..."
	@cd shared/tools/hook-collector && GOOS=darwin GOARCH=amd64 go build -o hook-collector_amd64 . && GOOS=darwin GOARCH=arm64 go build -o hook-collector_arm64 . &&  lipo -create -output hook-collector hook-collector_amd64 hook-collector_arm64 && rm hook-collector_amd64 hook-collector_arm64	
	@echo "✓ hook-collector built: hook-collector"

hook-browser:
	@echo "Building hook-browser..."
	@cd shared/tools/hook-browser && GOOS=darwin GOARCH=amd64 go build -o hook-browser_amd64 . && GOOS=darwin GOARCH=arm64 go build -o hook-browser_arm64 . &&  lipo -create -output hook-browser hook-browser_amd64 hook-browser_arm64 && rm hook-browser_amd64 hook-browser_arm64
	@echo "✓ hook-browser built: hook-browser"

# Documentation targets
.PHONY: docs
docs: ## Compile manuals to HTML
	@echo "Compiling manuals to HTML..."
	@mkdir -p server/docs
	@pandoc -s --toc --toc-depth=2 -f markdown docs/manuals/member.md -o server/docs/member.html
	@pandoc -s --toc --toc-depth=2 -f markdown docs/manuals/manager.md -o server/docs/manager.html
	@echo "✓ Member manual: server/docs/member.html"
	@echo "✓ Manager manual: server/docs/manager.html"
	@echo "Docs compilation completed"

# Docker targets
.PHONY: docker docker-server docker-manager docker-push docker-push-server docker-push-manager

DOCKER_TAG ?= latest
DOCKER_REGISTRY ?= genewoo
SERVER_IMAGE = $(DOCKER_REGISTRY)/ai-together-server
MANAGER_IMAGE = $(DOCKER_REGISTRY)/ai-together-manager

docker: docker-server docker-manager ## Build all Docker images

docker-server: ## Build server Docker image for current architecture
	@echo "Building server image (tag: $(DOCKER_TAG))..."
	docker build -f server/Dockerfile -t $(SERVER_IMAGE):$(DOCKER_TAG) .
	@echo "✓ Server image built: $(SERVER_IMAGE):$(DOCKER_TAG)"

docker-server-multi: ## Build server Docker image for multiple architectures (amd64, arm64)
	@echo "Building server image for linux/amd64,linux/arm64 (tag: $(DOCKER_TAG))..."
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f server/Dockerfile \
		-t $(SERVER_IMAGE):$(DOCKER_TAG) \
		--push .
	@echo "✓ Server multi-arch image built and pushed: $(SERVER_IMAGE):$(DOCKER_TAG)"

docker-manager: ## Build manager Docker image for current architecture
	@echo "Building manager image (tag: $(DOCKER_TAG))..."
	docker build -f manager/Dockerfile -t $(MANAGER_IMAGE):$(DOCKER_TAG) manager/
	@echo "✓ Manager image built: $(MANAGER_IMAGE):$(DOCKER_TAG)"

docker-manager-multi: ## Build manager Docker image for multiple architectures (amd64, arm64)
	@echo "Building manager image for linux/amd64,linux/arm64 (tag: $(DOCKER_TAG))..."
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f manager/Dockerfile \
		-t $(MANAGER_IMAGE):$(DOCKER_TAG) \
		--push .
	@echo "✓ Manager multi-arch image built and pushed: $(MANAGER_IMAGE):$(DOCKER_TAG)"

docker-multi: docker-server-multi docker-manager-multi ## Build and push all multi-arch Docker images

docker-push: docker-push-server docker-push-manager ## Push all Docker images

docker-push-server: ## Push server Docker image
	@echo "Pushing server image (tag: $(DOCKER_TAG))..."
	docker push $(SERVER_IMAGE):$(DOCKER_TAG)
	@echo "✓ Server image pushed: $(SERVER_IMAGE):$(DOCKER_TAG)"

docker-push-manager: ## Push manager Docker image
	@echo "Pushing manager image (tag: $(DOCKER_TAG))..."
	docker push $(MANAGER_IMAGE):$(DOCKER_TAG)
	@echo "✓ Manager image pushed: $(MANAGER_IMAGE):$(DOCKER_TAG)"
