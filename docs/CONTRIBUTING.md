# Contributing to git-wt

Thank you for your interest in contributing to git-wt! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git 2.20 or later

### Getting Started

1. Fork the repository on GitHub

2. Clone your fork:

   ```bash
   git clone git@github.com:YOUR_USERNAME/git-wt.git
   cd git-wt
   ```

3. Build and create symlink (one-time setup):

   ```bash
   go build -o git-wt ./cmd/git-wt
   sudo ln -sf $(pwd)/git-wt /usr/local/bin/git-wt
   ```

   This creates a symlink so `git wt` works as a git subcommand. Subsequent builds automatically update the binary.

4. Verify installation:

   ```bash
   git wt --version
   ```

5. Run tests:

   ```bash
   go test -v ./...
   ```

### Development Workflow

After the one-time setup, your workflow is simply:

```bash
# Make changes, then rebuild
go build -o git-wt ./cmd/git-wt

# Test your changes
git wt clone owner/repo
git wt list
```

The symlink means you don't need to reinstall after each build.

### Project Structure

```
git-wt/
├── cmd/
│   └── git-wt/
│       └── main.go             # Entry point
├── internal/
│   ├── commands/               # CLI commands (clone, new, list, delete, prune)
│   ├── config/                 # Configuration loading
│   ├── files/                  # File copy utilities
│   ├── git/                    # Git operations
│   ├── github/                 # GitHub CLI integration
│   ├── integrations/           # direnv, zoxide, Claude Code integrations
│   └── ui/                     # Terminal UI styles
├── go.mod
├── go.sum
└── .goreleaser.yaml
```

## Development Workflow

### Making Changes

1. Create a feature branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes

3. Run tests and linting:

   ```bash
   go test -v ./...
   go fmt ./...
   go vet ./...
   ```

4. Commit your changes following the commit conventions below

5. Push to your fork and open a pull request

### Commit Conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. Each commit message should have the format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes |
| `style` | Code style changes (formatting, semicolons, etc.) |
| `refactor` | Code refactoring without feature changes |
| `perf` | Performance improvements |
| `test` | Adding or updating tests |
| `build` | Build system or external dependency changes |
| `ci` | CI/CD configuration changes |
| `chore` | Other changes that don't modify src or test files |

**Examples:**

```bash
feat: add support for custom branch templates
fix: handle empty repository URL in clone command
docs: update README with zoxide integration details
refactor: extract git operations into separate package
test: add tests for worktree creation
```

**Breaking Changes:**

For breaking changes, add `!` after the type or include `BREAKING CHANGE:` in the footer:

```bash
feat!: change default worktree root location

BREAKING CHANGE: The default worktree root has changed from ~/worktrees to ~/DEV/worktrees
```

## Pull Request Guidelines

1. **Keep PRs focused** - Each PR should address a single concern

2. **Update documentation** - If your change affects user-facing behavior, update the README

3. **Add tests** - New features should include tests

4. **Follow existing patterns** - Match the code style and architecture of the existing codebase

5. **Write clear descriptions** - Explain what your PR does and why

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

- [ ] Tests pass locally
- [ ] Code follows project style
- [ ] Documentation updated (if applicable)
```

## Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/git/...
```

### Writing Tests

- Place test files alongside the code they test (e.g., `worktree.go` and `worktree_test.go`)
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
        {"multiple spaces", "too   many   spaces", "too-many-spaces"},
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

## Release Process

Releases are automated via GitHub Actions and GoReleaser. When a new tag is pushed:

1. GoReleaser builds binaries for all platforms
2. Checksums are generated
3. GitHub Release is created with artifacts
4. Homebrew formula is updated

To create a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Code of Conduct

Be respectful and constructive. We're all here to build great software together.

## Questions?

If you have questions, feel free to open an issue for discussion.
