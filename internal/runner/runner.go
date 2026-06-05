package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Spec describes a host process to manage on behalf of a provider service.
type Spec struct {
	Project string   // Compose project name
	Service string   // Compose service name
	Argv    []string // command + args to exec (Argv[0] resolved via $PATH)
	Env     []string // extra KEY=VALUE entries appended to the parent env
}

// Alive reports whether a process with the given pid currently exists.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, signal 0 probes for existence without delivering a signal:
	// nil => alive, EPERM => alive but not ours, ESRCH => gone.
	err = proc.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

// Running returns the PID of an already-running instance for this spec, or 0.
// A stale pid file (process gone) is cleaned up and reported as not running.
func Running(project, service string) (int, error) {
	pid, err := readPidFile(project, service)
	if err != nil {
		return 0, err
	}
	if pid == 0 {
		return 0, nil
	}
	if Alive(pid) {
		return pid, nil
	}
	_ = removePidFile(project, service)
	return 0, nil
}

// Start launches the spec's command as a detached host process in its own
// process group, redirecting its output to the service log file, and records
// its PID. The child outlives this (the provider `up`) process.
func Start(spec Spec) (int, error) {
	if len(spec.Argv) == 0 {
		return 0, fmt.Errorf("empty command")
	}

	logPath := LogFilePath(spec.Project, spec.Service)
	if err := os.MkdirAll(stateDir(), 0o755); err != nil {
		return 0, err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	cmd := exec.Command(spec.Argv[0], spec.Argv[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(), spec.Env...)
	// New process group so the child is detached from us and can be signalled
	// as a group on Stop (negative PID kill).
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start %q: %w", spec.Argv[0], err)
	}
	pid := cmd.Process.Pid

	// Release the child so it is not reaped when this process exits.
	if err := cmd.Process.Release(); err != nil {
		return pid, fmt.Errorf("release child: %w", err)
	}
	if err := writePidFile(spec.Project, spec.Service, pid); err != nil {
		return pid, fmt.Errorf("record pid: %w", err)
	}
	return pid, nil
}

// Stop terminates the host process recorded for this service: SIGTERM to the
// process group, then SIGKILL if it has not exited within grace. A missing or
// already-dead process is treated as success. The pid file is always removed.
func Stop(project, service string, grace time.Duration) error {
	pid, err := readPidFile(project, service)
	if err != nil {
		return err
	}
	defer removePidFile(project, service)

	if pid == 0 || !Alive(pid) {
		return nil
	}

	// Negative pid signals the whole process group created with Setpgid.
	_ = syscall.Kill(-pid, syscall.SIGTERM)

	deadline := time.Now().Add(grace)
	for time.Now().Before(deadline) {
		reapIfChild(pid)
		if !Alive(pid) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	_ = syscall.Kill(-pid, syscall.SIGKILL)
	// Give the kernel a moment, then reap if the dead process is our child so
	// it does not linger as a zombie (relevant when the caller forked it).
	time.Sleep(50 * time.Millisecond)
	reapIfChild(pid)
	return nil
}

// reapIfChild non-blockingly reaps pid if it is a child of this process. It is
// a no-op (ECHILD) when called from a different process than the one that
// forked the child, e.g. a separate `down` invocation.
func reapIfChild(pid int) {
	var ws syscall.WaitStatus
	_, _ = syscall.Wait4(pid, &ws, syscall.WNOHANG, nil)
}
