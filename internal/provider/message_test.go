package provider

import (
	"strings"
	"testing"
)

func TestEmitterMessages(t *testing.T) {
	var buf strings.Builder
	em := NewEmitter(&buf)
	em.Info("preparing")
	em.Setenv("ENDPOINT", "http://localhost:8100")
	em.Error("boom")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), buf.String())
	}
	want := []string{
		`{"type":"info","message":"preparing"}`,
		`{"type":"setenv","message":"ENDPOINT=http://localhost:8100"}`,
		`{"type":"error","message":"boom"}`,
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
}
