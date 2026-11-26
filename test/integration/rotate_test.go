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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// KeyRotationSuite rotates the keys of the suite's own agent with
// osapi pki key rotate, which runs on the agent's host with the agent's
// config. The lab's agents run on its VMs, not here, so the lab checks the
// rotation on a VM with just ssh, and this suite runs against the suite's
// own server only.
type KeyRotationSuite struct {
	suite.Suite
}

func (s *KeyRotationSuite) SetupSuite() {
	if deployment() {
		s.T().
			Skip("rotates the keys of the agent on this host: runs against the suite's own server")
	}
	skipWriteOp(s.T(), "KEY_ROTATE")
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

// agentFingerprint is the fingerprint of the keys in the agent's key
// directory, or in keyDir when it is set.
func (s *KeyRotationSuite) agentFingerprint(
	keyDir string,
) string {
	stdout, stderr, code := runAgentCLI(keyDir, "pki", "key", "get", "--name", "agent", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var key struct {
		Fingerprint string `json:"fingerprint"`
	}
	s.Require().NoError(parseJSON(stdout, &key), stdout)
	s.Require().True(strings.HasPrefix(key.Fingerprint, "SHA256:"), stdout)

	return key.Fingerprint
}

// admitted is the fingerprint of each admitted machine.
func (s *KeyRotationSuite) admitted() []string {
	stdout, stderr, code := runCLI("client", "admission", "list", "--state", "admitted", "--json")
	s.Require().Zero(code, stderr)
	var resp struct {
		Items []struct {
			Fingerprint string `json:"fingerprint"`
		} `json:"items"`
	}
	s.Require().NoError(parseJSON(stdout, &resp), stdout)
	out := make([]string, 0, len(resp.Items))
	for _, a := range resp.Items {
		out = append(out, a.Fingerprint)
	}

	return out
}

// TestRotateTheAgentsKeys: the agent's keys rotate with no new admission.
// The admission holds the new fingerprint, the running agent takes the new
// keys at its next heartbeat, and it runs a job after.
func (s *KeyRotationSuite) TestRotateTheAgentsKeys() {
	before := s.agentFingerprint("")
	s.Require().Contains(s.admitted(), before, "the suite's agent is admitted with its keys")

	stdout, stderr, code := runAgentCLI("", "pki", "key", "rotate", "--name", "agent", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var rotated struct {
		Key struct {
			Name        string `json:"name"`
			Fingerprint string `json:"fingerprint"`
		} `json:"key"`
	}
	s.Require().NoError(parseJSON(stdout, &rotated), stdout)
	after := rotated.Key.Fingerprint
	s.Equal("agent", rotated.Key.Name)
	s.NotEqual(before, after, "the agent has new keys")
	s.Equal(after, s.agentFingerprint(""), "the new keys are in place")

	fingerprints := s.admitted()
	s.Contains(fingerprints, after, "the admission holds the new keys")
	s.NotContains(fingerprints, before, "the admission no longer holds the old keys")

	// The running agent reads the new keys at its next heartbeat, 10s.
	host := s.agentWith(after, 30*time.Second)

	stdout, stderr, code = runCLI("client", "node", "hostname", "get", "--target", host, "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var got struct {
		Results []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &got), stdout)
	s.Require().Len(got.Results, 1, stdout)
	r := got.Results[0]
	s.Empty(r.Error, "the controller takes the agent's result: %s", stdout)
	s.NotEqual("refused", r.Status, stdout)
	s.NotEmpty(r.Name, stdout)
}

// agentWith waits until an agent reports the fingerprint fp, and returns
// its hostname.
func (s *KeyRotationSuite) agentWith(
	fp string,
	timeout time.Duration,
) string {
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		stdout, _, code := runCLI("client", "agent", "list", "--json")
		last = stdout
		if code == 0 {
			var resp struct {
				Items []struct {
					Hostname    string `json:"hostname"`
					Fingerprint string `json:"fingerprint"`
				} `json:"items"`
			}
			if parseJSON(stdout, &resp) == nil {
				for _, a := range resp.Items {
					if a.Fingerprint == fp {
						return a.Hostname
					}
				}
			}
		}
		time.Sleep(time.Second)
	}
	s.FailNow(fmt.Sprintf("no agent reports %s after %s: %s", fp, timeout, last))

	return ""
}

// TestKeysThatAreNotAdmitted: keys that the controller did not admit for
// this machine do not rotate, and their new keys are removed.
func (s *KeyRotationSuite) TestKeysThatAreNotAdmitted() {
	dir := filepath.Join(s.T().TempDir(), "agent")
	s.Require().NoError(os.MkdirAll(dir, 0o700))
	stdout, stderr, code := runAgentCLI(dir, "pki", "key", "create", "--name", "agent")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	before := s.agentFingerprint(dir)

	stdout, stderr, code = runAgentCLI(dir, "pki", "key", "rotate", "--name", "agent")
	s.NotZero(code, "a rotation of keys that are not admitted succeeded: %s", stdout)
	s.Contains(stderr+stdout, "the controller refused the rotation")
	s.Equal(before, s.agentFingerprint(dir), "the keys stay")
	_, err := os.Stat(filepath.Join(dir, "agent.next.key"))
	s.True(os.IsNotExist(err), "the refused keys are removed")
}

func TestKeyRotationSuite(
	t *testing.T,
) {
	suite.Run(t, new(KeyRotationSuite))
}
