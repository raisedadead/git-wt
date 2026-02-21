# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

This repo uses its own worktree structure. The project root contains:

```
wt/
├── .bare/       # Bare git repository
├── .git         # Pointer to .bare
├── CLAUDE.md    # Symlink to main/CLAUDE.md
└── main/        # Default branch worktree (all source code here)
```

**All source code and build commands must run from `main/`** (or another worktree directory).

## Build & Test

Run these from the `main/` worktree directory:

```bash
just setup          # First-time setup: install dev tools + git hooks
just test           # Run all tests
just test-unit      # Run unit tests only
just test-integration        # Run integration tests (builds first)
just test-integration-parallel  # Run integration tests in parallel
just test-match <pattern>    # Run specific test by pattern
just test-pkg <pkg>          # Run tests for specific package (e.g., git, hooks)
just lint           # Run go vet + golangci-lint
just build          # Build to ./bin/
just build-all      # Cross-platform build check (linux/darwin/windows)
just dev            # Build and show version
just clean          # Remove build artifacts and test fixtures
just fmt            # Format code (go fmt + goimports)
just setup-test-repo # Set up test repo for integration tests
```

Or directly with Go:

```bash
go build -v ./...              # Build all packages
go test -v ./...               # Run all tests
go test -v ./internal/git/...  # Run tests for a specific package
go vet ./...                   # Static analysis
```

## Development Workflow

**Two-binary setup:**

- Install released version via `brew install raisedadead/tap/wt` (includes man pages + shell completions)
- Build dev version with `just build` → `./bin/wt`
- Alias for convenience: `alias wt-dev='./bin/wt'`

**Typical workflow:**

```bash
# Use released wt for daily worktree management
wt add feature/my-feature   # Creates ../feature-my-feature/ worktree
wt list

# After code changes, rebuild and test
just dev            # Build and show version
./bin/wt list       # Test dev build
wt-dev list         # Same thing via alias

# Clean up
wt delete feature/my-feature
```

## Release Workflow

```bash
just release-check    # Validate goreleaser config
just release-snapshot # Build release locally (no publish)
just release-alpha    # Create alpha tag, CI releases (skips homebrew)
```

Alpha releases auto-increment: `v0.1.0-alpha.1` → `v0.1.0-alpha.2`

Goreleaser automatically generates completions via `scripts/completions.sh` and includes them in:

- Release archives (tar.gz/zip)
- Homebrew formula (auto-installed for users)

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
cmd/wt/main.go → commands.Execute()
                   ↓
             rootCmd.Execute() (Cobra)
                   ↓
         ┌── wt (no args) → tui.Run() → lazygit-style TUI
         │
         └── wt <command> [flags] → Subcommand (flag-only, no interactive prompts)
                   ↓
         internal/git/* for git operations
                   ↓
         hooks.RunWithTimeout() for post-operation hooks
```

Commands register in `init()` via `rootCmd.AddCommand()`.
Subcommands are flag-only -- no interactive prompts. All interactivity is in the TUI.

**Package responsibilities:**

| Package                   | Purpose                                                           |
| ------------------------- | ----------------------------------------------------------------- |
| `internal/tui/`           | lazygit-style TUI (bubbletea) - panels, overlays, keys            |
| `internal/tui/panels/`    | TUI panels: worktree list, detail (info/diff/log), header, footer |
| `internal/tui/overlays/`  | TUI overlays: help, confirm, input, menu                          |
| `internal/commands/`      | CLI layer (Cobra) - flag-only, no interactive prompts             |
| `internal/git/`           | Git operations - exec with timeouts, no user output               |
| `internal/hooks/`         | Hook execution with workflow support and helper protocol          |
| `internal/hooks/bundled/` | Embedded hook scripts (.sh) + helpers.sh library                  |
| `internal/config/`        | TOML config loading, hierarchical merge, workflow configs         |
| `internal/ui/`            | Terminal styling (lipgloss) and JSON output envelope              |

**Test structure:**

- `test/integration/` - Integration tests (use `just test-integration`)
- `internal/*/_test.go` - Unit tests co-located with source

**Key patterns:**

- Branch names with slashes flatten to dashes for directory names (`feature/auth` → `feature-auth`)
- `GetProjectRoot()` walks up to find `.bare/` directory
- JSON output via `--json` flag for scripting; check `IsJSONOutput()` before printing
- Running bare `wt` opens a lazygit-style TUI (bubbletea); subcommands are flag-only
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
runtime flag > .wt.toml (repo) > ~/.config/wt/config.toml (global) > defaults
```

**Commands:**

```bash
wt config init --global  # Create global config with defaults
wt config init           # Create repo config
wt config show           # Show effective config with sources
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
post_clone = ["zoxide add $WT_PATH"]
post_add = ["direnv allow"]
```

Hook env vars: `WT_PATH`, `WT_BRANCH`, `WT_PROJECT_ROOT`, `WT_DEFAULT_BRANCH`

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

Table-driven tests preferred. Unit tests co-located with source (`_test.go`), integration tests in `test/integration/`.

## Known Limitations

- Assumes single remote (`origin` by default, configurable via `default_remote`)
- Only git clone flags support passthrough via `--` separator (not gh flags)

## Dependencies

- Go 1.25+
- git 2.20+ (for worktree features)
- gh CLI (optional, for GitHub workflow hooks like `github-issue` and `github-pr`)

**Go packages:**

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework (Elm architecture)
- `github.com/charmbracelet/bubbles` - TUI components (list, viewport, help, key)
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/BurntSushi/toml` - Config parsing
