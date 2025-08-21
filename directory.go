package main

import (
	"log"
	"os"
	"strings"
)

type DirItem struct {
	isDir    bool
	isHidden bool
	name     string
}

func List(dir string, listHidden bool) []DirItem {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var items []DirItem

	for _, e := range entries {
		if !listHidden && e.Name()[0] == '.' {
			continue
		}

		if !e.IsDir() {
			continue
		}

		items = append(items, DirItem{name: e.Name(), isDir: true, isHidden: e.Name()[0] == '.'})
	}

	return items
}

func PopPath(dir string) string {
	parts := strings.Split(dir, "/")

	if len(parts) == 2 {
		return "/"
	}

	parts = parts[:len(parts)-1]

	return strings.Join(parts, "/")
}

func PushPath(dir string, path string) string {
	prefix := "/"
	if dir == "/" {
		prefix = ""
	}

	return dir + prefix + path
}
