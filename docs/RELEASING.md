# Releasing git-wt

This document describes how to release git-wt when the project is ready.

## Prerequisites

1. **Create Homebrew tap repository**
   ```bash
   # Create raisedadead/homebrew-tap on GitHub
   ```

2. **Add repository secret**
   - Go to repo Settings → Secrets → Actions
   - Add `HOMEBREW_TAP_GITHUB_TOKEN` - a PAT with repo access to `raisedadead/homebrew-tap`

## Release Workflow

Add `.github/workflows/release.yaml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

## Creating a Release

```bash
# Ensure you're on main with all changes committed
git checkout main
git pull

# Create and push tag
git tag v0.1.0
git push origin v0.1.0
```

## What Happens on Release

1. Tests run
2. GoReleaser builds binaries:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
3. GitHub Release created with:
   - Binary archives (.tar.gz, .zip for Windows)
   - Checksums file
   - Auto-generated changelog
4. Homebrew formula updated in `raisedadead/homebrew-tap`

## Installation After Release

```bash
# Homebrew
brew install raisedadead/tap/git-wt

# Go
go install github.com/raisedadead/git-wt/cmd/git-wt@latest

# Manual
# Download from GitHub Releases
```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- `v0.x.x` - Pre-release, API may change
- `v1.0.0` - First stable release
- `vX.Y.Z` - Major.Minor.Patch

## Local Testing

Test the release process locally without publishing:

```bash
goreleaser release --snapshot --clean
```

This creates binaries in `dist/` without pushing anything.
