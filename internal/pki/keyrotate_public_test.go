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
	"crypto/rand"
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

func (s *KeyRotatePublicTestSuite) TestResume() {
	moves := [][2]string{
		{"/keys/agent.next.ekey", "/keys/agent.ekey"},
		{"/keys/agent.next.key", "/keys/agent.key"},
		{"/keys/agent.next.pub", "/keys/agent.pub"},
	}
	tests := []struct {
		name      string
		moved     []int
		removed   string
		noNext    bool
		other     bool
		fail      avfs.FnVFS
		suffix    string
		garbage   string
		wantMoved bool
		wantErr   string
	}{
		{name: "with no next keys, it does nothing", noNext: true},
		{name: "with each next key beside the keys, it does nothing"},
		{
			name:  "after the encryption key moved, it moves the others",
			moved: []int{0}, wantMoved: true,
		},
		{
			name:  "after the private key moved, it moves the public key",
			moved: []int{0, 1}, wantMoved: true,
		},
		{
			name:    "a next key that is gone after one that is not is not a Promote",
			removed: "/keys/agent.next.key",
		},
		{
			name:  "keys that are not the pair of the next key",
			moved: []int{0, 1}, other: true,
			wantErr: "are halfway through a rotation, and are not the pair of the next key",
		},
		{
			name: "a key that does not read",
			fail: avfs.FnStat, wantErr: "read key /keys/agent.next.ekey",
		},
		{
			name:  "a private key whose file does not read",
			moved: []int{0}, fail: avfs.FnReadFile, suffix: "agent.next.key",
			wantErr: "are not the pair of the next key",
		},
		{
			name:  "a public key whose file does not read",
			moved: []int{0}, fail: avfs.FnReadFile, suffix: "agent.next.pub",
			wantErr: "are not the pair of the next key",
		},
		{
			name:  "a public key that does not decode",
			moved: []int{0}, garbage: "/keys/agent.next.pub",
			wantErr: "are not the pair of the next key",
		},
		{
			name:  "a key that does not move",
			moved: []int{0}, fail: avfs.FnRename,
			wantErr: "move /keys/agent.next.key to /keys/agent.key",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			base, m := made(s, pki.WithEncryptionKey())
			want := m.PublicKey()
			if !tt.noNext {
				next, err := m.Next()
				s.Require().NoError(err)
				want = next.PublicKey()
			}
			for _, i := range tt.moved {
				s.Require().NoError(base.Rename(moves[i][0], moves[i][1]))
			}
			if tt.removed != "" {
				s.Require().NoError(base.Remove(tt.removed))
			}
			if tt.garbage != "" {
				s.Require().NoError(base.WriteFile(tt.garbage, []byte("nope"), 0o600))
			}
			if tt.other {
				pub, _, err := ed25519.GenerateKey(rand.Reader)
				s.Require().NoError(err)
				want = pub
			}
			fs := base
			if tt.fail != 0 {
				fs = failing(base, tt.fail, tt.suffix)
			}
			resumed := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())
			moved, err := resumed.Resume(want)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.wantMoved, moved)
			if !tt.wantMoved {
				return
			}
			s.Equal(want, resumed.PublicKey())
			loaded := pki.New(base, "/keys", "agent", pki.WithEncryptionKey())
			s.Require().NoError(loaded.Load())
			s.Equal(want, loaded.PublicKey())
			for _, mv := range moves {
				_, err := base.Stat(mv[0])
				s.Error(err)
			}
		})
	}
}

func (s *KeyRotatePublicTestSuite) TestResumeNext() {
	moves := [][2]string{
		{"/keys/agent.next.ekey", "/keys/agent.ekey"},
		{"/keys/agent.next.key", "/keys/agent.key"},
		{"/keys/agent.next.pub", "/keys/agent.pub"},
	}
	tests := []struct {
		name      string
		moved     []int
		noNext    bool
		garbage   bool
		fail      bool
		wantMoved bool
		wantErr   string
	}{
		{name: "with no next keys, it does nothing", noNext: true},
		{name: "with each next key beside the keys, it does nothing"},
		{
			name:  "after the private key moved, it moves the public key",
			moved: []int{0, 1}, wantMoved: true,
		},
		{
			name:  "after the encryption key moved, it moves the others",
			moved: []int{0}, wantMoved: true,
		},
		{name: "with each key moved, it does nothing", moved: []int{0, 1, 2}},
		{
			name: "a next public key that does not decode", garbage: true,
			wantErr: "decode key /keys/agent.next.pub",
		},
		{
			name: "a next public key that does not read", fail: true,
			wantErr: "read key /keys/agent.next.pub",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			base, m := made(s, pki.WithEncryptionKey())
			want := m.PublicKey()
			if !tt.noNext {
				next, err := m.Next()
				s.Require().NoError(err)
				want = next.PublicKey()
			}
			for _, i := range tt.moved {
				s.Require().NoError(base.Rename(moves[i][0], moves[i][1]))
			}
			if tt.garbage {
				s.Require().NoError(base.WriteFile("/keys/agent.next.pub", []byte("nope"), 0o600))
			}
			fs := base
			if tt.fail {
				fs = failing(base, avfs.FnReadFile, "agent.next.pub")
			}
			resumed := pki.New(fs, "/keys", "agent", pki.WithEncryptionKey())
			moved, err := resumed.ResumeNext()
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.wantMoved, moved)
			loaded := pki.New(base, "/keys", "agent", pki.WithEncryptionKey())
			s.Require().NoError(loaded.Load(), "the keys are a pair")
			if tt.wantMoved {
				s.Equal(want, loaded.PublicKey())
			}
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
