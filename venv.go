package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func pythonVersion(venvPath string) string {
	out, err := exec.Command(pythonExe(venvPath), "--version").CombinedOutput()
	if err != nil {
		return "?"
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(out)), "Python "))
}

func scanDir(root string, maxDepth int) []string {
	var found []string
	root, _ = filepath.Abs(root)
	rootDepth := strings.Count(root, string(os.PathSeparator))
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		if depth > maxDepth {
			return filepath.SkipDir
		}
		name := d.Name()
		if depth > 0 && (name == "node_modules" || name == ".git") {
			return filepath.SkipDir
		}
		if isVenv(path) {
			found = append(found, path)
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

func defaultAlias(path string) string {
	abs, _ := filepath.Abs(path)
	base := filepath.Base(abs)
	if base == ".venv" || base == "venv" || base == "env" {
		return filepath.Base(filepath.Dir(abs))
	}
	return base
}
