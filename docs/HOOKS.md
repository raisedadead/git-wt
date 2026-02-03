# Hooks

git-wt supports `post_clone` and `post_add` hooks for running shell commands after worktree operations.

## Quick Start: Auto-Navigation with zoxide

git-wt integrates with [zoxide](https://github.com/ajeetdsouza/zoxide) for instant worktree navigation.

**Setup (one-time):**

```bash
git wt hooks enable zoxide
```

Or enable during initial setup—`git wt config init --global` will detect zoxide and offer to enable it automatically.

**Workflow:**

```bash
git wt add feature/auth    # Creates worktree, registers with zoxide
z auth                     # Jump to it instantly
z main                     # Jump back to main
```

This is the recommended way to navigate between worktrees. No shell wrappers or subshells needed.

---

## Hooks Ecosystem

git-wt provides an oh-my-zsh-style hooks ecosystem. Hooks can be:

- **Bundled hooks** - Pre-packaged scripts for common integrations
- **Custom hooks** - Your own scripts in `~/.config/git-wt/hooks/custom/`
- **Inline commands** - Traditional shell commands in config

### Setup

```bash
# Initialize hooks directory and install bundled hooks
git wt config init --global
```

This creates:

```
~/.config/git-wt/
├── config.toml
└── hooks/
    ├── community/      # Bundled and shared hooks
    │   ├── gh-default.sh
    │   ├── direnv.sh
    │   └── zoxide.sh
    └── custom/         # Your custom hooks
```

### Managing Hooks

```bash
# List available hooks and their status
git wt hooks list

# Enable a hook (uses its declared events)
git wt hooks enable zoxide

# Enable a hook for a specific event
git wt hooks enable my-hook --event post_add

# Disable a hook
git wt hooks disable zoxide

# Show hook details and content
git wt hooks show zoxide
```

### Bundled Hooks

| Hook       | Events               | Description                                      |
| ---------- | -------------------- | ------------------------------------------------ |
| zoxide     | post_clone, post_add | Auto-register worktrees for quick `z` navigation |
| gh-default | post_clone           | Auto-configure GitHub CLI default repo           |
| direnv     | post_add             | Auto-allow .envrc files in new worktrees         |

### Creating Custom Hooks

Create a script in `~/.config/git-wt/hooks/custom/`:

```bash
#!/bin/bash
# @name: my-hook
# @description: My custom hook
# @events: post_add
# @requires: some-tool

cd "$GIT_WT_PATH" || exit 0

# Your hook logic here
```

Metadata tags:

- `@name`: Hook name (optional, defaults to filename)
- `@description`: What the hook does
- `@events`: Comma-separated list (post_clone, post_add)
- `@requires`: Required external tool (optional)

### Resolution Order

When a hook name is referenced in config:

1. First checks `~/.config/git-wt/hooks/custom/<name>.sh`
2. Then checks `~/.config/git-wt/hooks/community/<name>.sh`
3. Falls back to treating it as an inline shell command

---

## direnv Integration

Auto-allow [direnv](https://direnv.net/) in new worktrees.

```bash
git wt hooks enable direnv
```

Or manually in config:

```toml
[hooks]
post_add = ["direnv"]
```

## File Copying

Copy environment files from the default branch worktree.

```toml
[hooks]
post_add = [
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.envrc $GIT_WT_PATH/ 2>/dev/null || true",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.env $GIT_WT_PATH/ 2>/dev/null || true",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.env.local $GIT_WT_PATH/ 2>/dev/null || true",
]
```

The `2>/dev/null || true` pattern silently skips missing files.

### Copy Directories

```toml
[hooks]
post_add = [
  "cp -r $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.vscode $GIT_WT_PATH/ 2>/dev/null || true",
]
```

## AI Context File Generation

Generate context files for AI coding assistants via external script.

**Hook configuration:**

```toml
[hooks]
post_add = [
  "~/.local/bin/generate-ai-context.sh",
]
```

**Example script (`~/.local/bin/generate-ai-context.sh`):**

```bash
#!/bin/bash
# Generate CLAUDE.md with worktree context

cat > "$GIT_WT_PATH/CLAUDE.md" << EOF
# Worktree Context

- **Branch:** $GIT_WT_BRANCH
- **Project Root:** $GIT_WT_PROJECT_ROOT
- **Default Branch:** $GIT_WT_DEFAULT_BRANCH
- **Created:** $(date +%Y-%m-%d)

This is an isolated git worktree. The main branch is at \`../$GIT_WT_DEFAULT_BRANCH/\`.
EOF
```

## Combined Setup

A comprehensive configuration using bundled hooks and custom inline commands:

```toml
# ~/.config/git-wt/config.toml
worktree_root = "~/DEV/worktrees"

[hooks]
post_clone = [
  "zoxide",      # Bundled: register with zoxide
  "gh-default",  # Bundled: set gh default repo
]

post_add = [
  "zoxide",  # Bundled: register with zoxide

  # Environment files (inline commands)
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.envrc $GIT_WT_PATH/ 2>/dev/null || true",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.env $GIT_WT_PATH/ 2>/dev/null || true",

  # IDE settings
  "cp -r $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.vscode $GIT_WT_PATH/ 2>/dev/null || true",

  "direnv",  # Bundled: allow direnv (must come after .envrc copy)

  # AI context
  "~/.local/bin/generate-ai-context.sh",
]
```

## GitHub CLI Integration

**Note:** GitHub CLI (`gh`) integration for `--issue` and `--pr` flags is built into git-wt and does not require hooks configuration. The `gh` CLI must be installed and authenticated:

```bash
gh auth login
```

## Tool Installation

| Tool   | Install                               |
| ------ | ------------------------------------- |
| zoxide | https://github.com/ajeetdsouza/zoxide |
| direnv | https://direnv.net/                   |
| gh     | https://cli.github.com/               |

All integrations are optional. Hooks that reference missing tools will log warnings but not block execution.

## Hook Behavior

- Hooks run in the order listed
- Each hook runs with the worktree path as the working directory
- A failing hook logs a warning but doesn't block subsequent hooks
- Each hook has a configurable timeout (default 30 seconds)
- Hooks exceeding the timeout are terminated (including child processes on Unix)

## Timeout Configuration

Override the default hook timeout in your config:

```toml
hook_timeout = 60  # seconds
```

Or per-command via flag:

```bash
git wt add feature/auth --hook-timeout 60
```

## Security

**Template values are shell-quoted** to prevent command injection. If a branch name contains special characters, they will be safely escaped.

Environment variables (`$GIT_WT_PATH`, etc.) are passed directly to the shell and are safe to use.

See [Configuration](CONFIGURATION.md) for all hook options.
