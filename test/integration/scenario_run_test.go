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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ScenarioSuite runs each scenario of scenarios_test.go against a live
// controller: each scenario is a subtest, and each of its steps one under
// it. A step that fails stops its scenario, since the later steps read what
// it made.
type ScenarioSuite struct {
	suite.Suite
}

// run is one scenario as it runs: its fixture, the values its steps
// captured, and the fixture's agents it started.
type run struct {
	s     *ScenarioSuite
	f     *fixtureServer
	base  string
	stamp string
	vars  map[string]string
	procs map[string]*fixtureProc
	// bg is each command of a background step, which the end stops.
	bg []*exec.Cmd
}

func (s *ScenarioSuite) TestScenarios() {
	for _, sc := range scenarios() {
		s.Run(sc.name, func() {
			s.runScenario(sc)
		})
	}
}

func (s *ScenarioSuite) runScenario(
	sc scenario,
) {
	t := s.T()
	switch {
	case sc.ownServer && deployment():
		t.Skip("changes the host that runs the suite: runs against the suite's own server")
	case sc.labOnly && !deployment():
		t.Skip("runs only against a deployment such as the lab: it takes the target down")
	case sc.rootLinux && (runtime.GOOS != "linux" || os.Geteuid() != 0):
		t.Skip("needs root on Linux, as on the CI runner")
	}
	if sc.rootLinux {
		if _, err := exec.LookPath("systemctl"); err != nil {
			t.Skip("needs systemd: this host has no systemctl")
		}
	}
	if sc.writeOp != "" {
		skipWriteOp(t, sc.writeOp)
	}
	r := &run{
		s: s, vars: map[string]string{}, procs: map[string]*fixtureProc{},
		stamp: fmt.Sprintf("%08d", time.Now().UnixNano()%100000000),
	}
	r.base = t.TempDir()
	if sc.fixture {
		r.f = startFixture(t, sc.fixtureAutoAdmit, func(cfg map[string]any) {
			agent, _ := cfg["agent"].(map[string]any)
			consumer, _ := agent["consumer"].(map[string]any)
			for k, v := range sc.fixtureConsumer {
				if v == nil {
					delete(consumer, k)
				} else {
					consumer[k] = v
				}
			}
			for k, v := range sc.fixtureAgent {
				agent[k] = v
			}
			if sc.fixtureOperator {
				// With no auth of their own, the controller and the
				// fixture's host agent connect with the credentials that the
				// server makes, as on a host where osapi runs its NATS.
				nats, _ := cfg["nats"].(map[string]any)
				server, _ := nats["server"].(map[string]any)
				server["auth"] = map[string]any{"type": "operator"}
				for _, c := range []map[string]any{cfg["controller"].(map[string]any), agent} {
					n, _ := c["nats"].(map[string]any)
					n["auth"] = map[string]any{"type": "none"}
				}
			}
		})
		r.base = r.f.dir
	}
	defer r.cleanup(sc.cleanup)

	for _, st := range sc.steps {
		var skip string
		ok := s.Run(st.name, func() {
			skip = r.step(st)
		})
		if skip != "" {
			t.Skip(skip)
		}
		if !ok {
			return
		}
	}
}

// cleanup runs each cleanup step whose placeholders a step captured, and
// stops each agent that the scenario started. Its failures are not the
// scenario's.
func (r *run) cleanup(
	steps []step,
) {
	for _, st := range steps {
		args, err := r.expandAll(st.cli)
		if err != nil {
			continue
		}
		_, _, _ = r.command(st, args, "")
	}
	for _, p := range r.procs {
		r.f.stopProc(p)
	}
	for _, c := range r.bg {
		_ = c.Process.Kill()
		_ = c.Wait()
	}
}

// builtin is the value of the placeholders that each scenario has.
func (r *run) builtin(
	name string,
	arg string,
) (string, bool) {
	switch name {
	case "target":
		return target, true
	case "stamp":
		return r.stamp, true
	case "dir":
		dir := filepath.Join(r.base, "agent-"+arg)
		r.s.Require().NoError(os.MkdirAll(dir, 0o700))

		return dir, true
	case "wrongfp":
		return wrongFingerprint(r.s.T(), r.vars[arg]), true
	case "urlq":
		return url.QueryEscape(r.vars[arg]), true
	}

	return "", false
}

func (r *run) expand(
	s string,
) (string, error) {
	return expand(s, r.vars, r.builtin)
}

func (r *run) expandAll(
	words []string,
) ([]string, error) {
	out := make([]string, 0, len(words))
	for _, w := range words {
		v, err := r.expand(w)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}

	return out, nil
}

// step runs st until its exit and its checks hold, or its wait ends, and
// returns the reason to skip the scenario, when its skipIf matched.
func (r *run) step(
	st step,
) string {
	s := r.s
	if st.ownServer && deployment() {
		s.T().Skip("reads what only the suite's own server makes")
	}
	keyDir, err := r.expand(st.keyDir)
	s.Require().NoError(err)
	switch {
	case st.start != "":
		join, err := r.expand(st.join)
		s.Require().NoError(err)
		p := r.f.startAgent(keyDir, join)
		r.procs[st.start] = p
		if !st.exits && st.pause > 0 {
			time.Sleep(st.pause)
		}
		if st.exits {
			code := p.exited(st.wait)
			s.NotEqual(
				-1,
				code,
				"the agent still runs after %s: the controller admitted it",
				st.wait,
			)
			s.NotZero(
				code,
				"the agent stopped with no error:\n%s",
				redact(tail(p.out.String(), 20)),
			)
			r.f.stopProc(p)
			delete(r.procs, st.start)
		}

		return ""
	case st.stop != "":
		r.f.stopProc(r.procs[st.stop])
		delete(r.procs, st.stop)

		return ""
	case st.kill != "":
		p := r.procs[st.kill]
		s.Require().NotNil(p, "no agent %s to kill", st.kill)
		_ = p.cmd.Process.Kill()
		<-p.done
		delete(r.procs, st.kill)

		return ""
	case st.background:
		args, err := r.expandAll(st.cli)
		s.Require().NoError(err)
		cmd := exec.Command(binaryPath, append([]string{"-f", r.f.cfg}, args...)...)
		cmd.Env = r.f.env
		s.Require().NoError(cmd.Start())
		r.bg = append(r.bg, cmd)
		if st.pause > 0 {
			time.Sleep(st.pause)
		}

		return ""
	}

	args, err := r.expandAll(append(append([]string{}, st.cli...), st.exec...))
	s.Require().NoError(err)
	stdin, err := r.expand(st.stdin)
	s.Require().NoError(err)

	deadline := time.Now().Add(st.wait)
	var why string
	for {
		var stdout, stderr string
		var code int
		if st.logs != "" {
			p := r.procs[st.logs]
			s.Require().NotNil(p, "no agent %s to read", st.logs)
			stdout = p.out.String()
		} else {
			stdout, stderr, code = r.command(st, args, keyDir, stdin)
		}
		why = r.judge(st, stdout, stderr, code)
		if why == "" && st.skipIf != nil {
			if skip := r.holds(*st.skipIf, stdout, stderr) == ""; skip {
				return fmt.Sprintf("%s: the host cannot run it (%s %s %s)",
					st.name, st.skipIf.path, st.skipIf.op, st.skipIf.want)
			}
		}
		if why == "" || time.Now().After(deadline) {
			break
		}
		time.Sleep(max(st.pause, 500*time.Millisecond))
	}
	s.Require().Empty(why, "%s", strings.Join(append([]string{}, args...), " "))
	if st.wait == 0 && st.pause > 0 {
		time.Sleep(st.pause)
	}

	return ""
}

// judge is the reason that one run of st did not hold, or "". It keeps the
// step's captures first, so a check can read them.
func (r *run) judge(
	st step,
	stdout string,
	stderr string,
	code int,
) string {
	if len(st.exec) == 0 {
		switch {
		case st.fails && code == 0:
			return "the command succeeded, but it must fail: " + redact(clip(stdout))
		case !st.fails && code != 0:
			return fmt.Sprintf(
				"the command exited %d: %s%s",
				code,
				redact(clip(stderr)),
				redact(clip(stdout)),
			)
		}
	}
	for name, path := range st.capture {
		p, err := r.expand(path)
		if err != nil {
			return err.Error()
		}
		vals := r.values(p, stdout, stderr)
		if len(vals) == 0 {
			return fmt.Sprintf("no value at %q to keep as %s", p, name)
		}
		r.vars[name] = vals[0]
	}
	for _, c := range st.checks {
		if why := r.holds(c, stdout, stderr); why != "" {
			return why
		}
	}

	return ""
}

// holds is the reason that c does not hold for a step's output, or "".
func (r *run) holds(
	c check,
	stdout string,
	stderr string,
) string {
	path, err := r.expand(c.path)
	if err != nil {
		return err.Error()
	}
	want, err := r.expand(c.want)
	if err != nil {
		return err.Error()
	}
	c.path, c.want = path, want

	return c.eval(r.values(path, stdout, stderr), func(p string) bool {
		_, err := os.Stat(p)

		return err == nil
	})
}

// values is each value at path in a step's output.
func (r *run) values(
	path string,
	stdout string,
	stderr string,
) []string {
	switch path {
	case "":
		return []string{strings.TrimSpace(stdout)}
	case "!stderr":
		return []string{strings.TrimSpace(stderr)}
	case "!all":
		return []string{strings.TrimSpace(stdout + stderr)}
	}
	var doc any
	if json.Unmarshal([]byte(strings.TrimSpace(stdout)), &doc) != nil {
		return nil
	}
	var out []string
	for _, v := range values(doc, path) {
		out = append(out, text(v))
	}

	return out
}

// command runs a step's command where the step says.
func (r *run) command(
	st step,
	args []string,
	keyDir string,
	stdin ...string,
) (string, string, int) {
	in := ""
	if len(stdin) > 0 {
		in = stdin[0]
	}
	if len(st.exec) > 0 {
		out, _ := exec.Command(args[0], args[1:]...).Output()

		return string(out), "", 0
	}
	switch st.on {
	case onHost:
		return runAgentCLI(keyDir, args...)
	case onFixture:
		if keyDir != "" {
			return r.f.cliWith([]string{"OSAPI_AGENT_PKI_KEY_DIR=" + keyDir}, args...)
		}

		return r.f.cli(args...)
	case onFixtureHost:
		return runFixtureHost(args...)
	case onSuite:
	}

	return runCLIStdin(in, args...)
}

// runFixtureHost runs the binary on the fixture's host with no config file
// and no OSAPI_ environment, as osapi agent service install runs.
func runFixtureHost(
	args ...string,
) (string, string, int) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = fixtureBaseEnv()
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

// secrets matches the parts of an output that a failure's message must not
// print: a join string, a token and a password.
var secrets = regexp.MustCompile(`(?i)("join"\s*:\s*"|token=|pass=|"token"\s*:\s*"|creds=)[^"&\s]*`)

// redact is s with each secret replaced.
func redact(
	s string,
) string {
	return secrets.ReplaceAllString(s, "${1}[redacted]")
}

// clip is the start of s, for a failure's message.
func clip(
	s string,
) string {
	if len(s) > 600 {
		return s[:600] + "..."
	}

	return s
}

// runAgentCLI runs the binary as a command on the agent's host: with the
// config and the environment of the suite's server, so it finds the agent's
// key directory and its NATS user. keyDir, when it is set, is the key
// directory in place of the agent's.
func runAgentCLI(
	keyDir string,
	args ...string,
) (string, string, int) {
	cmd := exec.Command(binaryPath, append([]string{"-f", configPath}, args...)...)
	cmd.Env = serverEnv()
	if keyDir != "" {
		cmd.Env = append(cmd.Env, "OSAPI_AGENT_PKI_KEY_DIR="+keyDir)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		code = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
	}

	return stdout.String(), stderr.String(), code
}

// wrongFingerprint is a join string with a certificate fingerprint that is
// not the server's.
func wrongFingerprint(
	t *testing.T,
	join string,
) string {
	t.Helper()

	u, err := url.Parse(join)
	if err != nil {
		t.Fatalf("parse the join string: %v", err)
	}
	q := u.Query()
	q.Set("fp", "SHA256:"+strings.Repeat("0", 64))
	u.RawQuery = q.Encode()

	return u.String()
}

func TestScenarioSuite(t *testing.T) {
	suite.Run(t, new(ScenarioSuite))
}
