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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ControllerKeyRotationSuite rotates the controller's key through the API,
// with no new admission of the agents. It runs against the suite's own
// server and against the lab: the agents take the next key at their next
// heartbeat, the controller switches when each of them has it, and they run
// a job after.
type ControllerKeyRotationSuite struct {
	suite.Suite
}

func (s *ControllerKeyRotationSuite) SetupSuite() {
	skipWriteOp(s.T(), "CONTROLLER_KEY_ROTATE")
}

// controllerKey is the controller's key and its rotation, as
// osapi client pki key get answers it.
type controllerKey struct {
	Current string `json:"current"`
	Next    string `json:"next"`
	Agents  []struct {
		Hostname string `json:"hostname"`
		HasNext  bool   `json:"has_next"`
	} `json:"agents"`
}

func (s *ControllerKeyRotationSuite) key() controllerKey {
	stdout, stderr, code := runCLI("client", "pki", "key", "get", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var k controllerKey
	s.Require().NoError(parseJSON(stdout, &k), stdout)

	return k
}

// TestRotateTheControllersKey: the controller's key rotates, each agent
// takes the new key and stays connected, and a job runs after.
func (s *ControllerKeyRotationSuite) TestRotateTheControllersKey() {
	before := s.key().Current
	s.Require().NotEmpty(before)

	stdout, stderr, code := runCLI("client", "pki", "key", "rotate", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var started controllerKey
	s.Require().NoError(parseJSON(stdout, &started), stdout)
	s.Require().NotEmpty(started.Next, "a rotation is in progress: %s", stdout)
	s.NotEqual(before, started.Next, "the next key is a new key")
	s.NotEmpty(started.Agents, "the switch waits for the admitted agents: %s", stdout)

	// Each agent takes the next key at its next heartbeat, 10s, and the
	// controller looks every 10s.
	k := s.switchedTo(started.Next, 90*time.Second)
	s.Empty(k.Next, "no rotation is in progress after the switch")

	// Each agent reports the new key as the one that it verifies with, after
	// its next heartbeat.
	s.agentsTrust(started.Next, 30*time.Second)

	stdout, stderr, code = runCLI("client", "node", "hostname", "get", "--target", target, "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var got struct {
		Results []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &got), stdout)
	s.Require().NotEmpty(got.Results, stdout)
	for _, r := range got.Results {
		s.Empty(r.Error, "the agent verifies the job that the new key signed: %s", stdout)
		s.NotEmpty(r.Name, stdout)
	}
}

// TestRotateAgain: a rotation in progress is not started again; rotate
// answers it.
func (s *ControllerKeyRotationSuite) TestRotateAgain() {
	stdout, stderr, code := runCLI("client", "pki", "key", "rotate", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var first controllerKey
	s.Require().NoError(parseJSON(stdout, &first), stdout)

	stdout, stderr, code = runCLI("client", "pki", "key", "rotate", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var second controllerKey
	s.Require().NoError(parseJSON(stdout, &second), stdout)
	if second.Next != "" {
		s.Equal(first.Next, second.Next, "the rotation in progress, not a new one")
	}
	s.switchedTo(first.Next, 90*time.Second)
}

// switchedTo waits until the controller signs with fp, and returns its key.
func (s *ControllerKeyRotationSuite) switchedTo(
	fp string,
	timeout time.Duration,
) controllerKey {
	deadline := time.Now().Add(timeout)
	var last controllerKey
	for time.Now().Before(deadline) {
		last = s.key()
		if last.Current == fp {
			return last
		}
		time.Sleep(2 * time.Second)
	}
	s.FailNow(fmt.Sprintf("the controller did not switch to %s after %s: %+v", fp, timeout, last))

	return last
}

// agentsTrust waits until each agent reports fp as its controller key.
func (s *ControllerKeyRotationSuite) agentsTrust(
	fp string,
	timeout time.Duration,
) {
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		stdout, _, code := runCLI("client", "agent", "list", "--json")
		last = stdout
		if code == 0 {
			var resp struct {
				Items []struct {
					Hostname      string `json:"hostname"`
					State         string `json:"state"`
					ControllerKey string `json:"controller_key"`
				} `json:"items"`
			}
			if parseJSON(stdout, &resp) == nil && len(resp.Items) > 0 {
				all := true
				for _, a := range resp.Items {
					if a.ControllerKey != fp {
						all = false
					}
				}
				if all {
					return
				}
			}
		}
		time.Sleep(time.Second)
	}
	s.FailNow(fmt.Sprintf("not each agent reports %s after %s: %s", fp, timeout, last))
}

func TestControllerKeyRotationSuite(
	t *testing.T,
) {
	suite.Run(t, new(ControllerKeyRotationSuite))
}
