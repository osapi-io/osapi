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

package agent_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

type JoinPublicTestSuite struct {
	suite.Suite
}

func TestJoinPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JoinPublicTestSuite))
}

// controller is a controller that answers a join with key, or with status
// when it is not 200, and checks that the request is signed by its key.
func (s *JoinPublicTestSuite) controller(
	key string,
	status int,
) *httptest.Server {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/api/agent/join", r.URL.Path)
		var body struct {
			Token         string `json:"token"`
			MachineID     string `json:"machine_id"`
			Hostname      string `json:"hostname"`
			PublicKey     string `json:"public_key"`
			EncryptionKey string `json:"encryption_key"`
			Signature     string `json:"signature"`
		}
		s.NoError(json.NewDecoder(r.Body).Decode(&body))
		pub, _ := base64.StdEncoding.DecodeString(body.PublicKey)
		sig, _ := base64.StdEncoding.DecodeString(body.Signature)
		req := pki.JoinRequest{
			Token: body.Token, MachineID: body.MachineID, Hostname: body.Hostname,
			PublicKey: pub, EncryptionKey: body.EncryptionKey, Signature: sig,
		}
		s.True(req.Verify(), "the agent signs its request with its own key")
		s.Equal("osapi_jt_t1.secret", body.Token)
		w.Header().Set("Content-Type", "application/json")
		if status != http.StatusOK {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"join token has expired"}`))

			return
		}
		_, _ = w.Write([]byte(`{"controller_public_key":"` + key + `"}`))
	}))
	s.T().Cleanup(srv.Close)

	return srv
}

// joinString is the join string of srv, with fp as its fingerprint.
func joinString(
	srv *httptest.Server,
	fp string,
) string {
	u, _ := url.Parse(srv.URL)
	q := url.Values{"token": {"osapi_jt_t1.secret"}, "fp": {fp}}

	return "osapi://" + u.Host + "?" + q.Encode()
}

// fingerprint is the fingerprint of srv's certificate, as the join string has it.
func fingerprint(
	srv *httptest.Server,
) string {
	sum := sha256.Sum256(srv.Certificate().Raw)

	return "SHA256:" + hex.EncodeToString(sum[:])
}

// enrollWithJoin runs the enrollment of an agent with join as its join string.
func enrollWithJoin(
	join string,
) (*agent.Agent, avfs.VFS) {
	fs := memfs.New()
	a := agent.New(fs, config.Config{Agent: config.AgentConfig{
		Join: join,
		PKI:  config.AgentPKI{KeyDir: "/keys"},
	}}, slog.New(slog.DiscardHandler), nil, "", nil, nil, nil, nil, nil, nil, "", "", nil, nil)
	_ = agent.ExportHandlePKIEnrollment(context.Background(), a)

	return a, fs
}

func (s *JoinPublicTestSuite) TestJoin() {
	ctrlKey, _, _ := ed25519.GenerateKey(rand.Reader)
	good := base64.StdEncoding.EncodeToString(ctrlKey)

	s.Run("when the controller admits, the agent keeps its key and is enrolled", func() {
		srv := s.controller(good, http.StatusOK)

		a, fs := enrollWithJoin(joinString(srv, fingerprint(srv)))

		s.Equal(ctrlKey, agent.GetAgentPKIManager(a).ControllerPublicKey())
		_, err := fs.Stat("/keys/" + agent.EnrolledControllerKey)
		s.NoError(err, "kept for the next start")
	})

	for _, tc := range []struct {
		name string
		join func(*httptest.Server) string
		key  string
		code int
	}{
		{
			name: "when the certificate is not the one that the string pins, it does not send the token",
			join: func(srv *httptest.Server) string { return joinString(srv, "SHA256:00") },
			key:  good, code: http.StatusOK,
		},
		{
			name: "when the controller refuses the token, it asks over NATS",
			join: func(srv *httptest.Server) string { return joinString(srv, fingerprint(srv)) },
			key:  good, code: http.StatusForbidden,
		},
		{
			name: "when the controller's key is not base64, it asks over NATS",
			join: func(srv *httptest.Server) string { return joinString(srv, fingerprint(srv)) },
			key:  "not base64!", code: http.StatusOK,
		},
		{
			name: "when the join string is not one, it asks over NATS",
			join: func(*httptest.Server) string { return "https://controller:8080?token=x" },
			key:  good, code: http.StatusOK,
		},
		{
			name: "when the join string has no fingerprint, it asks over NATS",
			join: func(*httptest.Server) string { return "osapi://controller:8080?token=x" },
			key:  good, code: http.StatusOK,
		},
	} {
		s.Run(tc.name, func() {
			srv := s.controller(tc.key, tc.code)

			a, _ := enrollWithJoin(tc.join(srv))

			s.Nil(agent.GetAgentPKIManager(a).ControllerPublicKey(), "not enrolled")
		})
	}
}

func (s *JoinPublicTestSuite) TestVerifyPin() {
	s.Error(agent.ExportVerifyPin("SHA256:00", nil), "a server with no certificate")
}
