// Package runner starts and stops the native host process behind a provider
// service. Because Compose invokes the provider binary once per lifecycle event
// (a fresh process for `up`, another for `down`), the running child cannot be
// held in memory across calls. We instead persist its PID to a state file keyed
// by project+service, so a later `down` can find and stop it.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// stateDir returns the directory used to track running host processes. It
// honors XDG_STATE_HOME, falling back to ~/.local/state, then the OS temp dir.
func stateDir() string {
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, "docker-gliner")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "state", "docker-gliner")
	}
	return filepath.Join(os.TempDir(), "docker-gliner")
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, s)
}

// instanceKey is the stable per-service identity used to name state files.
func instanceKey(project, service string) string {
	return sanitize(project) + "__" + sanitize(service)
}

func pidFilePath(project, service string) string {
	return filepath.Join(stateDir(), instanceKey(project, service)+".pid")
}

// LogFilePath returns where a service's host process writes its stdout/stderr.
func LogFilePath(project, service string) string {
	return filepath.Join(stateDir(), instanceKey(project, service)+".log")
}

func writePidFile(project, service string, pid int) error {
	if err := os.MkdirAll(stateDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(pidFilePath(project, service), []byte(strconv.Itoa(pid)), 0o644)
}

// readPidFile returns the recorded PID, or (0, nil) if no state file exists.
func readPidFile(project, service string) (int, error) {
	b, err := os.ReadFile(pidFilePath(project, service))
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, fmt.Errorf("corrupt pid file %s: %w", pidFilePath(project, service), err)
	}
	return pid, nil
}

func removePidFile(project, service string) error {
	err := os.Remove(pidFilePath(project, service))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
