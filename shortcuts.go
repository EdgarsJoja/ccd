package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Shortcut struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type shortcutFile struct {
	Shortcuts []Shortcut `json:"shortcuts"`
}

func shortcutsFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "ccd", "shortcuts.json"), nil
}

func LoadShortcuts() ([]Shortcut, error) {
	path, err := shortcutsFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var file shortcutFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse shortcuts: %w", err)
	}

	for i := range file.Shortcuts {
		file.Shortcuts[i].Path = filepath.Clean(file.Shortcuts[i].Path)
	}

	return file.Shortcuts, nil
}

func SaveShortcuts(shortcuts []Shortcut) error {
	path, err := shortcutsFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	file := shortcutFile{Shortcuts: shortcuts}
	if file.Shortcuts == nil {
		file.Shortcuts = []Shortcut{}
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "shortcuts-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func AddShortcut(shortcuts []Shortcut, name string, path string) []Shortcut {
	name = strings.TrimSpace(name)
	if name == "" {
		name = DefaultShortcutName(path)
	}
	name = UniqueShortcutName(shortcuts, name)

	updated := make([]Shortcut, len(shortcuts), len(shortcuts)+1)
	copy(updated, shortcuts)

	return append(updated, Shortcut{Name: name, Path: filepath.Clean(path)})
}

func RemoveShortcutAt(shortcuts []Shortcut, index int) []Shortcut {
	if index < 0 || index >= len(shortcuts) {
		return shortcuts
	}

	updated := make([]Shortcut, 0, len(shortcuts)-1)
	updated = append(updated, shortcuts[:index]...)

	return append(updated, shortcuts[index+1:]...)
}

func UniqueShortcutName(shortcuts []Shortcut, base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "shortcut"
	}

	used := make(map[string]struct{}, len(shortcuts))
	for _, shortcut := range shortcuts {
		used[shortcut.Name] = struct{}{}
	}

	if _, exists := used[base]; !exists {
		return base
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func DefaultShortcutName(path string) string {
	base := filepath.Base(filepath.Clean(path))
	if base == "/" || base == "." || base == "" {
		return "root"
	}

	return base
}
