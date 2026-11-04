package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Maidang1/hergo/internal/herdr"
)

func TestStartupPlain(t *testing.T) {
	installed := herdr.Installation{Path: "/usr/local/bin/herdr", Version: "herdr 0.9.5"}
	missing := herdr.Installation{}
	broken := herdr.Installation{Path: "/usr/local/bin/herdr", Error: "exit 1"}
	cases := []struct {
		mode         string
		plain, herdr bool
		install      herdr.Installation
		want         bool
	}{
		{"auto", false, false, installed, false},
		// A missing Herdr opens the setup window, which can install it.
		{"auto", false, false, missing, false},
		// A Herdr that does not answer still opens its window, which
		// explains the failure and offers Retry.
		{"auto", false, false, broken, false},
		{"plain", false, false, installed, true},
		{"herdr", false, false, missing, false},
		{"plain", false, true, installed, false},
		{"herdr", true, false, installed, true},
	}
	for _, c := range cases {
		if got := startupPlain(c.mode, c.plain, c.herdr, c.install); got != c.want {
			t.Fatalf("%+v: %v", c, got)
		}
	}
}

func TestInstallationWatchUse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in executable is a shell script")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "herdr")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho herdr 0.9.5\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	w := newInstallationWatch("missing-herdr-name", herdr.Installation{})
	w.Use(bin)
	found := w.Refresh(context.Background())
	if !found.Installed() || found.Version != "herdr 0.9.5" {
		t.Fatal(found)
	}
}

func TestPlainAccelerator(t *testing.T) {
	want := "Ctrl+Shift+T"
	if isDarwin {
		want = "Cmd+T"
	}
	if got := plainAccelerator("T"); got != want {
		t.Fatal(got)
	}
	if !isDarwin && plainAccelerator("Shift+]") != "Ctrl+Shift+]" {
		t.Fatal(plainAccelerator("Shift+]"))
	}
}
