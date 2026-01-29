# Configuration

git-wt uses a TOML configuration file with sensible defaults.

## Configuration File

| Location                              | Description       |
| ------------------------------------- | ----------------- |
| `$XDG_CONFIG_HOME/git-wt/config.toml` | Primary location  |
| `~/.config/git-wt/config.toml`        | Fallback location |

## Options

### worktree_root

Directory where projects are cloned.

```toml
worktree_root = "~/DEV/worktrees"
```

**Default:** `~/DEV/worktrees`

### [hooks]

Hooks run shell commands after worktree operations. Commands run in sequence; failures log warnings but do not block execution.

```toml
[hooks]
post_clone = [
  "zoxide add $GIT_WT_PATH",
]

post_add = [
  "zoxide add $GIT_WT_PATH",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.envrc $GIT_WT_PATH/ 2>/dev/null || true",
  "direnv allow",
]
```

**Available hooks:**

| Hook         | Trigger                          |
| ------------ | -------------------------------- |
| `post_clone` | After `git wt clone` completes   |
| `post_add`   | After `git wt add/new` completes |

## Environment Variables

Hooks have access to these environment variables:

| Variable                | Description                              | Example                              |
| ----------------------- | ---------------------------------------- | ------------------------------------ |
| `GIT_WT_PATH`           | Path to the new worktree                 | `/home/user/DEV/worktrees/repo/main` |
| `GIT_WT_BRANCH`         | Branch name of the new worktree          | `feature/auth`                       |
| `GIT_WT_PROJECT_ROOT`   | Project root (contains `.bare/`)         | `/home/user/DEV/worktrees/repo`      |
| `GIT_WT_DEFAULT_BRANCH` | Default branch name (main, master, etc.) | `main`                               |

## Template Syntax

Hook commands support Go template variables as an alternative to environment variables:

| Template             | Equivalent Variable      |
| -------------------- | ------------------------ |
| `{{.Path}}`          | `$GIT_WT_PATH`           |
| `{{.Branch}}`        | `$GIT_WT_BRANCH`         |
| `{{.ProjectRoot}}`   | `$GIT_WT_PROJECT_ROOT`   |
| `{{.DefaultBranch}}` | `$GIT_WT_DEFAULT_BRANCH` |

Example using templates:

```toml
[hooks]
post_add = [
  "echo 'Created worktree at {{.Path}} for branch {{.Branch}}'",
]
```

## Hook Behavior

- Hooks run in the order listed
- Each hook runs with the worktree path as the working directory
- A failing hook logs a warning but does not prevent subsequent hooks
- Hooks have a 30-second timeout per command

## Example Configurations

### Minimal Config

```toml
# ~/.config/git-wt/config.toml
worktree_root = "~/code/worktrees"
```

### Full Config

```toml
# ~/.config/git-wt/config.toml
worktree_root = "~/DEV/worktrees"

[hooks]
post_clone = [
  "zoxide add $GIT_WT_PATH",
]

post_add = [
  "zoxide add $GIT_WT_PATH",
  "cp $GIT_WT_PROJECT_ROOT/$GIT_WT_DEFAULT_BRANCH/.envrc $GIT_WT_PATH/ 2>/dev/null || true",
  "direnv allow",
  "~/.local/bin/generate-ai-context.sh $GIT_WT_PATH",
]
```

See [Hooks Examples](HOOKS.md) for more hook recipes.
