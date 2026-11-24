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
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type XkeyPublicTestSuite struct {
	suite.Suite
}

func (suite *XkeyPublicTestSuite) TearDownSubTest() {
	pki.ResetCurveFns()
}

// failingReader is entropy that is not there.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy exhausted")
}

// noPublicKey is a curve key pair whose public key does not encode.
type noPublicKey struct {
	nkeys.KeyPair
}

func (noPublicKey) PublicKey() (string, error) {
	return "", errors.New("no public key")
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
// encryption key, as an agent enrolled before the encryption key has.
func (suite *XkeyPublicTestSuite) withSigningKey() avfs.VFS {
	base := memfs.New()
	require.NoError(suite.T(), pki.New(base, "/keys", "agent").LoadOrGenerate())

	return base
}

func (suite *XkeyPublicTestSuite) TestLoadOrGenerate() {
	tests := []struct {
		name         string
		fs           func() avfs.VFS
		setupFn      func()
		wantContains string
		validateFunc func(m *pki.Manager, fs avfs.VFS)
	}{
		{
			name: "when no keys exist generates both",
			fs:   func() avfs.VFS { return memfs.New() },
			validateFunc: func(m *pki.Manager, fs avfs.VFS) {
				pub := m.EncryptionPublicKey()
				assert.True(suite.T(), pki.ValidEncryptionKey(pub))
				assert.NotNil(suite.T(), m.EncryptionKey())
				assert.Equal(
					suite.T(),
					pki.FingerprintOfKeys(m.PublicKey(), pub),
					m.Fingerprint(),
				)
				assert.NotEqual(suite.T(), pki.FingerprintOf(m.PublicKey()), m.Fingerprint())

				data, err := fs.ReadFile("/keys/agent.xkey")
				require.NoError(suite.T(), err)
				assert.True(suite.T(), strings.HasPrefix(string(data), "SX"))
				info, err := fs.Stat("/keys/agent.xkey")
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), os.FileMode(0o600), info.Mode().Perm())
			},
		},
		{
			name: "when the keys exist loads the same encryption key",
			fs: func() avfs.VFS {
				base := memfs.New()
				m := pki.New(base, "/keys", "agent", pki.WithEncryptionKey())
				require.NoError(suite.T(), m.LoadOrGenerate())
				require.NoError(suite.T(), base.WriteFile(
					"/keys/want", []byte(m.EncryptionPublicKey()), 0o600,
				))
				return base
			},
			validateFunc: func(m *pki.Manager, fs avfs.VFS) {
				want, err := fs.ReadFile("/keys/want")
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), string(want), m.EncryptionPublicKey())
			},
		},
		{
			name: "when only the signing key exists generates the encryption key next to it",
			fs:   suite.withSigningKey,
			validateFunc: func(m *pki.Manager, fs avfs.VFS) {
				signing := pki.New(fs, "/keys", "agent")
				require.NoError(suite.T(), signing.LoadOrGenerate())
				assert.Equal(suite.T(), signing.PublicKey(), m.PublicKey())
				assert.True(suite.T(), pki.ValidEncryptionKey(m.EncryptionPublicKey()))
			},
		},
		{
			name: "when the encryption key cannot be read it is not replaced",
			fs: func() avfs.VFS {
				base := memfs.New()
				m := pki.New(base, "/keys", "agent", pki.WithEncryptionKey())
				require.NoError(suite.T(), m.LoadOrGenerate())
				return failOn(base, avfs.FnReadFile, ".xkey")
			},
			wantContains: "does not replace a key it cannot read",
		},
		{
			name: "when the encryption key file is not a seed",
			fs: func() avfs.VFS {
				base := suite.withSigningKey()
				require.NoError(
					suite.T(),
					base.WriteFile("/keys/agent.xkey", []byte("junk"), 0o600),
				)
				return base
			},
			wantContains: "decode encryption key /keys/agent.xkey",
		},
		{
			name: "when the encryption key file is not a curve seed",
			fs: func() avfs.VFS {
				base := suite.withSigningKey()
				user, err := nkeys.CreateUser()
				require.NoError(suite.T(), err)
				seed, err := user.Seed()
				require.NoError(suite.T(), err)
				require.NoError(suite.T(), base.WriteFile("/keys/agent.xkey", seed, 0o600))
				return base
			},
			wantContains: "decode encryption key /keys/agent.xkey",
		},
		{
			name: "when the signing key does not load nothing else is made",
			fs: func() avfs.VFS {
				base := memfs.New()
				require.NoError(suite.T(), base.MkdirAll("/keys", 0o700))
				require.NoError(suite.T(), base.WriteFile("/keys/agent.key", []byte("x"), 0o600))
				return base
			},
			wantContains: "half a keypair",
			validateFunc: func(_ *pki.Manager, fs avfs.VFS) {
				_, err := fs.Stat("/keys/agent.xkey")
				assert.ErrorIs(suite.T(), err, os.ErrNotExist)
			},
		},
		{
			name: "when the entropy fails",
			fs:   func() avfs.VFS { return memfs.New() },
			setupFn: func() {
				pki.SetCurveRand(failingReader{})
			},
			wantContains: "generate encryption key",
		},
		{
			name: "when generating the seed fails",
			fs:   func() avfs.VFS { return memfs.New() },
			setupFn: func() {
				pki.SetGenerateCurveSeedFn(func() ([]byte, error) {
					return nil, errors.New("entropy exhausted")
				})
			},
			wantContains: "generate encryption key",
		},
		{
			name: "when the generated seed does not decode",
			fs:   func() avfs.VFS { return memfs.New() },
			setupFn: func() {
				pki.SetGenerateCurveSeedFn(func() ([]byte, error) {
					return []byte("junk"), nil
				})
			},
			wantContains: "decode encryption key",
		},
		{
			name: "when the public key does not encode",
			fs:   func() avfs.VFS { return memfs.New() },
			setupFn: func() {
				pki.SetFromCurveSeedFn(func(seed []byte) (nkeys.KeyPair, error) {
					kp, err := nkeys.FromCurveSeed(seed)
					return noPublicKey{kp}, err
				})
			},
			wantContains: "no public key",
		},
		{
			name: "when the key directory cannot be made",
			fs: func() avfs.VFS {
				return failOn(suite.withSigningKey(), avfs.FnMkdirAll, "/keys")
			},
			wantContains: "create key directory",
		},
		{
			name: "when the encryption key cannot be written",
			fs: func() avfs.VFS {
				return failOn(suite.withSigningKey(), avfs.FnOpenFile, "")
			},
			wantContains: "write encryption key",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			fs := tc.fs()
			if tc.setupFn != nil {
				tc.setupFn()
			}
			m := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())

			err := m.LoadOrGenerate()

			if tc.wantContains != "" {
				require.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.wantContains)
			} else {
				require.NoError(suite.T(), err)
			}
			if tc.validateFunc != nil {
				tc.validateFunc(m, fs)
			}
		})
	}
}

func (suite *XkeyPublicTestSuite) TestWithoutEncryptionKey() {
	tests := []struct {
		name string
	}{
		{name: "when the manager has no encryption key the fingerprint covers the signing key"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			fs := memfs.New()
			m := pki.New(fs, "/keys", "controller")

			require.NoError(suite.T(), m.LoadOrGenerate())

			assert.Empty(suite.T(), m.EncryptionPublicKey())
			assert.Nil(suite.T(), m.EncryptionKey())
			assert.Equal(suite.T(), pki.FingerprintOf(m.PublicKey()), m.Fingerprint())
			_, err := fs.Stat("/keys/controller.xkey")
			assert.ErrorIs(suite.T(), err, os.ErrNotExist)
		})
	}
}

func (suite *XkeyPublicTestSuite) TestFingerprintOfKeys() {
	curve, err := nkeys.CreateCurveKeys()
	require.NoError(suite.T(), err)
	encPub, err := curve.PublicKey()
	require.NoError(suite.T(), err)
	raw, err := nkeys.Decode(nkeys.PrefixByteCurve, []byte(encPub))
	require.NoError(suite.T(), err)
	signPub, _, err := ed25519.GenerateKey(nil)
	require.NoError(suite.T(), err)
	sum := sha256.Sum256(append(append([]byte{}, signPub...), raw...))

	tests := []struct {
		name       string
		signingKey ed25519.PublicKey
		encryption string
		want       string
	}{
		{
			name:       "when both keys are valid hashes the signing key then the encryption key",
			signingKey: signPub,
			encryption: encPub,
			want:       "SHA256:" + hex.EncodeToString(sum[:]),
		},
		{
			name:       "when the signing key is empty",
			encryption: encPub,
		},
		{
			name:       "when the encryption key is empty",
			signingKey: signPub,
		},
		{
			name:       "when the encryption key is not a curve key",
			signingKey: signPub,
			encryption: "UABC",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			assert.Equal(suite.T(), tc.want, pki.FingerprintOfKeys(tc.signingKey, tc.encryption))
		})
	}
}

func (suite *XkeyPublicTestSuite) TestValidEncryptionKey() {
	curve, err := nkeys.CreateCurveKeys()
	require.NoError(suite.T(), err)
	encPub, err := curve.PublicKey()
	require.NoError(suite.T(), err)

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{name: "when the key is a public curve key", key: encPub, want: true},
		{name: "when the key is empty", key: ""},
		{name: "when the key is not a curve key", key: "junk"},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			assert.Equal(suite.T(), tc.want, pki.ValidEncryptionKey(tc.key))
		})
	}
}

func TestXkeyPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(XkeyPublicTestSuite))
}
