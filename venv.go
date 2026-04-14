package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func pythonExe(venvPath string) string {
	return filepath.Join(venvPath, "Scripts", "python.exe")
}

func isVenv(dir string) bool {
	if _, err := os.Stat(pythonExe(dir)); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "pyvenv.cfg")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "Scripts", "activate.bat")); err == nil {
		return true
	}
	return false
}

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

func activatedEnv(venvPath string) []string {
	env := os.Environ()
	out := make([]string, 0, len(env)+2)
	venvAbs, _ := filepath.Abs(venvPath)
	scripts := filepath.Join(venvAbs, "Scripts")
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			out = append(out, "PATH="+scripts+string(os.PathListSeparator)+e[5:])
			continue
		}
		if strings.HasPrefix(e, "PYTHONHOME=") {
			continue
		}
		out = append(out, e)
	}
	out = append(out, "VIRTUAL_ENV="+venvAbs)
	return out
}
