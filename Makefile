# git-wt Makefile
# Development workflow:
#   make dev-mode      - Switch to development binary (~/go/bin)
#   make homebrew-mode - Switch to released binary (homebrew)
# Release workflow:
#   make release-check    - Validate goreleaser config
#   make release-snapshot - Build release locally (no publish)
#   make release-alpha    - Tag and release alpha version
#   make release          - Tag and release stable version

.PHONY: help build build-all install install-global uninstall clean test lint dev dev-mode homebrew-mode
.PHONY: release-check release-snapshot release-local release-alpha release
.DEFAULT_GOAL := help

BINARY := git-wt
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/raisedadead/git-wt/internal/commands.version=$(VERSION)"

# Default install location (Go-idiomatic)
PREFIX ?= $(shell go env GOPATH)
BINDIR := $(PREFIX)/bin

help:
	@echo "Development Workflow:"
	@echo "  make dev-mode       - Switch to dev binary (removes /usr/local/bin, installs to ~/go/bin)"
	@echo "  make homebrew-mode  - Switch to released binary (reinstalls from homebrew)"
	@echo ""
	@echo "Build & Install:"
	@echo "  make build          - Build to ./bin/"
	@echo "  make install        - Build and install to ~/go/bin"
	@echo "  make install-global - Build and install to /usr/local/bin"
	@echo "  make uninstall      - Remove from ~/go/bin"
	@echo ""
	@echo "Development:"
	@echo "  make dev            - Build and show version"
	@echo "  make test           - Run all tests"
	@echo "  make build-all      - Cross-platform build check (linux/darwin/windows)"
	@echo "  make lint           - Run go vet and golangci-lint"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Release (mirrors CI):"
	@echo "  make release-check    - Validate goreleaser config"
	@echo "  make release-snapshot - Build release locally without publishing"
	@echo "  make release-alpha    - Create alpha tag and release (skips homebrew)"
	@echo "  make release          - Create release tag and publish (includes homebrew)"

# Build to ./bin/
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/git-wt
	@echo "Built bin/$(BINARY) ($(VERSION))"

# Install to ~/go/bin (default) - for development
install: build
	@mkdir -p $(BINDIR)
	cp bin/$(BINARY) $(BINDIR)/$(BINARY)
	@echo "Installed to $(BINDIR)/$(BINARY)"

# Install to /usr/local/bin - overwrites homebrew version
install-global: build
	cp bin/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

uninstall:
	rm -f $(BINDIR)/$(BINARY)
	@echo "Removed $(BINDIR)/$(BINARY)"

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./...

# Cross-platform build check (catches platform-specific issues)
build-all:
	GOOS=linux GOARCH=amd64 go build ./...
	GOOS=linux GOARCH=arm64 go build ./...
	GOOS=darwin GOARCH=amd64 go build ./...
	GOOS=darwin GOARCH=arm64 go build ./...
	GOOS=windows GOARCH=amd64 go build ./...
	GOOS=windows GOARCH=arm64 go build ./...
	@echo "All platforms build successfully"

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# Quick rebuild and show version
dev: build
	@./bin/$(BINARY) --version

# Switch to development mode: use local build instead of homebrew
dev-mode: install
	@rm -f /usr/local/bin/$(BINARY)
	@echo "Removed /usr/local/bin/$(BINARY)"
	@echo "Now using: $$(which $(BINARY))"
	@$(BINARY) --version

# Switch to homebrew mode: use released version
homebrew-mode:
	@rm -f $(BINDIR)/$(BINARY)
	@brew reinstall raisedadead/tap/git-wt
	@echo "Now using: $$(which $(BINARY))"
	@$(BINARY) --version

# =============================================================================
# Release targets (mirror CI workflow)
# =============================================================================

# Validate goreleaser configuration
release-check:
	goreleaser check

# Build release locally without publishing (for testing)
release-snapshot:
	goreleaser release --snapshot --clean

# Release with local token (uses gh auth token)
# This is what CI does, but locally
release-local:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG is required. Usage: make release-local TAG=v0.1.0"; \
		exit 1; \
	fi
	GITHUB_TOKEN=$$(gh auth token) goreleaser release --clean

# Create alpha release (skips homebrew automatically via skip_upload: auto)
# Usage: make release-alpha [ALPHA=2]  (defaults to incrementing from last alpha)
release-alpha: test lint build-all
	@LAST_TAG=$$(git tag -l "v*-alpha.*" --sort=-v:refname | head -1); \
	if [ -z "$(ALPHA)" ]; then \
		if [ -z "$$LAST_TAG" ]; then \
			NEW_TAG="v0.1.0-alpha.1"; \
		else \
			BASE=$$(echo $$LAST_TAG | sed 's/-alpha\.[0-9]*$$//'); \
			NUM=$$(echo $$LAST_TAG | grep -o 'alpha\.[0-9]*' | grep -o '[0-9]*'); \
			NEW_TAG="$$BASE-alpha.$$((NUM + 1))"; \
		fi; \
	else \
		BASE=$$(git tag -l "v*" --sort=-v:refname | grep -v alpha | grep -v beta | grep -v rc | head -1 || echo "v0.1.0"); \
		if [ -z "$$BASE" ]; then BASE="v0.1.0"; fi; \
		NEW_TAG="$$BASE-alpha.$(ALPHA)"; \
	fi; \
	echo "Creating tag: $$NEW_TAG"; \
	git tag $$NEW_TAG && \
	git push origin $$NEW_TAG && \
	echo "Pushed $$NEW_TAG - GitHub Actions will create the release"

# Create stable release (includes homebrew)
# Usage: make release VERSION=0.1.0
release: test lint build-all
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=0.1.0"; \
		exit 1; \
	fi
	@echo "Creating tag: v$(VERSION)"
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"
	@echo "Pushed v$(VERSION) - GitHub Actions will create the release"
