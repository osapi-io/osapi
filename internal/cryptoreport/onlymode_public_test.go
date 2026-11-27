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

package cryptoreport_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/fips140"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// OnlyModePublicTestSuite holds each use's OnlyMode to what the code does.
// The fips140=only CI job runs it (go_fips_packages in the justfile), where
// each probe of a use that fails there must fail, so a fix upstream fails
// this suite until the use says that it works.
type OnlyModePublicTestSuite struct {
	suite.Suite
}

// probe runs f, and is its error or the panic that it raised: FIPS 140-3
// only mode refuses some calls with a panic.
func probe(
	f func() error,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return f()
}

// nkeysSign is what osapi does with nkeys: make a user and sign with it.
func nkeysSign() error {
	kp, err := nkeys.CreateUser()
	if err != nil {
		return err
	}
	_, err = kp.Sign([]byte("osapi"))

	return err
}

// storeSeal is the call that nats-server makes to encrypt a block of the
// JetStream store: AES-GCM with a nonce that it picks.
func storeSeal() error {
	block, err := aes.NewCipher(make([]byte, 32))
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	aead.Seal(nil, make([]byte, aead.NonceSize()), []byte("osapi"), nil)

	return nil
}

func (s *OnlyModePublicTestSuite) TestProbes() {
	tests := []struct {
		name  string
		use   cryptoreport.Algorithm
		probe func() error
	}{
		{name: "NATS JWTs sign with nkeys", use: cryptoreport.NATSAuth, probe: nkeysSign},
		{name: "the JetStream store seals with its own nonce", use: cryptoreport.Store, probe: storeSeal},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := probe(tc.probe)
			if !fips140.Enforced() {
				s.NoError(err, "it works outside only mode")
				return
			}
			s.Equal(tc.use.OnlyMode, err == nil,
				"%s: OnlyMode is %v, and in only mode the probe gave %v", tc.use.Name, tc.use.OnlyMode, err)
		})
	}
}

// TestOnlyModeIsTested: each use that claims to work in only mode is in a
// package that the fips140=only CI job tests.
func (s *OnlyModePublicTestSuite) TestOnlyModeIsTested() {
	data, err := os.ReadFile("../../justfile")
	s.Require().NoError(err)
	m := regexp.MustCompile(`(?m)^go_fips_packages := "([^"]+)"`).FindSubmatch(data)
	s.Require().NotNil(m, "the justfile has go_fips_packages")
	tested := strings.Fields(string(m[1]))
	tests := []struct {
		use     cryptoreport.Algorithm
		pkg     string
		reasons bool
	}{
		{use: cryptoreport.APITokens, pkg: "./internal/authtoken/"},
		{use: cryptoreport.Signatures, pkg: "./internal/pki/"},
		{use: cryptoreport.Seal, pkg: "./internal/secret/seal/"},
		{use: cryptoreport.Vault, pkg: "./internal/secret/vault/natsvault/"},
		{use: cryptoreport.NATSAuth, pkg: "./internal/cryptoreport/"},
		{use: cryptoreport.Store, pkg: "./internal/cryptoreport/"},
	}
	for _, tc := range tests {
		s.Run(tc.use.Name, func() {
			s.Contains(tested, tc.pkg, "the fips140=only job tests %s", tc.pkg)
			s.True(tc.use.Approved)
			if !tc.use.OnlyMode {
				s.NotEmpty(tc.use.Reason, "a use that fails in only mode says why")
			}
		})
	}
}

// TestMode: the report reads the mode that Go runs the process in.
func (s *OnlyModePublicTestSuite) TestMode() {
	cryptoreport.ResetMode()
	s.Equal(cryptoreport.ModeOf(fips140.Enabled(), fips140.Enforced()),
		cryptoreport.New(nil, nil).Mode)
	tests := []struct {
		name     string
		enabled  bool
		enforced bool
		want     types.FIPSMode
	}{
		{name: "when the module is not enabled", want: types.FIPSOff},
		{name: "when it is enabled", enabled: true, want: types.FIPSOn},
		{name: "when it is enforced", enabled: true, enforced: true, want: types.FIPSOnly},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, cryptoreport.ModeOf(tc.enabled, tc.enforced))
		})
	}
}

func TestOnlyModePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OnlyModePublicTestSuite))
}
