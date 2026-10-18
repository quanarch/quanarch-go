// Package tools finds the shop's root for the commands in cmd: artifacts, register, build and start.
package tools

import (
	"os"
	"path/filepath"
)

// Root is the shop's folder: the one with go.work, at or above the working directory.
func Root() string {
	dir, _ := os.Getwd()
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.work")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			return dir
		}
	}
}

// Quans lists the quans' folders, by name.
func Quans(root string) []string {
	entries, _ := os.ReadDir(filepath.Join(root, "quans"))
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// Env reads a variable, or falls back.
func Env(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
