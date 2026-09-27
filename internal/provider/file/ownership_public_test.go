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

package file_test

import (
	"errors"
	"io/fs"
	"os"
	osuser "os/user"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/file"
)

// OwnershipPublicTestSuite covers the real ownership helpers against the running
// system, which is the only place a file's uid and gid come from.
type OwnershipPublicTestSuite struct {
	suite.Suite
}

func (s *OwnershipPublicTestSuite) TestOwnershipOf() {
	s.Run("a real file reports the running user", func() {
		path := filepath.Join(s.T().TempDir(), "f")
		s.Require().NoError(os.WriteFile(path, []byte("x"), 0o600))

		uid, gid, err := file.OwnershipOf(path)
		s.Require().NoError(err)
		s.Equal(os.Getuid(), uid)
		s.Equal(os.Getgid(), gid)
	})

	s.Run("a path that does not exist", func() {
		_, _, err := file.OwnershipOf(filepath.Join(s.T().TempDir(), "absent"))
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "stat")
	})
}

func (s *OwnershipPublicTestSuite) TestOwnershipOfUnreadablePlatform() {
	defer file.ResetOwnershipFuncs()

	// A FileInfo whose Sys() carries nothing this package can read, which is what
	// a platform without a Unix stat structure hands back.
	file.SetStatFn(func(string) (fs.FileInfo, error) {
		return syslessFileInfo{}, nil
	})

	_, _, err := file.OwnershipOf("/etc/app/app.conf")
	s.Require().ErrorIs(err, file.ErrOwnershipUnreadable)
}

func (s *OwnershipPublicTestSuite) TestOwnershipMatches() {
	defer file.ResetOwnershipFuncs()

	tests := []struct {
		name         string
		fileUID      int
		fileGID      int
		wantUID      int
		wantGID      int
		validateFunc func(bool, error)
	}{
		{
			name:    "both match",
			fileUID: 33, fileGID: 33,
			wantUID: 33, wantGID: 33,
			validateFunc: func(ok bool, err error) {
				s.Require().NoError(err)
				s.True(ok)
			},
		},
		{
			name:    "the owner differs",
			fileUID: 0, fileGID: 33,
			wantUID: 33, wantGID: 33,
			validateFunc: func(ok bool, err error) {
				s.Require().NoError(err)
				s.False(ok)
			},
		},
		{
			name:    "the group differs",
			fileUID: 33, fileGID: 0,
			wantUID: 33, wantGID: 33,
			validateFunc: func(ok bool, err error) {
				s.Require().NoError(err)
				s.False(ok)
			},
		},
		{
			name:    "an owner the request did not name is not compared",
			fileUID: 0, fileGID: 33,
			wantUID: file.UnsetID, wantGID: 33,
			validateFunc: func(ok bool, err error) {
				s.Require().NoError(err)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			uid, gid := tt.fileUID, tt.fileGID
			file.SetOwnershipFuncs(
				func(string) (int, int, error) { return uid, gid, nil },
				func(string) (int, error) { return 33, nil },
				func(string) (int, error) { return 33, nil },
			)

			tt.validateFunc(file.OwnershipMatches("/etc/app/app.conf", tt.wantUID, tt.wantGID))
		})
	}
}

// syslessFileInfo is a FileInfo with no platform stat structure behind it.
type syslessFileInfo struct{}

func (syslessFileInfo) Name() string       { return "app.conf" }
func (syslessFileInfo) Size() int64        { return 0 }
func (syslessFileInfo) Mode() fs.FileMode  { return 0o644 }
func (syslessFileInfo) ModTime() time.Time { return time.Time{} }
func (syslessFileInfo) IsDir() bool        { return false }
func (syslessFileInfo) Sys() any           { return nil }

func (s *OwnershipPublicTestSuite) TestLookup() {
	current, err := osuser.Current()
	s.Require().NoError(err)

	s.Run("the running user resolves to its own uid", func() {
		uid, lookupErr := file.LookupUID(current.Username)
		s.Require().NoError(lookupErr)
		s.Equal(current.Uid, strconv.Itoa(uid))
	})

	s.Run("a user the host does not have", func() {
		_, lookupErr := file.LookupUID("osapi-no-such-user")
		s.Require().Error(lookupErr)
	})

	s.Run("the running user's group resolves to its own gid", func() {
		group, groupErr := osuser.LookupGroupId(current.Gid)
		s.Require().NoError(groupErr)

		gid, lookupErr := file.LookupGroupID(group.Name)
		s.Require().NoError(lookupErr)
		s.Equal(current.Gid, strconv.Itoa(gid))
	})

	s.Run("a group the host does not have", func() {
		_, lookupErr := file.LookupGroupID("osapi-no-such-group")
		s.Require().Error(lookupErr)
	})
}

func (s *OwnershipPublicTestSuite) TestResolveID() {
	unknown := func(string) (int, error) { return 0, errors.New("no such name") }
	known := func(string) (int, error) { return 33, nil }

	tests := []struct {
		name         string
		input        string
		lookup       func(string) (int, error)
		validateFunc func(int, error)
	}{
		{
			name:   "a name the host knows",
			input:  "nginx",
			lookup: known,
			validateFunc: func(id int, err error) {
				s.Require().NoError(err)
				s.Equal(33, id)
			},
		},
		{
			name:   "a numeric id, which a container image without a passwd entry needs",
			input:  "1000",
			lookup: unknown,
			validateFunc: func(id int, err error) {
				s.Require().NoError(err)
				s.Equal(1000, id)
			},
		},
		{
			name:   "nothing requested",
			input:  "",
			lookup: unknown,
			validateFunc: func(id int, err error) {
				s.Require().NoError(err)
				s.Equal(file.UnsetID, id)
			},
		},
		{
			name:   "a name that is neither known nor numeric",
			input:  "ngnix",
			lookup: unknown,
			validateFunc: func(_ int, err error) {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "unknown owner or group")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(file.ResolveID(tt.input, tt.lookup))
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestOwnershipPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OwnershipPublicTestSuite))
}
