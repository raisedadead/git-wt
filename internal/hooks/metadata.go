package hooks

import (
	"bufio"
	"strings"
)

// Metadata holds parsed hook script metadata
type Metadata struct {
	Name        string   // @name
	Description string   // @description
	Events      []string // @events (comma-separated)
	Requires    string   // @requires
}

// ParseMetadata extracts metadata from hook script header comments.
// Metadata tags are expected in comment lines at the top of the script:
//
//	#!/bin/bash
//	# @name: my-hook
//	# @description: What this hook does
//	# @events: post_clone, post_add
//	# @requires: gh
//
// Parsing stops at the first non-comment, non-empty, non-shebang line.
func ParseMetadata(script string) (*Metadata, error) {
	meta := &Metadata{}
	scanner := bufio.NewScanner(strings.NewReader(script))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Stop at first non-comment, non-empty, non-shebang line
		if line != "" && !strings.HasPrefix(line, "#") {
			break
		}

		// Skip shebang and empty lines
		if line == "" || strings.HasPrefix(line, "#!") {
			continue
		}

		// Parse @ tags from comment lines
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "@name:") {
			meta.Name = strings.TrimSpace(strings.TrimPrefix(line, "@name:"))
		} else if strings.HasPrefix(line, "@description:") {
			meta.Description = strings.TrimSpace(strings.TrimPrefix(line, "@description:"))
		} else if strings.HasPrefix(line, "@events:") {
			eventsStr := strings.TrimSpace(strings.TrimPrefix(line, "@events:"))
			for _, e := range strings.Split(eventsStr, ",") {
				meta.Events = append(meta.Events, strings.TrimSpace(e))
			}
		} else if strings.HasPrefix(line, "@requires:") {
			meta.Requires = strings.TrimSpace(strings.TrimPrefix(line, "@requires:"))
		}
	}

	return meta, scanner.Err()
}
