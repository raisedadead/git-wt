# Contributing to git-wt

Thank you for your interest in contributing to git-wt! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git 2.20 or later
- [goreleaser](https://goreleaser.com/) (optional, for release testing)

### Getting Started

1. Fork the repository on GitHub

2. Clone your fork:

   ```bash
   git clone git@github.com:YOUR_USERNAME/git-wt.git
   cd git-wt
   ```

3. Build and install:

   ```bash
   make install    # Install to ~/go/bin
   ```

4. Verify installation:

   ```bash
   git wt --version
   ```

5. Run tests:

   ```bash
   make test
   ```

### Development Workflow

Use the Makefile for common tasks:

```bash
make help           # Show all available targets

# Build & test
make build          # Build to ./bin/
make test           # Run all tests
make lint           # Run go vet + golangci-lint
make build-all      # Cross-platform build check

# Development
make dev            # Build and show version
make dev-mode       # Switch to local build (remove homebrew)
make install        # Install to ~/go/bin

# After making changes
make test && make lint && make build-all
```

### Project Structure

```
git-wt/
├── cmd/
│   └── git-wt/
│       └── main.go             # Entry point
├── internal/
│   ├── commands/               # CLI commands
│   │   ├── root.go            # Root command, global flags
│   │   ├── clone.go           # Clone bare repo
│   │   ├── new.go             # Create worktree (add/new)
│   │   ├── list.go            # List worktrees
│   │   ├── delete.go          # Remove worktree
│   │   ├── prune.go           # Clean stale worktrees
│   │   ├── config.go          # Config init/show
│   │   └── completion.go      # Shell completions
│   ├── config/                 # Configuration loading
│   │   └── config.go          # TOML config, hierarchical merge
│   ├── git/                    # Git operations
│   │   ├── exec.go            # Command execution with timeouts
│   │   ├── bare.go            # Bare repo operations
│   │   ├── worktree.go        # Worktree CRUD
│   │   └── validate.go        # Input validation
│   ├── github/                 # GitHub CLI integration
│   │   └── gh.go              # Issue/PR fetching
│   ├── hooks/                  # Hook execution
│   │   ├── hooks.go           # Run post-operation hooks
│   │   ├── hooks_unix.go      # Unix-specific (process groups)
│   │   └── hooks_windows.go   # Windows stub
│   └── ui/                     # Terminal UI
│       ├── styles.go          # Lipgloss styles
│       └── output.go          # JSON output envelope
├── Makefile                    # Build, test, release targets
├── .goreleaser.yaml           # Release configuration
└── go.mod
```

## Making Changes

1. Create a feature branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Run tests, lint, and cross-platform check:

   ```bash
   make test
   make lint
   make build-all    # Catches platform-specific issues
   ```

4. Commit your changes following the conventions below

5. Push to your fork and open a pull request

### Commit Conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**

| Type       | Description                                       |
| ---------- | ------------------------------------------------- |
| `feat`     | A new feature                                     |
| `fix`      | A bug fix                                         |
| `docs`     | Documentation changes                             |
| `style`    | Code style changes (formatting, etc.)             |
| `refactor` | Code refactoring without feature changes          |
| `perf`     | Performance improvements                          |
| `test`     | Adding or updating tests                          |
| `build`    | Build system or dependency changes                |
| `ci`       | CI/CD configuration changes                       |
| `chore`    | Other changes that don't modify src or test files |

**Examples:**

```bash
feat: add support for custom branch templates
fix: handle empty repository URL in clone command
docs: update README with zoxide integration details
refactor: extract git operations into separate package
test: add tests for worktree creation
ci: add cross-platform build check
```

## Pull Request Guidelines

1. **Keep PRs focused** - Each PR should address a single concern

2. **Update documentation** - If your change affects user-facing behavior

3. **Add tests** - New features should include tests

4. **Cross-platform** - Run `make build-all` to verify

5. **Follow existing patterns** - Match the code style of the codebase

### PR Template

```markdown
## Summary

Brief description of changes.

## Changes

- Change 1
- Change 2

## Testing

How was this tested?

## Checklist

- [ ] Tests pass (`make test`)
- [ ] Lint passes (`make lint`)
- [ ] Cross-platform build (`make build-all`)
- [ ] Documentation updated (if applicable)
```

## Testing

### Running Tests

```bash
make test                        # Run all tests
go test -v ./internal/git/...    # Run specific package
go test -cover ./...             # With coverage
```

### Writing Tests

- Place test files alongside the code they test
- Use table-driven tests where appropriate
- Test both success and error cases

Example:

```go
func TestSlugify(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"lowercase", "Hello World", "hello-world"},
        {"special chars", "Fix: bug #42", "fix-bug-42"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Slugify(tt.input)
            if result != tt.expected {
                t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

### Platform-Specific Code

Use build tags for platform-specific code:

```go
//go:build unix

package hooks

// Unix-specific implementation
```

```go
//go:build windows

package hooks

// Windows-specific implementation
```

Always verify with `make build-all` before submitting.

## Release Process

Releases are automated via GitHub Actions. See [RELEASING.md](RELEASING.md) for details.

**Quick reference:**

```bash
make release-alpha              # Create alpha release
make release VERSION=0.1.0      # Create stable release
```

## Code of Conduct

Be respectful and constructive. We're all here to build great software together.

## Questions?

If you have questions, feel free to open an issue for discussion.
