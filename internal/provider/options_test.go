package provider

import "testing"

func TestParseUp(t *testing.T) {
	// The exact shape Compose uses (PROTOCOL.md §1).
	args := []string{
		"compose", "--project-name", "myproj", "up",
		"--model=fastino/gliner2-base-v1", "--device=auto", "ner",
	}
	req, err := Parse(args)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if req.Command != CmdUp {
		t.Errorf("command = %q, want up", req.Command)
	}
	if req.ProjectName != "myproj" {
		t.Errorf("project = %q, want myproj", req.ProjectName)
	}
	if req.Service != "ner" {
		t.Errorf("service = %q, want ner", req.Service)
	}
	if got := req.Opt("model", ""); got != "fastino/gliner2-base-v1" {
		t.Errorf("model = %q", got)
	}
	if got := req.Opt("device", ""); got != "auto" {
		t.Errorf("device = %q", got)
	}
}

func TestParseDown(t *testing.T) {
	req, err := Parse([]string{"compose", "--project-name", "p", "down", "svc"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if req.Command != CmdDown {
		t.Errorf("command = %q, want down", req.Command)
	}
	if req.Service != "svc" {
		t.Errorf("service = %q, want svc", req.Service)
	}
}

func TestParseSpaceSeparatedFlag(t *testing.T) {
	req, err := Parse([]string{"compose", "--project-name", "p", "up", "--port", "9000", "svc"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := req.Opt("port", ""); got != "9000" {
		t.Errorf("port = %q, want 9000", got)
	}
	if req.Service != "svc" {
		t.Errorf("service = %q, want svc", req.Service)
	}
}

func TestParseErrors(t *testing.T) {
	cases := map[string][]string{
		"missing subcommand": {"compose", "--project-name", "p", "svc"},
		"missing service":    {"compose", "--project-name", "p", "up"},
		"extra positionals":  {"compose", "up", "a", "b"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(args); err == nil {
				t.Errorf("expected error for %v", args)
			}
		})
	}
}

func TestOptDefault(t *testing.T) {
	req := &Request{Options: map[string]string{"a": "x", "empty": ""}}
	if got := req.Opt("a", "def"); got != "x" {
		t.Errorf("got %q, want x", got)
	}
	if got := req.Opt("empty", "def"); got != "def" {
		t.Errorf("empty should fall back to default, got %q", got)
	}
	if got := req.Opt("missing", "def"); got != "def" {
		t.Errorf("missing should fall back to default, got %q", got)
	}
}
