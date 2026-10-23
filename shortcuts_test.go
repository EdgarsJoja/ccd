package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadShortcuts(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	want := []Shortcut{
		{Name: "project", Path: "/home/user/project"},
		{Name: "logs", Path: "/var/log"},
	}

	if err := SaveShortcuts(want); err != nil {
		t.Fatalf("SaveShortcuts() error = %v", err)
	}

	got, err := LoadShortcuts()
	if err != nil {
		t.Fatalf("LoadShortcuts() error = %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("LoadShortcuts() returned %d shortcuts, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("shortcut %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	path, err := shortcutsFilePath()
	if err != nil {
		t.Fatalf("shortcutsFilePath() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat shortcuts file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("shortcuts file permissions = %o, want 600", perm)
	}
}

func TestLoadShortcutsMissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := LoadShortcuts()
	if err != nil {
		t.Fatalf("LoadShortcuts() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("LoadShortcuts() returned %d shortcuts, want 0", len(got))
	}
}

func TestLoadShortcutsInvalidJSON(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	path := filepath.Join(configDir, "ccd", "shortcuts.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadShortcuts(); err == nil {
		t.Fatal("LoadShortcuts() error = nil, want parse error")
	}
}

func TestUniqueShortcutName(t *testing.T) {
	shortcuts := []Shortcut{
		{Name: "project"},
		{Name: "project-2"},
	}

	if got := UniqueShortcutName(shortcuts, "project"); got != "project-3" {
		t.Errorf("UniqueShortcutName() = %q, want %q", got, "project-3")
	}

	if got := UniqueShortcutName(shortcuts, "other"); got != "other" {
		t.Errorf("UniqueShortcutName() = %q, want %q", got, "other")
	}

	if got := UniqueShortcutName(shortcuts, "  "); got != "shortcut" {
		t.Errorf("UniqueShortcutName() = %q, want %q", got, "shortcut")
	}
}

func TestAddAndRemoveShortcut(t *testing.T) {
	shortcuts := AddShortcut(nil, "", "/tmp/project")
	if len(shortcuts) != 1 {
		t.Fatalf("AddShortcut() returned %d shortcuts, want 1", len(shortcuts))
	}
	if shortcuts[0].Name != "project" {
		t.Errorf("shortcut name = %q, want %q", shortcuts[0].Name, "project")
	}

	shortcuts = AddShortcut(shortcuts, "project", "/tmp/other")
	if shortcuts[1].Name != "project-2" {
		t.Errorf("shortcut name = %q, want %q", shortcuts[1].Name, "project-2")
	}

	shortcuts = RemoveShortcutAt(shortcuts, 0)
	if len(shortcuts) != 1 || shortcuts[0].Name != "project-2" {
		t.Fatalf("RemoveShortcutAt() = %+v, want only project-2", shortcuts)
	}

	shortcuts = RemoveShortcutAt(shortcuts, 5)
	if len(shortcuts) != 1 || shortcuts[0].Name != "project-2" {
		t.Fatalf("RemoveShortcutAt() with invalid index = %+v, want unchanged list", shortcuts)
	}
}

func TestDefaultShortcutName(t *testing.T) {
	tests := map[string]string{
		"/":              "root",
		".":              "root",
		"/home/user/dev": "dev",
		"relative":       "relative",
	}

	for path, want := range tests {
		if got := DefaultShortcutName(path); got != want {
			t.Errorf("DefaultShortcutName(%q) = %q, want %q", path, got, want)
		}
	}
}
