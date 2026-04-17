package main

import (
	"runtime"
	"testing"
)

func TestCanonicalPathWindowsCaseInsensitive(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only path normalization test")
	}

	got1 := canonicalPath(`C:\Users\Sanghee\proj\yoloe`)
	got2 := canonicalPath(`c:/Users/Sanghee/proj/yoloe`)
	if got1 != got2 {
		t.Fatalf("expected canonical paths to match, got %q != %q", got1, got2)
	}
}

func TestConfigFindByPathUsesCanonicalPath(t *testing.T) {
	cfg := &Config{Venvs: []Venv{{Alias: "demo", Path: "."}}}
	if got := cfg.FindByPath("./"); got == nil || got.Alias != "demo" {
		t.Fatalf("expected to find venv by equivalent path, got %#v", got)
	}
}
