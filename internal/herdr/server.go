package herdr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Maidang1/hergo/internal/localipc"
)

const (
	defaultServerReadyTimeout = 15 * time.Second
	defaultServerPollInterval = 50 * time.Millisecond
)

// serverReadyTimeout is how long EnsureServer waits for a server it just
// started. Tests shorten it.
var (
	serverReadyTimeout = defaultServerReadyTimeout
	serverPollInterval = defaultServerPollInterval
	startDetached      = startDetachedServer
	socketListening    = listenSocket
)

// ServerOptions describes the Herdr server Hergo may start.
type ServerOptions struct {
	// Binary is the Herdr executable.
	Binary string
	// Target is the socket that server should be listening on.
	Target Target
	// Start is false when the user pointed at a socket or a named session.
	// EnsureServer then leaves that session alone.
	Start bool
}

// OwnsDefaultServer reports whether Hergo may start the default Herdr
// session. An explicit socket, HERDR_SOCKET_PATH, or a named HERDR_SESSION
// selects a session the caller already chose, so Hergo only connects to it.
func OwnsDefaultServer(explicitSocket, socketPath, session string) bool {
	if explicitSocket != "" || socketPath != "" {
		return false
	}
	return session == "" || session == "default"
}

// EnsureServer connects when Target is already listening. When Start is set
// and nothing is listening, it runs `binary server` detached and waits until
// the socket accepts a connection. The process is not stopped when this
// function, or the window that called it, returns.
func EnsureServer(ctx context.Context, opts ServerOptions) error {
	if !opts.Start {
		return nil
	}
	if opts.Binary == "" || opts.Target.Socket == "" {
		return fmt.Errorf("cannot start Herdr without its executable and socket")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if socketListening(ctx, opts.Target.Socket) {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := startDetached(opts.Binary); err != nil {
		return fmt.Errorf("could not start Herdr: %w", err)
	}
	deadline := time.Now().Add(serverReadyTimeout)
	for {
		if socketListening(ctx, opts.Target.Socket) {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("Herdr did not start listening")
		}
		timer := time.NewTimer(serverPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func listenSocket(ctx context.Context, socket string) bool {
	if socket == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	conn, err := localipc.DialContext(ctx, socket)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// startDetachedServer runs `binary server` in its own session, with stdin,
// stdout and stderr discarded. The caller does not wait for it to exit.
func startDetachedServer(binary string) error {
	cmd := exec.Command(binary, "server")
	cmd.Env = serverEnv()
	cmd.SysProcAttr = detachAttr()
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

// serverEnv is the environment of a server Hergo starts. Caller identity and
// socket overrides stay with the parent, so the child uses its default session.
func serverEnv() []string {
	inherited := os.Environ()
	env := make([]string, 0, len(inherited)+1)
	for _, v := range inherited {
		k, _, _ := strings.Cut(v, "=")
		switch k {
		case "HERDR_SOCKET_PATH", "HERDR_SESSION", "HERDR_PANE_ID", "HERDR_TAB_ID", "HERDR_WORKSPACE_ID", "HERDR_ENV":
			continue
		}
		env = append(env, v)
	}
	if dir, err := os.Getwd(); err == nil && dir != "" {
		env = append(env, "HERDR_STARTUP_CWD="+dir)
	}
	return env
}
