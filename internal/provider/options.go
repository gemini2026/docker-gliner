package provider

import (
	"fmt"
	"strings"
)

// Command is the lifecycle subcommand Compose asked for.
type Command string

const (
	CmdUp   Command = "up"
	CmdDown Command = "down"
	CmdStop Command = "stop"
)

// Request is a parsed provider invocation.
//
// Compose calls us as:
//
//	docker-gliner compose --project-name <P> up   --k=v ... "<service>"
//	docker-gliner compose --project-name <P> down "<service>"
type Request struct {
	ProjectName string
	Command     Command
	Service     string
	Options     map[string]string
}

// Opt returns the option value for key, or def if unset/empty.
func (r *Request) Opt(key, def string) string {
	if v, ok := r.Options[key]; ok && v != "" {
		return v
	}
	return def
}

// Parse interprets the argv Compose passes (excluding the program name itself).
// It expects a leading "compose" super-command, then flags, the subcommand, more
// flags, and a trailing service positional. Flags may appear before or after the
// subcommand; only "--key=value" / "--key value" forms are accepted.
func Parse(args []string) (*Request, error) {
	req := &Request{Options: map[string]string{}}
	var positionals []string

	i := 0
	// Compose prefixes the call with a "compose" super-command; tolerate its
	// absence so the binary is testable/runnable directly.
	if i < len(args) && args[i] == "compose" {
		i++
	}

	for i < len(args) {
		arg := args[i]
		switch {
		case strings.HasPrefix(arg, "--"):
			key := strings.TrimPrefix(arg, "--")
			var val string
			if eq := strings.IndexByte(key, '='); eq >= 0 {
				key, val = key[:eq], key[eq+1:]
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				// "--key value" form: only consume the next token if it is not
				// itself a flag and not the subcommand/service we still need.
				i++
				val = args[i]
			}
			if key == "project-name" {
				req.ProjectName = val
			} else {
				req.Options[key] = val
			}
		case arg == string(CmdUp), arg == string(CmdDown), arg == string(CmdStop):
			if req.Command != "" {
				return nil, fmt.Errorf("multiple subcommands: %q and %q", req.Command, arg)
			}
			req.Command = Command(arg)
		default:
			positionals = append(positionals, arg)
		}
		i++
	}

	if req.Command == "" {
		return nil, fmt.Errorf("missing subcommand (up/down/stop)")
	}
	switch len(positionals) {
	case 0:
		return nil, fmt.Errorf("missing service name")
	case 1:
		req.Service = positionals[0]
	default:
		return nil, fmt.Errorf("unexpected extra arguments: %v", positionals[1:])
	}
	return req, nil
}
