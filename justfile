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

# Validate goreleaser config, or build a local snapshot
release target="check":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{target}}" in
        check)
            goreleaser check
            ;;
        snapshot)
            HOMEBREW_TAP_GITHUB_TOKEN=snapshot goreleaser release --snapshot --clean
            ;;
        alpha|patch|minor|major)
            just test lint build-all
            just _release-tag "{{target}}"
            ;;
        *)
            echo "Usage: just release [check|snapshot|alpha|patch|minor|major]"
            echo "  check    - validate goreleaser config (default)"
            echo "  snapshot - local dry-run build"
            echo "  alpha    - auto-increment alpha pre-release"
            echo "  patch    - bump patch version (e.g. 0.1.3 → 0.1.4)"
            echo "  minor    - bump minor version (e.g. 0.1.3 → 0.2.0)"
            echo "  major    - bump major version (e.g. 0.1.3 → 1.0.0)"
            exit 1
            ;;
    esac

# Internal: compute next tag and push (not meant to be called directly)
_release-tag type:
    #!/usr/bin/env bash
    set -euo pipefail

    # Get latest stable tag (ignore pre-releases)
    LATEST_STABLE=$(git tag -l "v[0-9]*" --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1)
    LATEST_STABLE="${LATEST_STABLE:-v0.0.0}"

    # Parse major.minor.patch
    VERSION="${LATEST_STABLE#v}"
    IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

    case "{{type}}" in
        alpha)
            # Find latest alpha for the *next* patch version
            NEXT_PATCH="v${MAJOR}.${MINOR}.$((PATCH + 1))"
            LAST_ALPHA=$(git tag -l "${NEXT_PATCH}-alpha.*" --sort=-v:refname | head -1)
            if [ -z "$LAST_ALPHA" ]; then
                NEW_TAG="${NEXT_PATCH}-alpha.1"
            else
                NUM=$(echo "$LAST_ALPHA" | grep -o 'alpha\.[0-9]*' | grep -o '[0-9]*')
                NEW_TAG="${NEXT_PATCH}-alpha.$((NUM + 1))"
            fi
            ;;
        patch)
            NEW_TAG="v${MAJOR}.${MINOR}.$((PATCH + 1))"
            ;;
        minor)
            NEW_TAG="v${MAJOR}.$((MINOR + 1)).0"
            ;;
        major)
            NEW_TAG="v$((MAJOR + 1)).0.0"
            ;;
    esac

    echo "Latest stable: $LATEST_STABLE"
    echo "Creating tag:  $NEW_TAG"
    echo ""
    git tag "$NEW_TAG"
    git push origin "$NEW_TAG"
    echo "Pushed $NEW_TAG — GitHub Actions will create the release"

# === Test repo setup ===
# Run once to set up experiments-by-mrugesh/test-repo for integration tests
setup-test-repo:
    ./scripts/setup-test-repo.sh
