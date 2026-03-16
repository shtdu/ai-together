# BDD Test Suite Makefile
# Behavioral-Driven Development tests using godog framework

.PHONY: help
help: ## Show this help message
	@echo "BDD Test Suite - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

.PHONY: test-bdd
test-bdd: ## Run all BDD tests
	@echo "Running BDD tests..."
	@$(MAKE) -s ensure-test-server
	@cd godog && go test -v -godog.format=pretty

.PHONY: test-bdd-ci
test-bdd-ci: ## Run BDD tests with JUnit output for CI/CD
	@echo "Running BDD tests with JUnit output..."
	@$(MAKE) -s ensure-test-server
	@cd godog && go test -v -godog.format=junit -godog.paths="../features" -godog.format=junit > test-results.xml 2>&1 || true
	@echo "Test results saved to test-results.xml"

.PHONY: test-bdd-coverage
test-bdd-coverage: ## Run BDD tests with coverage
	@echo "Running BDD tests with coverage..."
	@$(MAKE) -s ensure-test-server
	@cd godog && go test -v -godog.format=pretty -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-bdd-tags
test-bdd-tags: ## Run BDD tests with specific tags (use: make test-bdd-tags tags="@smoke")
	@echo "Running BDD tests with tags: $(tags)"
	@$(MAKE) -s ensure-test-server
	@cd godog && go test -v -godog.format=pretty -godog.tags="$(tags)"

.PHONY: ensure-test-server
ensure-test-server: ## Ensure test server is running
	@echo "Checking test server status..."
	@curl -s http://localhost:8088/health > /dev/null 2>&1 || \
		(echo "Starting test server..." && \
		 cd ../integration && \
		 ./test-server.sh > /tmp/test-server.log 2>&1 & \
		 sleep 3 && \
		 echo "Test server started")

.PHONY: clean
clean: ## Clean test artifacts
	@echo "Cleaning BDD test artifacts..."
	@rm -f test-results.xml
	@rm -f coverage.out
	@rm -f coverage.html
	@rm -rf /tmp/test-server.log
	@echo "Clean complete"

.PHONY: deps
deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Vet Go code
	@echo "Vetting Go code..."
	@go vet ./...

.PHONY: lint
lint: fmt vet ## Run linters (fmt + vet)

.PHONY: test-server-logs
test-server-logs: ## Show test server logs
	@echo "Test server logs:"
	@cat /tmp/test-server.log 2>/dev/null || echo "No test server logs found"

.PHONY: test-server-health
test-server-health: ## Check test server health
	@echo "Checking test server health..."
	@curl -s http://localhost:8088/health || echo "Test server is not responding"

.DEFAULT_GOAL := help
