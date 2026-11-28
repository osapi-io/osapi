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

package config_test

import (
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"

	"github.com/avfs/avfs"
)

type SigningKeyPublicTestSuite struct {
	suite.Suite
}

func (s *SigningKeyPublicTestSuite) TearDownTest() {
	config.SetRandRead(rand.Read)
}

func (s *SigningKeyPublicTestSuite) TestEnsureSigningKey() {
	dir := s.T().TempDir()
	var c config.Config
	c.Controller.PKI.KeyDir = filepath.Join(dir, "pki")

	s.Require().NoError(config.EnsureSigningKey(&c))
	s.Require().Len(c.Controller.API.Security.SigningKeys, 1)
	made := c.Controller.API.Security.SigningKeys[0]
	s.Equal(config.SigningKeyID, made.ID)
	s.Len(made.Key, 64, "32 bytes, as hex")
	info, err := testFS.Stat(filepath.Join(c.Controller.PKI.KeyDir, config.SigningKeyFile))
	s.Require().NoError(err)
	s.Equal(os.FileMode(0o600), info.Mode().Perm())

	var again config.Config
	again.Controller.PKI.KeyDir = c.Controller.PKI.KeyDir
	s.Require().NoError(config.EnsureSigningKey(&again))
	s.Equal(made, again.Controller.API.Security.SigningKeys[0], "the key is made once, then read")

	var own config.Config
	own.Controller.API.Security.SigningKeys = []config.SigningKey{
		{ID: "main", Key: strings.Repeat("k", 32)},
	}
	s.Require().NoError(config.EnsureSigningKey(&own))
	s.Equal("main", own.Controller.API.Security.SigningKeys[0].ID, "a key in the config wins")

	var unreadable config.Config
	unreadable.Controller.PKI.KeyDir = filepath.Join(dir, "dirkey")
	s.Require().
		NoError(testFS.MkdirAll(filepath.Join(unreadable.Controller.PKI.KeyDir, config.SigningKeyFile), 0o700))
	s.ErrorContains(config.EnsureSigningKey(&unreadable), "read the signing key")

	locked := filepath.Join(dir, "locked")
	s.Require().NoError(testFS.MkdirAll(locked, 0o500))
	var blocked config.Config
	blocked.Controller.PKI.KeyDir = filepath.Join(locked, "pki")
	s.ErrorContains(config.EnsureSigningKey(&blocked), "make the signing key")

	var readonly config.Config
	readonly.Controller.PKI.KeyDir = filepath.Join(dir, "ro")
	s.Require().NoError(testFS.MkdirAll(readonly.Controller.PKI.KeyDir, 0o500))
	s.ErrorContains(config.EnsureSigningKey(&readonly), "write the signing key")

	config.SetRandRead(func([]byte) (int, error) { return 0, errors.New("no entropy") })
	var noRand config.Config
	noRand.Controller.PKI.KeyDir = filepath.Join(dir, "norand")
	s.ErrorContains(config.EnsureSigningKey(&noRand), "make the signing key")
}

// TestShortSigningKey: a key that is empty or short is refused, from each
// source (GHSA). A controller that signed with an empty key would accept a
// token that anybody can sign.
func (s *SigningKeyPublicTestSuite) TestShortSigningKey() {
	tests := []struct {
		name  string
		setup func(c *config.Config)
	}{
		{
			// A first start that made the file and failed to write it.
			name: "an empty file in the key dir",
			setup: func(c *config.Config) {
				s.Require().NoError(testFS.MkdirAll(c.Controller.PKI.KeyDir, 0o700))
				s.Require().NoError(avfs.WriteFile(testFS,
					filepath.Join(c.Controller.PKI.KeyDir, config.SigningKeyFile), nil, 0o600))
			},
		},
		{
			name: "a short file in the key dir",
			setup: func(c *config.Config) {
				s.Require().NoError(testFS.MkdirAll(c.Controller.PKI.KeyDir, 0o700))
				s.Require().NoError(avfs.WriteFile(testFS,
					filepath.Join(
						c.Controller.PKI.KeyDir,
						config.SigningKeyFile,
					),
					[]byte("short\n"),
					0o600,
				))
			},
		},
		{
			// A secret reference resolves after the config check.
			name: "a short key in the config",
			setup: func(c *config.Config) {
				c.Controller.API.Security.SigningKeys = []config.SigningKey{{ID: "main", Key: ""}}
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var c config.Config
			c.Controller.PKI.KeyDir = filepath.Join(s.T().TempDir(), "pki")
			tt.setup(&c)
			s.ErrorIs(config.EnsureSigningKey(&c), config.ErrShortSigningKey)
		})
	}
}

// TestSigningKeyRace: two processes that make the key at the same time read
// the same key.
func (s *SigningKeyPublicTestSuite) TestSigningKeyRace() {
	dir := filepath.Join(s.T().TempDir(), "pki")
	const n = 8
	keys := make(chan string, n)
	errs := make(chan error, n)
	for range n {
		go func() {
			var c config.Config
			c.Controller.PKI.KeyDir = dir
			err := config.EnsureSigningKey(&c)
			errs <- err
			if err == nil {
				keys <- c.Controller.API.Security.SigningKeys[0].Key
			}
		}()
	}
	for range n {
		s.Require().NoError(<-errs)
	}
	first := <-keys
	for range n - 1 {
		s.Equal(first, <-keys)
	}
	entries, err := testFS.ReadDir(dir)
	s.Require().NoError(err)
	s.Len(entries, 1, "no temporary file is left")
}

// TestSigningKeyWriteErrors: each failed step of the write is an error, and
// leaves no key that a later start reads.
func (s *SigningKeyPublicTestSuite) TestSigningKeyWriteErrors() {
	boom := errors.New("boom")
	notThere := func(string) ([]byte, error) { return nil, os.ErrNotExist }
	tests := []struct {
		name    string
		read    func(string) ([]byte, error)
		write   func(string, []byte, os.FileMode) error
		link    func(string, string) error
		wantErr string
	}{
		{
			name:    "the temporary file cannot be written",
			write:   func(string, []byte, os.FileMode) error { return boom },
			wantErr: "write the signing key",
		},
		{
			name:    "the link fails",
			link:    func(string, string) error { return boom },
			wantErr: "write the signing key",
		},
		{
			name:    "the key cannot be read after the link",
			read:    notThere,
			wantErr: "read the signing key",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			defer config.SetSigningKeyFS(tt.read, tt.write, tt.link)()
			var c config.Config
			c.Controller.PKI.KeyDir = filepath.Join(s.T().TempDir(), "pki")
			s.ErrorContains(config.EnsureSigningKey(&c), tt.wantErr)
			entries, _ := testFS.ReadDir(c.Controller.PKI.KeyDir)
			for _, e := range entries {
				s.NotContains(e.Name(), ".tmp", "no temporary file is left")
			}
		})
	}
}

func TestSigningKeyPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SigningKeyPublicTestSuite))
}
