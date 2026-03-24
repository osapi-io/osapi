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
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/file"
	filemocks "github.com/retr0h/osapi/internal/provider/file/mocks"
)

type UndeployPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
	ctx    context.Context
}

func (suite *UndeployPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
}

func (suite *UndeployPublicTestSuite) TearDownTest() {}

func (suite *UndeployPublicTestSuite) TestUndeploy() {
	tests := []struct {
		name         string
		setupMock    func(*gomock.Controller, *jobmocks.MockKeyValue, afero.Fs)
		req          file.UndeployRequest
		want         *file.UndeployResult
		wantErr      bool
		wantErrMsg   string
		validateFunc func(afero.Fs)
	}{
		{
			name: "when file exists on disk with state",
			setupMock: func(
				ctrl *gomock.Controller,
				mockKV *jobmocks.MockKeyValue,
				appFs afero.Fs,
			) {
				_ = afero.WriteFile(appFs, "/etc/cron.d/backup", []byte("content"), 0o644)

				stateJSON, _ := json.Marshal(job.FileState{
					ObjectName: "backup-script",
					Path:       "/etc/cron.d/backup",
					SHA256:     "abc123",
					Mode:       "0644",
					DeployedAt: "2026-03-22T00:00:00Z",
				})

				mockEntry := jobmocks.NewMockKeyValueEntry(ctrl)
				mockEntry.EXPECT().Value().Return(stateJSON)

				mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
				mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			req: file.UndeployRequest{Path: "/etc/cron.d/backup"},
			want: &file.UndeployResult{
				Changed: true,
				Path:    "/etc/cron.d/backup",
			},
			validateFunc: func(appFs afero.Fs) {
				exists, _ := afero.Exists(appFs, "/etc/cron.d/backup")
				suite.False(exists, "file should be removed from disk")
			},
		},
		{
			name: "when file does not exist on disk",
			setupMock: func(
				_ *gomock.Controller,
				_ *jobmocks.MockKeyValue,
				_ afero.Fs,
			) {
			},
			req: file.UndeployRequest{Path: "/etc/cron.d/nonexistent"},
			want: &file.UndeployResult{
				Changed: false,
				Path:    "/etc/cron.d/nonexistent",
			},
		},
		{
			name: "when file exists but no state entry",
			setupMock: func(
				_ *gomock.Controller,
				mockKV *jobmocks.MockKeyValue,
				appFs afero.Fs,
			) {
				_ = afero.WriteFile(appFs, "/etc/cron.d/orphan", []byte("content"), 0o644)

				mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("key not found"))
			},
			req: file.UndeployRequest{Path: "/etc/cron.d/orphan"},
			want: &file.UndeployResult{
				Changed: true,
				Path:    "/etc/cron.d/orphan",
			},
			validateFunc: func(appFs afero.Fs) {
				exists, _ := afero.Exists(appFs, "/etc/cron.d/orphan")
				suite.False(exists, "file should be removed from disk")
			},
		},
		{
			name: "when file exists but state entry value is invalid JSON",
			setupMock: func(
				ctrl *gomock.Controller,
				mockKV *jobmocks.MockKeyValue,
				appFs afero.Fs,
			) {
				_ = afero.WriteFile(appFs, "/etc/cron.d/corrupt", []byte("content"), 0o644)

				mockEntry := jobmocks.NewMockKeyValueEntry(ctrl)
				mockEntry.EXPECT().Value().Return([]byte("not-valid-json"))

				mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			req: file.UndeployRequest{Path: "/etc/cron.d/corrupt"},
			want: &file.UndeployResult{
				Changed: true,
				Path:    "/etc/cron.d/corrupt",
			},
			validateFunc: func(appFs afero.Fs) {
				exists, _ := afero.Exists(appFs, "/etc/cron.d/corrupt")
				suite.False(exists, "file should be removed from disk")
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctrl := gomock.NewController(suite.T())
			defer ctrl.Finish()

			appFs := afero.NewMemMapFs()
			mockObj := filemocks.NewMockObjectStore(ctrl)
			mockKV := jobmocks.NewMockKeyValue(ctrl)

			tt.setupMock(ctrl, mockKV, appFs)

			provider := file.New(suite.logger, appFs, mockObj, mockKV, "test-host")

			got, err := provider.Undeploy(suite.ctx, tt.req)

			if tt.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tt.wantErrMsg)
				suite.Nil(got)
			} else {
				suite.NoError(err)
				suite.Require().NotNil(got)
				suite.Equal(tt.want.Changed, got.Changed)
				suite.Equal(tt.want.Path, got.Path)

				if tt.validateFunc != nil {
					tt.validateFunc(appFs)
				}
			}
		})
	}
}

func TestUndeployPublicTestSuite(t *testing.T) {
	suite.Run(t, new(UndeployPublicTestSuite))
}
