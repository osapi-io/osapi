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
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/file"
)

type StatusPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	logger *slog.Logger
	ctx    context.Context
	appFs  afero.Fs
	mockKV *jobmocks.MockKeyValue
}

func (suite *StatusPublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.ctx = context.Background()
	suite.appFs = afero.NewMemMapFs()
	suite.mockKV = jobmocks.NewMockKeyValue(suite.ctrl)
}

func (suite *StatusPublicTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *StatusPublicTestSuite) TestStatus() {
	fileContent := []byte("server { listen 80; }")
	fileSHA := computeTestSHA256(fileContent)
	driftedContent := []byte("server { listen 443; }")
	driftedSHA := computeTestSHA256(driftedContent)

	tests := []struct {
		name       string
		setupMock  func()
		req        file.StatusRequest
		want       *file.StatusResult
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "when file in sync",
			setupMock: func() {
				_ = afero.WriteFile(suite.appFs, "/etc/nginx/nginx.conf", fileContent, 0o644)

				existingState := job.FileState{
					SHA256: fileSHA,
					Path:   "/etc/nginx/nginx.conf",
				}
				stateBytes, _ := json.Marshal(existingState)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			req: file.StatusRequest{
				Path: "/etc/nginx/nginx.conf",
			},
			want: &file.StatusResult{
				Path:   "/etc/nginx/nginx.conf",
				Status: "in-sync",
				SHA256: fileSHA,
			},
		},
		{
			name: "when file drifted",
			setupMock: func() {
				_ = afero.WriteFile(suite.appFs, "/etc/nginx/nginx.conf", driftedContent, 0o644)

				existingState := job.FileState{
					SHA256: fileSHA,
					Path:   "/etc/nginx/nginx.conf",
				}
				stateBytes, _ := json.Marshal(existingState)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			req: file.StatusRequest{
				Path: "/etc/nginx/nginx.conf",
			},
			want: &file.StatusResult{
				Path:   "/etc/nginx/nginx.conf",
				Status: "drifted",
				SHA256: driftedSHA,
			},
		},
		{
			name: "when file missing on disk",
			setupMock: func() {
				existingState := job.FileState{
					SHA256: fileSHA,
					Path:   "/etc/nginx/nginx.conf",
				}
				stateBytes, _ := json.Marshal(existingState)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			req: file.StatusRequest{
				Path: "/etc/nginx/nginx.conf",
			},
			want: &file.StatusResult{
				Path:   "/etc/nginx/nginx.conf",
				Status: "missing",
			},
		},
		{
			name: "when state entry has invalid JSON",
			setupMock: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return([]byte("not-json"))

				suite.mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			req: file.StatusRequest{
				Path: "/etc/nginx/nginx.conf",
			},
			wantErr:    true,
			wantErrMsg: "failed to parse file state",
		},
		{
			name: "when no state entry",
			setupMock: func() {
				suite.mockKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			req: file.StatusRequest{
				Path: "/etc/nginx/nginx.conf",
			},
			want: &file.StatusResult{
				Path:   "/etc/nginx/nginx.conf",
				Status: "missing",
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Reset filesystem for each test case.
			suite.appFs = afero.NewMemMapFs()

			if tc.setupMock != nil {
				tc.setupMock()
			}

			provider := file.NewFileProvider(
				suite.logger,
				suite.appFs,
				&stubObjectStore{},
				suite.mockKV,
				"test-host",
			)

			got, err := provider.Status(suite.ctx, tc.req)

			if tc.wantErr {
				suite.Error(err)
				suite.ErrorContains(err, tc.wantErrMsg)
				suite.Nil(got)
			} else {
				suite.NoError(err)
				suite.Require().NotNil(got)
				suite.Equal(tc.want, got)
			}
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestStatusPublicTestSuite(t *testing.T) {
	suite.Run(t, new(StatusPublicTestSuite))
}
