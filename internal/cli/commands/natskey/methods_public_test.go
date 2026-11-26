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

package natskey_test

import (
	"context"
	"encoding/hex"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/avfs/avfs/vfs/rofs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/natskey"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

// failFS is a file system whose OpenFile, or the Write or Close of the file
// that it opens, fails.
type failFS struct {
	avfs.VFS
	open      bool
	write     bool
	failClose bool
}

func (f *failFS) OpenFile(
	name string,
	flag int,
	perm fs.FileMode,
) (avfs.File, error) {
	if f.open {
		return nil, errors.New("no open")
	}
	file, err := f.VFS.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return &failFile{File: file, write: f.write, failClose: f.failClose}, nil
}

// failFile is a file whose Write or Close fails.
type failFile struct {
	avfs.File
	write     bool
	failClose bool
}

func (f *failFile) Write(
	b []byte,
) (int, error) {
	if f.write {
		return 0, errors.New("no write")
	}

	return f.File.Write(b)
}

func (f *failFile) Close() error {
	err := f.File.Close()
	if f.failClose {
		return errors.New("no close")
	}

	return err
}

func (s *MethodsPublicTestSuite) TestCreate() {
	const keyFile = "/etc/osapi/nats/jetstream.key"

	tests := []struct {
		name     string
		fs       func() avfs.VFS
		keyFile  string
		randRead func([]byte) (int, error)
		wantErr  string
	}{
		{
			name:    "when the file does not exist writes a key",
			fs:      func() avfs.VFS { return memfs.New() },
			keyFile: keyFile,
		},
		{
			name:    "when no key file is set returns the error",
			fs:      func() avfs.VFS { return memfs.New() },
			wantErr: "nats.server.jetstream.key_file is not set",
		},
		{
			name: "when the file exists refuses to write it",
			fs: func() avfs.VFS {
				fs := memfs.New()
				s.Require().NoError(fs.MkdirAll("/etc/osapi/nats", 0o700))
				s.Require().NoError(fs.WriteFile(keyFile, []byte("old\n"), 0o600))
				return fs
			},
			keyFile: keyFile,
			wantErr: "exists: a new key in its place makes each stream unreadable",
		},
		{
			name: "when the file cannot be checked returns the error",
			fs: func() avfs.VFS {
				fs := memfs.New()
				s.Require().NoError(fs.MkdirAll("/etc/osapi", 0o700))
				s.Require().NoError(fs.WriteFile("/etc/osapi/nats", []byte("a file"), 0o600))
				return fs
			},
			keyFile: keyFile,
			wantErr: "check " + keyFile,
		},
		{
			name:     "when no random bytes can be read returns the error",
			fs:       func() avfs.VFS { return memfs.New() },
			keyFile:  keyFile,
			randRead: func([]byte) (int, error) { return 0, errors.New("no entropy") },
			wantErr:  "make the key: no entropy",
		},
		{
			name:    "when the directory cannot be made returns the error",
			fs:      func() avfs.VFS { return rofs.New(memfs.New()) },
			keyFile: keyFile,
			wantErr: "make the directory of " + keyFile,
		},
		{
			name:    "when the file cannot be opened returns the error",
			fs:      func() avfs.VFS { return &failFS{VFS: memfs.New(), open: true} },
			keyFile: keyFile,
			wantErr: "write " + keyFile + ": no open",
		},
		{
			name:    "when the key cannot be written returns the error",
			fs:      func() avfs.VFS { return &failFS{VFS: memfs.New(), write: true} },
			keyFile: keyFile,
			wantErr: "write " + keyFile + ": no write",
		},
		{
			name:    "when the file cannot be closed returns the error",
			fs:      func() avfs.VFS { return &failFS{VFS: memfs.New(), failClose: true} },
			keyFile: keyFile,
			wantErr: "write " + keyFile + ": no close",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.randRead != nil {
				natskey.SetRandRead(tc.randRead)
				defer natskey.ResetRandRead()
			}
			vfs := tc.fs()

			var entries []audit.Entry
			got, err := natskey.Create(
				context.Background(),
				natskey.Deps{FS: vfs, KeyFile: tc.keyFile, Audit: func(
					_ context.Context, e audit.Entry,
				) error {
					entries = append(entries, e)

					return nil
				}},
				natskey.CreateInput{},
			)
			// A key that was made is a local change in the audit log, with its
			// file and not the key (osapi-io/osapi#1453).
			if err == nil {
				s.Require().Len(entries, 1)
				s.Equal("natskey.create", entries[0].Action)
				s.Equal(tc.keyFile, entries[0].Target)
			} else {
				s.Empty(entries)
			}

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				s.Empty(got.KeyFile)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.keyFile, got.KeyFile)

			b, err := vfs.ReadFile(tc.keyFile)
			s.Require().NoError(err)
			raw, err := hex.DecodeString(strings.TrimSpace(string(b)))
			s.Require().NoError(err)
			s.Len(raw, 32)

			info, err := vfs.Stat(tc.keyFile)
			s.Require().NoError(err)
			s.Equal("-rw-------", info.Mode().Perm().String())

			dir, err := vfs.Stat("/etc/osapi/nats")
			s.Require().NoError(err)
			s.Equal("-rwx------", dir.Mode().Perm().String())
		})
	}
}
