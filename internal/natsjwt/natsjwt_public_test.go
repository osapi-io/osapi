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

package natsjwt_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/natsjwt"
)

type NatsJWTPublicTestSuite struct {
	suite.Suite
	dir     string
	account string
	keyFile string
}

func TestNatsJWTPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(NatsJWTPublicTestSuite))
}

func (s *NatsJWTPublicTestSuite) SetupTest() {
	s.dir = s.T().TempDir()
	acct, err := nkeys.CreateAccount()
	s.Require().NoError(err)
	s.account, _ = acct.PublicKey()
	signing, err := nkeys.CreateAccount()
	s.Require().NoError(err)
	seed, _ := signing.Seed()
	s.keyFile = filepath.Join(s.dir, "signing.nk")
	s.Require().NoError(os.WriteFile(s.keyFile, append(seed, '\n'), 0o600))
}

func (s *NatsJWTPublicTestSuite) TestMint() {
	signer, err := natsjwt.Load(s.keyFile, s.account)
	s.Require().NoError(err)
	seed, pub, err := natsjwt.NewUser()
	s.Require().NoError(err)
	s.True(strings.HasPrefix(string(seed), "SU"))

	expires := time.Now().Add(time.Hour).Truncate(time.Second)
	token, err := signer.Mint(pub, "m-1", bus.Permissions{
		Publish:   []string{"osapi.join.request"},
		Subscribe: []string{"_INBOX.>"},
		Responses: true,
	}, expires)
	s.Require().NoError(err)

	uc, err := jwt.DecodeUserClaims(token)
	s.Require().NoError(err)
	s.Equal(pub, uc.Subject)
	s.Equal("m-1", uc.Name)
	s.Equal(s.account, uc.IssuerAccount)
	s.Equal(jwt.StringList{"osapi.join.request"}, uc.Pub.Allow)
	s.Equal(jwt.StringList{"_INBOX.>"}, uc.Sub.Allow)
	s.NotNil(uc.Resp)
	s.Equal(expires.Unix(), uc.Expires)

	forever, err := signer.Mint(pub, "m-1", bus.Permissions{}, time.Time{})
	s.Require().NoError(err)
	uc, _ = jwt.DecodeUserClaims(forever)
	s.Zero(uc.Expires)
	s.Nil(uc.Resp)
	s.True(uc.HasEmptyPermissions(), "a user of a scoped key has no permission or limit")
	scope := jwt.NewUserScope()
	scope.Key = uc.Issuer
	s.NoError(scope.ValidateScopedSigner(uc), "a scoped key may sign it")

	_, err = signer.Mint("not-a-key", "m-1", bus.Permissions{}, time.Time{})
	s.ErrorContains(err, "is not a user public key")

	creds, err := natsjwt.Creds(token, seed)
	s.Require().NoError(err)
	s.Contains(string(creds), token)
	s.Contains(string(creds), string(seed))
}

func (s *NatsJWTPublicTestSuite) TestNewCreds() {
	signer, err := natsjwt.Load(s.keyFile, s.account)
	s.Require().NoError(err)
	creds, err := signer.NewCreds(
		"osapi-join",
		bus.Permissions{Publish: []string{"a"}},
		time.Time{},
	)
	s.Require().NoError(err)
	userJWT, err := jwt.ParseDecoratedJWT(creds)
	s.Require().NoError(err)
	uc, err := jwt.DecodeUserClaims(userJWT)
	s.Require().NoError(err)
	s.Equal("osapi-join", uc.Name)

	_, err = natsjwt.BrokenSigner(s.account).NewCreds("x", bus.Permissions{}, time.Time{})
	s.Error(err, "a key with no seed cannot sign")

	defer natsjwt.SetCreateUser(func() (nkeys.KeyPair, error) {
		return nil, errors.New("no entropy")
	})()
	_, err = signer.NewCreds("x", bus.Permissions{}, time.Time{})
	s.ErrorContains(err, "no entropy")
}

func (s *NatsJWTPublicTestSuite) TestNewSigner() {
	kp, _ := nkeys.CreateAccount()
	signer := natsjwt.NewSigner(kp, s.account)
	_, pub, err := natsjwt.NewUser()
	s.Require().NoError(err)
	token, err := signer.Mint(pub, "m-1", bus.Permissions{}, time.Time{})
	s.Require().NoError(err)
	uc, err := jwt.DecodeUserClaims(token)
	s.Require().NoError(err)
	s.Equal(s.account, uc.IssuerAccount)
}

func (s *NatsJWTPublicTestSuite) TestNewUserFails() {
	defer natsjwt.SetCreateUser(func() (nkeys.KeyPair, error) {
		return nil, errors.New("no entropy")
	})()
	_, _, err := natsjwt.NewUser()
	s.ErrorContains(err, "no entropy")
}

// TestLoadAccountPub: with no account, Load reads account.pub beside the
// signing key, which osapi's NATS server writes at its first start.
func (s *NatsJWTPublicTestSuite) TestLoadAccountPub() {
	acct, _ := nkeys.CreateAccount()
	seed, _ := acct.Seed()
	pub, _ := acct.PublicKey()
	dir := s.T().TempDir()
	file := filepath.Join(dir, "agent-signing.nk")
	s.Require().NoError(os.WriteFile(file, seed, 0o600))

	_, err := natsjwt.Load(file, "")
	s.ErrorContains(err, "no account.pub beside the signing key")

	s.Require().NoError(os.WriteFile(filepath.Join(dir, "account.pub"), []byte(pub+"\n"), 0o600))
	signer, err := natsjwt.Load(file, "")
	s.Require().NoError(err)
	s.NotNil(signer)
}

func (s *NatsJWTPublicTestSuite) TestLoad() {
	user, _ := nkeys.CreateUser()
	userSeed, _ := user.Seed()
	userFile := filepath.Join(s.dir, "user.nk")
	s.Require().NoError(os.WriteFile(userFile, userSeed, 0o600))
	badFile := filepath.Join(s.dir, "bad.nk")
	s.Require().NoError(os.WriteFile(badFile, []byte("nope"), 0o600))

	tests := []struct {
		name    string
		file    string
		account string
		wantErr string
	}{
		{
			name:    "a file that is not there",
			file:    filepath.Join(s.dir, "none"),
			account: s.account,
			wantErr: "read the signing key",
		},
		{
			name:    "a file that holds no seed",
			file:    badFile,
			account: s.account,
			wantErr: "holds no account seed",
		},
		{
			name:    "a seed that is not an account's",
			file:    userFile,
			account: s.account,
			wantErr: "holds no account seed",
		},
		{
			name:    "an account that is not a key",
			file:    s.keyFile,
			account: "x",
			wantErr: "is not an account public key",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := natsjwt.Load(tt.file, tt.account)
			s.ErrorContains(err, tt.wantErr)
		})
	}
}
