//go:build integration

// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package integration_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// fixtureServer is a second server of the suite's own: a NATS server and a
// controller in their own processes, with their own ports, store and key
// directories, and a short life. A test that spends an agent, such as a
// join, an admission or a revoke, runs against it, so the suite's own
// agent, which each other test needs, stays as it is. It runs where the
// suite runs, also against a deployment such as the lab: so the commands
// that read the files of a component on its host run against it there too.
type fixtureServer struct {
	t        *testing.T
	dir      string
	cfg      string
	apiPort  int
	natsPort int
	token    string
	env      []string
	procs    []*fixtureProc
}

// hostTarget, when it is set, is the fixture server that each command of
// runCLIEnv runs against in place of the suite's: a command that reads the
// files of a component on its host, and its read-backs, against a
// deployment.
var hostTarget *fixtureServer

// fixtureProc is one process of a fixture, with its output kept for a
// failure's message.
type fixtureProc struct {
	name string
	cmd  *exec.Cmd
	out  *lockedBuffer
	done chan struct{}
	err  error
}

// lockedBuffer is a buffer that a process writes and a test reads.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(
	p []byte,
) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// startFixture starts a fixture's NATS server and controller, and stops
// them when the test ends. autoAdmit is controller.pki.auto_admit: off, an
// agent that asks to join without a token waits as pending.
func startFixture(
	t *testing.T,
	autoAdmit bool,
	edits ...func(cfg map[string]any),
) *fixtureServer {
	t.Helper()

	dir, err := os.MkdirTemp(tempDir, "fixture-*")
	if err != nil {
		t.Fatalf("make the fixture's directory: %v", err)
	}
	apiPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	natsPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	tok, err := authtoken.New(slog.New(slog.DiscardHandler)).
		Generate(signingKey, []types.Role{types.RoleAdmin}, "integration-fixture@test", nil, nil)
	if err != nil {
		t.Fatalf("make the fixture's token: %v", err)
	}

	cfgPath, err := writeConfig(dir, true, edits...)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixtureServer{
		t: t, dir: dir, cfg: cfgPath, apiPort: apiPort, natsPort: natsPort, token: tok,
	}
	f.env = append(fixtureBaseEnv(),
		fmt.Sprintf("OSAPI_NATS_SERVER_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_NATS_SERVER_STORE_DIR=%s", filepath.Join(dir, "jetstream")),
		fmt.Sprintf("OSAPI_NATS_SERVER_KEY_DIR=%s", filepath.Join(dir, "pki", "nats")),
		fmt.Sprintf("OSAPI_NATS_SERVER_URL=tls://localhost:%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_API_PORT=%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_API_URL=https://127.0.0.1:%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_AGENT_NATS_PORT=%d", natsPort),
		fmt.Sprintf("OSAPI_CONTROLLER_PKI_KEY_DIR=%s", filepath.Join(dir, "pki", "controller")),
		// The key directory of the fixture's host agent; startAgent gives
		// each other agent its own, which the last value in the environment
		// sets.
		fmt.Sprintf("OSAPI_AGENT_PKI_KEY_DIR=%s", filepath.Join(dir, "pki", "agent")),
		fmt.Sprintf("OSAPI_CONTROLLER_PKI_AUTO_ADMIT=%t", autoAdmit),
		// A job that no agent takes, as for a revoked or a pending agent,
		// fails after this, not after the suite's 2m.
		"OSAPI_CONTROLLER_API_JOB_TIMEOUT=15s",
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_URL=https://127.0.0.1:%d", apiPort),
		fmt.Sprintf("OSAPI_CONTROLLER_CLIENT_SECURITY_BEARER_TOKEN=%s", tok),
	)
	t.Cleanup(f.stop)

	nats := f.start("nats", nil, "nats", "start")
	// The controller reads the certificate that the NATS server makes at
	// its first start.
	cert := filepath.Join(dir, "pki", "nats", "nats.crt")
	if !waitFor(30*time.Second, func() bool {
		_, err := os.Stat(cert)

		return err == nil
	}) {
		f.stop()
		t.Fatalf("the fixture's NATS made no certificate:\n%s", tail(nats.out.String(), 30))
	}
	f.start("controller", nil, "controller", "start")
	f.waitReady(30 * time.Second)

	return f
}

// fixtureBaseEnv is the environment of the suite's process without the
// settings of osapi, which the fixture sets itself: a deployment's, such as
// the lab's in OSAPI_INTEGRATION_OSAPIFILE, must not reach it.
func fixtureBaseEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "OSAPI_") {
			env = append(env, kv)
		}
	}

	return env
}

// start runs one process of the fixture: the binary with the suite's
// config and the fixture's environment, plus extra.
func (f *fixtureServer) start(
	name string,
	extra []string,
	args ...string,
) *fixtureProc {
	f.t.Helper()

	cmd := exec.Command(binaryPath, append([]string{"-f", f.cfg}, args...)...)
	cmd.Env = append(append([]string{}, f.env...), extra...)
	out := &lockedBuffer{}
	cmd.Stdout, cmd.Stderr = out, out
	if err := cmd.Start(); err != nil {
		f.t.Fatalf("start the fixture's %s: %v", name, err)
	}
	p := &fixtureProc{name: name, cmd: cmd, out: out, done: make(chan struct{})}
	go func() {
		p.err = cmd.Wait()
		close(p.done)
	}()
	f.procs = append(f.procs, p)

	return p
}

// startAgent starts an agent of the fixture with its own key directory,
// and the join string join when it is not empty. Each agent of a fixture
// has the machine ID of this host, so a test runs one at a time.
func (f *fixtureServer) startAgent(
	keyDir string,
	join string,
) *fixtureProc {
	f.t.Helper()

	extra := []string{"OSAPI_AGENT_PKI_KEY_DIR=" + keyDir}
	if join != "" {
		extra = append(extra, "OSAPI_AGENT_JOIN="+join)
	}

	return f.start("agent", extra, "agent", "start")
}

// stopProc stops one process of the fixture and waits for it.
func (f *fixtureServer) stopProc(
	p *fixtureProc,
) {
	select {
	case <-p.done:
		return
	default:
	}
	_ = p.cmd.Process.Signal(os.Interrupt)
	select {
	case <-p.done:
	case <-time.After(10 * time.Second):
		_ = p.cmd.Process.Kill()
		<-p.done
	}
}

// stop stops each process of the fixture, the last one first.
func (f *fixtureServer) stop() {
	for i := len(f.procs) - 1; i >= 0; i-- {
		f.stopProc(f.procs[i])
	}
	if f.t.Failed() {
		for _, p := range f.procs {
			f.t.Logf("fixture %s output:\n%s", p.name, tail(p.out.String(), 40))
		}
	}
}

// exited waits up to timeout for a process to stop, and reports its exit
// code, or -1 when it still runs.
func (p *fixtureProc) exited(
	timeout time.Duration,
) int {
	select {
	case <-p.done:
		var exitErr *exec.ExitError
		if errors.As(p.err, &exitErr) {
			return exitErr.ExitCode()
		}
		if p.err != nil {
			return -2
		}

		return 0
	case <-time.After(timeout):
		return -1
	}
}

// cli runs the binary against the fixture: as a client of its API, and as a
// command on its host, such as join token create.
func (f *fixtureServer) cli(
	args ...string,
) (string, string, int) {
	cmd := exec.Command(binaryPath, append([]string{"-f", f.cfg}, args...)...)
	cmd.Env = f.env
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return stdout.String(), stderr.String(), 0
	case errors.As(err, &exitErr):
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	default:
		return stdout.String(), err.Error(), -1
	}
}

// waitReady waits until the fixture's API answers its readiness.
func (f *fixtureServer) waitReady(
	timeout time.Duration,
) {
	f.t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, _, code := f.cli("client", "health", "ready"); code == 0 {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	f.stop()
	f.t.Fatalf("the fixture is not ready after %s", timeout)
}

// waitFor runs check every half second until it is true or timeout passes.
func waitFor(
	timeout time.Duration,
	check func() bool,
) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}

	return check()
}

// tail is the last n lines of s.
func tail(
	s string,
	n int,
) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return strings.Join(lines, "\n")
}

// cliWith is cli with extra in the environment, such as the key directory
// of one agent.
func (f *fixtureServer) cliWith(
	extra []string,
	args ...string,
) (string, string, int) {
	saved := f.env
	f.env = append(append([]string{}, saved...), extra...)
	defer func() { f.env = saved }()

	return f.cli(args...)
}

// writeConfig writes the suite's config into dir with the key file of the
// default vault in dir too, and returns its path. The default key file is
// /etc/osapi/secrets.key, which only root may write: a fixture on the host
// of a person, such as the lab's, and the suite's own server under a user
// that is not root, need one of their own.
//
// A fixture's NATS server serves no metrics: the suite's own takes the
// port.
func writeConfig(
	dir string,
	fixture bool,
	edits ...func(cfg map[string]any),
) (string, error) {
	raw, err := os.ReadFile(filepath.Join(repoRoot(), "test", "integration", "osapi.yaml"))
	if err != nil {
		return "", fmt.Errorf("read the suite's config: %w", err)
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return "", fmt.Errorf("parse the suite's config: %w", err)
	}
	controller, _ := cfg["controller"].(map[string]any)
	if controller == nil {
		controller = map[string]any{}
		cfg["controller"] = controller
	}
	controller["secrets"] = map[string]any{"vaults": []map[string]any{{
		"name": "default", "type": "nats", "key_file": filepath.Join(dir, "secrets.key"),
	}}}
	if fixture {
		nats, _ := cfg["nats"].(map[string]any)
		server, _ := nats["server"].(map[string]any)
		if server == nil {
			return "", fmt.Errorf("the suite's config has no nats.server")
		}
		server["metrics"] = map[string]any{"enabled": false}
	}
	for _, edit := range edits {
		edit(cfg)
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("write the config: %w", err)
	}
	path := filepath.Join(dir, "osapi.yaml")
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return "", fmt.Errorf("write the config: %w", err)
	}

	return path, nil
}
