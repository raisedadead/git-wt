package bundled

import "embed"

//go:embed *.sh
var Scripts embed.FS

// List returns the names of all bundled hooks (without .sh extension)
func List() []string {
	return []string{"gh-default", "direnv", "zoxide", "github-issue", "github-pr"}
}

// HelpersScript is the name of the helpers library
const HelpersScript = "helpers.sh"

// GetHelpers returns the contents of the helpers library
func GetHelpers() ([]byte, error) {
	return Scripts.ReadFile(HelpersScript)
}
