package hooks

import "testing"

func TestParseMetadata(t *testing.T) {
	script := `#!/bin/bash
# @name: gh-default
# @description: Auto-configure GitHub CLI default repository
# @events: post_clone
# @requires: gh

cd "$WT_PATH" || exit 0
`

	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Name != "gh-default" {
		t.Errorf("expected name gh-default, got %s", meta.Name)
	}
	if meta.Description != "Auto-configure GitHub CLI default repository" {
		t.Errorf("unexpected description: %s", meta.Description)
	}
	if len(meta.Events) != 1 || meta.Events[0] != "post_clone" {
		t.Errorf("unexpected events: %v", meta.Events)
	}
	if meta.Requires != "gh" {
		t.Errorf("expected requires gh, got %s", meta.Requires)
	}
}

func TestParseMetadata_MultipleEvents(t *testing.T) {
	script := `#!/bin/bash
# @name: test
# @events: post_clone, post_add
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meta.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(meta.Events))
	}
	if meta.Events[0] != "post_clone" || meta.Events[1] != "post_add" {
		t.Errorf("unexpected events: %v", meta.Events)
	}
}

func TestParseMetadata_NoMetadata(t *testing.T) {
	script := `#!/bin/bash
echo "hello"
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "" {
		t.Errorf("expected empty name, got %s", meta.Name)
	}
}

func TestParseMetadata_EmptyScript(t *testing.T) {
	meta, err := ParseMetadata("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "" || meta.Description != "" || len(meta.Events) != 0 || meta.Requires != "" {
		t.Errorf("expected empty metadata for empty script")
	}
}

func TestParseMetadata_OnlyShebang(t *testing.T) {
	script := `#!/bin/bash
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "" {
		t.Errorf("expected empty name for shebang-only script, got %s", meta.Name)
	}
}

func TestParseMetadata_WhitespaceHandling(t *testing.T) {
	script := `#!/bin/bash
#   @name:   spacy-name
#@description:no space before tag
# @events:   post_clone  ,  post_add  ,post_delete
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "spacy-name" {
		t.Errorf("expected trimmed name 'spacy-name', got '%s'", meta.Name)
	}
	if meta.Description != "no space before tag" {
		t.Errorf("expected trimmed description, got '%s'", meta.Description)
	}
	if len(meta.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(meta.Events))
	}
	for i, expected := range []string{"post_clone", "post_add", "post_delete"} {
		if i < len(meta.Events) && meta.Events[i] != expected {
			t.Errorf("event[%d] expected '%s', got '%s'", i, expected, meta.Events[i])
		}
	}
}

func TestParseMetadata_StopsAtNonComment(t *testing.T) {
	script := `#!/bin/bash
# @name: before-code
echo "some code"
# @description: after code - should be ignored
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "before-code" {
		t.Errorf("expected name 'before-code', got '%s'", meta.Name)
	}
	if meta.Description != "" {
		t.Errorf("expected empty description (after code), got '%s'", meta.Description)
	}
}

func TestParseMetadata_CommentsWithoutTags(t *testing.T) {
	script := `#!/bin/bash
# This is a regular comment
# Another comment
# @name: my-hook
# Yet another comment without tag
# @description: My description
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "my-hook" {
		t.Errorf("expected name 'my-hook', got '%s'", meta.Name)
	}
	if meta.Description != "My description" {
		t.Errorf("expected description 'My description', got '%s'", meta.Description)
	}
}

func TestParseMetadata_RealBundledScript(t *testing.T) {
	// Test with actual bundled script content
	script := `#!/bin/bash
# @name: direnv
# @description: Auto-allow .envrc files in worktrees
# @events: post_add
# @requires: direnv

cd "$WT_PATH" || exit 0

# Skip if direnv not installed
command -v direnv &>/dev/null || exit 0

# Skip if no .envrc
[[ -f .envrc ]] || exit 0

direnv allow
`
	meta, err := ParseMetadata(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Name != "direnv" {
		t.Errorf("expected name 'direnv', got '%s'", meta.Name)
	}
	if meta.Description != "Auto-allow .envrc files in worktrees" {
		t.Errorf("unexpected description: '%s'", meta.Description)
	}
	if len(meta.Events) != 1 || meta.Events[0] != "post_add" {
		t.Errorf("unexpected events: %v", meta.Events)
	}
	if meta.Requires != "direnv" {
		t.Errorf("expected requires 'direnv', got '%s'", meta.Requires)
	}
}
