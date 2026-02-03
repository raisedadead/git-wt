# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

This repo uses its own worktree structure. The project root contains:

```
git-wt/
├── .bare/       # Bare git repository
├── .git         # Pointer to .bare
├── CLAUDE.md    # This file (at project root)
└── main/        # Default branch worktree (all source code here)
```

**All source code and build commands must run from `main/`** (or another worktree directory).

## Build & Test

Run these from the `main/` worktree directory (prefer `just` over `make`):

```bash
just test           # Run all tests
just lint           # Run go vet + golangci-lint
just build          # Build to ./bin/
just build-all      # Cross-platform build check (linux/darwin/windows)
just dev            # Build and show version
```

Or directly with Go:

```bash
go build -v ./...              # Build all packages
go test -v ./...               # Run all tests
go test -v ./internal/git/...  # Run tests for a specific package
go vet ./...                   # Static analysis
```

## Development Workflow

**Switch between dev and released binary:**

```bash
make dev-mode       # Use local build (~/go/bin), remove homebrew version
make homebrew-mode  # Use released version (reinstall from homebrew)
```

**Typical workflow:**

```bash
# From within main/ worktree:
make install        # Install to ~/go/bin

# Use git-wt itself for development (from project root)
git-wt add feature/my-feature   # Creates ../feature-my-feature/ worktree
git-wt list

# After changes in any worktree, rebuild from that worktree
make dev            # Rebuild and show version

# Clean up
git-wt delete feature/my-feature
```

## Release Workflow

```bash
make release-check    # Validate goreleaser config
make release-snapshot # Build release locally (no publish)
make release-alpha    # Create alpha tag, CI releases (skips homebrew)
make release VERSION=0.1.0  # Create stable release (includes homebrew)
```

Alpha releases auto-increment: `v0.1.0-alpha.1` → `v0.1.0-alpha.2`

## Architecture

CLI wrapper around `git` and `gh` for managing git worktrees using the bare repo workflow.

**Bare repo layout:**

```
project/
├── .bare/       # bare git repository
├── .git         # file pointing to .bare (gitdir: ./.bare)
├── main/        # default branch worktree
└── feature-x/   # feature branch worktree (slashes become dashes)
```

**Code flow:**

```
cmd/git-wt/main.go → commands.Execute()
                   ↓
             rootCmd.Execute() (Cobra)
                   ↓
         Subcommand (clone, add, list, delete, prune)
                   ↓
         git/* for git operations, github/* for gh CLI
                   ↓
         hooks.RunWithTimeout() for post-operation hooks
```

Commands register in `init()` via `rootCmd.AddCommand()`.

**Package responsibilities:**

| Package          | Purpose                                              |
| ---------------- | ---------------------------------------------------- |
| `commands/`      | CLI layer (Cobra) - args, flags, prompts, output     |
| `git/`           | Git operations - exec with timeouts, no user output  |
| `github/`        | GitHub CLI integration - issue/PR fetching           |
| `hooks/`         | Post-operation shell commands with env vars          |
| `hooks/bundled/` | Embedded hook scripts (.sh) + embed.go registry      |
| `config/`        | TOML config loading, hierarchical merge              |
| `ui/`            | Terminal styling (lipgloss) and JSON output envelope |

**Key patterns:**

- Branch names with slashes flatten to dashes for directory names (`feature/auth` → `feature-auth`)
- `GetProjectRoot()` walks up to find `.bare/` directory
- JSON output via `--json` flag for scripting; check `IsJSONOutput()` before printing
- Interactive TUI forms via `charmbracelet/huh` when args not provided
- `add` and `new` are aliases (both implemented in `new.go`)
- Timeouts: 2min default, 10min for clone/fetch, 30sec for hooks (all configurable)
- Platform-specific code uses build tags (`hooks_unix.go`, `hooks_windows.go`)

**Error handling:**

- `git/*` functions return errors with context (e.g., `git worktree add: <stderr>`)
- Commands should use `ui.GetExitCode(err)` for proper exit codes
- JSON output wraps errors in envelope: `{"success": false, "error": "..."}`

## Configuration

**Hierarchical config (highest priority first):**

```
runtime flag > .git-wt.toml (repo) > ~/.config/git-wt/config.toml (global) > defaults
```

**Commands:**

```bash
git wt config init --global  # Create global config with defaults
git wt config init           # Create repo config
git wt config show           # Show effective config with sources
```

**Config options:**

| Option                | Default  | Description                |
| --------------------- | -------- | -------------------------- |
| `worktree_root`       | (none)   | Override clone location    |
| `default_remote`      | `origin` | Remote for fetch/prune     |
| `default_base_branch` | (none)   | Base for new worktrees     |
| `branch_template`     | (none)   | Branch naming template     |
| `git_timeout`         | `120`    | Default git timeout (sec)  |
| `git_long_timeout`    | `600`    | Clone/fetch timeout (sec)  |
| `hook_timeout`        | `30`     | Hook command timeout (sec) |

**Hooks:**

```toml
[hooks]
post_clone = ["zoxide add $GIT_WT_PATH"]
post_add = ["direnv allow"]
```

Hook env vars: `GIT_WT_PATH`, `GIT_WT_BRANCH`, `GIT_WT_PROJECT_ROOT`, `GIT_WT_DEFAULT_BRANCH`

## Global Flags

| Flag        | Description                    |
| ----------- | ------------------------------ |
| `--json`    | JSON output for scripting      |
| `--timeout` | Override git timeout (seconds) |

## Input Validation

The `git/validate.go` module rejects:

- Path traversal (`..`, leading `/`)
- Git-reserved names (`.git`, `.bare`, `@`)
- Invalid git ref characters

## Commit Conventions

Uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for custom branch templates
fix: handle empty repository URL in clone command
docs: update README with zoxide integration details
refactor: extract git operations into separate package
test: add tests for worktree creation
ci: add cross-platform build check
```

## Testing Pattern

Table-driven tests preferred:

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

## Known Limitations

- Assumes single remote (`origin` by default, configurable via `default_remote`)
- Only git clone flags support passthrough via `--` separator (not gh flags)

## Dependencies

- Go 1.23+
- git 2.20+ (for worktree features)
- gh CLI (optional, for `--issue` and `--pr` flags)

**Go packages:**

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/huh` - Interactive TUI forms
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/BurntSushi/toml` - Config parsing
