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
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// AgentServiceSuite installs the agent as a systemd service with a join
// string of a fixture server, and removes it. It runs as root on Linux with
// systemd, against the suite's own server only: a deployment's suite runs on
// the host of a person, such as the lab's Mac, where it must install no
// service.
type AgentServiceSuite struct {
	suite.Suite

	f *fixtureServer
}

func (s *AgentServiceSuite) SetupSuite() {
	switch {
	case deployment():
		s.T().
			Skip("installs a service on the host that runs the suite: runs against the suite's own server")
	case runtime.GOOS != "linux" || os.Geteuid() != 0:
		s.T().Skip("installs a systemd service: runs as root on Linux, as on the CI runner")
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		s.T().Skip("installs a systemd service: this host has no systemctl")
	}
	skipWriteOp(s.T(), "AGENT_SERVICE")
	s.f = startFixture(s.T(), true)
}

// service runs osapi agent service with the default config path, which the
// service reads, and none of the suite's settings: the join string carries
// what the agent needs.
func (s *AgentServiceSuite) service(
	args ...string,
) (string, string, int) {
	cmd := exec.Command(binaryPath, append([]string{"agent", "service"}, args...)...)
	cmd.Env = fixtureBaseEnv()
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

func (s *AgentServiceSuite) TestInstallAndUninstall() {
	out, errOut, code := s.f.cli("controller", "join", "token", "create", "--json")
	s.Require().Zero(code, errOut)
	var tok struct {
		Join string `json:"join"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &tok))

	out, errOut, code = s.service("install", "--join", tok.Join, "--json")
	s.Require().Zero(code, "install: %s%s", errOut, out)
	var installed struct {
		Service    string `json:"service"`
		ConfigFile string `json:"config_file"`
		Joined     bool   `json:"joined"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &installed), out)
	s.True(installed.Joined, "the service's agent joined")
	s.FileExists(installed.Service)
	s.FileExists(installed.ConfigFile)
	enabled, _ := exec.Command("systemctl", "is-enabled", "osapi-agent").Output()
	s.Equal("enabled", string(bytes.TrimSpace(enabled)))
	s.True(waitFor(60*time.Second, func() bool {
		out, _, code := s.f.cli("client", "agent", "list", "--json")
		if code != 0 {
			return false
		}
		var resp struct {
			Items []struct {
				State string `json:"state"`
			} `json:"items"`
		}

		return json.Unmarshal([]byte(out), &resp) == nil &&
			len(resp.Items) == 1 && resp.Items[0].State == "Ready"
	}), "the service's agent is not Ready in the fixture")

	out, errOut, code = s.service("uninstall", "--purge", "--json")
	s.Require().Zero(code, "uninstall: %s%s", errOut, out)
	s.NoFileExists(installed.Service)
	s.NoFileExists(installed.ConfigFile)
	active, _ := exec.Command("systemctl", "is-active", "osapi-agent").Output()
	s.NotEqual("active", string(bytes.TrimSpace(active)))
}

func TestAgentServiceSuite(t *testing.T) {
	suite.Run(t, new(AgentServiceSuite))
}
