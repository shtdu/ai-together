# Makefile for shared tools
# Builds all CLI tools into ../bin directory

# Module paths
MODULE := github.com/code-together/shared
TOOLS_DIR := ./tools
BIN_DIR := ../bin

# Tools to build (directory name = binary name)
TOOLS := hook-browser hook-collector hook-schema-analyzer hook-events-tool

# Tools that are separate Go modules (have their own go.mod)
MODULE_TOOLS := hook-browser hook-collector

# Tools that are packages in the main module
PACKAGE_TOOLS := hook-schema-analyzer hook-events-tool

# Go build flags
GO := go
GOFLAGS := -v
LDFLAGS := -s -w

# Platform defaults
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: all build clean help

all: build

## build: Build all tools
build:
	@mkdir -p $(BIN_DIR)
	@echo "Building module tools..."
	@for tool in $(MODULE_TOOLS); do \
		echo "Building $$tool..."; \
		(cd $(TOOLS_DIR)/$$tool && $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o ../../$(BIN_DIR)/$$tool .); \
	done
	@echo "Building package tools..."
	@for tool in $(PACKAGE_TOOLS); do \
		echo "Building $$tool..."; \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$$tool ./tools/$$tool; \
	done

## clean: Remove all built binaries
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)

## help: Show this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
