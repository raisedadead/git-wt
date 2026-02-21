# wt

[![Go Version](https://img.shields.io/github/go-mod/go-version/raisedadead/wt)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/raisedadead/wt)](https://github.com/raisedadead/wt/releases)

A CLI and TUI for managing git worktrees with the bare repository workflow.

wt wraps `git` and `gh` to streamline worktree operations—cloning as bare repos, creating worktrees from branches or GitHub issues, and running post-create hooks. Running bare `wt` launches a lazygit-style terminal interface for visual worktree management.

## Why Worktrees?

| Problem                                  | Solution                              |
| ---------------------------------------- | ------------------------------------- |
| Context switching requires stashing      | Each worktree is isolated             |
| Can't run tests on main while developing | Parallel worktrees                    |
| Branch switching breaks IDE state        | Each worktree is a separate directory |

## Install

```bash
# Homebrew
brew install raisedadead/tap/wt

# Go
go install github.com/raisedadead/wt/cmd/wt@latest
```

## Requirements

- Git 2.20+
- [GitHub CLI](https://cli.github.com/) (`gh`) - optional, for GitHub issue/PR workflows
- [zoxide](https://github.com/ajeetdsouza/zoxide) (optional) - for quick worktree navigation

## Quick Start

```bash
# Clone as bare repo with worktree structure
wt clone owner/repo

# Launch TUI (lazygit-style worktree manager)
wt

# Create worktrees
wt add my-experiment              # Plain branch
wt add --feature auth             # Feature workflow (feat/auth)
wt add --bugfix --issue 42        # Bugfix linked to GitHub issue
wt add --pr-review 123            # Review a PR (uses PR's actual branch)

# List worktrees
wt list

# Switch to a worktree (with shell completions sourced)
wt switch feature/auth

# Clean up
wt delete feature/auth
wt prune
```

All commands also work as `git wt <command>` when shell completions are sourced.

### Quick Navigation with zoxide

Enable [zoxide](https://github.com/ajeetdsouza/zoxide) integration for instant worktree switching:

```bash
wt hooks enable zoxide            # One-time setup

wt add --feature auth             # Creates worktree
z auth                            # Jump to it instantly
z main                            # Jump back
```

## TUI

Running bare `wt` opens a full-screen terminal interface:

- Worktree list with branch status (merged, remote gone)
- Detail panel with info, diff, and log views
- Single-key actions: `Enter` to switch, `n` for new, `d` for delete
- Overlay dialogs for confirmations and text input

All operations available in the TUI are also available as CLI subcommands for scripting.

## Commands

| Command           | Description                                           | Aliases |
| ----------------- | ----------------------------------------------------- | ------- |
| `clone <repo>`    | Clone as bare repo with initial worktree              |         |
| `add [branch]`    | Create worktree with workflow support                 | `new`   |
| `list`            | List worktrees                                        | `ls`    |
| `switch [branch]` | Switch to a worktree (auto-cd with shell completions) |         |
| `delete [branch]` | Remove worktree and branch (interactive if no branch) | `rm`    |
| `prune`           | Remove stale worktrees                                |         |
| `repair`          | Repair worktree paths after moving a repository       |         |
| `config init`     | Create config file with documented defaults           |         |
| `config show`     | Show effective configuration with sources             |         |
| `hooks`           | Manage hooks (enable/disable/list/show)               |         |
| `completion`      | Print shell completion setup instructions             |         |

### Workflow Flags (for `add`)

| Flag              | Description                                |
| ----------------- | ------------------------------------------ |
| `--feature`, `-f` | Feature workflow (branch: `feat/{slug}`)   |
| `--bugfix`, `-b`  | Bugfix workflow (branch: `fix/{slug}`)     |
| `--pr-review`     | PR review workflow (uses PR's head branch) |
| `--issue <n>`     | Pass issue number to hooks                 |
| `--pr <n>`        | Pass PR number to hooks                    |
| `--workflow`      | Use custom workflow from config            |
| `--base`          | Base branch to create worktree from        |
| `--track`         | Track existing remote branch               |
| `--new`           | Force create new local branch              |
| `--fetch`         | Fetch all remotes before checking          |
| `--no-hooks`      | Skip all hooks                             |

### Global Flags

| Flag     | Description                                      |
| -------- | ------------------------------------------------ |
| `--json` | Output in JSON format (for scripting/automation) |

### Common Flags

| Flag             | Commands                 | Description                    |
| ---------------- | ------------------------ | ------------------------------ |
| `--yes`, `-y`    | `delete`                 | Skip confirmation prompt       |
| `--force`, `-f`  | `delete`, `clone`        | Force operation                |
| `--dry-run`      | `delete`, `prune`        | Show what would happen         |
| `--timeout`      | `clone`, `add`, `delete` | Override git operation timeout |
| `--hook-timeout` | `clone`, `add`           | Override hook timeout          |
| `--remote`       | `add`, `prune`           | Override default remote        |
| `--no-hooks`     | `clone`, `add`           | Skip all hooks                 |

### Passthrough Flags

Pass git flags after `--`:

```bash
wt clone owner/repo -- --depth=1 --single-branch
```

## Directory Structure

```
project/
├── .bare/          # Bare git repository
├── .git            # Pointer to .bare
├── main/           # Stable worktree
├── feature-auth/   # Feature worktree
└── issue-42/       # Issue worktree
```

## Configuration

wt supports hierarchical configuration:

```
runtime flag > .wt.toml (repo) > ~/.config/wt/config.toml (global) > defaults
```

Create a config file with documented options:

```bash
wt config init --global  # ~/.config/wt/config.toml
wt config init           # .wt.toml in project root
```

View effective configuration:

```bash
wt config show
```

Example config:

```toml
default_remote = "upstream"
default_base_branch = "develop"
branch_template = "{{type}}-{{number}}-{{slug}}"
hook_timeout = 30

[hooks]
post_clone = ["zoxide add $WT_PATH"]
post_add = ["direnv allow"]
```

See [Configuration](docs/CONFIGURATION.md) for all options.

## Development

```bash
just setup          # First-time setup: install dev tools + git hooks
just build          # Build to ./bin/
just dev            # Build and show version + alias hint
just test           # Run all tests
just test-unit      # Unit tests only
just test-integration  # Integration tests (builds first)
just lint           # Run go vet + golangci-lint
just build-all      # Cross-platform build check
just fmt            # Format code
```

Keep the brew install as your daily driver, alias the dev build for testing:

```bash
alias wt-dev='./bin/wt'    # after just build
wt list                    # released version
wt-dev list                # dev build
```

See [Contributing](docs/CONTRIBUTING.md) for the full development guide.

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Design goals and internals
- [Configuration](docs/CONFIGURATION.md) - Config options and hooks
- [Hooks Examples](docs/HOOKS.md) - Common hook recipes
- [Contributing](docs/CONTRIBUTING.md) - Development setup
- [Releasing](docs/RELEASING.md) - Release process

## Links

- [Git Worktrees Documentation](https://git-scm.com/docs/git-worktree) - Official git worktree reference
- [Bare Repo + Worktree Workflow](https://nicknisi.com/posts/git-worktrees/) - The workflow wt implements
- [GitHub CLI](https://cli.github.com/) - Required for issue/PR integration

## License

MIT - see [LICENSE](LICENSE)
