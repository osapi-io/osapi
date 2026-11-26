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
	"errors"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/pki"
)

type KeyRotatePublicTestSuite struct {
	suite.Suite
}

func TestKeyRotatePublicTestSuite(t *testing.T) {
	suite.Run(t, new(KeyRotatePublicTestSuite))
}

// failing is fs, failing fn on a path that ends with suffix.
func failing(
	fs avfs.VFS,
	fn avfs.FnVFS,
	suffix string,
) avfs.VFS {
	vfs := failfs.New(fs)
	_ = vfs.SetFailFunc(func(_ avfs.VFSBase, f avfs.FnVFS, fp *failfs.FailParam) error {
		if f == fn && strings.HasSuffix(fp.Path, suffix) {
			return errors.New("permission denied")
		}

		return nil
	})

	return vfs
}

// made is a file system with the agent's keys, and the manager that made
// them.
func made(
	s *KeyRotatePublicTestSuite,
	opts ...pki.Option,
) (avfs.VFS, *pki.Manager) {
	fs := memfs.New()
	m := pki.New(fs, "/keys", "agent", opts...)
	s.Require().NoError(m.LoadOrGenerate())

	return fs, m
}

func (s *KeyRotatePublicTestSuite) TestLoad() {
	tests := []struct {
		name    string
		fs      func() avfs.VFS
		opts    []pki.Option
		wantErr string
		isNone  bool
	}{
		{
			name: "loads the keys on the host",
			fs: func() avfs.VFS {
				fs, _ := made(s, pki.WithEncryptionKey())

				return fs
			},
			opts: []pki.Option{pki.WithEncryptionKey()},
		},
		{
			name: "loads the signing keys of a manager with no encryption key",
			fs: func() avfs.VFS {
				fs, _ := made(s)

				return fs
			},
		},
		{
			name:    "the keys are not there",
			fs:      func() avfs.VFS { return memfs.New() },
			wantErr: "the keys are not on this host",
			isNone:  true,
		},
		{
			name: "a key cannot be read",
			fs: func() avfs.VFS {
				fs, _ := made(s)

				return failing(fs, avfs.FnReadFile, ".pub")
			},
			wantErr: "read key /keys/agent.pub",
		},
		{
			name: "the private key is not a key",
			fs: func() avfs.VFS {
				fs, _ := made(s)
				s.Require().NoError(fs.WriteFile("/keys/agent.key", []byte("no"), 0o600))

				return fs
			},
			wantErr: "decode private key PEM",
		},
		{
			name: "the private key is not the pair of the public key",
			fs: func() avfs.VFS {
				fs, _ := made(s)
				other, _ := made(s)
				pub, _ := other.ReadFile("/keys/agent.pub")
				s.Require().NoError(fs.WriteFile("/keys/agent.pub", pub, 0o644))

				return fs
			},
			wantErr: "is not the pair of",
		},
		{
			name: "the encryption key is not a key",
			fs: func() avfs.VFS {
				fs, _ := made(s, pki.WithEncryptionKey())
				s.Require().NoError(fs.WriteFile("/keys/agent.ekey", []byte("no"), 0o600))

				return fs
			},
			opts:    []pki.Option{pki.WithEncryptionKey()},
			wantErr: "decode encryption key /keys/agent.ekey",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			m := pki.New(tt.fs(), "/keys", "agent", tt.opts...)
			err := m.Load()
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Equal(tt.isNone, errors.Is(err, pki.ErrNoKeys))

				return
			}
			s.Require().NoError(err)
			s.NotEmpty(m.Fingerprint())
		})
	}
}

func (s *KeyRotatePublicTestSuite) TestReload() {
	tests := []struct {
		name        string
		change      func(fs avfs.VFS)
		opts        []pki.Option
		wantChanged bool
		wantErr     string
		wantPrev    bool
	}{
		{
			name:   "keys that did not change",
			change: func(avfs.VFS) {},
			opts:   []pki.Option{pki.WithEncryptionKey()},
		},
		{
			name: "a rotation replaced both keys",
			change: func(fs avfs.VFS) {
				m := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())
				s.Require().NoError(m.Load())
				next, err := m.Next()
				s.Require().NoError(err)
				s.Require().NoError(m.Promote(next))
			},
			opts:        []pki.Option{pki.WithEncryptionKey()},
			wantChanged: true,
			wantPrev:    true,
		},
		{
			name: "a rotation replaced the signing key of a manager with no encryption key",
			change: func(fs avfs.VFS) {
				m := pki.New(fs, "/keys", "agent")
				s.Require().NoError(m.Load())
				next, err := m.Next()
				s.Require().NoError(err)
				s.Require().NoError(m.Promote(next))
			},
			wantChanged: true,
		},
		{
			name: "the keys are halfway through the renames of a rotation",
			change: func(fs avfs.VFS) {
				otherFS, _ := made(s)
				priv, _ := otherFS.ReadFile("/keys/agent.key")
				s.Require().NoError(fs.WriteFile("/keys/agent.key", priv, 0o600))
			},
			opts:    []pki.Option{pki.WithEncryptionKey()},
			wantErr: "is not the pair of",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			fs, m := made(s, tt.opts...)
			before := m.Fingerprint()
			tt.change(fs)
			changed, err := m.Reload()
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.False(changed)
				s.Equal(before, m.Fingerprint())

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.wantChanged, changed)
			s.Equal(tt.wantChanged, before != m.Fingerprint())
			s.Equal(tt.wantPrev, m.PreviousEncryptionKey() != nil)
		})
	}
}

func (s *KeyRotatePublicTestSuite) TestNext() {
	tests := []struct {
		name    string
		fs      func(fs avfs.VFS) avfs.VFS
		wantErr string
	}{
		{
			name: "makes new keys, then gives the same keys again",
			fs:   func(fs avfs.VFS) avfs.VFS { return fs },
		},
		{
			name:    "the new keys cannot be read or written",
			fs:      func(fs avfs.VFS) avfs.VFS { return failing(fs, avfs.FnOpenFile, ".next.key") },
			wantErr: "/keys/agent.next.key",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			fs, _ := made(s, pki.WithEncryptionKey())
			m := pki.New(tt.fs(fs), "/keys", "agent", pki.WithEncryptionKey())
			s.Require().NoError(m.Load())
			next, err := m.Next()
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.Require().NoError(err)
			s.NotEqual(m.Fingerprint(), next.Fingerprint())
			again, err := m.Next()
			s.Require().NoError(err)
			s.Equal(next.Fingerprint(), again.Fingerprint())
			_, err = fs.Stat("/keys/agent.next.ekey")
			s.NoError(err)
		})
	}
}

func (s *KeyRotatePublicTestSuite) TestPromote() {
	tests := []struct {
		name    string
		fail    string
		wantErr string
	}{
		{name: "moves the new keys in place of the old ones"},
		{
			name:    "a key cannot be moved",
			fail:    ".next.ekey",
			wantErr: "move /keys/agent.next.ekey to /keys/agent.ekey",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			base, _ := made(s, pki.WithEncryptionKey())
			fs := base
			if tt.fail != "" {
				fs = failing(base, avfs.FnRename, tt.fail)
			}
			m := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())
			s.Require().NoError(m.Load())
			next, err := m.Next()
			s.Require().NoError(err)
			err = m.Promote(next)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(next.Fingerprint(), m.Fingerprint())
			loaded := pki.New(base, "/keys", "agent", pki.WithEncryptionKey())
			s.Require().NoError(loaded.Load())
			s.Equal(next.Fingerprint(), loaded.Fingerprint())
			_, err = base.Stat("/keys/agent.next.key")
			s.Error(err)
		})
	}
}

func (s *KeyRotatePublicTestSuite) TestDrop() {
	tests := []struct {
		name    string
		fail    bool
		wantErr string
	}{
		{name: "removes the keys"},
		{name: "a key cannot be removed", fail: true, wantErr: "remove /keys/agent.next.key"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			base, _ := made(s, pki.WithEncryptionKey())
			fs := base
			if tt.fail {
				fs = failing(base, avfs.FnRemove, ".next.key")
			}
			m := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())
			s.Require().NoError(m.Load())
			next, err := m.Next()
			s.Require().NoError(err)
			err = next.Drop()
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.Require().NoError(err)
			for _, f := range []string{"key", "pub", "ekey"} {
				_, err := base.Stat("/keys/agent.next." + f)
				s.Error(err, f)
			}
			s.NoError(next.Drop(), "a second drop finds nothing to remove")
		})
	}
}
