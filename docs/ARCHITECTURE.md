# Architecture

This document describes the design goals, architecture, and internals of wt.

## Design Goals

### 1. Wrapper, Not Replacement

wt wraps existing CLI tools rather than reimplementing their functionality:

- **git** - All git operations (`clone`, `worktree`, `branch`, `fetch`) delegate to the git CLI
- **gh** - GitHub issue/PR metadata is fetched via hooks that call the GitHub CLI (optional)

This keeps wt simple, avoids reimplementing complex git internals, and ensures compatibility with git's evolution.

### 2. Bare Repository Workflow

wt implements the bare repository + worktree pattern popularized by [Nick Nisi](https://nicknisi.com/posts/git-worktrees/) and [Josh Medeski](https://www.joshmedeski.com/posts/how-to-use-git-worktrees/).

**Why this pattern?**

| Problem                                  | Solution                              |
| ---------------------------------------- | ------------------------------------- |
| Context switching requires stashing      | Each worktree is isolated             |
| Can't run tests on main while developing | Parallel worktrees                    |
| Branch switching breaks IDE state        | Each worktree is a separate directory |
| Forgetting to stash loses work           | Changes stay in their worktree        |

**Structure:**

```
project/
├── .bare/              # Bare git repository (hidden)
├── .git                # Pointer file: "gitdir: ./.bare"
├── main/               # Stable worktree
├── feature-auth/       # Feature worktree
└── issue-42-fix/       # Issue worktree
```

The bare repo (`.bare/`) contains all git objects and refs. Worktrees are siblings that share this history. Switching context = switching directories.

### 3. Hooks for Extensibility

wt uses a hooks system for post-operation customization:

```toml
[hooks]
post_clone = ["zoxide add $WT_PATH"]
post_add = ["zoxide add $WT_PATH", "direnv allow"]
```

Hooks are user-configurable shell commands that run after worktree operations. This replaces hardcoded integrations with a flexible, user-controlled system.

Core functionality (clone, add, list, delete, prune) works without any hooks configured.

### 4. TUI-First

Running bare `wt` launches a lazygit-style TUI (bubbletea):

- Full-screen terminal interface with worktree list and detail panels
- Single-key contextual actions (enter to switch, n for new, d for delete, etc.)
- Overlay dialogs for confirmations and text input
- All subcommands remain flag-only for scripting (`wt switch <branch>`, `wt delete --force <branch>`)

### 5. Convention Over Configuration

Sensible defaults that work out of the box:

| Setting           | Default    | Override                  |
| ----------------- | ---------- | ------------------------- |
| Worktree location | (cwd)      | `worktree_root` in config |
| Remote            | `origin`   | `default_remote`          |
| Git timeout       | 2 minutes  | `git_timeout`             |
| Hook timeout      | 30 seconds | `hook_timeout`            |
| Hooks             | None       | `[hooks]` in config       |

Hierarchical config: `runtime flag > .wt.toml (repo) > ~/.config/wt/config.toml (global) > defaults`

### 6. Extensibility

wt supports passthrough flags to underlying git commands:

```bash
# Pass --depth to git clone
git wt clone owner/repo -- --depth=1

# Pass --single-branch flag
git wt clone owner/repo -- --single-branch --depth=1
```

### 7. Machine-Readable Output

All commands support `--json` for scripting and automation:

```bash
git wt list --json | jq '.data.worktrees[].branch'
git wt clone owner/repo --json
```

JSON envelope format:

```json
{
  "success": true,
  "command": "list",
  "data": { ... },
  "error": null
}
```

## Architecture

### Package Structure

```
cmd/wt/
└── main.go                 # Entry point

internal/
├── tui/                    # Lazygit-style TUI (bubbletea)
│   ├── tui.go             # Entry point: Run()
│   ├── model.go           # Root model, Update/View, key routing
│   ├── keys.go            # Keybindings
│   ├── styles.go          # Shared styles and colors
│   ├── messages.go        # Custom tea.Msg types
│   ├── commands.go        # Async tea.Cmd wrappers
│   ├── panels/            # Panels: worktree list, detail, header, footer
│   └── overlays/          # Overlays: help, confirm, input, menu
│
├── commands/               # CLI layer (Cobra) - flag-only, no interactive prompts
│   ├── root.go            # Root command, version, global flags, TUI launch
│   ├── clone.go           # Clone bare repo
│   ├── new.go             # Create worktree with workflows (add/new aliases)
│   ├── list.go            # List worktrees
│   ├── delete.go          # Remove worktree
│   ├── prune.go           # Clean stale worktrees
│   ├── config.go          # Config init/show subcommands
│   ├── hooks.go           # Hooks management subcommands
│   └── completion.go      # Shell completions
│
├── git/                    # Git operations
│   ├── exec.go            # Command execution with timeouts
│   ├── bare.go            # Bare repo operations
│   ├── worktree.go        # Worktree CRUD
│   ├── branch.go          # Branch name utilities
│   └── validate.go        # Input validation
│
├── hooks/                  # Hook execution
│   ├── hooks.go           # Run hooks with workflow support
│   ├── resolver.go        # Hook name resolution
│   ├── metadata.go        # Hook metadata parsing
│   ├── hooks_unix.go      # Unix process groups
│   ├── hooks_windows.go   # Windows stub
│   └── bundled/           # Embedded hook scripts
│       ├── embed.go       # Go embed for scripts
│       ├── helpers.sh     # Helper library for hooks
│       ├── github-issue.sh # GitHub issue integration
│       ├── github-pr.sh   # GitHub PR integration
│       ├── gh-default.sh  # Auto-configure gh CLI
│       ├── direnv.sh      # Auto-allow .envrc
│       └── zoxide.sh      # Auto-register with zoxide
│
├── config/                 # Configuration
│   └── config.go          # TOML config loading + workflows
│
└── ui/                     # Terminal UI
    ├── styles.go          # Lipgloss styles
    └── output.go          # JSON output envelope
```

### Layer Responsibilities

**Commands Layer** (`internal/commands/`)

- Parse CLI arguments and flags
- Handle interactive prompts
- Orchestrate operations
- Format output for users

**Git Layer** (`internal/git/`)

- Execute git commands with timeouts
- Parse git output (porcelain format)
- Validate input (branch names, project names)
- No user-facing output

**Hooks Layer** (`internal/hooks/`)

- Execute user-configured shell commands
- Set environment variables for hooks
- Handle hook failures gracefully

**Config Layer** (`internal/config/`)

- Load TOML config from XDG locations
- Merge hierarchical config (repo > global > defaults)
- Provide `config init` and `config show` commands

### Data Flow

#### Clone Command

```
User: git wt clone owner/repo
         │
         ▼
    ┌─────────────────┐
    │  commands/clone │  Parse args, expand shorthand
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  config/config  │  Load global config (worktree_root)
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  git/validate   │  Validate project name
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │   git/bare      │  git clone --bare
    └────────┬────────┘  git config remote.origin.fetch
             │           git fetch origin
             ▼
    ┌─────────────────┐
    │  git/worktree   │  git worktree add main/
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │   hooks/hooks   │  Run post_clone hooks
    └─────────────────┘
```

#### Add Command (with workflow)

```
User: git wt add --bugfix --issue 42
         │
         ▼
    ┌─────────────────┐
    │  commands/add   │  Parse flags, select workflow
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │   git/bare      │  Find project root (.bare/)
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  config/config  │  Load workflow config (bugfix)
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  hooks/hooks    │  Run pre_create hooks (github-issue.sh)
    └────────┬────────┘  Hook fetches issue, suggests branch name
             │
             ▼
    ┌─────────────────┐
    │  git/validate   │  Validate suggested branch name
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  git/worktree   │  git worktree add -b branch
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │   hooks/hooks   │  Run post_add hooks (direnv, zoxide)
    └─────────────────┘
```

The workflow system uses a hook helper protocol where pre_create hooks can:

- Fetch external metadata (GitHub issues, PRs)
- Suggest branch names
- Pass metadata to subsequent hooks

#### Switch Command

```
User: wt switch feature/auth
         │
         ▼
    ┌─────────────────┐
    │ commands/switch │  Parse args, find worktree
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  git/worktree   │  List worktrees, match by branch or dir
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │     stdout      │  Output path (shell wrapper does cd)
    └─────────────────┘
```

The switch command outputs the worktree path to stdout. Shell completions include a wrapper function that captures this output and performs `cd` automatically.

### Security Considerations

**Input Validation:**

- Project names validated against path traversal (`..`, `/`)
- Branch names validated against git restrictions
- Reserved names (`.git`, `.bare`) rejected

**Command Execution:**

- All git commands use context with timeout (2min default, 10min for clone/fetch)
- Git stderr captured and included in error messages
- Hook commands run with configurable timeout (default 30 seconds)
- Hook template variables are shell-quoted to prevent injection
- Process groups used on Unix to kill child processes on timeout

**File Operations:**

- TOCTOU race mitigated in clone with atomic `os.Mkdir`

### Testing Strategy

**Unit Tests:**

- `git/validate_test.go` - Input validation
- `git/worktree_test.go` - Porcelain parsing
- `git/exec_test.go` - Command execution
- `git/branch_test.go` - Branch name utilities
- `github/gh_test.go` - Issue/PR fetching
- `config/config_test.go` - Config loading
- `hooks/hooks_test.go` - Hook execution

**Integration Tests:**

- TBD: End-to-end tests with test repositories

### Known Limitations

1. **Single remote** - Assumes `origin` remote
2. **gh flags not passthrough** - Only git clone flags are supported via `--` separator

## Dependencies

| Package                              | Purpose          |
| ------------------------------------ | ---------------- |
| `github.com/spf13/cobra`             | CLI framework    |
| `github.com/charmbracelet/bubbletea` | TUI framework    |
| `github.com/charmbracelet/bubbles`   | TUI components   |
| `github.com/charmbracelet/lipgloss`  | Terminal styling |
| `github.com/BurntSushi/toml`         | Config parsing   |
