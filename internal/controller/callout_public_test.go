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
package controller_test

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/nats-io/nkeys"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/pki"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
)

// testFS is the host's filesystem, through avfs. The certificates, keys and NATS configs here are read by path, by crypto/tls and the NATS server as well as by the code under test.
var testFS avfs.VFS = osfs.NewWithNoIdm()

type CalloutPublicTestSuite struct {
	suite.Suite
}

func TestCalloutPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CalloutPublicTestSuite))
}

// issuerFile is a file with the seed of key.
func (s *CalloutPublicTestSuite) issuerFile(
	seed []byte,
) string {
	f := filepath.Join(s.T().TempDir(), "issuer.nk")
	s.Require().NoError(avfs.WriteFile(testFS, f, seed, 0o600))

	return f
}

func (s *CalloutPublicTestSuite) TestCalloutService() {
	acct, _ := nkeys.CreateAccount()
	acctSeed, _ := acct.Seed()
	user, _ := nkeys.CreateUser()
	userSeed, _ := user.Seed()
	tests := []struct {
		name    string
		file    func() string
		wantErr string
	}{
		{
			name: "an account seed is the issuer",
			file: func() string { return s.issuerFile(acctSeed) },
		},
		{
			name:    "a file that is not a seed",
			file:    func() string { return s.issuerFile([]byte("nope")) },
			wantErr: "the issuer key",
		},
		{
			name:    "a seed that is not an account's",
			file:    func() string { return s.issuerFile(userSeed) },
			wantErr: "not an account key",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var cfg config.Config
			cfg.Controller.Join.Callout.IssuerKeyFile = tt.file()
			err := controller.ExportCalloutService(cfg)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.NoError(err)
		})
	}
}

func (s *CalloutPublicTestSuite) TestCalloutConnect() {
	s.Run("TLS that does not load is an error", func() {
		var cfg config.Config
		cfg.Controller.NATS.TLS = config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.pem"}
		s.Error(controller.ExportCalloutConnect(cfg))
	})
}

func (s *CalloutPublicTestSuite) TestJoinGivesNoUser() {
	acct, _ := nkeys.CreateAccount()
	seed, _ := acct.Seed()
	var cfg config.Config
	cfg.Controller.Join.Callout.IssuerKeyFile = s.issuerFile(seed)
	user, mint, err := controller.ExportJoinAgentUser(cfg, "osapi")
	s.Require().NoError(err)
	s.Nil(mint, "the callout renews nothing: it decides each connection")
	_, err = user(pki.JoinRequest{})
	s.ErrorContains(err, "sent no NATS user key")
	got, err := user(pki.JoinRequest{NATSUserKey: "UA1"})
	s.Require().NoError(err)
	s.Equal(pki.JoinNATS{Callout: true}, got)
}

// TestCalloutConnectFails: a callout that cannot connect stops its own
// background task, and the controller starts.
func (s *CalloutPublicTestSuite) TestCalloutConnectFails() {
	restore := controller.SetCalloutConnect(
		func(context.Context, *slog.Logger, config.Config) (*natsclient.Client, error) {
			return nil, errors.New("no route")
		},
	)
	defer restore()
	cfg := startServer(s.T())
	acct, _ := nkeys.CreateAccount()
	seed, _ := acct.Seed()
	cfg.Controller.Join.Callout.IssuerKeyFile = s.issuerFile(seed)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := controller.New(ctx, slog.Default(), cfg)
	s.Require().NoError(err)
	s.Require().NoError(c.Start())
	c.Stop(context.Background())
}
