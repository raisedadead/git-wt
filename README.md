# git-wt

[![Go Version](https://img.shields.io/github/go-mod/go-version/raisedadead/git-wt)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/raisedadead/git-wt)](https://github.com/raisedadead/git-wt/releases)

A CLI for managing git worktrees with the bare repository workflow.

git-wt wraps `git` and `gh` to streamline worktree operations—cloning as bare repos, creating worktrees from branches or GitHub issues, and running post-create hooks.

## Why Worktrees?

| Problem                                  | Solution                              |
| ---------------------------------------- | ------------------------------------- |
| Context switching requires stashing      | Each worktree is isolated             |
| Can't run tests on main while developing | Parallel worktrees                    |
| Branch switching breaks IDE state        | Each worktree is a separate directory |

## Install

```bash
# Homebrew
brew install raisedadead/tap/git-wt

# Go
go install github.com/raisedadead/git-wt/cmd/git-wt@latest
```

## Requirements

- Git 2.20+
- [GitHub CLI](https://cli.github.com/) (`gh`) - required for `--issue` and `--pr` flags

## Quick Start

```bash
# Clone as bare repo with worktree structure
git wt clone owner/repo

# Create a feature worktree
git wt add feature/auth

# Create from GitHub issue
git wt add --issue 42

# List worktrees
git wt list

# Clean up
git wt delete feature/auth
git wt prune
```

## Commands

| Command           | Description                                                |
| ----------------- | ---------------------------------------------------------- |
| `clone <repo>`    | Clone as bare repo with initial worktree                   |
| `add [branch]`    | Create worktree (supports `--issue`, `--pr`, alias: `new`) |
| `list`            | List worktrees                                             |
| `delete [branch]` | Remove worktree and branch (interactive if no branch)      |
| `prune`           | Remove stale worktrees                                     |
| `config init`     | Create config file with documented defaults                |
| `config show`     | Show effective configuration with sources                  |
| `completion`      | Print shell completion setup instructions                  |

### Global Flags

| Flag     | Description                                      |
| -------- | ------------------------------------------------ |
| `--json` | Output in JSON format (for scripting/automation) |

### Common Flags

| Flag            | Commands          | Description                    |
| --------------- | ----------------- | ------------------------------ |
| `--yes`, `-y`   | `delete`, `prune` | Skip confirmation prompt       |
| `--force`, `-f` | `delete`, `clone` | Force operation                |
| `--dry-run`     | `delete`, `prune` | Show what would happen         |
| `--timeout`     | all               | Override git operation timeout |
| `--remote`      | `add`, `prune`    | Override default remote        |

### Passthrough Flags

Pass git flags after `--`:

```bash
git wt clone owner/repo -- --depth=1 --single-branch
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

git-wt supports hierarchical configuration:

```
runtime flag > .git-wt.toml (repo) > ~/.config/git-wt/config.toml (global) > defaults
```

Create a config file with documented options:

```bash
git wt config init --global  # ~/.config/git-wt/config.toml
git wt config init           # .git-wt.toml in project root
```

View effective configuration:

```bash
git wt config show
```

Example config:

```toml
default_remote = "upstream"
default_base_branch = "develop"
branch_template = "{{type}}-{{number}}-{{slug}}"
hook_timeout = 30

[hooks]
post_clone = ["zoxide add $GIT_WT_PATH"]
post_add = ["direnv allow"]
```

See [Configuration](docs/CONFIGURATION.md) for all options.

## Development

```bash
go build -v ./...      # Build
go test -v ./...       # Run tests
go vet ./...           # Static analysis
golangci-lint run      # Lint
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Design goals and internals
- [Configuration](docs/CONFIGURATION.md) - Config options and hooks
- [Hooks Examples](docs/HOOKS.md) - Common hook recipes
- [Contributing](docs/CONTRIBUTING.md) - Development setup
- [Releasing](docs/RELEASING.md) - Release process

## Links

- [Git Worktrees Documentation](https://git-scm.com/docs/git-worktree) - Official git worktree reference
- [Bare Repo + Worktree Workflow](https://nicknisi.com/posts/git-worktrees/) - The workflow git-wt implements
- [GitHub CLI](https://cli.github.com/) - Required for issue/PR integration

## License

MIT - see [LICENSE](LICENSE)
