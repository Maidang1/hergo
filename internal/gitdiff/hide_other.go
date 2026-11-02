//go:build !windows

package gitdiff

import "os/exec"

func hideCommand(*exec.Cmd) {}
