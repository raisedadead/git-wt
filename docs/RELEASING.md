# Releasing git-wt

This document describes the release process for git-wt.

## Prerequisites

- [goreleaser](https://goreleaser.com/) installed locally (for snapshots)
- [gh CLI](https://cli.github.com/) authenticated (for local releases)
- Push access to the repository

## Release Types

| Type   | Tag Format       | Homebrew | Use Case                  |
| ------ | ---------------- | -------- | ------------------------- |
| Alpha  | `v0.1.0-alpha.1` | Skipped  | Testing, early feedback   |
| Beta   | `v0.1.0-beta.1`  | Skipped  | Feature complete, testing |
| RC     | `v0.1.0-rc.1`    | Skipped  | Release candidate         |
| Stable | `v0.1.0`         | Updated  | Production release        |

## Quick Reference

```bash
# Validate config before releasing
make release-check

# Test release locally (no publish)
make release-snapshot

# Create alpha release (auto-increments)
make release-alpha

# Create stable release
make release VERSION=0.1.0
```

## Alpha Releases

Alpha releases are for testing and don't update Homebrew.

```bash
# Auto-increment from last alpha tag
make release-alpha
# v0.1.0-alpha.1 → v0.1.0-alpha.2 → v0.1.0-alpha.3

# Or specify a number
make release-alpha ALPHA=5
# Creates v0.1.0-alpha.5
```

**What happens:**

1. Runs tests and lint
2. Cross-platform build check (linux/darwin/windows)
3. Creates and pushes tag
4. GitHub Actions runs goreleaser
5. GitHub Release created with binaries
6. Homebrew is **skipped** (via `skip_upload: auto`)

## Stable Releases

Stable releases update Homebrew.

```bash
make release VERSION=0.1.0
```

**What happens:**

1. Runs tests and lint
2. Cross-platform build check
3. Creates and pushes `v0.1.0` tag
4. GitHub Actions runs goreleaser
5. GitHub Release created
6. Homebrew formula updated in `raisedadead/homebrew-tap`

## Local Release (Manual)

For debugging or when CI isn't available:

```bash
# Test build without publishing
make release-snapshot

# Full release with local token
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
```

## CI Workflow

The release is automated via `.github/workflows/release.yaml`:

1. Triggered by tag push (`v*`)
2. Runs tests
3. Runs goreleaser with `GITHUB_TOKEN`
4. Creates GitHub Release with:
   - Binary archives (tar.gz, zip for Windows)
   - Checksums file
   - Auto-generated changelog

## Versioning

Follow [Semantic Versioning](https://semver.org/):

| Version  | Meaning                     |
| -------- | --------------------------- |
| `v0.x.x` | Pre-release, API may change |
| `v1.0.0` | First stable release        |
| `vX.Y.Z` | Major.Minor.Patch           |

**Increment rules:**

- **Major**: Breaking changes
- **Minor**: New features (backward compatible)
- **Patch**: Bug fixes

## Troubleshooting

### Release failed - platform-specific code

```bash
# Check all platforms build before releasing
make build-all
```

### goreleaser config error

```bash
# Validate configuration
make release-check
```

### Missing GitHub token

```bash
# Use gh CLI token
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
```

## Post-Release

After a stable release:

1. Verify GitHub Release page has all artifacts
2. Test Homebrew installation:
   ```bash
   brew update
   brew install raisedadead/tap/git-wt
   git-wt --version
   ```
3. Update any version references in documentation
