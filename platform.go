package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, _ := os.UserHomeDir()
		if runtime.GOOS == "windows" {
			dir = filepath.Join(home, "AppData", "Roaming")
		} else {
			dir = filepath.Join(home, ".config")
		}
	}
	return filepath.Join(dir, "pvm", "config.json")
}

func venvBinDir(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts")
	}
	return filepath.Join(venvPath, "bin")
}

func pythonExe(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "python.exe")
	}
	return filepath.Join(venvPath, "bin", "python")
}

func activationScript(venvPath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "activate.bat")
	}
	return filepath.Join(venvPath, "bin", "activate")
}

func isVenv(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "pyvenv.cfg")); err != nil {
		return false
	}
	if _, err := os.Stat(pythonExe(dir)); err == nil {
		return true
	}
	if runtime.GOOS != "windows" {
		if _, err := os.Stat(filepath.Join(dir, "bin", "python3")); err == nil {
			return true
		}
	}
	if _, err := os.Stat(activationScript(dir)); err == nil {
		return true
	}
	return false
}

func activatedEnv(venvPath string) []string {
	env := os.Environ()
	out := make([]string, 0, len(env)+2)
	venvAbs, _ := filepath.Abs(venvPath)
	binDir := venvBinDir(venvAbs)
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			out = append(out, "PATH="+binDir+string(os.PathListSeparator)+e[5:])
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

func shellCommand(venvPath string) (*exec.Cmd, error) {
	projectDir := filepath.Dir(venvPath)
	activate := activationScript(venvPath)

	if runtime.GOOS == "windows" {
		c := exec.Command("cmd", "/K", activate)
		c.Dir = projectDir
		c.Env = activatedEnv(venvPath)
		return c, nil
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	home, _ := os.UserHomeDir()
	base := filepath.Base(shell)

	switch base {
	case "zsh":
		tmpDir, err := os.MkdirTemp("", "pvm-zsh-*")
		if err != nil {
			return nil, err
		}
		origZdotdir := os.Getenv("ZDOTDIR")
		if origZdotdir == "" {
			origZdotdir = home
		}
		zshenv := fmt.Sprintf("[ -f %q/.zshenv ] && . %q/.zshenv\n", origZdotdir, origZdotdir)
		if err := os.WriteFile(filepath.Join(tmpDir, ".zshenv"), []byte(zshenv), 0644); err != nil {
			return nil, err
		}
		zshrc := fmt.Sprintf("export ZDOTDIR=%q\n[ -f %q/.zshrc ] && . %q/.zshrc\nsource %q\ncd %q\n", origZdotdir, origZdotdir, origZdotdir, activate, projectDir)
		if err := os.WriteFile(filepath.Join(tmpDir, ".zshrc"), []byte(zshrc), 0644); err != nil {
			return nil, err
		}
		c := exec.Command(shell, "-i")
		env := activatedEnv(venvPath)
		filtered := make([]string, 0, len(env)+1)
		for _, e := range env {
			if !strings.HasPrefix(e, "ZDOTDIR=") {
				filtered = append(filtered, e)
			}
		}
		filtered = append(filtered, "ZDOTDIR="+tmpDir)
		c.Env = filtered
		c.Dir = projectDir
		return c, nil
	case "bash":
		tmpFile, err := os.CreateTemp("", "pvm-bashrc-*")
		if err != nil {
			return nil, err
		}
		content := fmt.Sprintf("[ -f %q/.bashrc ] && . %q/.bashrc\nsource %q\ncd %q\n", home, home, activate, projectDir)
		if _, err := tmpFile.WriteString(content); err != nil {
			return nil, err
		}
		if err := tmpFile.Close(); err != nil {
			return nil, err
		}
		c := exec.Command(shell, "--rcfile", tmpFile.Name(), "-i")
		c.Env = activatedEnv(venvPath)
		c.Dir = projectDir
		return c, nil
	default:
		c := exec.Command(shell, "-i")
		c.Env = activatedEnv(venvPath)
		c.Dir = projectDir
		return c, nil
	}
}

func commandFromString(line string, venvPath string) (*exec.Cmd, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}
	if runtime.GOOS == "windows" {
		c := exec.Command("cmd", "/C", line)
		c.Env = activatedEnv(venvPath)
		return c, nil
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	c := exec.Command(shell, "-lc", line)
	c.Env = activatedEnv(venvPath)
	return c, nil
}

func commandFromArgs(args []string, venvPath string) (*exec.Cmd, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command given")
	}
	env := activatedEnv(venvPath)
	name := args[0]
	if !strings.ContainsRune(name, os.PathSeparator) {
		binDir := venvBinDir(venvPath)
		candidate := filepath.Join(binDir, name)
		if _, err := os.Stat(candidate); err == nil {
			name = candidate
		}
	}
	c := exec.Command(name, args[1:]...)
	c.Env = env
	return c, nil
}
