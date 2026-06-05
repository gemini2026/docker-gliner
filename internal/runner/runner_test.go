package runner

import (
	"testing"
	"time"
)

// useTempState points the state dir at a per-test temp dir via XDG_STATE_HOME.
func useTempState(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
}

func TestStartStopLifecycle(t *testing.T) {
	useTempState(t)

	pid, err := Start(Spec{
		Project: "proj",
		Service: "svc",
		Argv:    []string{"sleep", "30"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !Alive(pid) {
		t.Fatalf("process %d not alive after Start", pid)
	}

	// Running should report the same pid from the persisted state file.
	got, err := Running("proj", "svc")
	if err != nil {
		t.Fatalf("Running: %v", err)
	}
	if got != pid {
		t.Errorf("Running = %d, want %d", got, pid)
	}

	if err := Stop("proj", "svc", 2*time.Second); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	// Give the kernel a moment to reap.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && Alive(pid) {
		time.Sleep(20 * time.Millisecond)
	}
	if Alive(pid) {
		t.Errorf("process %d still alive after Stop", pid)
	}

	// State file removed -> Running reports nothing.
	if got, _ := Running("proj", "svc"); got != 0 {
		t.Errorf("Running after Stop = %d, want 0", got)
	}
}

func TestStopNoInstance(t *testing.T) {
	useTempState(t)
	// Stopping a service that was never started is a no-op success.
	if err := Stop("proj", "missing", time.Second); err != nil {
		t.Errorf("Stop on missing instance: %v", err)
	}
}

func TestRunningStaleState(t *testing.T) {
	useTempState(t)
	// A pid file pointing at a long-dead pid should be cleaned and report 0.
	if err := writePidFile("proj", "svc", 999999); err != nil {
		t.Fatalf("writePidFile: %v", err)
	}
	got, err := Running("proj", "svc")
	if err != nil {
		t.Fatalf("Running: %v", err)
	}
	if got != 0 {
		t.Errorf("Running with stale pid = %d, want 0", got)
	}
}

func TestStartEmptyArgv(t *testing.T) {
	useTempState(t)
	if _, err := Start(Spec{Project: "p", Service: "s"}); err == nil {
		t.Error("expected error for empty argv")
	}
}
