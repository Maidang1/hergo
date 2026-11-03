package herdr

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// installScriptURL is Herdr's own Unix installer. It installs the
	// classic herdr release, the program Hergo speaks to.
	installScriptURL = "https://herdr.dev/install.sh"
	// installPowerShellURL is the same installer for Windows.
	installPowerShellURL = "https://herdr.dev/install.ps1"
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
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(string(out))
	if len(detail) > 500 {
		detail = detail[len(detail)-500:]
	}
	if detail == "" {
		return fmt.Errorf("could not install Herdr: %w", err)
	}
	return fmt.Errorf("could not install Herdr: %s", detail)
}
