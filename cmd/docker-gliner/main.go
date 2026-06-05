// Command docker-gliner is a Docker Compose provider plugin that runs a native
// host GLiNER server (or any host process) so it can use the Apple Metal GPU,
// which is unreachable from inside a container. Compose invokes this binary for
// a service declaring `provider: { type: gliner }`. See docs/PROTOCOL.md.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/amichel/docker-gliner/internal/provider"
	"github.com/amichel/docker-gliner/internal/runner"
)

func main() {
	em := provider.NewEmitter(os.Stdout)

	req, err := provider.Parse(os.Args[1:])
	if err != nil {
		em.Errorf("invalid invocation: %v", err)
		os.Exit(1)
	}

	switch req.Command {
	case provider.CmdUp:
		if err := up(em, req); err != nil {
			em.Errorf("%v", err)
			os.Exit(1)
		}
	case provider.CmdDown, provider.CmdStop:
		if err := down(em, req); err != nil {
			em.Errorf("%v", err)
			os.Exit(1)
		}
	default:
		em.Errorf("unsupported command %q", req.Command)
		os.Exit(1)
	}
}

// resolveCommand returns the argv for the host process. If a `command` option is
// given it is used verbatim; otherwise we default to the bundled GLiNER server.
func resolveCommand(req *provider.Request, port string) []string {
	if c := req.Opt("command", ""); c != "" {
		return splitCommand(c)
	}
	// Default: the bundled server next to this binary (server/serve_gliner.py),
	// run via the interpreter from the `python` option (default python3).
	py := req.Opt("python", "python3")
	script := req.Opt("server", defaultServerPath())
	argv := []string{py, script, "--port", port}
	if d := req.Opt("device", ""); d != "" {
		argv = append(argv, "--device", d)
	}
	if m := req.Opt("model", ""); m != "" {
		argv = append(argv, "--model", m)
	}
	return argv
}

// defaultServerPath locates the bundled server relative to the binary, then
// falls back to a conventional install location.
func defaultServerPath() string {
	if exe, err := os.Executable(); err == nil {
		cand := filepath.Join(filepath.Dir(exe), "server", "serve_gliner.py")
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return "serve_gliner.py"
}

func up(em *provider.Emitter, req *provider.Request) error {
	port := req.Opt("port", "8100")
	host := req.Opt("host", "localhost")
	endpointVar := req.Opt("endpoint_var", "ENDPOINT")
	endpoint := fmt.Sprintf("http://%s:%s", host, port)
	healthURL := req.Opt("health", endpoint+"/health")

	// Idempotency: if an instance is already running, just re-publish its
	// endpoint without starting another. (PROTOCOL.md §1: up MUST be idempotent.)
	if pid, err := runner.Running(req.ProjectName, req.Service); err != nil {
		return err
	} else if pid != 0 {
		em.Infof("%s already running (pid %d)", req.Service, pid)
		em.Setenv(endpointVar, endpoint)
		return nil
	}

	argv := resolveCommand(req, port)
	em.Infof("starting %s: %v", req.Service, argv)
	pid, err := runner.Start(runner.Spec{
		Project: req.ProjectName,
		Service: req.Service,
		Argv:    argv,
	})
	if err != nil {
		return err
	}
	em.Debug(fmt.Sprintf("pid %d, logs at %s", pid, runner.LogFilePath(req.ProjectName, req.Service)))

	timeout := parseSeconds(req.Opt("health_timeout", "120"), 120*time.Second)
	em.Infof("waiting for %s to become healthy at %s", req.Service, healthURL)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := runner.WaitHealthy(ctx, healthURL, time.Second); err != nil {
		// Tear down the half-started process so a retry starts clean.
		_ = runner.Stop(req.ProjectName, req.Service, 5*time.Second)
		return fmt.Errorf("%s did not become healthy (see %s): %w",
			req.Service, runner.LogFilePath(req.ProjectName, req.Service), err)
	}

	em.Infof("%s healthy at %s", req.Service, endpoint)
	em.Setenv(endpointVar, endpoint)
	return nil
}

func down(em *provider.Emitter, req *provider.Request) error {
	em.Infof("stopping %s", req.Service)
	grace := parseSeconds(req.Opt("stop_grace", "10"), 10*time.Second)
	if err := runner.Stop(req.ProjectName, req.Service, grace); err != nil {
		return err
	}
	em.Infof("%s stopped", req.Service)
	return nil
}

func parseSeconds(s string, def time.Duration) time.Duration {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return def
}
