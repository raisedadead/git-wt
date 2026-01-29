# Hooks Examples

git-wt supports `post_clone` and `post_add` hooks for running shell commands after worktree operations. This document provides common recipes.

## zoxide Integration

Add worktrees to [zoxide](https://github.com/ajeetdsouza/zoxide) for quick navigation.

```toml
[hooks]
post_clone = [
  "zoxide add $GIT_WT_PATH",
]

post_add = [
  "zoxide add $GIT_WT_PATH",
]
```

**Usage after setup:**

```bash
z feature-auth    # Jump to feature/auth worktree
z fcc main        # Jump to freeCodeCamp main worktree
```

## direnv Integration

Auto-allow [direnv](https://direnv.net/) in new worktrees.

```toml
[hooks]
post_add = [
  "direnv allow",
]
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

A comprehensive configuration combining multiple integrations:

```toml
# ~/.config/git-wt/config.toml
worktree_root = "~/DEV/worktrees"

[hooks]
post_clone = [
  "zoxide add $GIT_WT_PATH",
]

post_add = [
  # Navigation
  "zoxide add $GIT_WT_PATH",

  # Environment files
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.envrc $GIT_WT_PATH/ 2>/dev/null || true",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.env $GIT_WT_PATH/ 2>/dev/null || true",

  # IDE settings
  "cp -r $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.vscode $GIT_WT_PATH/ 2>/dev/null || true",

  # direnv (must come after .envrc copy)
  "direnv allow",

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
