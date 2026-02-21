# Contributing to wt

Thank you for your interest in contributing to wt! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git 2.20 or later
- [goreleaser](https://goreleaser.com/) (optional, for release testing)

### Getting Started

1. Fork the repository on GitHub

2. Clone your fork:

   ```bash
   git clone git@github.com:YOUR_USERNAME/wt.git
   cd wt
   ```

3. Build and install:

   ```bash
   just install    # Install to ~/go/bin
   ```

4. Verify installation:

   ```bash
   git wt --version
   ```

5. Run tests:

   ```bash
   just test
   ```

### Development Workflow

Use the justfile for common tasks:

```bash
just                # Show all available targets

# Build & test
just build          # Build to ./bin/
just test           # Run all tests
just lint           # Run go vet + golangci-lint
just build-all      # Cross-platform build check

# Development
just dev            # Build and show version + alias hint

# After making changes
just test && just lint && just build-all
```

**Two-binary setup:** Install the released version via `brew install raisedadead/tap/wt`, then alias the dev build:

```bash
alias wt-dev='/path/to/wt/main/bin/wt'

wt list             # released version
wt-dev list         # dev build
```

### Project Structure

```
wt/
├── cmd/wt/
│   └── main.go                 # Entry point
├── internal/
│   ├── tui/                    # Lazygit-style TUI (bubbletea)
│   │   ├── panels/            # Worktree list, detail, header, footer
│   │   └── overlays/          # Help, confirm, input, menu
│   ├── commands/               # CLI commands (Cobra, flag-only)
│   ├── config/                 # TOML config, hierarchical merge
│   ├── git/                    # Git operations with timeouts
│   ├── hooks/                  # Hook execution with workflow support
│   │   └── bundled/           # Embedded hook scripts + helpers.sh
│   └── ui/                     # Terminal styling, tables, JSON envelope
├── test/integration/           # Integration tests
├── justfile                    # Build, test, release targets
├── .goreleaser.yaml            # Release configuration
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
   just test
   just lint
   just build-all    # Catches platform-specific issues
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

4. **Cross-platform** - Run `just build-all` to verify

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

- [ ] Tests pass (`just test`)
- [ ] Lint passes (`just lint`)
- [ ] Cross-platform build (`just build-all`)
- [ ] Documentation updated (if applicable)
```

## Testing

### Running Tests

```bash
just test                        # Run all tests
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

Always verify with `just build-all` before submitting.

## Release Process

Releases are automated via GitHub Actions. See [RELEASING.md](RELEASING.md) for details.

**Quick reference:**

```bash
just release-alpha              # Create alpha release
just release VERSION=0.1.0      # Create stable release
```

## Code of Conduct

Be respectful and constructive. We're all here to build great software together.

## Questions?

If you have questions, feel free to open an issue for discussion.
