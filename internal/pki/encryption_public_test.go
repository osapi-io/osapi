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

package pki_test

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/fips140"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type EncryptionPublicTestSuite struct {
	suite.Suite
}

func TestEncryptionPublicTestSuite(t *testing.T) {
	suite.Run(t, new(EncryptionPublicTestSuite))
}

func (s *EncryptionPublicTestSuite) TearDownSubTest() {
	pki.ResetEncryptionKeyFns()
}

// failOn makes a failfs over base that fails fn on a path with suffix, or,
// with no suffix, fails fn when it creates a file.
func failOn(
	base avfs.VFS,
	fn avfs.FnVFS,
	suffix string,
) avfs.VFS {
	vfs := failfs.New(base)
	_ = vfs.SetFailFunc(func(
		_ avfs.VFSBase,
		got avfs.FnVFS,
		fp *failfs.FailParam,
	) error {
		switch {
		case got != fn:
			return nil
		case suffix == "" && fp.Flag&os.O_CREATE != 0:
			return errors.New("disk full")
		case suffix != "" && strings.HasSuffix(fp.Path, suffix):
			return errors.New("permission denied")
		}

		return nil
	})

	return vfs
}

// withSigningKey is a memfs that has an agent signing key and no
// encryption key.
func (s *EncryptionPublicTestSuite) withSigningKey() avfs.VFS {
	base := memfs.New()
	s.Require().NoError(pki.New(base, "/keys", "agent").LoadOrGenerate())

	return base
}

// pemOf is key as the PKCS #8 PEM block that the agent writes.
func (s *EncryptionPublicTestSuite) pemOf(
	key any,
) []byte {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	s.Require().NoError(err)

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func (s *EncryptionPublicTestSuite) TestLoadOrGenerate() {
	p384, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	_, ed, _ := ed25519.GenerateKey(rand.Reader)
	tests := []struct {
		name         string
		fs           func() avfs.VFS
		setupFn      func()
		wantContains string
		validateFunc func(m *pki.Manager, fs avfs.VFS)
	}{
		{
			name: "when no keys exist it makes both, and removes an old X25519 key",
			fs: func() avfs.VFS {
				base := s.withSigningKey()
				s.Require().NoError(base.WriteFile("/keys/agent.xkey", []byte("SXA..."), 0o600))

				return base
			},
			validateFunc: func(m *pki.Manager, fs avfs.VFS) {
				s.NotNil(m.EncryptionKey())
				s.True(pki.ValidEncryptionKey(m.EncryptionPublicKey()))
				data, err := fs.ReadFile("/keys/agent.ekey")
				s.Require().NoError(err)
				s.Contains(string(data), "BEGIN PRIVATE KEY")
				_, err = fs.Stat("/keys/agent.xkey")
				s.ErrorIs(err, os.ErrNotExist, "the X25519 key is gone")
			},
		},
		{
			name: "when the key exists it loads the same key",
			fs: func() avfs.VFS {
				base := s.withSigningKey()
				first := pki.New(base, "/keys", "agent", pki.WithEncryptionKey())
				s.Require().NoError(first.LoadOrGenerate())
				s.Require().NoError(base.WriteFile("/keys/first.pub",
					[]byte(first.EncryptionPublicKey()), 0o600))

				return base
			},
			validateFunc: func(m *pki.Manager, fs avfs.VFS) {
				want, _ := fs.ReadFile("/keys/first.pub")
				s.Equal(string(want), m.EncryptionPublicKey())
			},
		},
		{
			name:         "when the key file cannot be read it is not replaced",
			fs:           func() avfs.VFS { return failOn(s.withSigningKey(), avfs.FnReadFile, ".ekey") },
			wantContains: "does not replace a key it cannot read",
		},
		{
			name: "when the key file is not PEM",
			fs: func() avfs.VFS {
				base := s.withSigningKey()
				s.Require().NoError(base.WriteFile("/keys/agent.ekey", []byte("junk"), 0o600))

				return base
			},
			wantContains: "no PRIVATE KEY PEM block",
		},
		{
			name: "when the key file is not PKCS #8",
			fs: func() avfs.VFS {
				base := s.withSigningKey()
				s.Require().NoError(base.WriteFile("/keys/agent.ekey",
					pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("x")}), 0o600))

				return base
			},
			wantContains: "decode encryption key",
		},
		{
			name: "when the key file is a P-384 key",
			fs: func() avfs.VFS {
				base := s.withSigningKey()
				s.Require().NoError(base.WriteFile("/keys/agent.ekey", s.pemOf(p384), 0o600))

				return base
			},
			wantContains: "not a P-256 key",
		},
		{
			name: "when the key file is an Ed25519 key",
			fs: func() avfs.VFS {
				base := s.withSigningKey()
				s.Require().NoError(base.WriteFile("/keys/agent.ekey", s.pemOf(ed), 0o600))

				return base
			},
			wantContains: "not a P-256 key",
		},
		{
			name: "when making the key fails",
			fs:   s.withSigningKey,
			setupFn: func() {
				pki.SetGenerateEncryptionKeyFn(func() (*ecdh.PrivateKey, error) {
					return nil, errors.New("entropy exhausted")
				})
			},
			wantContains: "generate encryption key: entropy exhausted",
		},
		{
			name: "when encoding the key fails",
			fs:   s.withSigningKey,
			setupFn: func() {
				pki.SetMarshalEncryptionKeyFn(func(any) ([]byte, error) {
					return nil, errors.New("no encoding")
				})
			},
			wantContains: "encode encryption key: no encoding",
		},
		{
			name:         "when the key directory cannot be made",
			fs:           func() avfs.VFS { return failOn(s.withSigningKey(), avfs.FnMkdirAll, "/keys") },
			wantContains: "create key directory",
		},
		{
			name:         "when the key cannot be written",
			fs:           func() avfs.VFS { return failOn(s.withSigningKey(), avfs.FnOpenFile, "") },
			wantContains: "write encryption key",
		},
		{
			name:         "when the old X25519 key cannot be removed",
			fs:           func() avfs.VFS { return failOn(s.withSigningKey(), avfs.FnRemove, ".xkey") },
			wantContains: "remove the old encryption key",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setupFn != nil {
				tt.setupFn()
			}
			fs := tt.fs()
			m := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())
			err := m.LoadOrGenerate()
			if tt.wantContains != "" {
				s.ErrorContains(err, tt.wantContains)

				return
			}
			s.Require().NoError(err)
			tt.validateFunc(m, fs)
		})
	}
}

func (s *EncryptionPublicTestSuite) TestFingerprintOfKeys() {
	sig, _, _ := ed25519.GenerateKey(rand.Reader)
	enc, _ := ecdh.P256().GenerateKey(rand.Reader)
	encPub := pki.EncodeEncryptionKey(enc.PublicKey())
	h := sha256.New()
	h.Write(sig)
	h.Write(enc.PublicKey().Bytes())
	want := "SHA256:" + hex.EncodeToString(h.Sum(nil))

	s.Equal(want, pki.FingerprintOfKeys(sig, encPub),
		"the signing key, then the 65 bytes of the encryption key")
	s.Empty(pki.FingerprintOfKeys(nil, encPub), "no signing key")
	s.Empty(pki.FingerprintOfKeys(sig, ""), "no encryption key")
	s.Empty(pki.FingerprintOfKeys(sig, "junk!"), "an encryption key that is not one")
}

func (s *EncryptionPublicTestSuite) TestValidEncryptionKey() {
	enc, _ := ecdh.P256().GenerateKey(rand.Reader)
	var x *ecdh.PrivateKey
	fips140.WithoutEnforcement(func() { x, _ = ecdh.X25519().GenerateKey(rand.Reader) })
	s.True(pki.ValidEncryptionKey(pki.EncodeEncryptionKey(enc.PublicKey())))
	s.False(pki.ValidEncryptionKey(""))
	s.False(pki.ValidEncryptionKey("not base64!"))
	s.False(pki.ValidEncryptionKey(pki.EncryptionKeyEncoding.EncodeToString(x.PublicKey().Bytes())),
		"an X25519 key is not a P-256 point")
}

func (s *EncryptionPublicTestSuite) TestFingerprint() {
	s.Run("a manager with an encryption key covers both keys", func() {
		m := pki.New(memfs.New(), "/keys", "agent", pki.WithEncryptionKey())
		s.Require().NoError(m.LoadOrGenerate())
		s.Equal(pki.FingerprintOfKeys(m.PublicKey(), m.EncryptionPublicKey()), m.Fingerprint())
	})
	s.Run("a manager without one covers the signing key", func() {
		fs := memfs.New()
		m := pki.New(fs, "/keys", "controller")
		s.Require().NoError(m.LoadOrGenerate())
		s.Empty(m.EncryptionPublicKey())
		s.Nil(m.EncryptionKey())
		s.Equal(pki.FingerprintOf(m.PublicKey()), m.Fingerprint())
		_, err := fs.Stat("/keys/controller.ekey")
		s.ErrorIs(err, os.ErrNotExist)
	})
}
