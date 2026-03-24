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

package agent

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
)

type SeedTestSuite struct {
	suite.Suite

	logger *slog.Logger
	ctx    context.Context
}

func (suite *SeedTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
}

func (suite *SeedTestSuite) TearDownTest() {
	embeddedFS = systemTemplates
	readEmbeddedFile = func(path string) ([]byte, error) {
		return systemTemplates.ReadFile(path)
	}
}

func (suite *SeedTestSuite) TestSeedSystemTemplates() {
	tests := []struct {
		name       string
		setupFunc  func()
		objStore   jetstream.ObjectStore
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "when WalkDir callback receives error",
			setupFunc: func() {
				// Create an FS where "templates" exists as a dir but
				// contains an entry that triggers a walk error.
				embeddedFS = &errorWalkFS{}
			},
			objStore:   &seedStubObjStore{},
			wantErr:    true,
			wantErrMsg: "walk error",
		},
		{
			name: "when ReadFile fails returns error",
			setupFunc: func() {
				readEmbeddedFile = func(_ string) ([]byte, error) {
					return nil, fmt.Errorf("read failure")
				}
			},
			objStore:   &seedStubObjStore{},
			wantErr:    true,
			wantErrMsg: "read embedded template",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			defer func() {
				embeddedFS = systemTemplates
				readEmbeddedFile = func(path string) ([]byte, error) {
					return systemTemplates.ReadFile(path)
				}
			}()

			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			err := SeedSystemTemplates(suite.ctx, suite.logger, tc.objStore)

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrMsg)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func TestSeedTestSuite(t *testing.T) {
	suite.Run(t, new(SeedTestSuite))
}

// errorWalkFS is an fs.FS that returns a walk error for the templates dir.
type errorWalkFS struct{}

func (e *errorWalkFS) Open(
	name string,
) (fs.File, error) {
	if name == "templates" {
		return &errorDir{}, nil
	}

	return nil, fmt.Errorf("walk error: %s", name)
}

// errorDir is an fs.File that is a directory but returns an error on ReadDir.
type errorDir struct{}

func (d *errorDir) Stat() (fs.FileInfo, error) {
	return &dirInfo{}, nil
}

func (d *errorDir) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("not a file")
}

func (d *errorDir) Close() error { return nil }

func (d *errorDir) ReadDir(_ int) ([]fs.DirEntry, error) {
	return nil, fmt.Errorf("walk error")
}

// dirInfo satisfies fs.FileInfo for a directory.
type dirInfo struct{}

func (i *dirInfo) Name() string      { return "templates" }
func (i *dirInfo) Size() int64       { return 0 }
func (i *dirInfo) Mode() fs.FileMode { return fs.ModeDir | 0o755 }
func (i *dirInfo) ModTime() time.Time  { return time.Time{} }
func (i *dirInfo) IsDir() bool         { return true }
func (i *dirInfo) Sys() interface{}    { return nil }

// seedStubObjStore is a minimal stub for internal seed tests.
type seedStubObjStore struct {
	jetstream.ObjectStore
}

func (s *seedStubObjStore) GetBytes(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectOpt,
) ([]byte, error) {
	return nil, fmt.Errorf("not found")
}

func (s *seedStubObjStore) PutBytes(
	_ context.Context,
	_ string,
	_ []byte,
) (*jetstream.ObjectInfo, error) {
	return nil, nil
}

func (s *seedStubObjStore) GetInfo(
	_ context.Context,
	_ string,
	_ ...jetstream.GetObjectInfoOpt,
) (*jetstream.ObjectInfo, error) {
	return nil, fmt.Errorf("not found")
}
