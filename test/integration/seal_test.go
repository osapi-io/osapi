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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

// SealSmokeSuite runs a job with a secret end to end: the controller seals
// the value to the P-256 key of the agent with AES-256-GCM, and the agent
// opens it and gives it to the method.
type SealSmokeSuite struct {
	suite.Suite
}

// TestAdmittedKeys: each admitted agent joined with a P-256 encryption key,
// so the controller can seal a job to it.
func (s *SealSmokeSuite) TestAdmittedKeys() {
	stdout, stderr, code := runCLI("client", "admission", "list", "--state", "admitted", "--json")
	s.Require().Equal(0, code, stderr)

	var resp struct {
		Items []struct {
			Hostname      string `json:"hostname"`
			EncryptionKey string `json:"encryption_key"`
			Fingerprint   string `json:"fingerprint"`
		} `json:"items"`
	}
	s.Require().NoError(parseJSON(stdout, &resp))
	s.Require().NotEmpty(resp.Items, "the suite needs an admitted agent")
	for _, a := range resp.Items {
		s.True(pki.ValidEncryptionKey(a.EncryptionKey),
			"%s: the encryption key is not a P-256 key: %q", a.Hostname, a.EncryptionKey)
		s.True(strings.HasPrefix(a.Fingerprint, "SHA256:"), a.Hostname)
	}
}

// TestSealedEnv: a vault secret in the env of docker exec reaches the
// command, and the job that the controller stores does not have it.
func (s *SealSmokeSuite) TestSealedEnv() {
	skipWriteOp(s.T(), "SEALED_SECRET")

	stamp := fmt.Sprintf("%d", time.Now().UnixNano())
	key := "int-sealed-" + stamp
	value := "sealed-value-" + stamp
	name := "osapi-int-sealed-" + stamp[len(stamp)-8:]

	_, stderr, code := runCLIStdin(value,
		"client", "secret", "update", "--vault", "default", "--key", key, "--value", "-")
	s.Require().Equal(0, code, stderr)
	s.T().Cleanup(func() {
		_, _, _ = runCLI("client", "secret", "delete", "--vault", "default", "--key", key)
	})

	stdout, stderr, code := runCLI("client", "node", "docker", "create",
		"--target", target, "--name", name, "--image", "alpine:3.20",
		"--command", "sleep", "--command", "300", "--auto-start", "--json")
	s.Require().Equal(0, code, "%s %s", stdout, stderr)
	var created struct {
		Results []struct {
			Hostname string `json:"hostname"`
			Status   string `json:"status"`
			Error    string `json:"error"`
		} `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &created))
	s.Require().Len(created.Results, 1, stdout)
	r := created.Results[0]
	if r.Status == "skipped" {
		s.T().Skipf("%s has no docker: %s", r.Hostname, r.Error)
	}
	s.Require().NotEqual("failed", r.Status, stdout)
	host := r.Hostname
	s.T().Cleanup(func() {
		_, _, _ = runCLI("client", "node", "docker", "stop", "--target", host, "--name", name)
		_, _, _ = runCLI("client", "node", "docker", "delete", "--target", host, "--name", name)
	})

	stdout, stderr, code = runCLI("client", "node", "docker", "exec",
		"--target", host, "--name", name,
		"--env", "SEALED=secret://default/"+key,
		"--command", "printenv", "--command", "SEALED", "--json")
	skipIfCommandOff(s.T(), stdout)
	s.Require().Equal(0, code, "%s %s", stdout, stderr)
	var exec struct {
		JobID   string `json:"job_id"`
		Results []struct {
			Status   string `json:"status"`
			Stdout   string `json:"stdout"`
			ExitCode int    `json:"exit_code"`
		} `json:"results"`
	}
	s.Require().NoError(parseJSON(stdout, &exec))
	s.Require().Len(exec.Results, 1, stdout)
	s.Equal(0, exec.Results[0].ExitCode, stdout)
	s.Equal(value, strings.TrimSpace(exec.Results[0].Stdout),
		"the agent opened the sealed value and gave it to the command")

	// The result has the value, because the command wrote it. The operation
	// that the controller stored has only the sealed value.
	job, stderr, code := runCLI("client", "job", "get", "--id", exec.JobID, "--json")
	s.Require().Equal(0, code, stderr)
	var stored struct {
		Operation struct {
			Data struct {
				Env map[string]string `json:"env"`
			} `json:"data"`
		} `json:"operation"`
	}
	s.Require().NoError(parseJSON(job, &stored))
	sealed := stored.Operation.Data.Env["SEALED"]
	s.True(strings.HasPrefix(sealed, "sealed:"), "the stored env is sealed: %q", sealed)
	s.NotContains(sealed, value)
}

func TestSealSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(SealSmokeSuite))
}
