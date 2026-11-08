package herdr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("HERGO_FAKE_SERVER") == "1" {
		marker := os.Getenv("HERGO_FAKE_MARKER")
		if marker == "" {
			os.Exit(2)
		}
		if err := os.WriteFile(marker, []byte("ok"), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestOwnsDefaultServer(t *testing.T) {
	if !OwnsDefaultServer("", "", "") || !OwnsDefaultServer("", "", "default") {
		t.Fatal("default session")
	}
	if OwnsDefaultServer("/tmp/herdr.sock", "", "") || OwnsDefaultServer("", "/tmp/herdr.sock", "") || OwnsDefaultServer("", "", "work") {
		t.Fatal("explicit session must be left alone")
	}
}

func TestEnsureServerLeavesAnExplicitSessionAlone(t *testing.T) {
	called := false
	restoreServerFns(t)
	startDetached = func(string) error { called = true; return nil }
	socketListening = func(context.Context, string) bool { return false }
	if err := EnsureServer(context.Background(), ServerOptions{Start: false, Binary: "herdr", Target: Target{Socket: "/tmp/custom.sock"}}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("started a server for a session the caller chose")
	}
}

func TestEnsureServerDoesNotStartASecondServer(t *testing.T) {
	called := false
	restoreServerFns(t)
	startDetached = func(string) error { called = true; return nil }
	socketListening = func(context.Context, string) bool { return true }
	if err := EnsureServer(context.Background(), ServerOptions{Start: true, Binary: "herdr", Target: Target{Socket: "/tmp/herdr.sock"}}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("started a server that was already listening")
	}
}

func TestEnsureServerWaitsUntilTheSocketListens(t *testing.T) {
	listening := false
	restoreServerFns(t)
	startDetached = func(binary string) error {
		if binary != "/usr/local/bin/herdr" {
			t.Errorf("binary %s", binary)
		}
		listening = true
		return nil
	}
	socketListening = func(context.Context, string) bool { return listening }
	if err := EnsureServer(context.Background(), ServerOptions{Start: true, Binary: "/usr/local/bin/herdr", Target: Target{Socket: "/tmp/herdr.sock"}}); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureServerReportsAServerThatNeverListens(t *testing.T) {
	restoreServerFns(t)
	serverReadyTimeout = 80 * time.Millisecond
	serverPollInterval = 20 * time.Millisecond
	startDetached = func(string) error { return nil }
	socketListening = func(context.Context, string) bool { return false }
	err := EnsureServer(context.Background(), ServerOptions{Start: true, Binary: "herdr", Target: Target{Socket: "/tmp/herdr.sock"}})
	if err == nil || err.Error() != "Herdr did not start listening" {
		t.Fatal(err)
	}
}

func TestEnsureServerDoesNotStartAfterCancel(t *testing.T) {
	called := false
	restoreServerFns(t)
	startDetached = func(string) error { called = true; return nil }
	socketListening = func(context.Context, string) bool { return false }
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := EnsureServer(ctx, ServerOptions{Start: true, Binary: "herdr", Target: Target{Socket: "/tmp/herdr.sock"}})
	if err == nil || called {
		t.Fatal(err, called)
	}
}

func TestEnsureServerReportsAStartFailure(t *testing.T) {
	restoreServerFns(t)
	startDetached = func(string) error { return fmt.Errorf("exec: no such file") }
	socketListening = func(context.Context, string) bool { return false }
	err := EnsureServer(context.Background(), ServerOptions{Start: true, Binary: "herdr", Target: Target{Socket: "/tmp/herdr.sock"}})
	if err == nil || err.Error() != "could not start Herdr: exec: no such file" {
		t.Fatal(err)
	}
}

func TestStartDetachedServerRunsTheBinary(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	t.Setenv("HERGO_FAKE_SERVER", "1")
	t.Setenv("HERGO_FAKE_MARKER", marker)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := startDetachedServer(exe); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("detached server did not run")
}

func TestInstallCommandUsesTheClassicInstaller(t *testing.T) {
	name, args := InstallCommand("darwin")
	if name != "sh" || len(args) != 2 || args[0] != "-c" {
		t.Fatalf("darwin %s %v", name, args)
	}
	if !strings.Contains(args[1], "https://herdr.dev/install.sh") || strings.Contains(args[1], "releases.herdr.dev") {
		t.Fatal(args[1])
	}
	name, args = InstallCommand("linux")
	if name != "sh" || !strings.Contains(args[1], "https://herdr.dev/install.sh") {
		t.Fatalf("linux %s %v", name, args)
	}
	name, args = InstallCommand("windows")
	joined := name + " " + strings.Join(args, " ")
	if name != "powershell" || !strings.Contains(joined, "https://herdr.dev/install.ps1") || strings.Contains(joined, "releases.herdr.dev") {
		t.Fatal(joined)
	}
}

func TestVersionOutdatedComparesAgainstCompatibleVersion(t *testing.T) {
	for version, want := range map[string]bool{
		"herdr 0.8.0":            true,
		"0.8.0":                  true,
		"herdr 0.9.4":            true,
		"herdr 0.9.5":            false,
		"herdr 0.9.10":           false,
		"herdr 0.10.0-preview.3": false,
		"v1.0.0":                 false,
		"":                       false,
		"herdr nightly":          false,
	} {
		if got := VersionOutdated(version); got != want {
			t.Errorf("VersionOutdated(%q) = %v", version, got)
		}
	}
	if (Installation{Path: "/bin/herdr", Version: "herdr 0.8.0", Error: "exit 1"}).Outdated() {
		t.Fatal("a broken executable is not an outdated one")
	}
}

func TestUpdateCommandFollowsTheInstallation(t *testing.T) {
	direct := filepath.Join(string(filepath.Separator)+"home", "a", ".local", "bin", "herdr")
	for _, goos := range []string{"darwin", "linux"} {
		name, args, handoff := UpdateCommand(goos, direct)
		if name != direct || strings.Join(args, " ") != "update --handoff" || handoff {
			t.Fatalf("%s direct: %s %v %v", goos, name, args, handoff)
		}
	}
	name, args, handoff := UpdateCommand("windows", direct)
	if name != direct || strings.Join(args, " ") != "update" || handoff {
		t.Fatalf("windows: %s %v %v", name, args, handoff)
	}
	prefix := filepath.Join(string(filepath.Separator)+"opt", "homebrew")
	cellar := filepath.Join(prefix, "Cellar", "herdr", "0.8.0", "bin", "herdr")
	name, args, handoff = UpdateCommand("darwin", cellar)
	if name != "sh" || len(args) != 3 || !strings.Contains(args[1], "upgrade herdr") || args[2] != filepath.Join(prefix, "bin", "brew") || !handoff {
		t.Fatalf("homebrew: %s %v %v", name, args, handoff)
	}
}

func TestCandidateDirsIncludeOfficialInstallLocations(t *testing.T) {
	t.Setenv("LOCALAPPDATA", filepath.Join("C:", "Users", "A", "AppData", "Local"))
	home := filepath.Join("C:", "Users", "A")
	unix := candidateDirs(home, "linux")
	if !containsPath(unix, filepath.Join(home, ".local", "bin")) {
		t.Fatal(unix)
	}
	windows := candidateDirs(home, "windows")
	if !containsPath(windows, filepath.Join("C:", "Users", "A", "AppData", "Local", "Programs", "Herdr", "bin")) {
		t.Fatal(windows)
	}
}

func restoreServerFns(t *testing.T) {
	t.Helper()
	timeout, poll := serverReadyTimeout, serverPollInterval
	start, listen := startDetached, socketListening
	t.Cleanup(func() {
		serverReadyTimeout, serverPollInterval = timeout, poll
		startDetached, socketListening = start, listen
	})
}

func containsPath(dirs []string, want string) bool {
	for _, dir := range dirs {
		if dir == want {
			return true
		}
	}
	return false
}
