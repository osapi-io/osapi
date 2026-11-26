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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior/command"
)

// CommandSmokeSuite runs each CLI command that is not a behavior from its
// definition, as BehaviorSmokeSuite runs the behaviors: each refusal of a
// method's Fails must fail, and each Example must succeed, with the value
// that it names replaced by a thing that the suite made. cliPolicies, which
// the unit gate checks, says which examples cannot run and why.
type CommandSmokeSuite struct {
	suite.Suite
	fx *fixtures
}

func (s *CommandSmokeSuite) SetupSuite() {
	s.fx = &fixtures{s: s, dir: s.T().TempDir(), made: map[string]string{}}
}

func (s *CommandSmokeSuite) TearDownSuite() {
	s.fx.cleanup()
	s.fx.settled()
}

// deployment is whether the suite runs against a deployment, such as the
// lab, whose component files are not on this host.
func deployment() bool {
	return os.Getenv("OSAPI_INTEGRATION_OSAPIFILE") != ""
}

// TestRefusals runs each input of each Fails of each command through the
// binary: it must fail, and not for a reason other than its input. A refusal
// that the CLI does not catch, such as a path parameter, reaches the API: it
// goes to the controller, not through the spec proxy, which fails the suite
// for each request that breaks the spec.
func (s *CommandSmokeSuite) TestRefusals() {
	count := 0
	for _, c := range cliCases() {
		for i, r := range c.refusals {
			count++
			s.Run(fmt.Sprintf("%s %d", c.path, i+1), func() {
				args := append(strings.Fields(c.path), command.Args(r)...)
				stdout, stderr, code := runDirect(s.fx.dir, command.Stdin(r), args...)
				s.NotZero(code, "%s: a refusal succeeded: %s", strings.Join(args, " "), stdout)
				s.NotRegexp(connectionError, stderr,
					"%s: it failed for its connection, not its input", strings.Join(args, " "))
			})
		}
	}
	s.NotZero(count)
}

// TestExamples runs each example of each command whose policy lets it run
// here, in an order that makes a thing before it reads or removes it.
func (s *CommandSmokeSuite) TestExamples() {
	for _, c := range byGroup(cliCases()) {
		p := cliPolicies[c.path]
		for i, flags := range c.examples {
			s.Run(fmt.Sprintf("%s %d", c.path, i+1), func() {
				switch {
				case p.skip != "":
					s.T().Skip(p.skip)
				case p.host && deployment():
					s.T().
						Skip("reads the files of a component on its host: runs against the suite's own server")
				case (p.write || c.api && c.verb != "GET") && !runWrites:
					s.T().Skip("a write: set OSAPI_INTEGRATION_WRITES=1")
				}
				words, stdin := s.fx.example(c, flags)
				args := append(strings.Fields(c.path), words...)
				s.fx.before(c, i, args)
				stdout, stderr, code := runRetried(s.fx.dir, stdin, p.env, args...)
				s.Require().Zero(code, "%s: %s%s", strings.Join(args, " "), stderr, stdout)
				if slices.Contains(args, "--json") {
					s.True(json.Valid([]byte(strings.TrimSpace(stdout))),
						"%s: --json printed no JSON: %s", strings.Join(args, " "), stdout)
				}
				s.fx.after(c)
			})
		}
	}
}

// connectionError is an error of the connection, not of the input. It does
// not match a URL in the help, such as tls://nats.example.com:4222.
var connectionError = regexp.MustCompile(
	`(?i)connection refused|no such host|x509: |tls: (handshake|bad|failed)|remote error: tls`,
)

// runDirect is runCLIIn with the controller's URL in place of the spec
// proxy's.
func runDirect(
	dir string,
	stdin string,
	args ...string,
) (string, string, int) {
	cmd := exec.Command(binaryPath, append([]string{"-f", configPath}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(clientEnv(), "OSAPI_CONTROLLER_CLIENT_URL="+controllerURL)
	cmd.Stdin = strings.NewReader(stdin)
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

// runRetried is runCLIIn, again for up to 30 seconds while the command says
// to try again, as a trace does until its spans reach the backend.
func runRetried(
	dir string,
	stdin string,
	env []string,
	args ...string,
) (string, string, int) {
	deadline := time.Now().Add(30 * time.Second)
	for {
		stdout, stderr, code := runCLIWith(dir, stdin, env, args...)
		if code == 0 || !strings.Contains(stderr, "try again") || time.Now().After(deadline) {
			return stdout, stderr, code
		}
		time.Sleep(2 * time.Second)
	}
}

// byGroup keeps the cases of a group together, in the order of the first
// of each group, so the order within a group holds.
func byGroup(
	cases []cliCase,
) []cliCase {
	var order []string
	groups := map[string][]cliCase{}
	for _, c := range cases {
		if _, ok := groups[c.group]; !ok {
			order = append(order, c.group)
		}
		groups[c.group] = append(groups[c.group], c)
	}
	out := make([]cliCase, 0, len(cases))
	for _, g := range order {
		in := groups[g]
		slices.SortStableFunc(in, func(a, b cliCase) int { return weight(a.name) - weight(b.name) })
		out = append(out, in...)
	}

	return out
}

// weight orders the commands of a group: a command that makes a thing,
// then those that read or use it, then one that removes it.
func weight(
	name string,
) int {
	switch name {
	case "create", "upload", "update", "generate", "drain":
		return 0
	case "delete", "revoke":
		return 2
	default:
		return 1
	}
}

// runCLIIn is runCLIStdin with dir as the command's working directory, so
// a command that writes where it runs, such as skill install, writes there.
func runCLIIn(
	dir string,
	stdin string,
	args ...string,
) (string, string, int) {
	return runCLIWith(dir, stdin, nil, args...)
}

// runCLIWith is runCLIIn with env added to the client's environment.
func runCLIWith(
	dir string,
	stdin string,
	env []string,
	args ...string,
) (string, string, int) {
	prev, err := os.Getwd()
	if err != nil {
		return "", err.Error(), -1
	}
	if err := os.Chdir(dir); err != nil {
		return "", err.Error(), -1
	}
	defer func() { _ = os.Chdir(prev) }()

	return runCLIEnv(stdin, env, args...)
}

// fixtures are the things that the examples name, which the suite makes:
// the agent it targets, a job, a file, a secret, a join token. Each is made
// once, the first time that an example names it, and removed at the end.
type fixtures struct {
	s     *CommandSmokeSuite
	dir   string
	made  map[string]string
	undo  []func()
	stamp string
}

func (f *fixtures) cleanup() {
	for i := len(f.undo) - 1; i >= 0; i-- {
		f.undo[i]()
	}
}

// name is a name that is the suite's own, for a thing that it makes.
func (f *fixtures) name(
	base string,
) string {
	if f.stamp == "" {
		f.stamp = fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000)
	}

	return "int-cli-" + f.stamp + "-" + base
}

// once is the value of key, which build makes the first time.
func (f *fixtures) once(
	key string,
	build func() string,
) string {
	if v, ok := f.made[key]; ok {
		return v
	}
	v := build()
	f.made[key] = v

	return v
}

// json runs args with --json and decodes its output into v.
func (f *fixtures) json(
	v any,
	args ...string,
) {
	stdout, stderr, code := runCLI(append(args, "--json")...)
	f.s.Require().Zero(code, "%s: %s", strings.Join(args, " "), stderr)
	f.s.Require().NoError(parseJSON(stdout, v), stdout)
}

// agent is the hostname and the machine ID of the agent the suite targets.
func (f *fixtures) agent() (string, string) {
	host := f.once("hostname", func() string {
		var resp struct {
			Items []struct {
				Hostname  string `json:"hostname"`
				MachineID string `json:"machine_id"`
			} `json:"items"`
		}
		f.json(&resp, "client", "agent", "list")
		for _, a := range resp.Items {
			if target == "_any" || a.Hostname == target {
				f.made["machine"] = a.MachineID

				return a.Hostname
			}
		}
		f.s.FailNow("no agent matches the target " + target)

		return ""
	})

	return host, f.made["machine"]
}

var traceID = regexp.MustCompile(`trace ([0-9a-f]{32})`)

// value is the value that the suite puts in place of v, a value that an
// example of c names, or v.
func (f *fixtures) value(
	c cliCase,
	v string,
) string {
	switch {
	case v == "web-01":
		h, _ := f.agent()

		return h
	case v == "d530664edc1c431eb5ef1e9dfd6617a5":
		_, m := f.agent()

		return m
	case v == "550e8400-e29b-41d4-a716-446655440000" && c.group == "client job":
		return f.once("job", func() string {
			h, _ := f.agent()
			var resp struct {
				JobID string `json:"job_id"`
			}
			f.json(&resp, "client", "node", "hostname", "get", "--target", h)

			return resp.JobID
		})
	case v == "550e8400-e29b-41d4-a716-446655440000" && c.group == "client audit":
		return f.once("audit", func() string {
			var resp struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			}
			f.json(&resp, "client", "audit", "list", "--limit", "1")
			f.s.Require().NotEmpty(resp.Items, "the audit log is empty")

			return resp.Items[0].ID
		})
	case v == "26e448c9dbe1c4a8d216a98c675079c9":
		return f.once("trace", func() string {
			stdout, stderr, _ := runCLI("client", "agent", "list")
			m := traceID.FindStringSubmatch(stdout + stderr)
			// The CLI prints a trace only when tracing is on, as in the
			// lab, which runs Jaeger. The suite's own server has no backend.
			if m == nil && !deployment() {
				f.s.T().
					Skip("the suite's server has no tracing backend: just int in the lab runs trace get")
			}
			f.s.Require().NotNil(m, "the CLI printed no trace: %s%s", stdout, stderr)

			return m[1]
		})
	case v == "app.conf" || v == "app.conf.tmpl":
		return f.once("file "+v, func() string {
			name := f.name(v)
			f.undo = append(
				f.undo,
				func() { _, _, _ = runCLI("client", "file", "delete", "--name", name) },
			)

			return name
		})
	case v == "/tmp/app.conf" || v == "/tmp/app.conf.tmpl":
		return f.once(v, func() string {
			path := filepath.Join(f.dir, filepath.Base(v))
			body := "made by CommandSmokeSuite\n"
			if strings.HasSuffix(v, ".tmpl") {
				body = "host {{ .Hostname }}\n"
			}
			f.s.Require().NoError(os.WriteFile(path, []byte(body), 0o600))

			return path
		})
	case v == "registry-token":
		return f.once("secret", func() string {
			key := f.name("token")
			f.undo = append(f.undo, func() {
				_, _, _ = runCLI("client", "secret", "delete", "--vault", "default", "--key", key)
			})

			return key
		})
	case v == "1a2b3c4d":
		return f.once("join token", func() string {
			var resp struct {
				ID string `json:"id"`
			}
			f.json(&resp, "controller", "join", "token", "create")

			return resp.ID
		})
	case v == "/etc/osapi/osapi.yaml" || v == "./osapi.yaml":
		return configPath
	case v == "/etc/osapi/nats" || v == "./ops-repo":
		return filepath.Join(f.dir, filepath.Base(v)+"-"+f.name("out"))
	}

	return v
}

// example is the words of an example of c with the suite's values, and what
// it reads from standard input.
func (f *fixtures) example(
	c cliCase,
	flags string,
) ([]string, string) {
	split := splitWords(flags)
	words := make([]string, 0, len(split))
	stdin := ""
	for _, w := range split {
		if w == "-" {
			stdin = f.stdin(c)
		}
		words = append(words, f.value(c, w))
	}

	return words, stdin
}

// stdin is what an example of c that reads a value from standard input
// reads.
func (f *fixtures) stdin(
	c cliCase,
) string {
	if c.group == "controller token" {
		return token
	}

	return "int-cli-value"
}

// before makes again what an earlier example of c removed: each example of
// a delete after the first removes a thing that the first already removed.
func (f *fixtures) before(
	c cliCase,
	i int,
	args []string,
) {
	if i == 0 || c.path != "client file delete" {
		return
	}
	name := args[slices.Index(args, "--name")+1]
	_, stderr, code := runCLI("client", "file", "upload", "--name", name,
		"--file", f.value(c, "/tmp/app.conf"))
	f.s.Require().Zero(code, "upload %s again: %s", name, stderr)
}

// settled waits until the agent that the suite targets applied its newest
// config, so a later suite does not send a job while the agent restarts its
// consumers on new labels (osapi-io/osapi#1487).
func (f *fixtures) settled() {
	if _, ok := f.made["hostname"]; !ok {
		return
	}
	host, _ := f.agent()
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		var resp struct {
			Items []struct {
				Hostname     string `json:"hostname"`
				ConfigBehind bool   `json:"config_behind"`
			} `json:"items"`
		}
		f.json(&resp, "client", "agent", "list")
		behind := false
		for _, a := range resp.Items {
			behind = behind || a.Hostname == host && a.ConfigBehind
		}
		if !behind {
			return
		}
		time.Sleep(2 * time.Second)
	}
	f.s.Fail(host + " did not apply its config in 45 seconds")
}

// after undoes what an example changed on the agent that other tests use.
func (f *fixtures) after(
	c cliCase,
) {
	u, ok := undoes[c.path]
	if !ok {
		return
	}
	host, _ := f.agent()
	f.undo = append(f.undo, func() {
		args := append(slices.Clone(u.args), "--hostname", host)
		_, stderr, code := runCLI(args...)
		if u.nothing == "" || !strings.Contains(stderr, u.nothing) {
			f.s.Zero(code, "%s: %s", strings.Join(args, " "), stderr)
		}
	})
}

// splitWords splits flags into words as a shell does: quotes group words,
// a backslash at the end of a line continues it, and a redirect ends the
// command, as in "> osapi.yaml".
func splitWords(
	flags string,
) []string {
	flags = strings.ReplaceAll(flags, "\\\n", " ")
	var words []string
	var cur strings.Builder
	quote := rune(0)
	in := false
	for _, r := range flags {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote, in = r, true
		case r == ' ' || r == '\t' || r == '\n':
			if in {
				words = append(words, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteRune(r)
			in = true
		}
	}
	if in {
		words = append(words, cur.String())
	}
	if i := slices.Index(words, ">"); i >= 0 {
		words = words[:i]
	}

	return words
}

func TestCommandSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(CommandSmokeSuite))
}
