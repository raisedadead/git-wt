package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
)

type claudeHookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type claudeHookMatcher struct {
	Hooks []claudeHookEntry `json:"hooks"`
}

type claudeSettings struct {
	Hooks map[string][]claudeHookMatcher `json:"hooks"`
}

func generateClaudeSettings() ([]byte, error) {
	settings := claudeSettings{
		Hooks: map[string][]claudeHookMatcher{
			"WorktreeCreate": {
				{
					Hooks: []claudeHookEntry{
						{
							Type:    "command",
							Command: `bash -c 'set -euo pipefail; NAME=$(cat | jq -r .name); wt add "$NAME" --json | jq -r .data.path'`,
						},
					},
				},
			},
			"WorktreeRemove": {
				{
					Hooks: []claudeHookEntry{
						{
							Type:    "command",
							Command: `bash -c 'set -euo pipefail; WORKTREE_PATH=$(cat | jq -r .worktree_path); wt delete --path "$WORKTREE_PATH" --force --yes || true'`,
						},
					},
				},
			},
		},
	}

	return json.MarshalIndent(settings, "", "  ")
}

func symlinkClaudeDir(projectRoot, worktreePath string) {
	srcDir := filepath.Join(projectRoot, ".claude")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return
	}

	target := filepath.Join(worktreePath, ".claude")
	info, err := os.Lstat(target)
	if err == nil {
		// Target exists -- either symlink or directory, skip
		_ = info
		return
	}

	if err := os.Symlink("../.claude", target); err != nil {
		if !IsJSONOutput() {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("Could not symlink .claude/ into %s: %v", filepath.Base(worktreePath), err)))
		}
	}
}

func migrateClaudeDir(projectRoot string) error {
	claudeDir := filepath.Join(projectRoot, ".claude")
	info, err := os.Lstat(claudeDir)
	if err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return nil
	}

	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	for _, wt := range worktrees {
		wtClaude := filepath.Join(wt.Path, ".claude")
		wtInfo, err := os.Lstat(wtClaude)
		if err != nil {
			continue
		}
		if !wtInfo.IsDir() || wtInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if err := os.Rename(wtClaude, claudeDir); err != nil {
			return fmt.Errorf("migrate .claude from %s: %w", wt.Path, err)
		}

		if err := os.Symlink("../.claude", wtClaude); err != nil {
			return fmt.Errorf("create symlink at %s: %w", wtClaude, err)
		}

		if !IsJSONOutput() {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Migrated .claude/ from %s to project root", filepath.Base(wt.Path))))
		}
		return nil
	}

	return nil
}

func generateClaudeCommand() string {
	return `This project uses a bare repo worktree layout managed by the ` + "`wt`" + ` CLI.

## Project structure

` + "```" + `
project/
├── .bare/           # bare git repository
├── .git             # file pointing to .bare (gitdir: ./.bare)
├── .claude/         # shared Claude config (symlinked into each worktree)
├── main/            # default branch worktree
├── feature-auth/    # feature branch worktree
└── bugfix-login/    # another worktree
` + "```" + `

Branch names with slashes flatten to dashes for directory names: ` + "`feature/auth`" + ` → ` + "`feature-auth/`" + `.

The ` + "`.claude/`" + ` directory at the project root is shared across all worktrees via symlinks.

## Key commands

| Command | Description |
|---------|-------------|
| ` + "`wt list --json`" + ` | List all worktrees (JSON output for scripting) |
| ` + "`wt add <branch> --new --json`" + ` | Create new branch + worktree |
| ` + "`wt add <branch> --json`" + ` | Check out existing branch into a worktree |
| ` + "`wt delete <branch>`" + ` | Delete worktree by branch name |
| ` + "`wt delete --path <path>`" + ` | Delete worktree by filesystem path |

## How it works

- ` + "`wt`" + ` discovers the project root by walking up to find the ` + "`.bare/`" + ` directory
- You can run ` + "`wt`" + ` from any worktree directory within the project
- All worktree directories are siblings at the project root level
- Use ` + "`--json`" + ` on any command for machine-readable output
`
}

func installClaudeCommand(projectRoot string) error {
	commandsDir := filepath.Join(projectRoot, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("create .claude/commands directory: %w", err)
	}

	commandPath := filepath.Join(commandsDir, "wt.md")
	if _, err := os.Stat(commandPath); err == nil {
		if !IsJSONOutput() {
			fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("  Skipped: %s (already exists, delete to regenerate)", commandPath)))
		}
		return nil
	}

	if err := os.WriteFile(commandPath, []byte(generateClaudeCommand()), 0644); err != nil {
		return fmt.Errorf("write wt.md: %w", err)
	}

	return nil
}

func setupClaudeIntegration(projectRoot string) error {
	_ = migrateClaudeDir(projectRoot)

	claudeDir := filepath.Join(projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("create .claude directory: %w", err)
	}

	settingsBytes, err := generateClaudeSettings()
	if err != nil {
		return fmt.Errorf("generate settings: %w", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	existing, readErr := os.ReadFile(settingsPath)
	if readErr == nil && len(existing) > 0 {
		var existingMap map[string]interface{}
		if err := json.Unmarshal(existing, &existingMap); err == nil {
			var generated map[string]interface{}
			if err := json.Unmarshal(settingsBytes, &generated); err == nil {
				hooksExisting, ok := existingMap["hooks"].(map[string]interface{})
				if !ok {
					hooksExisting = make(map[string]interface{})
				}
				hooksGenerated, ok := generated["hooks"].(map[string]interface{})
				if !ok {
					hooksGenerated = make(map[string]interface{})
				}
				for k, v := range hooksGenerated {
					if _, exists := hooksExisting[k]; exists && !IsJSONOutput() {
						fmt.Println(ui.WarningMsg(fmt.Sprintf("Overwriting existing %s hook", k)))
					}
					hooksExisting[k] = v
				}
				existingMap["hooks"] = hooksExisting

				merged, err := json.MarshalIndent(existingMap, "", "  ")
				if err == nil {
					settingsBytes = merged
				}
			}
		}
	}

	if err := os.WriteFile(settingsPath, settingsBytes, 0644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	if err := installClaudeCommand(projectRoot); err != nil {
		return fmt.Errorf("install command file: %w", err)
	}

	worktrees, err := git.ListWorktrees(projectRoot)
	if err == nil {
		for _, wt := range worktrees {
			symlinkClaudeDir(projectRoot, wt.Path)
		}
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SuccessMsg("Claude Code integration configured"))
		fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("  Settings: %s", settingsPath)))
	}

	return nil
}

func resolveBranchFromPath(projectRoot, worktreePath string) (string, error) {
	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return "", fmt.Errorf("list worktrees: %w", err)
	}

	absPath, err := filepath.Abs(worktreePath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	absPath = filepath.Clean(absPath)

	for _, wt := range worktrees {
		wtAbs, err := filepath.Abs(wt.Path)
		if err != nil {
			continue
		}
		if filepath.Clean(wtAbs) == absPath {
			return wt.Branch, nil
		}
	}

	return "", fmt.Errorf("worktree not found at path: %s", absPath)
}

func isClaudeIntegrated(projectRoot string) bool {
	data, err := os.ReadFile(filepath.Join(projectRoot, ".claude", "settings.json"))
	if err != nil {
		return false
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return false
	}

	hooks, ok := parsed["hooks"].(map[string]interface{})
	if !ok {
		return false
	}

	wtCreate, ok := hooks["WorktreeCreate"]
	if !ok {
		return false
	}

	arr, ok := wtCreate.([]interface{})
	return ok && len(arr) > 0
}
