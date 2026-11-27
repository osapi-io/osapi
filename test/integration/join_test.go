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
	"encoding/json"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// JoinSuite runs an agent through its join, its admission and its removal,
// against a fixture server of its own with auto_admit off, so the suite's
// own agent stays as it is (osapi-io/osapi#1536). Each agent of the fixture
// has this host's machine ID, so the tests run in order, one agent at a
// time.
type JoinSuite struct {
	suite.Suite

	f        *fixtureServer
	hostname string
}

func (s *JoinSuite) SetupSuite() {
	skipWriteOp(s.T(), "JOIN")
	s.f = startFixture(s.T(), false)
}

// keyDir is a new key directory for one agent of the fixture.
func (s *JoinSuite) keyDir(
	name string,
) string {
	return filepath.Join(s.f.dir, "agent-"+name)
}

// joinToken makes a join token on the fixture and returns its ID and its
// join string.
func (s *JoinSuite) joinToken(
	args ...string,
) (string, string) {
	out, errOut, code := s.f.cli(append([]string{
		"controller", "join", "token", "create", "--json",
	}, args...)...)
	s.Require().Zero(code, "join token create: %s%s", errOut, out)
	var tok struct {
		ID   string `json:"id"`
		Join string `json:"join"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &tok), out)
	s.Require().NotEmpty(tok.Join)

	return tok.ID, tok.Join
}

type fixtureAgent struct {
	Hostname  string            `json:"hostname"`
	MachineID string            `json:"machine_id"`
	State     string            `json:"state"`
	Labels    map[string]string `json:"labels"`
}

// agents is the fixture's agent list.
func (s *JoinSuite) agents() []fixtureAgent {
	out, _, code := s.f.cli("client", "agent", "list", "--json")
	if code != 0 {
		return nil
	}
	var resp struct {
		Items []fixtureAgent `json:"items"`
	}
	_ = json.Unmarshal([]byte(out), &resp)

	return resp.Items
}

// readyAgent waits for the fixture's one agent to be Ready, and returns it.
func (s *JoinSuite) readyAgent(
	agent *fixtureProc,
) fixtureAgent {
	var got fixtureAgent
	ok := waitFor(60*time.Second, func() bool {
		for _, a := range s.agents() {
			if a.State == "Ready" {
				got = a

				return true
			}
		}

		return false
	})
	s.Require().True(ok, "no Ready agent in the fixture:\n%s", tail(agent.out.String(), 30))
	s.hostname = got.Hostname

	return got
}

// admissionState is the state of the machine's admission record, or "" when
// it has none.
func (s *JoinSuite) admissionState(
	machineID string,
) string {
	out, _, code := s.f.cli("client", "admission", "get", "--machine-id", machineID, "--json")
	if code != 0 {
		return ""
	}
	var rec struct {
		State string `json:"state"`
	}
	_ = json.Unmarshal([]byte(out), &rec)

	return rec.State
}

// waitState waits for the machine's admission record to be in state.
func (s *JoinSuite) waitState(
	machineID string,
	state string,
) {
	s.Require().True(waitFor(30*time.Second, func() bool {
		return s.admissionState(machineID) == state
	}), "the admission of %s is %q, not %q", machineID, s.admissionState(machineID), state)
}

// jobWorks is whether a job reaches the fixture's agent and completes.
func (s *JoinSuite) jobWorks() bool {
	_, _, code := s.f.cli("client", "node", "hostname", "get", "--target", s.hostname, "--json")

	return code == 0
}

// auditHas is whether the fixture's audit log has an entry with each text.
func (s *JoinSuite) auditHas(
	texts ...string,
) bool {
	out, _, code := s.f.cli("client", "audit", "list", "--limit", "100", "--json")
	if code != 0 {
		return false
	}
	for _, t := range texts {
		if !strings.Contains(out, t) {
			return false
		}
	}

	return true
}

// TestJoin is the life of one machine: a join with a token, the refusals of
// a token that may not admit it, a revoke that cuts it off, and the delete
// of its record; then the join of an agent with no token, which waits as
// pending, an admit, and a reject.
func (s *JoinSuite) TestJoin() {
	// A token admits the agent at once, and gives it the token's labels.
	_, join := s.joinToken("--uses", "1", "--expires", "10m", "--labels", "group=fixture")
	agent := s.f.startAgent(s.keyDir("token"), join)
	a := s.readyAgent(agent)
	s.Equal("fixture", a.Labels["group"], "the agent has the token's labels")
	s.Equal("admitted", s.admissionState(a.MachineID))
	s.True(s.jobWorks(), "a job reaches the joined agent")
	s.True(s.auditHas(`"agent.join"`), "the join is in the audit log")

	// A used token, a revoked token, an expired token and a wrong
	// fingerprint admit no other key of this machine: each agent stops with
	// the reason, and the admitted key stays.
	s.f.stopProc(agent)
	revokedID, revoked := s.joinToken("--uses", "1")
	_, revokeErr, code := s.f.cli("controller", "join", "token", "revoke", "--id", revokedID)
	s.Require().Zero(code, revokeErr)
	_, expired := s.joinToken("--uses", "1", "--expires", "1s")
	_, valid := s.joinToken("--uses", "1")
	time.Sleep(2 * time.Second)
	for name, j := range map[string]string{
		"used":        join,
		"revoked":     revoked,
		"expired":     expired,
		"fingerprint": wrongFingerprint(s.T(), valid),
	} {
		p := s.f.startAgent(s.keyDir(name), j)
		s.NotZero(p.exited(60*time.Second),
			"%s: the agent joined with a token that may not admit it:\n%s",
			name, tail(p.out.String(), 20))
		s.f.stopProc(p)
	}
	s.Equal("admitted", s.admissionState(a.MachineID), "a refused join leaves the admission")

	// A revoke cuts the agent off: it gets no job. A delete then removes
	// its record.
	agent = s.f.startAgent(s.keyDir("token"), "")
	s.readyAgent(agent)
	_, errOut, code := s.f.cli("client", "admission", "revoke", "--machine-id", a.MachineID)
	s.Require().Zero(code, errOut)
	s.waitState(a.MachineID, "revoked")
	s.True(waitFor(60*time.Second, func() bool { return !s.jobWorks() }),
		"a job still reaches the revoked agent")
	s.f.stopProc(agent)
	_, errOut, code = s.f.cli("client", "admission", "delete", "--machine-id", a.MachineID)
	s.Require().Zero(code, errOut)
	s.waitState(a.MachineID, "")

	// With auto_admit off, an agent with no token waits as pending, and
	// takes no job until an admin admits it with its fingerprint.
	pendingDir := s.keyDir("pending")
	agent = s.f.startAgent(pendingDir, "")
	s.waitState(a.MachineID, "pending")
	s.False(s.jobWorks(), "a job reached a pending agent")
	fp := s.fingerprint(pendingDir)
	_, errOut, code = s.f.cli("client", "admission", "admit",
		"--machine-id", a.MachineID, "--fingerprint", fp)
	s.Require().Zero(code, errOut)
	s.waitState(a.MachineID, "admitted")
	s.readyAgent(agent)
	s.True(s.jobWorks(), "a job reaches the admitted agent")

	// A reject of a pending machine keeps it out, with its reason.
	s.f.stopProc(agent)
	_, errOut, code = s.f.cli("client", "admission", "delete", "--machine-id", a.MachineID)
	s.Require().Zero(code, errOut)
	agent = s.f.startAgent(s.keyDir("rejected"), "")
	s.waitState(a.MachineID, "pending")
	_, errOut, code = s.f.cli("client", "admission", "reject",
		"--machine-id", a.MachineID, "--reason", "integration test")
	s.Require().Zero(code, errOut)
	s.waitState(a.MachineID, "rejected")
	s.False(s.jobWorks(), "a job reached a rejected agent")
	s.f.stopProc(agent)

	s.True(s.auditHas("admission"), "the admission changes are in the audit log")
}

// fingerprint is the fingerprint of the agent's keys in keyDir, as an admin
// gives it to admit.
func (s *JoinSuite) fingerprint(
	keyDir string,
) string {
	var fp string
	s.Require().True(waitFor(30*time.Second, func() bool {
		out, _, code := s.f.cliWith([]string{"OSAPI_AGENT_PKI_KEY_DIR=" + keyDir},
			"pki", "key", "get", "--name", "agent", "--json")
		if code != 0 {
			return false
		}
		var k struct {
			Fingerprint string `json:"fingerprint"`
		}
		_ = json.Unmarshal([]byte(out), &k)
		fp = k.Fingerprint

		return fp != ""
	}), "no fingerprint for the agent's keys in %s", keyDir)

	return fp
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

func TestJoinSuite(t *testing.T) {
	suite.Run(t, new(JoinSuite))
}
