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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// SecurityDefaultsSuite checks the defaults of osapi-io/osapi#1413 from the
// outside, as an attacker on the network sees them: TLS on each port, NATS
// auth, and the API's tokens and roles. It runs against the suite's own
// server, whose ports it knows, and against a fixture server of its own
// when the suite runs against a deployment.
//
// It is not a scenario: it dials the ports with raw TLS, HTTP and NATS
// clients, not the CLI, to see what a client that is not osapi gets.
type SecurityDefaultsSuite struct {
	suite.Suite

	dir      string
	base     string
	apiPort  int
	natsPort int
}

func (s *SecurityDefaultsSuite) SetupSuite() {
	s.dir, s.apiPort, s.natsPort = tempDir, apiPort, natsPort
	// A deployment's server, such as the lab's, has its own config: the
	// suite checks the defaults on a fixture server of its own there.
	if deployment() {
		f := startFixture(s.T(), true)
		s.dir, s.apiPort, s.natsPort = f.dir, f.apiPort, f.natsPort
	}
	s.base = fmt.Sprintf("https://127.0.0.1:%d", s.apiPort)
}

// client is an HTTP client that trusts the API of the server under test.
func (s *SecurityDefaultsSuite) client() *http.Client {
	return apiClientFor(filepath.Join(s.dir, "pki", "controller", "api.crt"))
}

// get is the status of a GET of url with token, with client.
func (s *SecurityDefaultsSuite) get(
	client *http.Client,
	url string,
	token string,
) int {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	s.Require().NoError(err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	return resp.StatusCode
}

// tokenFor is a token with roles, signed with key.
func (s *SecurityDefaultsSuite) tokenFor(
	key authtoken.Key,
	roles ...types.Role,
) string {
	t, err := authtoken.New(nil).Generate(key, roles, "security@test", nil, nil)
	s.Require().NoError(err)

	return t
}

func (s *SecurityDefaultsSuite) TestAPI() {
	health := s.base + "/api/health/ready"
	audit := s.base + "/api/audit"
	other := authtoken.Key{ID: "other", Secret: "0123456789abcdef0123456789abcdef0123456789abcdef"}

	s.Run("plain HTTP to the API is refused", func() {
		plain := fmt.Sprintf("http://127.0.0.1:%d/api/health/ready", s.apiPort)
		s.Equal(http.StatusBadRequest, s.get(&http.Client{Timeout: 5 * time.Second}, plain, ""))
	})
	s.Run("HTTPS with the self-signed certificate is served", func() {
		s.Equal(http.StatusOK, s.get(s.client(), health, ""))
	})
	s.Run("HTTPS that does not trust the certificate is refused by the client", func() {
		_, err := (&http.Client{Timeout: 5 * time.Second}).Get(health)
		s.ErrorContains(err, "certificate")
	})
	s.Run("no token is 401", func() {
		s.Equal(http.StatusUnauthorized, s.get(s.client(), audit, ""))
	})
	s.Run("a token of another signing key is 401", func() {
		s.Equal(
			http.StatusUnauthorized,
			s.get(s.client(), audit, s.tokenFor(other, types.RoleAdmin)),
		)
	})
	s.Run("a role without the permission is 403", func() {
		s.Equal(
			http.StatusForbidden,
			s.get(s.client(), audit, s.tokenFor(signingKey, types.RoleObserver)),
		)
	})
	s.Run("a role with the permission is 200", func() {
		s.Equal(http.StatusOK, s.get(s.client(), audit, s.tokenFor(signingKey, types.RoleAuditor)))
	})
}

func (s *SecurityDefaultsSuite) TestNATS() {
	url := fmt.Sprintf("nats://127.0.0.1:%d", s.natsPort)
	pool := x509.NewCertPool()
	pem, err := os.ReadFile(filepath.Join(s.dir, "pki", "nats", "nats.crt"))
	s.Require().NoError(err, "the NATS server made its self-signed certificate")
	s.Require().True(pool.AppendCertsFromPEM(pem))
	secure := nats.Secure(
		&tls.Config{RootCAs: pool, ServerName: "localhost", MinVersion: tls.VersionTLS12},
	)

	s.Run("a connection without TLS is refused", func() {
		_, err := nats.Connect(url, nats.UserInfo("controller", "integration-controller"),
			nats.Timeout(5*time.Second), nats.MaxReconnects(0))
		s.Error(err)
	})
	s.Run("a connection over TLS with no credentials is refused", func() {
		_, err := nats.Connect(url, secure, nats.Timeout(5*time.Second), nats.MaxReconnects(0))
		s.ErrorContains(err, "Authorization Violation")
	})
	s.Run("the controller cannot read the answers that agents get", func() {
		s.violates("controller", "integration-controller", secure, url, func(nc *nats.Conn) error {
			_, err := nc.SubscribeSync("enroll.response.integration-nosuch")

			return err
		})
	})
	s.Run("a join user can send no job", func() {
		s.violates("join", "integration-join", secure, url, func(nc *nats.Conn) error {
			return nc.Publish("jobs.query._any", []byte("{}"))
		})
	})
	s.Run("a join user cannot read the answers that agents get", func() {
		s.violates("join", "integration-join", secure, url, func(nc *nats.Conn) error {
			_, err := nc.SubscribeSync("enroll.response.integration-nosuch")

			return err
		})
	})
	s.Run("an agent cannot send a job", func() {
		nc, err := nats.Connect(url, secure, nats.UserInfo("agent", "integration-agent"),
			nats.Timeout(5*time.Second), nats.MaxReconnects(0))
		s.Require().NoError(err)
		defer nc.Close()
		errs := make(chan error, 1)
		nc.SetErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) { errs <- err })
		s.Require().NoError(nc.Publish("jobs.query._any", []byte("{}")))
		_ = nc.Flush()
		select {
		case err := <-errs:
			s.ErrorIs(err, nats.ErrPermissionViolation)
		case <-time.After(3 * time.Second):
			s.Fail("the server let an agent publish a job")
		}
	})
}

// violates connects as user and runs do, which the server must refuse with
// a permission violation.
func (s *SecurityDefaultsSuite) violates(
	user string,
	password string,
	secure nats.Option,
	url string,
	do func(nc *nats.Conn) error,
) {
	nc, err := nats.Connect(url, secure, nats.UserInfo(user, password),
		nats.Timeout(5*time.Second), nats.MaxReconnects(0))
	s.Require().NoError(err)
	defer nc.Close()
	errs := make(chan error, 1)
	nc.SetErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
		select {
		case errs <- err:
		default:
		}
	})
	s.Require().NoError(do(nc))
	_ = nc.Flush()
	select {
	case err := <-errs:
		s.ErrorIs(err, nats.ErrPermissionViolation)
	case <-time.After(3 * time.Second):
		s.Failf("no permission violation", "the server let %s do it", user)
	}
}

func TestSecurityDefaultsSuite(
	t *testing.T,
) {
	suite.Run(t, new(SecurityDefaultsSuite))
}
