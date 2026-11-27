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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// InFlightSuite holds that a job in flight sends its result when the agent
// restarts its consumers: a config with new labels does, and before
// osapi-io/osapi#1487 it cancelled the job, whose caller then timed out.
//
// It is not a scenario: its check reads the result of a command that runs
// while the next step runs, and a background step of a scenario keeps no
// output for a check.
type InFlightSuite struct {
	suite.Suite
}

// hostname is the agent that the suite targets.
func (s *InFlightSuite) hostname() string {
	var resp struct {
		Items []struct {
			Hostname string `json:"hostname"`
		} `json:"items"`
	}
	out, stderr, code := runCLI("client", "agent", "list", "--json")
	s.Require().Zero(code, stderr)
	s.Require().NoError(parseJSON(out, &resp))
	for _, a := range resp.Items {
		if target == "_any" || a.Hostname == target {
			return a.Hostname
		}
	}
	s.FailNow("no agent matches the target " + target)

	return ""
}

// applied waits until host applied its newest config.
func (s *InFlightSuite) applied(
	host string,
) {
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		var resp struct {
			Items []struct {
				Hostname     string `json:"hostname"`
				ConfigBehind bool   `json:"config_behind"`
			} `json:"items"`
		}
		out, _, code := runCLI("client", "agent", "list", "--json")
		if code == 0 && parseJSON(out, &resp) == nil {
			behind := false
			for _, a := range resp.Items {
				behind = behind || a.Hostname == host && a.ConfigBehind
			}
			if !behind {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	s.Fail(host + " did not apply its config in 45 seconds")
}

func (s *InFlightSuite) TestJobOutlivesNewLabels() {
	skipWrite(s.T())
	host := s.hostname()
	defer func() {
		_, stderr, code := runCLI("client", "agent", "config", "update", "--hostname", host)
		s.Zero(code, stderr)
		s.applied(host)
	}()

	// The job runs for 14 seconds; the agent applies the new labels at its
	// next heartbeat, within 10, while the job runs.
	type result struct {
		out, stderr string
		code        int
	}
	done := make(chan result, 1)
	go func() {
		out, stderr, code := runCLI("client", "node", "command", "exec", "--command", "sleep",
			"--args", "14", "--timeout", "60", "--target", host, "--json")
		done <- result{out, stderr, code}
	}()
	time.Sleep(time.Second)
	_, stderr, code := runCLI("client", "agent", "config", "update", "--hostname", host,
		"--labels", "inflight=1")
	s.Require().Zero(code, stderr)

	r := <-done
	skipIfCommandOff(s.T(), r.out)
	s.Require().Zero(r.code, "%s%s", r.stderr, r.out)
	var body struct {
		Results []map[string]any `json:"results"`
	}
	s.Require().NoError(parseJSON(r.out, &body), r.out)
	s.Require().NotEmpty(body.Results)
	s.Equal("ok", body.Results[0]["status"], "the job sent its result: %v", body.Results[0])
	s.False(strings.Contains(r.out, "timeout"), r.out)
}

func TestInFlightSuite(
	t *testing.T,
) {
	suite.Run(t, new(InFlightSuite))
}
