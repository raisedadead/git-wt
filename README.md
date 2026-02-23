# wt

[![Go Version](https://img.shields.io/github/go-mod/go-version/raisedadead/wt)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/raisedadead/wt)](https://github.com/raisedadead/wt/releases)

A git worktree manager built around the bare repository workflow. Use it as a
CLI, a TUI, or both.

`wt` wraps `git` and `gh` to handle the boilerplate of cloning bare repos,
creating worktrees, and running post-create hooks. Running `wt` with no
arguments opens a lazygit-style terminal interface; every operation is also
available as a flag-only subcommand for scripting.

## Why worktrees?

Git worktrees let you check out multiple branches simultaneously, each in its
own directory, sharing a single repository. No more stashing, no more broken IDE
state, no more waiting for CI on main while you develop in another branch.

```
project/
├── .bare/            # one shared repository
├── main/             # stable branch
├── feature-auth/     # feature in progress
└── fix-42-login/     # bugfix from issue #42
```

Switching context is just `cd ../feature-auth`.

## Install

```bash
# Homebrew (includes man page + shell completions)
brew install raisedadead/tap/wt

# Go
go install github.com/raisedadead/wt/cmd/wt@latest
```

Homebrew also installs `git-wt`, so `git wt` works as a subcommand.

### Requirements

- Git 2.20+
- [GitHub CLI](https://cli.github.com/) (`gh`) — optional, for issue/PR workflows
- [zoxide](https://github.com/ajeetdsouza/zoxide) — optional, for quick directory jumping

## Quick start

```bash
# Clone a repo into the bare worktree structure
wt clone owner/repo

# Open the TUI
wt

# Create worktrees
wt add my-experiment              # plain branch
wt add --feature auth             # feature workflow → feat/auth
wt add --bugfix --issue 42        # bugfix from GitHub issue → fix/42-title
wt add --pr-review 123            # check out a PR's actual branch

# Navigate
wt switch feature/auth            # prints path; shell wrapper does cd
wt list                           # table view with status

# Clean up
wt delete feature/auth
wt prune                          # remove worktrees for deleted/merged branches
```

## Terminal UI

Running `wt` with no arguments launches a full-screen TUI:

- **Worktree list** — browse, filter (`/`), multi-select (`space`)
- **Detail panel** — info, diff, and log tabs for the selected worktree
- **Single-key actions** — `enter` switch, `n` new, `N` workflow, `d` delete, `p` prune, `f` fetch
- **Overlay dialogs** — confirmation prompts, text input, workflow menu

Press `?` for the full keybinding reference. Press `tab` to move focus between panels.

## Commands

| Command              | Description                                         | Alias   |
| -------------------- | --------------------------------------------------- | ------- |
| `clone <repo>`       | Clone as bare repo with initial worktree            |         |
| `add [branch]`       | Create worktree (with optional workflow)             | `new`   |
| `list`               | List worktrees with status                          | `ls`    |
| `switch [branch]`    | Switch to a worktree directory                      |         |
| `delete [branch...]` | Remove worktrees and their branches                 | `rm`    |
| `prune`              | Remove stale worktrees (deleted remote, merged)     |         |
| `repair`             | Fix worktree paths after moving a repository        |         |
| `config init\|show`  | Create or inspect configuration                     |         |
| `hooks`              | Manage hooks (`list`, `enable`, `disable`, `show`)  |         |
| `completion`         | Generate shell completions (bash/zsh/fish/pwsh)     |         |

All commands support `--json` for machine-readable output.

### Workflows

Workflows combine branch naming conventions with hook execution. Use them with
`wt add`:

| Flag              | Workflow   | Branch template     | Description                         |
| ----------------- | ---------- | ------------------- | ----------------------------------- |
| `--feature`, `-f` | `feature`  | `feat/{slug}`       | New feature development             |
| `--bugfix`, `-b`  | `bugfix`   | `fix/{slug}`        | Bug fix                             |
| `--pr-review`     | `pr-review`| *(PR's own branch)* | Review a pull request               |
| `--workflow <w>`  | *(custom)* | *(from config)*     | Use a workflow defined in your TOML |

Combine with `--issue <n>` or `--pr <n>` to pass GitHub metadata to hooks:

```bash
wt add --bugfix --issue 42       # hook fetches issue title → fix/42-login-timeout
wt add --pr-review 123           # hook resolves PR's head branch
```

### Common flags

| Flag             | Commands                 | Description                    |
| ---------------- | ------------------------ | ------------------------------ |
| `--json`         | all                      | JSON output for scripting      |
| `--yes`, `-y`    | `delete`                 | Skip confirmation              |
| `--force`, `-f`  | `delete`, `clone`        | Force the operation            |
| `--dry-run`      | `delete`, `prune`        | Preview without changing state |
| `--timeout`      | `clone`, `add`, `delete` | Override git timeout (seconds) |
| `--hook-timeout` | `clone`, `add`           | Override hook timeout          |
| `--no-hooks`     | `clone`, `add`           | Skip all hooks                 |
| `--fetch`        | `add`, `prune`           | Fetch remotes first            |
| `--base`         | `add`                    | Base branch for new worktree   |
| `--track`        | `add`                    | Track existing remote branch   |
| `--new`          | `add`                    | Force new local branch         |
| `--merged`       | `prune`                  | Also prune merged branches     |
| `--path`         | `list`                   | Output paths only              |

Passthrough git flags after `--`:

```bash
wt clone owner/repo -- --depth=1 --single-branch
```

## Configuration

Hierarchical TOML config, highest priority first:

```
runtime flag → .wt.toml (repo) → ~/.config/wt/config.toml (global) → defaults
```

```bash
wt config init --global   # create global config
wt config init            # create repo-level config
wt config show            # show effective config with sources
```

Example configuration:

```toml
worktree_root = "~/DEV/worktrees"
default_owner = "myorg"
default_remote = "origin"
default_base_branch = "main"
branch_template = "{{type}}/{{number}}-{{slug}}"

git_timeout = 120
git_long_timeout = 600
hook_timeout = 30

[hooks]
post_clone = ["zoxide", "gh-default"]
post_add = ["zoxide", "direnv"]
```

See [Configuration](docs/CONFIGURATION.md) for the full options reference,
custom workflows, and template syntax.

## Hooks

wt has an oh-my-zsh-style hook system. Hooks can be bundled scripts, custom
scripts, or inline shell commands.

### Bundled hooks

| Hook           | Events               | Description                                       |
| -------------- | -------------------- | ------------------------------------------------- |
| `zoxide`       | post_clone, post_add | Register worktrees for quick `z` navigation       |
| `gh-default`   | post_clone           | Auto-configure `gh` CLI default repo              |
| `direnv`       | post_add             | Auto-allow `.envrc` in new worktrees              |
| `github-issue` | pre_create           | Fetch issue metadata, suggest branch name         |
| `github-pr`    | pre_create           | Fetch PR metadata, use the PR's actual branch     |

### Quick setup

```bash
# Enable zoxide integration (one-time)
wt hooks enable zoxide

# Now every new worktree registers itself
wt add --feature auth
z auth                      # jump there instantly
z main                      # jump back
```

### Managing hooks

```bash
wt hooks list               # see available hooks and status
wt hooks enable direnv      # enable using declared events
wt hooks disable zoxide     # disable
wt hooks show github-issue  # view hook details and source
```

### Custom hooks

Place scripts in `~/.config/wt/hooks/custom/`:

```bash
#!/bin/bash
# @name: my-setup
# @description: Project-specific setup
# @events: post_add

cd "$WT_PATH" || exit 0
npm install
cp .env.example .env
```

Hooks have access to environment variables: `WT_PATH`, `WT_BRANCH`,
`WT_PROJECT_ROOT`, `WT_DEFAULT_BRANCH`, `WT_WORKFLOW`, `WT_ISSUE`, `WT_PR`,
and more. Pre-create hooks can communicate back to wt through the
[hook helper protocol](docs/HOOKS.md#hook-helper-protocol) to suggest branch
names and pass metadata.

See [Hooks](docs/HOOKS.md) for recipes, the helper library, and the full
protocol reference.

## Shell completions

Completions include a shell wrapper so `wt switch` automatically `cd`s into the
target worktree. Generated completions also cover branch names, workflow names,
and hook names.

```bash
# Bash
source <(wt completion bash)

# Zsh
wt completion zsh > "${fpath[1]}/_wt"

# Fish
wt completion fish > ~/.config/fish/completions/wt.fish

# PowerShell
wt completion powershell >> $PROFILE
```

Homebrew installs completions automatically.

## JSON output

Every command supports `--json` for scripting and automation:

```bash
wt list --json | jq '.data.worktrees[].branch'
wt add --feature auth --json | jq '.data.path'
```

Output follows a consistent envelope:

```json
{
  "success": true,
  "command": "list",
  "data": { "..." },
  "error": null
}
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — design goals and internals
- [Configuration](docs/CONFIGURATION.md) — all options, workflows, and templates
- [Hooks](docs/HOOKS.md) — recipes, custom hooks, and the helper protocol
- [Contributing](docs/CONTRIBUTING.md) — development setup and guidelines
- [Releasing](docs/RELEASING.md) — release process

## Development

```bash
just setup          # install dev tools + git hooks
just build          # build to ./bin/
just test           # run all tests
just lint           # go vet + golangci-lint
just build-all      # cross-platform build check (linux/darwin/windows)
```

See [Contributing](docs/CONTRIBUTING.md) for the full development guide.

## Links

- [Git Worktrees Documentation](https://git-scm.com/docs/git-worktree) — official reference
- [Bare Repo + Worktree Workflow](https://nicknisi.com/posts/git-worktrees/) — the workflow wt implements
- [GitHub CLI](https://cli.github.com/) — required for issue/PR integration

## License

MIT — see [LICENSE](LICENSE)
