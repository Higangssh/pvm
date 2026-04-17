package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Venv struct {
	Alias       string              `json:"alias"`
	Path        string              `json:"path"`
	Commands    map[string]string   `json:"commands,omitempty"`
	CommandArgs map[string][]string `json:"command_args,omitempty"`
}

type Config struct {
	Venvs []Venv `json:"venvs"`
}

func loadConfig() (*Config, error) {
	p := configPath()
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) Save() error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

func (c *Config) Find(alias string) (*Venv, int) {
	for i := range c.Venvs {
		if c.Venvs[i].Alias == alias {
			return &c.Venvs[i], i
		}
	}
	return nil, -1
}

func canonicalPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if runtime.GOOS == "windows" {
		abs = strings.ToLower(abs)
	}
	return abs
}

func (c *Config) FindByPath(path string) *Venv {
	canon := canonicalPath(path)
	for i := range c.Venvs {
		if canonicalPath(c.Venvs[i].Path) == canon {
			return &c.Venvs[i]
		}
	}
	return nil
}

func (c *Config) Add(v Venv) error {
	if existing, _ := c.Find(v.Alias); existing != nil {
		return fmt.Errorf("alias %q already exists", v.Alias)
	}
	if existing := c.FindByPath(v.Path); existing != nil {
		return fmt.Errorf("path already registered as %q", existing.Alias)
	}
	c.Venvs = append(c.Venvs, v)
	return nil
}
