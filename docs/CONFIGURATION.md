# Configuration

wt uses TOML configuration files with hierarchical overrides.

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
runtime flag > .wt.toml (repo) > ~/.config/wt/config.toml (global) > defaults
```

| Source        | Location                       | Scope             |
| ------------- | ------------------------------ | ----------------- |
| Runtime flag  | `--timeout`, `--remote`        | Single command    |
| Repo config   | `.wt.toml` in project root | Single repository |
| Global config | `~/.config/wt/config.toml` | All repositories  |
| Defaults      | Built into wt              | Fallback          |

## Configuration Files

| Location                              | Description              |
| ------------------------------------- | ------------------------ |
| `$XDG_CONFIG_HOME/wt/config.toml` | Global config (XDG)      |
| `~/.config/wt/config.toml`        | Global config (fallback) |
| `.wt.toml`                        | Repo-specific config     |

## Options Reference

### Core Options

| Option                | Type   | Default  | Description                                  |
| --------------------- | ------ | -------- | -------------------------------------------- |
| `worktree_root`       | string | (none)   | Directory where projects are cloned          |
| `default_remote`      | string | `origin` | Remote for fetch/push/prune operations       |
| `default_base_branch` | string | (none)   | Base branch for new worktrees                |
| `branch_template`     | string | (none)   | Template for generated branch names          |
| `auto_track`          | bool   | `false`  | Auto-track remote branches without prompting |

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
# ~/.config/wt/config.toml

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
# Named hooks (resolved from hooks directory) or inline commands
post_clone = [
  "gh-default",                   # Bundled hook: auto-configure gh CLI
  "zoxide add $WT_PATH",      # Inline command
]

# Run after 'git wt add/new'
post_add = [
  "direnv",                       # Bundled hook: auto-allow .envrc
  "zoxide add $WT_PATH",
  "cp $WT_PROJECT_ROOT/$WT_DEFAULT_BRANCH/.envrc $WT_PATH/ 2>/dev/null || true",
]
```

## Workflows

wt provides workflow presets that combine branch naming conventions with hook execution.

### Built-in Workflows

| Workflow    | Flag              | Branch Template | Description                         |
| ----------- | ----------------- | --------------- | ----------------------------------- |
| `feature`   | `--feature`, `-f` | `feat/{slug}`   | New feature work                    |
| `bugfix`    | `--bugfix`, `-b`  | `fix/{slug}`    | Bug fixes                           |
| `pr-review` | `--pr-review`     | `{branch}`      | PR review (uses PR's actual branch) |
| `branch`    | (default)         | `{name}`        | Plain branch                        |

### Workflow Hooks

Each workflow can define hooks that run at different stages:

| Hook Stage   | When it runs              | Purpose                        |
| ------------ | ------------------------- | ------------------------------ |
| `pre_create` | Before worktree creation  | Fetch metadata, suggest branch |
| `post_add`   | After worktree is created | Setup environment              |

Pre-create hooks can communicate back to wt using the [hook helper protocol](HOOKS.md#hook-helper-protocol).

### Custom Workflow Configuration

Override or extend workflows in your config:

```toml
[workflows.feature]
description = "New feature development"
branch_template = "feat/{slug}"

[workflows.feature.hooks]
pre_create = ["github-issue"]  # Fetch issue metadata for --issue flag
post_add = ["direnv", "zoxide"]

[workflows.bugfix]
description = "Bug fix"
branch_template = "fix/{slug}"

[workflows.bugfix.hooks]
pre_create = ["github-issue"]
post_add = ["direnv", "zoxide"]

[workflows.pr-review]
description = "Review a pull request"
branch_template = "{branch}"

[workflows.pr-review.hooks]
pre_create = ["github-pr"]  # Fetches PR's actual branch
post_add = ["direnv", "zoxide"]

# Custom workflow
[workflows.hotfix]
description = "Emergency hotfix"
branch_template = "hotfix/{slug}"

[workflows.hotfix.hooks]
pre_create = ["github-issue"]
post_add = ["direnv", "zoxide"]
```

### Using Workflows

```bash
# Use built-in workflows
git wt add --feature auth
git wt add --bugfix --issue 42
git wt add --pr-review 123

# Use custom workflow
git wt add --workflow hotfix security-patch
```

## Repo-Specific Config

Create `.wt.toml` in your project root to override global settings:

```toml
# .wt.toml (in project root)

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

| Variable                 | Description                      | Example                              |
| ------------------------ | -------------------------------- | ------------------------------------ |
| `WT_PATH`            | Path to the new worktree         | `/home/user/DEV/worktrees/repo/main` |
| `WT_BRANCH`          | Branch name                      | `feature/auth`                       |
| `WT_PROJECT_ROOT`    | Project root (contains `.bare/`) | `/home/user/DEV/worktrees/repo`      |
| `WT_DEFAULT_BRANCH`  | Default branch name              | `main`                               |
| `WT_WORKFLOW`        | Current workflow name            | `feature`, `bugfix`, `pr-review`     |
| `WT_WORKFLOW_PREFIX` | Branch prefix from workflow      | `feat`, `fix`                        |
| `WT_ISSUE`           | GitHub issue number (if passed)  | `42`                                 |
| `WT_PR`              | GitHub PR number (if passed)     | `123`                                |
| `WT_OUTPUT`          | Output file for hook protocol    | (internal path)                      |
| `WT_LIB`             | Directory containing helpers.sh  | (internal path)                      |

## Template Syntax

Hook commands support Go template variables:

| Template             | Equivalent Variable      |
| -------------------- | ------------------------ |
| `{{.Path}}`          | `$WT_PATH`           |
| `{{.Branch}}`        | `$WT_BRANCH`         |
| `{{.ProjectRoot}}`   | `$WT_PROJECT_ROOT`   |
| `{{.DefaultBranch}}` | `$WT_DEFAULT_BRANCH` |

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

default_remote = "upstream"     # .wt.toml
default_base_branch = "develop" # .wt.toml
git_timeout = 120               # default
git_long_timeout = 600          # ~/.config/wt/config.toml
hook_timeout = 30               # default

[hooks]
post_clone = ["zoxide add $WT_PATH"]  # ~/.config/wt/config.toml
post_add = ["npm install"]                 # .wt.toml
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

## Hooks Ecosystem

wt supports both named hooks (scripts) and inline commands.

### Named Hooks vs Inline Commands

```toml
[hooks]
# Named hook - resolved from custom/ or community/ directory
post_clone = ["gh-default"]

# Inline command - executed as shell command
post_add = ["direnv allow", "echo done"]

# Mix both
post_add = ["gh-default", "zoxide add $WT_PATH"]
```

### Hook Directories

| Directory                           | Purpose              |
| ----------------------------------- | -------------------- |
| `~/.config/wt/hooks/custom/`    | Your custom hooks    |
| `~/.config/wt/hooks/community/` | Bundled/shared hooks |

Custom hooks take precedence over community hooks with the same name.

### Installing Bundled Hooks

```bash
git wt config init --global
```

This installs bundled hooks to the community directory:

- `gh-default.sh` - Auto-configure GitHub CLI default repo
- `direnv.sh` - Auto-allow .envrc files
- `zoxide.sh` - Register worktrees with zoxide
- `github-issue.sh` - Fetch GitHub issue metadata (pre_create hook)
- `github-pr.sh` - Fetch GitHub PR metadata and branch (pre_create hook)

See [Hooks Examples](HOOKS.md) for the full hooks ecosystem documentation.

## Remote Branch Tracking

When running `git wt add <branch>`, wt checks if the branch already exists on any remote.

**Behavior:**

- If found on one remote: prompts to track or create new (unless `auto_track = true`)
- If found on multiple remotes: errors with list, requires `--remote` flag
- If not found: creates new local branch

**Flags:**

| Flag       | Description                                                      |
| ---------- | ---------------------------------------------------------------- |
| `--track`  | Track the remote branch (required in scripts when remote exists) |
| `--new`    | Force create new local branch even if remote exists              |
| `--fetch`  | Fetch all remotes before checking for branches                   |
| `--remote` | Specify which remote to use when branch exists on multiple       |

**Config:**

```toml
# Auto-track remote branches without prompting (like git checkout)
auto_track = false
```

When `auto_track = true`, wt automatically tracks the remote branch without prompting in interactive mode. In JSON/script mode, `--track` is still required for explicit intent.
