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

package natsauth_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/osfs"
	"github.com/nats-io/jwt/v2"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/natsauth"
	"github.com/osapi-io/osapi/internal/natsjwt"
)

type NatsAuthPublicTestSuite struct {
	suite.Suite
	dir string
	cfg config.NATS
}

func TestNatsAuthPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(NatsAuthPublicTestSuite))
}

func (s *NatsAuthPublicTestSuite) SetupTest() {
	s.dir = s.T().TempDir()
	s.cfg = config.NATS{Stream: config.NATSStream{Name: "JOBS", Subjects: "jobs.>"}}
}

// mode is the permission bits of the file at path.
func (s *NatsAuthPublicTestSuite) mode(
	path string,
) string {
	info, err := os.Stat(path)
	s.Require().NoError(err)

	return info.Mode().Perm().String()
}

func (s *NatsAuthPublicTestSuite) TestPassword() {
	r, err := natsauth.Password(context.Background(), osfs.New(), s.dir, "osapi", s.cfg)
	s.Require().NoError(err)
	s.Require().Len(r.Files, 2)
	for _, f := range r.Files {
		s.Equal("-rw-------", s.mode(f.Path), f.Path)
	}
	y, _ := os.ReadFile(filepath.Join(s.dir, natsauth.OsapiConfig))
	for _, role := range []string{"setup", "controller", "agent", "join"} {
		s.Contains(string(y), "role: "+role)
	}
	s.Contains(r.Server, "allow_responses", "the controller answers requests")
	s.Contains(r.Server, `"osapi.join.request"`)
	_, err = natsserver.ProcessConfigFile(filepath.Join(s.dir, "nats-server.conf"))
	s.NoError(err, "a NATS server reads the block")

	again, err := natsauth.Password(context.Background(), osfs.New(), s.dir, "osapi", s.cfg)
	s.Require().NoError(err)
	s.NotEqual(r.Server, again.Server, "each create makes new passwords")
}

func (s *NatsAuthPublicTestSuite) TestOperator() {
	r, err := natsauth.Operator(context.Background(), osfs.New(), s.dir, "osapi", s.cfg)
	s.Require().NoError(err)
	names := make([]string, 0, len(r.Files))
	for _, f := range r.Files {
		names = append(names, filepath.Base(f.Path))
		s.Equal("-rw-------", s.mode(f.Path), f.Path)
	}
	s.ElementsMatch([]string{
		"agent.creds", "controller.creds", "setup.creds", "resolver.conf",
		"operator.nk", "account.nk", "agent-signing.nk", "join-signing.nk", natsauth.OsapiConfig,
	}, names)

	opts, err := natsserver.ProcessConfigFile(filepath.Join(s.dir, "resolver.conf"))
	s.Require().NoError(err, "a NATS server reads the resolver file")
	s.Len(opts.TrustedOperators, 1)

	creds, _ := os.ReadFile(filepath.Join(s.dir, "agent.creds"))
	userJWT, err := jwt.ParseDecoratedJWT(creds)
	s.Require().NoError(err)
	uc, err := jwt.DecodeUserClaims(userJWT)
	s.Require().NoError(err)
	s.Contains(uc.Pub.Allow, "osapi.enroll.request")
	s.Zero(uc.Expires)

	y, _ := os.ReadFile(filepath.Join(s.dir, natsauth.OsapiConfig))
	s.Contains(string(y), "signing_key_file: "+filepath.Join(s.dir, "agent-signing.nk"))
	s.Contains(string(y), "join_signing_key_file: "+filepath.Join(s.dir, "join-signing.nk"))
	s.Contains(string(y), "scoped: true")

	// Each signing key that the controller gets is scoped to its role.
	acctJWT := ""
	for _, line := range strings.Split(r.Server, "\n") {
		if f := strings.Fields(
			line,
		); len(f) == 2 && strings.HasPrefix(f[0], "A") &&
			strings.HasSuffix(f[0], ":") {
			acctJWT = f[1]
		}
	}
	ac, err := jwt.DecodeAccountClaims(acctJWT)
	s.Require().NoError(err)
	s.Equal("osapi", ac.Name)
	roles := map[string]jwt.StringList{}
	for _, file := range []string{"agent-signing.nk", "join-signing.nk"} {
		seed, _ := os.ReadFile(filepath.Join(s.dir, file))
		kp, err := nkeys.FromSeed(seed)
		s.Require().NoError(err)
		pub, _ := kp.PublicKey()
		scope, ok := ac.SigningKeys.GetScope(pub)
		s.Require().True(ok, file)
		us := scope.(*jwt.UserScope)
		roles[us.Role] = us.Template.Pub.Allow
	}
	s.Contains(roles["agent"], "osapi.enroll.request")
	s.Equal(jwt.StringList{"osapi.join.request"}, roles["join"])
	s.True(strings.Contains(r.Server, "resolver_preload"))
}

func (s *NatsAuthPublicTestSuite) TestFailures() {
	ctx := context.Background()
	boom := errors.New("boom")
	file := filepath.Join(s.dir, "file")
	s.Require().NoError(os.WriteFile(file, nil, 0o600))

	tests := []struct {
		name    string
		run     func() error
		wantErr string
	}{
		{
			name: "a password with no random bytes",
			run: func() error {
				defer natsauth.SetFns(func([]byte) (int, error) { return 0, boom }, nil, nil)()
				_, err := natsauth.Password(ctx, osfs.New(), s.dir, "osapi", s.cfg)
				return err
			},
			wantErr: "make a password: boom",
		},
		{
			name: "password files that cannot be written",
			run: func() error {
				_, err := natsauth.Password(ctx, osfs.New(), file, "osapi", s.cfg)
				return err
			},
			wantErr: "file",
		},
		{
			name: "keys that cannot be made",
			run: func() error {
				defer natsauth.SetFns(
					nil,
					func(int) ([]nkeys.KeyPair, error) { return nil, boom },
					nil,
				)()
				_, err := natsauth.Operator(ctx, osfs.New(), s.dir, "osapi", s.cfg)
				return err
			},
			wantErr: "boom",
		},
		{
			name: "credentials that cannot be minted",
			run: func() error {
				defer natsauth.SetFns(
					nil,
					nil,
					func(*natsjwt.Signer, string, bus.Permissions, time.Time) ([]byte, error) {
						return nil, boom
					},
				)()
				_, err := natsauth.Operator(ctx, osfs.New(), s.dir, "osapi", s.cfg)
				return err
			},
			wantErr: "boom",
		},
		{
			name: "an operator key that cannot be made",
			run: func() error {
				defer natsauth.SetKeyMakers(
					func() (nkeys.KeyPair, error) { return nil, boom },
					nkeys.CreateAccount,
				)()
				_, err := natsauth.Operator(ctx, osfs.New(), s.dir, "osapi", s.cfg)
				return err
			},
			wantErr: "boom",
		},
		{
			name: "an account key that cannot be made",
			run: func() error {
				defer natsauth.SetKeyMakers(
					nkeys.CreateOperator,
					func() (nkeys.KeyPair, error) { return nil, boom },
				)()
				_, err := natsauth.Operator(ctx, osfs.New(), s.dir, "osapi", s.cfg)
				return err
			},
			wantErr: "boom",
		},
		{
			name: "operator files that cannot be written",
			run: func() error {
				_, err := natsauth.Operator(ctx, osfs.New(), file, "osapi", s.cfg)
				return err
			},
			wantErr: "file",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.ErrorContains(tt.run(), tt.wantErr)
		})
	}
}

// TestBlockedFiles: a file that a create writes after others, whose path is
// a directory, is the error.
func (s *NatsAuthPublicTestSuite) TestBlockedFiles() {
	ctx := context.Background()
	tests := []struct {
		blocked string
		run     func(dir string) error
	}{
		{"nats-server.conf", func(dir string) error {
			_, err := natsauth.Password(ctx, osfs.New(), dir, "osapi", s.cfg)
			return err
		}},
		{"agent.creds", func(dir string) error {
			_, err := natsauth.Operator(ctx, osfs.New(), dir, "osapi", s.cfg)
			return err
		}},
		{"resolver.conf", func(dir string) error {
			_, err := natsauth.Operator(ctx, osfs.New(), dir, "osapi", s.cfg)
			return err
		}},
		{natsauth.OsapiConfig, func(dir string) error {
			_, err := natsauth.Operator(ctx, osfs.New(), dir, "osapi", s.cfg)
			return err
		}},
	}
	for _, tt := range tests {
		s.Run(tt.blocked, func() {
			dir := s.T().TempDir()
			s.Require().NoError(os.MkdirAll(filepath.Join(dir, tt.blocked, "x"), 0o700))
			s.Error(tt.run(dir))
		})
	}
}
