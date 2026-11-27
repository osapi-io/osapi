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

package integration_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// NATSCertRotationSuite rotates the NATS server's certificate through the
// API, with no restart and no new join of the agents. It runs against the
// suite's own server and against the lab: the agents trust the next
// certificate at their next heartbeat, the controller renames it into place
// when each of them does, NATS serves it at its next handshake, and the
// agents stay connected and run a job after.
type NATSCertRotationSuite struct {
	suite.Suite
}

func (s *NATSCertRotationSuite) SetupSuite() {
	skipWriteOp(s.T(), "NATS_CERT_ROTATE")
}

// natsCert is the NATS server's certificate and its rotation, as
// osapi client pki nats get answers it.
type natsCert struct {
	Current string `json:"current"`
	Next    string `json:"next"`
	Agents  []struct {
		Hostname string `json:"hostname"`
		HasNext  bool   `json:"has_next"`
	} `json:"agents"`
}

func (s *NATSCertRotationSuite) cert() natsCert {
	stdout, stderr, code := runCLI("client", "pki", "nats", "get", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var c natsCert
	s.Require().NoError(parseJSON(stdout, &c), stdout)

	return c
}

// TestRotateTheNATSCertificate: the certificate rotates, each agent trusts
// the new one and stays connected, a job runs after, and a join string made
// after pins the new certificate.
func (s *NATSCertRotationSuite) TestRotateTheNATSCertificate() {
	before := s.cert().Current
	s.Require().NotEmpty(before)

	stdout, stderr, code := runCLI("client", "pki", "nats", "rotate", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var started natsCert
	s.Require().NoError(parseJSON(stdout, &started), stdout)
	s.Require().NotEmpty(started.Next, "a rotation is in progress: %s", stdout)
	s.NotEqual(before, started.Next, "the next certificate is a new one")
	s.NotEmpty(started.Agents, "the switch waits for the admitted agents: %s", stdout)

	// Each agent trusts the next certificate at its next heartbeat, 10s, and
	// the controller looks every 10s.
	c := s.switchedTo(started.Next, 90*time.Second)
	s.Empty(c.Next, "no rotation is in progress after the switch")

	// The agents' connections stay up: each agent still reports, and runs a
	// job, over a NATS server that serves the new certificate.
	s.agentsReady(30 * time.Second)
	stdout, stderr, code = runCLI("client", "node", "hostname", "get", "--target", target, "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var got struct {
		Results []struct {
			Name  string `json:"name"`
			Error string `json:"error"`
		} `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &got), stdout)
	s.Require().NotEmpty(got.Results, stdout)
	for _, r := range got.Results {
		s.Empty(r.Error, "the agent runs a job after the switch: %s", stdout)
		s.NotEmpty(r.Name, stdout)
	}

	// A join string is made on the controller's host, which the lab's suite
	// does not run on.
	if deployment() {
		return
	}
	stdout, stderr, code = runCLI("controller", "join", "token", "create", "--json")
	s.Require().Zero(code, "%s%s", stdout, stderr)
	var token struct {
		Join string `json:"join"`
	}
	s.Require().NoError(parseJSON(stdout, &token), stdout)
	if strings.HasPrefix(token.Join, "tls://") {
		s.Contains(token.Join, "fp="+started.Next,
			"a join string made after the switch pins the new certificate")
	}
}

// switchedTo waits until the NATS server serves fp, and returns its
// certificate.
func (s *NATSCertRotationSuite) switchedTo(
	fp string,
	timeout time.Duration,
) natsCert {
	deadline := time.Now().Add(timeout)
	var last natsCert
	for time.Now().Before(deadline) {
		last = s.cert()
		if last.Current == fp {
			return last
		}
		time.Sleep(2 * time.Second)
	}
	s.FailNow(fmt.Sprintf("the NATS server did not switch to %s after %s: %+v", fp, timeout, last))

	return last
}

// agentsReady waits until each agent is Ready.
func (s *NATSCertRotationSuite) agentsReady(
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
					State string `json:"state"`
				} `json:"items"`
			}
			if parseJSON(stdout, &resp) == nil && len(resp.Items) > 0 {
				all := true
				for _, a := range resp.Items {
					if a.State != "Ready" {
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
	s.FailNow(fmt.Sprintf("not each agent is Ready after %s: %s", timeout, last))
}

func TestNATSCertRotationSuite(
	t *testing.T,
) {
	suite.Run(t, new(NATSCertRotationSuite))
}
