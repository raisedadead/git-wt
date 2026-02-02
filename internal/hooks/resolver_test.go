package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveHook_Inline(t *testing.T) {
	// Inline command (no matching file)
	result, isScript := ResolveHook("echo hello", "/nonexistent", "/nonexistent")
	if isScript {
		t.Error("expected inline command, not script")
	}
	if result != "echo hello" {
		t.Errorf("expected 'echo hello', got %s", result)
	}
}

func TestResolveHook_CustomDir(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(customDir, "my-hook.sh")
	if err := os.WriteFile(hookPath, []byte("#!/bin/bash\necho custom"), 0755); err != nil {
		t.Fatal(err)
	}

	result, isScript := ResolveHook("my-hook", customDir, "/nonexistent")
	if !isScript {
		t.Error("expected script, not inline")
	}
	if result != hookPath {
		t.Errorf("expected %s, got %s", hookPath, result)
	}
}

func TestResolveHook_CommunityDir(t *testing.T) {
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	if err := os.MkdirAll(communityDir, 0755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(communityDir, "gh-default.sh")
	if err := os.WriteFile(hookPath, []byte("#!/bin/bash\necho community"), 0755); err != nil {
		t.Fatal(err)
	}

	result, isScript := ResolveHook("gh-default", "/nonexistent", communityDir)
	if !isScript {
		t.Error("expected script, not inline")
	}
	if result != hookPath {
		t.Errorf("expected %s, got %s", hookPath, result)
	}
}

func TestResolveHook_CustomOverridesCommunity(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom")
	communityDir := filepath.Join(tmpDir, "community")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(communityDir, 0755); err != nil {
		t.Fatal(err)
	}

	customPath := filepath.Join(customDir, "test-hook.sh")
	communityPath := filepath.Join(communityDir, "test-hook.sh")
	if err := os.WriteFile(customPath, []byte("#!/bin/bash\necho custom"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(communityPath, []byte("#!/bin/bash\necho community"), 0755); err != nil {
		t.Fatal(err)
	}

	result, _ := ResolveHook("test-hook", customDir, communityDir)
	if result != customPath {
		t.Errorf("expected custom to override community, got %s", result)
	}
}

func TestResolveHook_EmptyDirs(t *testing.T) {
	// Both dirs empty - should fall back to inline
	result, isScript := ResolveHook("some-command", "", "")
	if isScript {
		t.Error("expected inline command when dirs are empty")
	}
	if result != "some-command" {
		t.Errorf("expected 'some-command', got %s", result)
	}
}

func TestResolveHook_PathTraversal(t *testing.T) {
	// Path traversal attempts should be treated as inline commands, not scripts
	tests := []struct {
		name string
		hook string
	}{
		{"dot-dot", "../etc/passwd"},
		{"slash", "foo/bar"},
		{"backslash", "foo\\bar"},
		{"complex traversal", "../../.ssh/id_rsa"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, isScript := ResolveHook(tt.hook, "/some/dir", "/other/dir")
			if isScript {
				t.Errorf("hook %q should not be treated as script (path traversal risk)", tt.hook)
			}
			if result != tt.hook {
				t.Errorf("expected %q returned as-is for inline, got %q", tt.hook, result)
			}
		})
	}
}

func TestResolveHooks(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(customDir, "my-hook.sh")
	if err := os.WriteFile(hookPath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	resolved := ResolveHooks([]string{"my-hook", "echo inline"}, customDir, "")
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved hooks, got %d", len(resolved))
	}
	if !resolved[0].IsScript || resolved[0].Path != hookPath {
		t.Errorf("first hook should be script at %s", hookPath)
	}
	if resolved[1].IsScript {
		t.Error("second hook should be inline")
	}
}

func TestResolveHooks_Empty(t *testing.T) {
	resolved := ResolveHooks([]string{}, "", "")
	if len(resolved) != 0 {
		t.Errorf("expected 0 resolved hooks for empty input, got %d", len(resolved))
	}
}

func TestResolveHooks_PreservesOrder(t *testing.T) {
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create multiple hooks
	for _, name := range []string{"first", "second", "third"} {
		hookPath := filepath.Join(customDir, name+".sh")
		if err := os.WriteFile(hookPath, []byte("#!/bin/bash\necho "+name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	resolved := ResolveHooks([]string{"first", "second", "third"}, customDir, "")
	if len(resolved) != 3 {
		t.Fatalf("expected 3 resolved hooks, got %d", len(resolved))
	}

	expectedNames := []string{"first", "second", "third"}
	for i, expected := range expectedNames {
		if resolved[i].Name != expected {
			t.Errorf("hook %d: expected name %s, got %s", i, expected, resolved[i].Name)
		}
	}
}
