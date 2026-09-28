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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	execMocks "github.com/osapi-io/osapi/internal/exec/mocks"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/provider/file"
	filemocks "github.com/osapi-io/osapi/internal/provider/file/mocks"
)

type DeployPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
	ctx    context.Context
}

func (suite *DeployPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
}

func (suite *DeployPublicTestSuite) TearDownSubTest() {
	file.ResetMarshalJSON()
	file.ResetOwnershipFuncs()
}

func (suite *DeployPublicTestSuite) TearDownTest() {}

func (suite *DeployPublicTestSuite) TestDeploy() {
	fileContent := []byte("server { listen 80; }")
	existingSHA := computeTestSHA256(fileContent)
	differentContent := []byte("server { listen 443; }")

	tests := []struct {
		name         string
		setupFunc    func()
		setupMock    func(*gomock.Controller, *filemocks.MockObjectStore, *jobmocks.MockKeyValue, *avfs.VFS)
		setupExec    func(*execMocks.MockManager)
		req          file.DeployRequest
		want         *file.DeployResult
		wantErr      bool
		wantErrMsg   string
		validateFunc func(avfs.VFS)
	}{
		{
			name: "when marshal state fails returns error",
			setupFunc: func() {
				file.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				})
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return([]byte("server { listen 80; }"), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to marshal file state",
		},
		{
			name: "when deploy succeeds (new file)",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0644",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				data, err := appFs.ReadFile("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal(fileContent, data)
			},
		},
		{
			name: "when deploy succeeds (changed content)",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0644",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				data, err := appFs.ReadFile("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal(fileContent, data)
			},
		},
		{
			name: "when deploy skips (unchanged)",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: false,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
		},
		{
			name: "when file is deleted but state exists redeploys",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0644",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				data, err := appFs.ReadFile("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal(fileContent, data)
			},
		},
		{
			name: "when Object Store get fails",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			req: file.DeployRequest{
				ObjectName:  "missing.conf",
				Path:        "/etc/missing.conf",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to get object",
		},
		{
			name: "when content type is template",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return([]byte("server {{ .Vars.host }}"), nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "template",
				Vars:        map[string]any{"host": "10.0.0.1"},
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  computeTestSHA256([]byte("server 10.0.0.1")),
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				data, err := appFs.ReadFile("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal("server 10.0.0.1", string(data))
			},
		},
		{
			name: "when content type is empty resolves template from object header",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return([]byte("server {{ .Vars.host }}"), nil)

				mockObj.EXPECT().
					GetInfo(gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{
						ObjectMeta: jetstream.ObjectMeta{
							Name: "nginx.conf",
							Headers: nats.Header{
								"Osapi-Content-Type": []string{"template"},
							},
						},
					}, nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName: "nginx.conf",
				Path:       "/etc/nginx/nginx.conf",
				Vars:       map[string]any{"host": "10.0.0.1"},
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  computeTestSHA256([]byte("server 10.0.0.1")),
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				data, err := appFs.ReadFile("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal("server 10.0.0.1", string(data))
			},
		},
		{
			name: "when content type is empty and GetInfo fails defaults to raw",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				mockObj.EXPECT().
					GetInfo(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("info error"))

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName: "nginx.conf",
				Path:       "/etc/nginx/nginx.conf",
				Mode:       "0644",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
		},
		{
			name: "when file write fails",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				// Use failfs to block the temp file the atomic write opens
				// beside the target.
				vfs := failfs.New(memfs.New())
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					param *failfs.FailParam,
				) error {
					if fn == avfs.FnOpenFile &&
						strings.Contains(param.Path, ".osapi-") {
						return errors.New("write failed")
					}

					return nil
				})
				*appFs = vfs
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to write file",
		},
		{
			name: "when mkdir fails",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				vfs := failfs.New(memfs.New())
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnMkdirAll {
						return errors.New("mkdir failed")
					}

					return nil
				})
				*appFs = vfs
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to create directory",
		},
		{
			name: "when state KV put fails",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(0), assert.AnError)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to update file state",
		},
		{
			name: "when the mode cannot be parsed on a file already correct",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "not-octal",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "invalid file mode",
		},
		{
			name: "when the mode cannot be parsed the deploy fails",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "not-octal",
				ContentType: "raw",
			},
			wantErr: true,
			// It used to become 0644, which granted more than was asked for and
			// reported success for a mode nobody applied.
			wantErrMsg: "invalid file mode",
		},
		{
			name: "when mode is set",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "script.sh",
				Path:        "/usr/local/bin/script.sh",
				Mode:        "0755",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/usr/local/bin/script.sh",
			},
			validateFunc: func(appFs avfs.VFS) {
				info, err := appFs.Stat("/usr/local/bin/script.sh")
				suite.Require().NoError(err)
				suite.Equal(os.FileMode(0o755), info.Mode())
			},
		},
		{
			name: "when the file on disk was edited the content is restored",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				// The record says this content was deployed; the disk says
				// somebody has been editing it.
				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", differentContent, 0o644)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				content, err := appFs.ReadFile("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal(string(fileContent), string(content))
			},
		},
		{
			name: "when the owner was changed outside osapi it is restored",
			setupFunc: func() {
				// A file owned by root where nginx was asked for.
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 0, 0, nil },
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			setupExec: func(mockExec *execMocks.MockManager) {
				mockExec.EXPECT().
					RunPrivilegedCmd(
						gomock.Any(),
						"chown",
						[]string{"nginx:www-data", "/etc/nginx/nginx.conf"},
					).
					Return("", nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "nginx",
				Group:       "www-data",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
		},
		{
			name: "when the file already has the requested ownership nothing runs",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 33, 33, nil },
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "nginx",
				Group:       "www-data",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: false,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
		},
		{
			name: "when the host does not know the requested owner",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 0, 0, nil },
					func(string) (int, error) { return 0, errors.New("no such user") },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "ngnix",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to resolve owner",
		},
		{
			name: "when the host does not know the requested group",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 0, 0, nil },
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 0, errors.New("no such group") },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Group:       "wwwdata",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to resolve group",
		},
		{
			name: "when the file ownership cannot be read at all",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 0, 0, errors.New("stat failed") },
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "nginx",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to read ownership",
		},
		{
			name: "when the platform cannot report ownership it is applied anyway",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) {
						return 0, 0, file.ErrOwnershipUnreadable
					},
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			setupExec: func(mockExec *execMocks.MockManager) {
				mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "chown", gomock.Any()).
					Return("", nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "nginx",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
		},
		{
			name: "when only the mode changed the file is chmodded",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0600",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: true,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				info, err := appFs.Stat("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal(os.FileMode(0o600), info.Mode().Perm())
			},
		},
		{
			name: "when an absent mode leaves the permissions alone",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o600)
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
			},
			want: &file.DeployResult{
				Changed: false,
				SHA256:  existingSHA,
				Path:    "/etc/nginx/nginx.conf",
			},
			validateFunc: func(appFs avfs.VFS) {
				info, err := appFs.Stat("/etc/nginx/nginx.conf")
				suite.Require().NoError(err)
				suite.Equal(os.FileMode(0o600), info.Mode().Perm())
			},
		},
		{
			name: "when chown fails the deploy fails",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 0, 0, nil },
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				_ *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)
			},
			setupExec: func(mockExec *execMocks.MockManager) {
				mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "chown", gomock.Any()).
					Return("", errors.New("operation not permitted"))
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "nginx",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to set ownership",
		},
		{
			name: "when the file cannot be stat'd while enforcing the mode",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				base := memfs.New()
				_ = base.MkdirAll("/etc/nginx", 0o755)
				_ = base.WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)

				vfs := failfs.New(base)

				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnStat {
						return errors.New("stat failed")
					}

					return nil
				})

				*appFs = vfs
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0600",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to stat file",
		},
		{
			name: "when the mode cannot be applied",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				base := memfs.New()
				_ = base.MkdirAll("/etc/nginx", 0o755)
				_ = base.WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)

				vfs := failfs.New(base)
				_ = vfs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnChmod {
						return errors.New("chmod failed")
					}

					return nil
				})

				*appFs = vfs
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0600",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to set mode",
		},
		{
			name: "when chown fails on a file whose content is already correct",
			setupFunc: func() {
				file.SetOwnershipFuncs(
					func(string) (int, int, error) { return 0, 0, nil },
					func(string) (int, error) { return 33, nil },
					func(string) (int, error) { return 33, nil },
				)
			},
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				_ *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)
			},
			setupExec: func(mockExec *execMocks.MockManager) {
				mockExec.EXPECT().
					RunPrivilegedCmd(gomock.Any(), "chown", gomock.Any()).
					Return("", errors.New("operation not permitted"))
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Owner:       "nginx",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to set ownership",
		},
		{
			name: "when recording a permission change fails",
			setupMock: func(
				_ *gomock.Controller,
				mockObj *filemocks.MockObjectStore,
				mockKV *jobmocks.MockKeyValue,
				appFs *avfs.VFS,
			) {
				mockObj.EXPECT().
					GetBytes(gomock.Any(), gomock.Any()).
					Return(fileContent, nil)

				_ = (*appFs).MkdirAll("/etc/nginx", 0o755)
				_ = (*appFs).WriteFile("/etc/nginx/nginx.conf", fileContent, 0o644)

				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv down"))
			},
			req: file.DeployRequest{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				Mode:        "0600",
				ContentType: "raw",
			},
			wantErr:    true,
			wantErrMsg: "failed to update file state",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctrl := gomock.NewController(suite.T())
			defer ctrl.Finish()

			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			var appFs avfs.VFS = memfs.New()
			mockKV := jobmocks.NewMockKeyValue(ctrl)
			mockObj := filemocks.NewMockObjectStore(ctrl)

			mockExec := execMocks.NewMockManager(ctrl)

			if tc.setupMock != nil {
				tc.setupMock(ctrl, mockObj, mockKV, &appFs)
			}

			if tc.setupExec != nil {
				tc.setupExec(mockExec)
			}

			provider := file.New(
				suite.logger,
				appFs,
				mockObj,
				mockKV,
				"test-host",
				mockExec,
			)

			got, err := provider.Deploy(suite.ctx, tc.req)

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrMsg)
				suite.Nil(got)
			} else {
				suite.NoError(err)
				suite.Require().NotNil(got)
				suite.Equal(tc.want, got)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(appFs)
			}
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDeployPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DeployPublicTestSuite))
}
