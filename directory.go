package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type DirItem struct {
	isDir bool
	name  string
}

func List(dir string) []DirItem {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var items []DirItem

	for _, e := range entries {
		if e.IsDir() {
			items = append(items, DirItem{name: e.Name(), isDir: true})
		} else {
			// items = append(items, DirItem{name: e.Name()})
		}
	}

	return items
}

func PopPath(dir string) string {
	parts := strings.Split(dir, "/")

	parts = parts[:len(parts)-1]

	return strings.Join(parts, "/")
}

func PushPath(dir string, path string) string {
	return dir + fmt.Sprintf("/%s", path)
}
