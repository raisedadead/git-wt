# wt development tasks

set shell := ["bash", "-uc"]

# Default recipe - show help
default:
    @just --list

# === Setup (run once after clone) ===

# Install dev tools and git hooks
setup: tools
    lefthook install
    @echo "✓ Git hooks installed"

# Install dev tool dependencies
tools:
    go install github.com/evilmartians/lefthook@latest
    go install gotest.tools/gotestsum@latest
    @echo "✓ Tools installed"

# Build binary to ./bin/
build:
    @mkdir -p bin
    go build -ldflags "-X github.com/raisedadead/wt/internal/commands.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o bin/wt ./cmd/wt
    @echo "Built bin/wt"

# Run unit tests
test-unit:
    gotestsum --format testdox -- ./internal/...

# Run integration tests (builds first)
test-integration: build
    gotestsum --format testdox -- ./test/integration/... -count=1

# Run all tests
test: test-unit test-integration

# Run integration tests in parallel
test-integration-parallel: build
    gotestsum --format testdox -- ./test/integration/... -count=1 -parallel=4

# Run specific test by pattern
test-match pattern: build
    gotestsum --format testdox -- ./test/integration/... -run "{{pattern}}" -count=1

# Run unit tests for specific package
test-pkg pkg:
    gotestsum --format testdox -- ./internal/{{pkg}}/...

# Clean build artifacts and test fixtures
clean:
    rm -rf bin/
    rm -rf test/integration/testdata/fixture
    go clean

# Format code
fmt:
    go fmt ./...
    @command -v goimports >/dev/null 2>&1 && goimports -w . || true

# Lint code
lint:
    go vet ./...
    @command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

# Build and show version
dev: build
    @echo ""
    @echo "Dev binary: ./bin/wt"
    @echo "Version:    $(./bin/wt --version)"
    @echo ""
    @echo "Test with:  ./bin/wt <command>"
    @echo "Or alias:   alias wt-dev='$(pwd)/bin/wt'"

# Cross-platform build check
build-all:
    GOOS=linux GOARCH=amd64 go build ./...
    GOOS=linux GOARCH=arm64 go build ./...
    GOOS=darwin GOARCH=amd64 go build ./...
    GOOS=darwin GOARCH=arm64 go build ./...
    GOOS=windows GOARCH=amd64 go build ./...
    GOOS=windows GOARCH=arm64 go build ./...
    @echo "All platforms build successfully"

# === Release targets ===

# Validate goreleaser config
release-check:
    goreleaser check

# Build release locally (no publish)
release-snapshot:
    HOMEBREW_TAP_GITHUB_TOKEN=snapshot goreleaser release --snapshot --clean

# Release with local token (uses gh auth token)
release-local:
    GITHUB_TOKEN=$(gh auth token) HOMEBREW_TAP_GITHUB_TOKEN=$(gh auth token) goreleaser release --clean

# Create alpha release (auto-increments from last alpha tag)
release-alpha: test lint build-all
    #!/usr/bin/env bash
    set -euo pipefail
    LAST_TAG=$(git tag -l "v*-alpha.*" --sort=-v:refname | head -1)
    if [ -z "$LAST_TAG" ]; then
        NEW_TAG="v0.1.0-alpha.1"
    else
        BASE=$(echo $LAST_TAG | sed 's/-alpha\.[0-9]*$//')
        NUM=$(echo $LAST_TAG | grep -o 'alpha\.[0-9]*' | grep -o '[0-9]*')
        NEW_TAG="$BASE-alpha.$((NUM + 1))"
    fi
    echo "Creating tag: $NEW_TAG"
    git tag $NEW_TAG
    git push origin $NEW_TAG
    echo "Pushed $NEW_TAG - GitHub Actions will create the release"

# Create stable release (includes homebrew)
release version: test lint build-all
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Creating tag: v{{version}}"
    git tag "v{{version}}"
    git push origin "v{{version}}"
    echo "Pushed v{{version}} - GitHub Actions will create the release"

# === Test repo setup ===
# Run once to set up experiments-by-mrugesh/test-repo for integration tests
setup-test-repo:
    ./scripts/setup-test-repo.sh
