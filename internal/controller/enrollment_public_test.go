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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/pki"
)

type EnrollmentPublicTestSuite struct {
	suite.Suite
}

func TestEnrollmentPublicTestSuite(t *testing.T) {
	suite.Run(t, new(EnrollmentPublicTestSuite))
}

// issuerFile is a file with the seed of key.
func (s *EnrollmentPublicTestSuite) issuerFile(
	seed []byte,
) string {
	f := filepath.Join(s.T().TempDir(), "issuer.nk")
	s.Require().NoError(os.WriteFile(f, seed, 0o600))

	return f
}

func (s *EnrollmentPublicTestSuite) TestCalloutService() {
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

func (s *EnrollmentPublicTestSuite) TestCalloutConnect() {
	s.Run("TLS that does not load is an error", func() {
		var cfg config.Config
		cfg.Controller.NATS.TLS = config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.pem"}
		s.Error(controller.ExportCalloutConnect(cfg))
	})
}

func (s *EnrollmentPublicTestSuite) TestJoinGivesNoUser() {
	acct, _ := nkeys.CreateAccount()
	seed, _ := acct.Seed()
	var cfg config.Config
	cfg.Controller.Join.Callout.IssuerKeyFile = s.issuerFile(seed)
	user, mint, err := controller.ExportJoinAgentUser(cfg, "osapi")
	s.Require().NoError(err)
	s.Nil(mint)
	_, err = user(pki.JoinRequest{})
	s.ErrorContains(err, "sent no NATS user key")
	got, err := user(pki.JoinRequest{NATSUserKey: "UA1"})
	s.Require().NoError(err)
	s.Equal(pki.JoinNATS{Callout: true}, got)
}

// TestCalloutConnectFails: a callout that cannot connect stops its own
// background task, and the controller starts.
func (s *EnrollmentPublicTestSuite) TestCalloutConnectFails() {
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

func (s *EnrollmentPublicTestSuite) TestJoinAgentUser() {
	s.Run("static users give each agent the agent user", func() {
		var cfg config.Config
		cfg.Controller.Join.Agent = config.NATSUser{Username: "agent", Password: "ap"}
		user, _, err := controller.ExportJoinAgentUser(cfg, "osapi")
		s.Require().NoError(err)
		got, err := user(pki.JoinRequest{})
		s.Require().NoError(err)
		s.Equal(pki.JoinNATS{Username: "agent", Password: "ap"}, got)
	})

	s.Run("operator mode mints a JWT for the agent's own key", func() {
		acct, _ := nkeys.CreateAccount()
		acctPub, _ := acct.PublicKey()
		seed, _ := acct.Seed()
		file := filepath.Join(s.T().TempDir(), "signing.nk")
		s.Require().NoError(os.WriteFile(file, seed, 0o600))
		var cfg config.Config
		cfg.Controller.Join.SigningKeyFile = file
		cfg.Controller.Join.Account = acctPub
		cfg.NATS.Stream = config.NATSStream{Name: "JOBS", Subjects: "jobs.>"}
		user, _, err := controller.ExportJoinAgentUser(cfg, "osapi")
		s.Require().NoError(err)

		kp, _ := nkeys.CreateUser()
		pub, _ := kp.PublicKey()
		got, err := user(pki.JoinRequest{MachineID: "m-1", NATSUserKey: pub})
		s.Require().NoError(err)
		uc, err := jwt.DecodeUserClaims(got.JWT)
		s.Require().NoError(err)
		s.Equal(pub, uc.Subject)
		s.Equal("m-1", uc.Name)
		s.Contains(uc.Pub.Allow, "osapi.enroll.request")
		s.Empty(got.Password)

		_, err = user(pki.JoinRequest{MachineID: "m-1"})
		s.ErrorContains(err, "sent no NATS user key")

		_, err = user(pki.JoinRequest{MachineID: "m-1", NATSUserKey: "bad"})
		s.ErrorContains(err, "is not a user public key")
		s.Zero(uc.Expires)

		cfg.Controller.Join.AgentTTL = "1h"
		_, mint, err := controller.ExportJoinAgentUser(cfg, "osapi")
		s.Require().NoError(err)
		token, err := mint(pub, "m-1")
		s.Require().NoError(err)
		uc, err = jwt.DecodeUserClaims(token)
		s.Require().NoError(err)
		s.InDelta(time.Now().Add(time.Hour).Unix(), uc.Expires, 5)
	})
}

// seedFile is a file with the seed of a key that newKey makes, or with
// data when newKey is nil.
func (s *EnrollmentPublicTestSuite) seedFile(
	newKey func() (nkeys.KeyPair, error),
	data []byte,
) string {
	if newKey != nil {
		kp, err := newKey()
		s.Require().NoError(err)
		data, _ = kp.Seed()
	}
	f := filepath.Join(s.T().TempDir(), "operator-signing.nk")
	s.Require().NoError(os.WriteFile(f, data, 0o600))

	return f
}

// TestStartRevocation: a controller in operator mode with a revocation
// setup starts its revoker, and refuses a setup that it cannot use.
func (s *EnrollmentPublicTestSuite) TestStartRevocation() {
	tests := []struct {
		name    string
		key     func() string
		connect func(context.Context, *slog.Logger, config.Config) (*natsclient.Client, error)
		wantErr string
	}{
		{
			name: "an operator signing key and a system user start the revoker",
			key:  func() string { return s.seedFile(nkeys.CreateOperator, nil) },
		},
		{
			name:    "no operator signing key",
			key:     func() string { return filepath.Join(s.T().TempDir(), "none.nk") },
			wantErr: "read the operator signing key",
		},
		{
			name:    "a file that is not a seed",
			key:     func() string { return s.seedFile(nil, []byte("nope")) },
			wantErr: "the operator signing key",
		},
		{
			name:    "a seed that is not an operator's",
			key:     func() string { return s.seedFile(nkeys.CreateAccount, nil) },
			wantErr: "not an operator key (SO...)",
		},
		{
			name: "a system user that does not connect",
			key:  func() string { return s.seedFile(nkeys.CreateOperator, nil) },
			connect: func(context.Context, *slog.Logger, config.Config) (*natsclient.Client, error) {
				return nil, errors.New("no route")
			},
			wantErr: "no route",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := startServer(s.T())
			connect := tt.connect
			if connect == nil {
				// The test's server is not in operator mode: the revoker
				// connects as the controller, and its sends fail and log.
				connect = func(ctx context.Context, log *slog.Logger, c config.Config) (*natsclient.Client, error) {
					opts, err := bus.ConnectionOptions(c.Controller.NATS, "osapi-revocation")
					s.Require().NoError(err)

					return bus.Dial(ctx, log, opts)
				}
			}
			defer controller.SetRevocationConnect(connect)()
			r := &cfg.Controller.Join.Revocation
			r.AccountJWTFile = filepath.Join(s.T().TempDir(), "account.jwt")
			s.Require().NoError(os.WriteFile(r.AccountJWTFile, []byte("eyJ"), 0o600))
			r.OperatorSigningKeyFile = tt.key()
			r.SystemCredsFile = filepath.Join(s.T().TempDir(), "system.creds")
			cfg.Controller.Join.AgentTTL = "1h"
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			c, err := controller.New(ctx, slog.Default(), cfg)
			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), "NATS user revocation")
				s.Contains(err.Error(), tt.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Require().NoError(c.Start())
			c.Stop(context.Background())
		})
	}
}

// TestNoAccountJWT: a setup from before the revocation, with no account
// JWT, starts with the revocation off.
func (s *EnrollmentPublicTestSuite) TestNoAccountJWT() {
	defer controller.SetRevocationConnect(
		func(context.Context, *slog.Logger, config.Config) (*natsclient.Client, error) {
			s.Fail("a controller with no account JWT connects no revoker")

			return nil, errors.New("not called")
		},
	)()
	cfg := startServer(s.T())
	cfg.Controller.Join.Revocation.AccountJWTFile = filepath.Join(s.T().TempDir(), "account.jwt")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := controller.New(ctx, slog.Default(), cfg)
	s.Require().NoError(err)
	s.Require().NoError(c.Start())
	c.Stop(context.Background())
}

func (s *EnrollmentPublicTestSuite) TestRevocationConnect() {
	s.Run("TLS that does not load is an error", func() {
		var cfg config.Config
		cfg.Controller.NATS.TLS = config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.pem"}
		s.Error(controller.ExportRevocationConnect(cfg))
	})
	s.Run("a server that is not there is an error", func() {
		var cfg config.Config
		cfg.Controller.NATS = config.NATSConnection{Host: "127.0.0.1", Port: 1}
		cfg.Controller.Join.Revocation.SystemCredsFile = filepath.Join(
			s.T().TempDir(),
			"system.creds",
		)
		s.Error(controller.ExportRevocationConnect(cfg))
	})
}
