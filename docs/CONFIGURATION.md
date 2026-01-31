# Configuration

git-wt uses TOML configuration files with hierarchical overrides.

## Quick Start

```bash
# Create global config with documented defaults
git wt config init --global

# Create repo-specific config
git wt config init

# View effective configuration with sources
git wt config show
```

## Config Hierarchy

Configuration is merged from multiple sources (highest priority first):

```
runtime flag > .git-wt.toml (repo) > ~/.config/git-wt/config.toml (global) > defaults
```

| Source        | Location                       | Scope             |
| ------------- | ------------------------------ | ----------------- |
| Runtime flag  | `--timeout`, `--remote`        | Single command    |
| Repo config   | `.git-wt.toml` in project root | Single repository |
| Global config | `~/.config/git-wt/config.toml` | All repositories  |
| Defaults      | Built into git-wt              | Fallback          |

## Configuration Files

| Location                              | Description              |
| ------------------------------------- | ------------------------ |
| `$XDG_CONFIG_HOME/git-wt/config.toml` | Global config (XDG)      |
| `~/.config/git-wt/config.toml`        | Global config (fallback) |
| `.git-wt.toml`                        | Repo-specific config     |

## Options Reference

### Core Options

| Option                | Type   | Default  | Description                            |
| --------------------- | ------ | -------- | -------------------------------------- |
| `worktree_root`       | string | (none)   | Directory where projects are cloned    |
| `default_remote`      | string | `origin` | Remote for fetch/push/prune operations |
| `default_base_branch` | string | (none)   | Base branch for new worktrees          |
| `branch_template`     | string | (none)   | Template for generated branch names    |

### Timeout Options

| Option             | Type | Default | Description                             |
| ------------------ | ---- | ------- | --------------------------------------- |
| `git_timeout`      | int  | `120`   | Default git operation timeout (seconds) |
| `git_long_timeout` | int  | `600`   | Clone/fetch timeout (seconds)           |
| `hook_timeout`     | int  | `30`    | Hook command timeout (seconds)          |

### Hooks

| Option             | Type     | Default | Description                   |
| ------------------ | -------- | ------- | ----------------------------- |
| `hooks.post_clone` | []string | `[]`    | Commands to run after clone   |
| `hooks.post_add`   | []string | `[]`    | Commands to run after add/new |

## Full Example

```toml
# ~/.config/git-wt/config.toml

# Where to clone repositories (optional)
# If not set, clones to current directory
worktree_root = "~/DEV/worktrees"

# Remote configuration
default_remote = "origin"
default_base_branch = "main"

# Branch naming template (for --issue/--pr)
# Available: {{type}}, {{number}}, {{slug}}
branch_template = "{{type}}/{{number}}-{{slug}}"

# Timeouts (seconds)
git_timeout = 120        # Default operations
git_long_timeout = 600   # Clone/fetch
hook_timeout = 30        # Each hook command

[hooks]
# Run after 'git wt clone'
post_clone = [
  "zoxide add $GIT_WT_PATH",
]

# Run after 'git wt add/new'
post_add = [
  "zoxide add $GIT_WT_PATH",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.envrc $GIT_WT_PATH/ 2>/dev/null || true",
  "direnv allow",
]
```

## Repo-Specific Config

Create `.git-wt.toml` in your project root to override global settings:

```toml
# .git-wt.toml (in project root)

# This repo uses 'upstream' instead of 'origin'
default_remote = "upstream"

# Use 'develop' as base for new branches
default_base_branch = "develop"

# Longer timeout for this large repo
git_long_timeout = 900

[hooks]
# Project-specific hooks
post_add = [
  "npm install",
  "cp .env.example .env",
]
```

## Environment Variables

Hooks have access to these environment variables:

| Variable                | Description                      | Example                              |
| ----------------------- | -------------------------------- | ------------------------------------ |
| `GIT_WT_PATH`           | Path to the new worktree         | `/home/user/DEV/worktrees/repo/main` |
| `GIT_WT_BRANCH`         | Branch name                      | `feature/auth`                       |
| `GIT_WT_PROJECT_ROOT`   | Project root (contains `.bare/`) | `/home/user/DEV/worktrees/repo`      |
| `GIT_WT_DEFAULT_BRANCH` | Default branch name              | `main`                               |

## Template Syntax

Hook commands support Go template variables:

| Template             | Equivalent Variable      |
| -------------------- | ------------------------ |
| `{{.Path}}`          | `$GIT_WT_PATH`           |
| `{{.Branch}}`        | `$GIT_WT_BRANCH`         |
| `{{.ProjectRoot}}`   | `$GIT_WT_PROJECT_ROOT`   |
| `{{.DefaultBranch}}` | `$GIT_WT_DEFAULT_BRANCH` |

Example:

```toml
[hooks]
post_add = [
  "echo 'Created {{.Branch}} at {{.Path}}'",
]
```

**Note:** Template values are automatically shell-quoted for security.

## Hook Behavior

- Hooks run in the order listed
- Each hook runs with the worktree path as working directory
- A failing hook logs a warning but doesn't block subsequent hooks
- Each hook command has a configurable timeout (default 30 seconds)
- Hooks that exceed the timeout are terminated

## Viewing Configuration

```bash
# Show effective configuration with sources
git wt config show
```

Output shows which file each setting comes from:

```
Effective Configuration:

default_remote = "upstream"     # .git-wt.toml
default_base_branch = "develop" # .git-wt.toml
git_timeout = 120               # default
git_long_timeout = 600          # ~/.config/git-wt/config.toml
hook_timeout = 30               # default

[hooks]
post_clone = ["zoxide add $GIT_WT_PATH"]  # ~/.config/git-wt/config.toml
post_add = ["npm install"]                 # .git-wt.toml
```

## Runtime Overrides

Override any timeout via command flags:

```bash
# Override git timeout for slow networks
git wt clone owner/repo --timeout 900

# Override hook timeout
git wt add feature/auth --hook-timeout 60
```

See [Hooks Examples](HOOKS.md) for more hook recipes.
