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

package fsutil_test

import (
	"fmt"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/fsutil"
)

type AtomicPublicTestSuite struct {
	suite.Suite
}

func (s *AtomicPublicTestSuite) TestWriteFileAtomic() {
	const target = "/etc/app/app.conf"

	tests := []struct {
		name         string
		setupFS      func() avfs.VFS
		content      []byte
		validateFunc func(avfs.VFS, error)
	}{
		{
			name: "a new file is written with the requested mode",
			setupFS: func() avfs.VFS {
				vfs := memfs.New()
				_ = vfs.MkdirAll("/etc/app", 0o755)

				return vfs
			},
			content: []byte("listen 80\n"),
			validateFunc: func(vfs avfs.VFS, err error) {
				s.Require().NoError(err)

				content, readErr := vfs.ReadFile(target)
				s.Require().NoError(readErr)
				s.Equal("listen 80\n", string(content))

				info, statErr := vfs.Stat(target)
				s.Require().NoError(statErr)
				s.Equal("-rw-------", info.Mode().Perm().String())
			},
		},
		{
			name: "an existing file is replaced rather than truncated",
			setupFS: func() avfs.VFS {
				vfs := memfs.New()
				_ = vfs.MkdirAll("/etc/app", 0o755)
				_ = vfs.WriteFile(target, []byte("listen 8080\n"), 0o644)

				return vfs
			},
			content: []byte("listen 80\n"),
			validateFunc: func(vfs avfs.VFS, err error) {
				s.Require().NoError(err)

				content, readErr := vfs.ReadFile(target)
				s.Require().NoError(readErr)
				s.Equal("listen 80\n", string(content))
			},
		},
		{
			name: "the temp file cannot be created",
			setupFS: func() avfs.VFS {
				return failfsFailing(avfs.FnOpenFile)
			},
			content: []byte("listen 80\n"),
			validateFunc: func(_ avfs.VFS, err error) {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "create temp file")
			},
		},
		{
			name: "the content cannot be written",
			setupFS: func() avfs.VFS {
				return failfsFailing(avfs.FnFileWrite)
			},
			content: []byte("listen 80\n"),
			validateFunc: func(vfs avfs.VFS, err error) {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "write temp file")
				s.assertNoLeftovers(vfs)
			},
		},
		{
			name: "the mode cannot be set",
			setupFS: func() avfs.VFS {
				return failfsFailing(avfs.FnChmod)
			},
			content: []byte("listen 80\n"),
			validateFunc: func(vfs avfs.VFS, err error) {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "set mode on")
				s.assertNoLeftovers(vfs)
			},
		},
		{
			name: "the temp file cannot be closed",
			setupFS: func() avfs.VFS {
				return failfsFailing(avfs.FnFileClose)
			},
			content: []byte("listen 80\n"),
			validateFunc: func(_ avfs.VFS, err error) {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "close temp file")
			},
		},
		{
			name: "the rename over the target fails",
			setupFS: func() avfs.VFS {
				return failfsFailing(avfs.FnRename)
			},
			content: []byte("listen 80\n"),
			validateFunc: func(vfs avfs.VFS, err error) {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "rename")
				s.assertNoLeftovers(vfs)

				// The target is untouched, which is the point of the rename.
				_, statErr := vfs.Stat(target)
				s.Require().Error(statErr)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			vfs := tt.setupFS()

			tt.validateFunc(vfs, fsutil.WriteFileAtomic(vfs, target, tt.content, 0o600))
		})
	}
}

// assertNoLeftovers checks that a failed write left no temporary file behind.
func (s *AtomicPublicTestSuite) assertNoLeftovers(
	vfs avfs.VFS,
) {
	entries, err := vfs.ReadDir("/etc/app")
	s.Require().NoError(err)

	for _, e := range entries {
		s.NotContains(e.Name(), ".osapi-")
	}
}

// failfsFailing returns a filesystem where one operation fails and the rest
// behave normally.
func failfsFailing(
	failing avfs.FnVFS,
) avfs.VFS {
	vfs := failfs.New(memfs.New())
	_ = vfs.MkdirAll("/etc/app", 0o755)

	_ = vfs.SetFailFunc(func(
		_ avfs.VFSBase,
		fn avfs.FnVFS,
		_ *failfs.FailParam,
	) error {
		if fn == failing {
			return fmt.Errorf("injected %v error", fn)
		}

		return nil
	})

	return vfs
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestAtomicPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AtomicPublicTestSuite))
}
