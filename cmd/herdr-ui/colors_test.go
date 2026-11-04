package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Maidang1/herdr-ui/internal/design"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

func TestMonokaiProTerminalColors(t *testing.T) {
	for _, tc := range []struct {
		name, background, foreground, sidebar, row, selection string
		dark                                                  bool
		ansi                                                  [16]string
	}{
		{
			"monokai-pro", "#2d2a2e", "#fcfcfa", "#221f22", "#3a373a", "#434144", true,
			[16]string{
				"#221f22", "#ff6188", "#a9dc76", "#ffd866", "#fc9867", "#ab9df2", "#78dce8", "#fcfcfa",
				"#727072", "#ff6188", "#a9dc76", "#ffd866", "#fc9867", "#ab9df2", "#78dce8", "#fcfcfa",
			},
		},
		{
			"monokai-pro-light", "#faf4f2", "#29242a", "#ede7e5", "#d7d2d0", "#e5dfde", false,
			[16]string{
				"#ede7e5", "#e14775", "#269d69", "#cc7a0a", "#e16032", "#7058be", "#1c8ca8", "#29242a",
				"#a59fa0", "#e14775", "#269d69", "#cc7a0a", "#e16032", "#7058be", "#1c8ca8", "#29242a",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			light, dark, err := design.HerdrThemes("", tc.name)
			if err != nil {
				t.Fatal(err)
			}
			for _, resolved := range []*design.HerdrTheme{light, dark} {
				if resolved.Theme.Dark != tc.dark || resolved.Colors.Shell != ui.Hex(tc.background) ||
					resolved.Colors.Sidebar != ui.Hex(tc.sidebar) || resolved.Colors.RowSelected != ui.Hex(tc.row) {
					t.Fatalf("incorrect Monokai Pro surfaces: %+v", resolved.Colors)
				}
				colors := terminalColorTheme(resolved)
				if colors.Background != ui.Hex(tc.background) || colors.Foreground != ui.Hex(tc.foreground) ||
					colors.Selection != ui.Hex(tc.selection) {
					t.Fatalf("incorrect Monokai Pro terminal canvas: %+v", colors)
				}
				for i, want := range tc.ansi {
					if colors.Palette[i] != ui.Hex(want) {
						t.Errorf("ANSI %d = %v, want %s", i, colors.Palette[i], want)
					}
				}
			}
		})
	}
	other, _, err := design.HerdrThemes("", "gruvbox")
	if err != nil {
		t.Fatal(err)
	}
	if got := terminalColorTheme(other); got.Palette != terminal.DarkTheme().Palette {
		t.Fatal("switching away from Monokai Pro retained its ANSI palette")
	}
}

func TestWatchHerdrColorsAtomicSaveDeleteAndCancel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[theme]"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changes := make(chan struct{}, 8)
	done := make(chan struct{})
	go func() {
		defer close(done)
		watchHerdrColors(ctx, path, 5*time.Millisecond, func() { changes <- struct{}{} })
	}()
	next := func() {
		t.Helper()
		select {
		case <-changes:
		case <-time.After(2 * time.Second):
			t.Fatal("configuration change was not detected")
		}
	}
	next()
	replacement := path + ".tmp"
	if err := os.WriteFile(replacement, []byte("[theme]\nname = \"gruvbox\""), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	next()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	next()
	if err := os.WriteFile(path, []byte("[theme]\nname = \"nord\""), 0600); err != nil {
		t.Fatal(err)
	}
	next()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop")
	}
}
