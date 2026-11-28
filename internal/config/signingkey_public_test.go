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
	"sync"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type SigningKeyPublicTestSuite struct {
	suite.Suite

	// fs is the host's file system: EnsureSigningKey reads and writes the
	// key on it.
	fs avfs.VFS
}

func (s *SigningKeyPublicTestSuite) SetupTest() {
	s.fs = osfs.NewWithNoIdm()
}

func (s *SigningKeyPublicTestSuite) TearDownTest() {
	config.SetRandRead(rand.Read)
}

func (s *SigningKeyPublicTestSuite) TearDownSubTest() {
	config.SetRandRead(rand.Read)
	config.ResetSigningKeyFSFn()
}

func (s *SigningKeyPublicTestSuite) TestEnsureSigningKey() {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// setup makes the config of the call, in dir.
		setup func(dir string) *config.Config
		// concurrent, when set, is how many calls run at the same time.
		concurrent   int
		validateFunc func(dir string, configs []*config.Config, errs []error)
	}{
		{
			name: "when the key dir has no key, a key of 32 bytes as hex is made with mode 0600, then read",
			setup: func(dir string) *config.Config {
				return keyDirConfig(filepath.Join(dir, "pki"))
			},
			validateFunc: func(dir string, configs []*config.Config, errs []error) {
				s.Require().NoError(errs[0])
				s.Require().Len(configs[0].Controller.API.Security.SigningKeys, 1)
				made := configs[0].Controller.API.Security.SigningKeys[0]
				s.Equal(config.SigningKeyID, made.ID)
				s.Len(made.Key, 64)
				info, err := s.fs.Stat(filepath.Join(dir, "pki", config.SigningKeyFile))
				s.Require().NoError(err)
				s.Equal(os.FileMode(0o600), info.Mode().Perm())

				again := keyDirConfig(filepath.Join(dir, "pki"))
				s.Require().NoError(config.EnsureSigningKey(again))
				s.Equal(made, again.Controller.API.Security.SigningKeys[0])
			},
		},
		{
			name: "when the config has a key, it wins",
			setup: func(_ string) *config.Config {
				var c config.Config
				c.Controller.API.Security.SigningKeys = []config.SigningKey{
					{ID: "main", Key: strings.Repeat("k", 32)},
				}

				return &c
			},
			validateFunc: func(_ string, configs []*config.Config, errs []error) {
				s.Require().NoError(errs[0])
				s.Equal("main", configs[0].Controller.API.Security.SigningKeys[0].ID)
			},
		},
		{
			name: "when the key file cannot be read, it is an error",
			setup: func(dir string) *config.Config {
				keyDir := filepath.Join(dir, "dirkey")
				s.Require().
					NoError(s.fs.MkdirAll(filepath.Join(keyDir, config.SigningKeyFile), 0o700))

				return keyDirConfig(keyDir)
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "read the signing key")
			},
		},
		{
			name: "when the key dir cannot be made, it is an error",
			setup: func(dir string) *config.Config {
				locked := filepath.Join(dir, "locked")
				s.Require().NoError(s.fs.MkdirAll(locked, 0o500))

				return keyDirConfig(filepath.Join(locked, "pki"))
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "make the signing key")
			},
		},
		{
			name: "when the key dir cannot be written, it is an error",
			setup: func(dir string) *config.Config {
				keyDir := filepath.Join(dir, "ro")
				s.Require().NoError(s.fs.MkdirAll(keyDir, 0o500))

				return keyDirConfig(keyDir)
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "write the signing key")
			},
		},
		{
			name: "when no random bytes can be read, it is an error",
			setup: func(dir string) *config.Config {
				config.SetRandRead(func([]byte) (int, error) { return 0, errors.New("no entropy") })

				return keyDirConfig(filepath.Join(dir, "norand"))
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "make the signing key")
			},
		},
		{
			// A first start that made the file and failed to write it. A
			// controller that signed with an empty key would accept a token
			// that anybody can sign.
			name: "when the key file is empty, it is a short key",
			setup: func(dir string) *config.Config {
				return s.keyFileConfig(dir, nil)
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorIs(errs[0], config.ErrShortSigningKey)
			},
		},
		{
			name: "when the key file is short, it is a short key",
			setup: func(dir string) *config.Config {
				return s.keyFileConfig(dir, []byte("short\n"))
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorIs(errs[0], config.ErrShortSigningKey)
			},
		},
		{
			// A secret reference resolves after the config check.
			name: "when the key in the config is short, it is a short key",
			setup: func(_ string) *config.Config {
				var c config.Config
				c.Controller.API.Security.SigningKeys = []config.SigningKey{{ID: "main", Key: ""}}

				return &c
			},
			validateFunc: func(_ string, _ []*config.Config, errs []error) {
				s.ErrorIs(errs[0], config.ErrShortSigningKey)
			},
		},
		{
			name: "when processes make the key at the same time, each reads the same key and no temporary file is left",
			setup: func(dir string) *config.Config {
				return keyDirConfig(filepath.Join(dir, "pki"))
			},
			concurrent: 8,
			validateFunc: func(dir string, configs []*config.Config, errs []error) {
				for _, err := range errs {
					s.Require().NoError(err)
				}
				first := configs[0].Controller.API.Security.SigningKeys[0].Key
				for _, c := range configs[1:] {
					s.Equal(first, c.Controller.API.Security.SigningKeys[0].Key)
				}
				entries, err := s.fs.ReadDir(filepath.Join(dir, "pki"))
				s.Require().NoError(err)
				s.Len(entries, 1)
			},
		},
		{
			name: "when the temporary file cannot be written, it is an error and leaves no file",
			setup: func(dir string) *config.Config {
				config.SetSigningKeyFSFn(
					nil,
					func(string, []byte, os.FileMode) error { return boom },
					nil,
				)

				return keyDirConfig(filepath.Join(dir, "pki"))
			},
			validateFunc: func(dir string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "write the signing key")
				s.noTempFile(filepath.Join(dir, "pki"))
			},
		},
		{
			name: "when the link fails, it is an error and leaves no file",
			setup: func(dir string) *config.Config {
				config.SetSigningKeyFSFn(nil, nil, func(string, string) error { return boom })

				return keyDirConfig(filepath.Join(dir, "pki"))
			},
			validateFunc: func(dir string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "write the signing key")
				s.noTempFile(filepath.Join(dir, "pki"))
			},
		},
		{
			name: "when the key cannot be read after the link, it is an error",
			setup: func(dir string) *config.Config {
				config.SetSigningKeyFSFn(
					func(string) ([]byte, error) { return nil, os.ErrNotExist },
					nil,
					nil,
				)

				return keyDirConfig(filepath.Join(dir, "pki"))
			},
			validateFunc: func(dir string, _ []*config.Config, errs []error) {
				s.ErrorContains(errs[0], "read the signing key")
				s.noTempFile(filepath.Join(dir, "pki"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			dir := s.T().TempDir()
			n := max(tt.concurrent, 1)
			configs := make([]*config.Config, n)
			errs := make([]error, n)
			var wg sync.WaitGroup
			for i := range n {
				configs[i] = tt.setup(dir)
				wg.Go(func() {
					errs[i] = config.EnsureSigningKey(configs[i])
				})
			}
			wg.Wait()

			tt.validateFunc(dir, configs, errs)
		})
	}
}

// keyDirConfig is a config whose key dir is keyDir, with no key of its own.
func keyDirConfig(
	keyDir string,
) *config.Config {
	var c config.Config
	c.Controller.PKI.KeyDir = keyDir

	return &c
}

// keyFileConfig is a config whose key dir, in dir, holds a key file with
// content.
func (s *SigningKeyPublicTestSuite) keyFileConfig(
	dir string,
	content []byte,
) *config.Config {
	keyDir := filepath.Join(dir, "pki")
	s.Require().NoError(s.fs.MkdirAll(keyDir, 0o700))
	s.Require().
		NoError(s.fs.WriteFile(filepath.Join(keyDir, config.SigningKeyFile), content, 0o600))

	return keyDirConfig(keyDir)
}

// noTempFile fails when keyDir holds a temporary file of the key.
func (s *SigningKeyPublicTestSuite) noTempFile(
	keyDir string,
) {
	entries, _ := s.fs.ReadDir(keyDir)
	for _, e := range entries {
		s.NotContains(e.Name(), ".tmp")
	}
}

func TestSigningKeyPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SigningKeyPublicTestSuite))
}
