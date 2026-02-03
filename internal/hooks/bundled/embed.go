package bundled

import "embed"

//go:embed *.sh
var Scripts embed.FS

// List returns the names of all bundled hooks (without .sh extension)
func List() []string {
	return []string{"gh-default", "direnv", "zoxide"}
}
