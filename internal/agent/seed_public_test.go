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

package agent_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"testing"
	"testing/fstest"
	"time"

	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
)

// SeedPublicTestSuite tests the exported SeedSystemTemplates function.
type SeedPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	logger *slog.Logger
}

func (s *SeedPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = slog.Default()
}

func (s *SeedPublicTestSuite) TearDownSubTest() {
	agent.ResetEmbeddedFS()
	agent.ResetReadEmbeddedFile()
}

var testTemplateData = []byte("#!/bin/sh\necho hello\n")

// testTemplateFS returns a fake fs.FS with a single template file.
func testTemplateFS() fstest.MapFS {
	return fstest.MapFS{
		"resources/templates/osapi/test.tmpl": &fstest.MapFile{
			Data: testTemplateData,
		},
	}
}

// setupTestTemplateFS sets both the embedded FS and the read function
// to use the fake template FS.
func setupTestTemplateFS() {
	fakeFS := testTemplateFS()
	agent.SetEmbeddedFS(fakeFS)
	agent.SetReadEmbeddedFile(func(path string) ([]byte, error) {
		return fs.ReadFile(fakeFS, path)
	})
}

func (s *SeedPublicTestSuite) TestSeedSystemTemplates() {
	templateData := testTemplateData

	tests := []struct {
		name         string
		setupFunc    func()
		setupMock    func(ctrl *gomock.Controller, mockObj *agentMocks.MockTemplateStore, putNames *[]string)
		validateFunc func(error, *[]string)
	}{
		{
			name: "when WalkDir callback receives error",
			setupFunc: func() {
				agent.SetEmbeddedFS(&errorWalkFS{})
			},
			setupMock: func(
				_ *gomock.Controller,
				_ *agentMocks.MockTemplateStore,
				_ *[]string,
			) {
			},
			validateFunc: func(err error, putNames *[]string) {
				s.Error(err)
				s.Contains(err.Error(), "walk error")

				s.Len(*putNames, 0)
			},
		},
		{
			name: "when directory contains only .gitkeep skips it",
			setupFunc: func() {
				agent.SetEmbeddedFS(fstest.MapFS{
					"resources/templates/.gitkeep": &fstest.MapFile{Data: []byte{}},
				})
			},
			setupMock: func(
				_ *gomock.Controller,
				_ *agentMocks.MockTemplateStore,
				_ *[]string,
			) {
				// No ObjectGetBytes or ObjectPutBytes calls expected — .gitkeep is skipped.
			},
			validateFunc: func(err error, putNames *[]string) {
				s.NoError(err)

				s.Len(*putNames, 0)
			},
		},
		{
			name: "when ReadFile fails returns error",
			setupFunc: func() {
				setupTestTemplateFS()
				agent.SetReadEmbeddedFile(func(_ string) ([]byte, error) {
					return nil, fmt.Errorf("read failure")
				})
			},
			setupMock: func(
				_ *gomock.Controller,
				_ *agentMocks.MockTemplateStore,
				_ *[]string,
			) {
			},
			validateFunc: func(err error, putNames *[]string) {
				s.Error(err)
				s.Contains(err.Error(), "read embedded template")

				s.Len(*putNames, 0)
			},
		},
		{
			name: "when template not found in store uploads it",
			setupFunc: func() {
				setupTestTemplateFS()
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *agentMocks.MockTemplateStore,
				putNames *[]string,
			) {
				mockObj.EXPECT().
					ObjectGetBytes(gomock.Any(), "templates", gomock.Any()).
					Return(nil, errors.New("not found"))

				mockObj.EXPECT().
					ObjectPutBytes(gomock.Any(), "templates", gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						name string,
						_ []byte,
					) (*jetstream.ObjectInfo, error) {
						*putNames = append(*putNames, name)

						return &jetstream.ObjectInfo{}, nil
					})
			},
			validateFunc: func(err error, putNames *[]string) {
				s.NoError(err)

				s.Len(*putNames, 1)

				s.Equal("osapi/test.tmpl", (*putNames)[0])
			},
		},
		{
			name: "when template unchanged in store skips upload",
			setupFunc: func() {
				setupTestTemplateFS()
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *agentMocks.MockTemplateStore,
				_ *[]string,
			) {
				// Return same content as embedded — SHA will match, no ObjectPutBytes call.
				mockObj.EXPECT().
					ObjectGetBytes(gomock.Any(), "templates", gomock.Any()).
					Return(templateData, nil)
			},
			validateFunc: func(err error, putNames *[]string) {
				s.NoError(err)

				s.Len(*putNames, 0)
			},
		},
		{
			name: "when template changed in store overwrites it",
			setupFunc: func() {
				setupTestTemplateFS()
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *agentMocks.MockTemplateStore,
				putNames *[]string,
			) {
				// Return different content — SHA will differ.
				mockObj.EXPECT().
					ObjectGetBytes(gomock.Any(), "templates", gomock.Any()).
					Return([]byte("old content"), nil)

				mockObj.EXPECT().
					ObjectPutBytes(gomock.Any(), "templates", gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						name string,
						_ []byte,
					) (*jetstream.ObjectInfo, error) {
						*putNames = append(*putNames, name)

						return &jetstream.ObjectInfo{}, nil
					})
			},
			validateFunc: func(err error, putNames *[]string) {
				s.NoError(err)

				s.Len(*putNames, 1)

				s.Equal("osapi/test.tmpl", (*putNames)[0])
			},
		},
		{
			name: "when ObjectPutBytes fails returns wrapped error",
			setupFunc: func() {
				setupTestTemplateFS()
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *agentMocks.MockTemplateStore,
				putNames *[]string,
			) {
				mockObj.EXPECT().
					ObjectGetBytes(gomock.Any(), "templates", gomock.Any()).
					Return(nil, errors.New("not found"))

				mockObj.EXPECT().
					ObjectPutBytes(gomock.Any(), "templates", gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						name string,
						_ []byte,
					) (*jetstream.ObjectInfo, error) {
						*putNames = append(*putNames, name)

						return nil, errors.New("object store unavailable")
					})
			},
			validateFunc: func(err error, putNames *[]string) {
				s.Error(err)
				s.Contains(err.Error(), "upload osapi template")

				s.Len(*putNames, 1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			mockObj := agentMocks.NewMockTemplateStore(ctrl)
			putNames := &[]string{}

			if tt.setupMock != nil {
				tt.setupMock(ctrl, mockObj, putNames)
			}

			err := agent.SeedSystemTemplates(s.ctx, s.logger, mockObj, "templates")
			tt.validateFunc(err, putNames)
		})
	}
}

func (s *SeedPublicTestSuite) TestDefaultReadEmbeddedFile() {
	tests := []struct {
		name         string
		path         string
		validateFunc func([]byte, error)
	}{
		{
			name: "when the file is embedded",
			path: "resources/templates/.gitkeep",
			validateFunc: func(_ []byte, err error) {
				s.NoError(err)
			},
		},
		{
			name: "when the file is not embedded",
			path: "resources/templates/missing",
			validateFunc: func(data []byte, err error) {
				s.Error(err)
				s.Nil(data)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(agent.ExportDefaultReadEmbeddedFile(tc.path))
		})
	}
}

func TestSeedPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SeedPublicTestSuite))
}

// errorWalkFS is an fs.FS that returns a walk error for the templates dir.
type errorWalkFS struct{}

func (e *errorWalkFS) Open(
	name string,
) (fs.File, error) {
	if name == "resources/templates" {
		return &errorDir{}, nil
	}

	return nil, fmt.Errorf("walk error: %s", name)
}

// errorDir is an fs.File that is a directory but returns an error on ReadDir.
type errorDir struct{}

func (d *errorDir) Stat() (fs.FileInfo, error) {
	return &dirInfo{}, nil
}

func (d *errorDir) Read(
	_ []byte,
) (int, error) {
	return 0, fmt.Errorf("not a file")
}

func (d *errorDir) Close() error { return nil }

func (d *errorDir) ReadDir(
	_ int,
) ([]fs.DirEntry, error) {
	return nil, fmt.Errorf("walk error")
}

// dirInfo satisfies fs.FileInfo for a directory.
type dirInfo struct{}

func (i *dirInfo) Name() string       { return "templates" }
func (i *dirInfo) Size() int64        { return 0 }
func (i *dirInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o755 }
func (i *dirInfo) ModTime() time.Time { return time.Time{} }
func (i *dirInfo) IsDir() bool        { return true }
func (i *dirInfo) Sys() interface{}   { return nil }
