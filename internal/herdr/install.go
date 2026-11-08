package herdr

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// installScriptURL is Herdr's own Unix installer. It installs the
	// classic herdr release, the program Hergo speaks to.
	installScriptURL = "https://herdr.dev/install.sh"
	// installPowerShellURL is the same installer for Windows.
	installPowerShellURL = "https://herdr.dev/install.ps1"
	// CompatibleVersion is the oldest Herdr Hergo is validated against. The
	// endpoint behind image paste, commands and notifications is missing
	// before 0.9.0, and 0.9.5 is the release the integration suite runs.
	CompatibleVersion = "0.9.5"
	// updateTimeout bounds a download or a Homebrew update.
	updateTimeout = 10 * time.Minute
)

// ErrSetupCanceled is returned when the user closes the file dialog without
// choosing a Herdr executable.
var ErrSetupCanceled = errors.New("setup canceled")

// InstallCommand is the official installer for goos. It does not run it.
func InstallCommand(goos string) (string, []string) {
	if goos == "windows" {
		return "powershell", []string{"-NoProfile", "-Command", "irm " + installPowerShellURL + " | iex"}
	}
	return "sh", []string{"-c", "curl -fsSL " + installScriptURL + " | sh"}
}

// Install runs Herdr's official installer. Call it only after the user asks
// to install Herdr.
func Install(ctx context.Context) error {
	name, args := InstallCommand(runtime.GOOS)
	return runSetup(exec.CommandContext(ctx, name, args...), "install")
}

// VersionOutdated reports whether version, a `herdr --version` line or a bare
// version, is older than CompatibleVersion. A version it cannot read is not
// called outdated, so an unexpected format never asks for an update.
func VersionOutdated(version string) bool {
	have, ok := parseVersion(version)
	want, _ := parseVersion(CompatibleVersion)
	if !ok {
		return false
	}
	for i := range have {
		if have[i] != want[i] {
			return have[i] < want[i]
		}
	}
	return false
}

func parseVersion(version string) ([3]int, bool) {
	var parsed [3]int
	for _, field := range strings.Fields(version) {
		field = strings.TrimPrefix(field, "v")
		field, _, _ = strings.Cut(field, "-")
		parts := strings.Split(field, ".")
		if len(parts) != len(parsed) {
			continue
		}
		ok := true
		for i, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil {
				ok = false
				break
			}
			parsed[i] = n
		}
		if ok {
			return parsed, true
		}
	}
	return parsed, false
}

// homebrewPrefix returns the Homebrew prefix owning path, an executable with
// its symlinks resolved: <prefix>/Cellar/herdr/<version>/bin/herdr.
func homebrewPrefix(path string) (string, bool) {
	formula := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	cellar := filepath.Dir(formula)
	if filepath.Base(formula) != "herdr" || filepath.Base(cellar) != "Cellar" {
		return "", false
	}
	return filepath.Dir(cellar), true
}

// UpdateCommand is how goos updates the Herdr at path, with symlinks resolved.
// It does not run it. Herdr refuses to update itself under Homebrew. Elsewhere
// on Unix, --handoff is required: without a terminal a plain `herdr update`
// exits successfully without installing while a server runs. Windows has no
// live handoff. The bool says the caller must hand off running sessions itself.
func UpdateCommand(goos, path string) (string, []string, bool) {
	if goos == "windows" {
		return path, []string{"update"}, false
	}
	if prefix, ok := homebrewPrefix(path); ok {
		return "sh", []string{"-c", `"$0" update && "$0" upgrade herdr`, filepath.Join(prefix, "bin", "brew")}, true
	}
	return path, []string{"update", "--handoff"}, false
}

// Update replaces the Herdr at path with the latest release, the way its
// installation calls for. Call it only after the user asks to update Herdr.
func Update(ctx context.Context, target Target, path string) error {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	name, args, handoff := UpdateCommand(runtime.GOOS, resolved)
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	// Herdr refuses to update from inside one of its own sessions.
	cmd.Env = serverEnv()
	if err := runSetup(cmd, "update"); err != nil {
		return err
	}
	if handoff && socketListening(ctx, target.Socket) {
		if err := LiveHandoff(ctx, target, path); err != nil {
			return fmt.Errorf("Herdr updated, but running sessions still use the old version: %w", err)
		}
	}
	return nil
}

func runSetup(cmd *exec.Cmd, action string) error {
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(string(out))
	if len(detail) > 500 {
		detail = detail[len(detail)-500:]
	}
	if detail == "" {
		return fmt.Errorf("could not %s Herdr: %w", action, err)
	}
	return fmt.Errorf("could not %s Herdr: %s", action, detail)
}
