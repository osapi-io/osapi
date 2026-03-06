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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apifile "github.com/retr0h/osapi/internal/api/file"
	"github.com/retr0h/osapi/internal/api/file/gen"
	"github.com/retr0h/osapi/internal/api/file/mocks"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
)

type FileUploadPublicTestSuite struct {
	suite.Suite

	mockCtrl     *gomock.Controller
	mockObjStore *mocks.MockObjectStoreManager
	handler      *apifile.File
	ctx          context.Context
	appConfig    config.Config
	logger       *slog.Logger
}

func (s *FileUploadPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockObjStore = mocks.NewMockObjectStoreManager(s.mockCtrl)
	s.handler = apifile.New(slog.Default(), s.mockObjStore)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *FileUploadPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FileUploadPublicTestSuite) TestPostFile() {
	tests := []struct {
		name         string
		request      gen.PostFileRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostFileResponseObject)
	}{
		{
			name: "success",
			request: gen.PostFileRequestObject{
				Body: &gen.FileUploadRequest{
					Name:    "nginx.conf",
					Content: []byte("server { listen 80; }"),
				},
			},
			setupMock: func() {
				s.mockObjStore.EXPECT().
					PutBytes(gomock.Any(), "nginx.conf", []byte("server { listen 80; }")).
					Return(&jetstream.ObjectInfo{
						ObjectMeta: jetstream.ObjectMeta{Name: "nginx.conf"},
						Size:       21,
					}, nil)
			},
			validateFunc: func(resp gen.PostFileResponseObject) {
				r, ok := resp.(gen.PostFile201JSONResponse)
				s.True(ok)
				s.Equal("nginx.conf", r.Name)
				s.Equal(21, r.Size)
				s.NotEmpty(r.Sha256)
			},
		},
		{
			name: "validation error empty name",
			request: gen.PostFileRequestObject{
				Body: &gen.FileUploadRequest{
					Name:    "",
					Content: []byte("data"),
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostFileResponseObject) {
				r, ok := resp.(gen.PostFile400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "validation error empty content",
			request: gen.PostFileRequestObject{
				Body: &gen.FileUploadRequest{
					Name:    "test.txt",
					Content: nil,
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostFileResponseObject) {
				r, ok := resp.(gen.PostFile400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "object store error",
			request: gen.PostFileRequestObject{
				Body: &gen.FileUploadRequest{
					Name:    "nginx.conf",
					Content: []byte("server { listen 80; }"),
				},
			},
			setupMock: func() {
				s.mockObjStore.EXPECT().
					PutBytes(gomock.Any(), "nginx.conf", []byte("server { listen 80; }")).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostFileResponseObject) {
				_, ok := resp.(gen.PostFile500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostFile(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *FileUploadPublicTestSuite) TestPostFileHTTP() {
	tests := []struct {
		name         string
		body         string
		setupMock    func() *mocks.MockObjectStoreManager
		wantCode     int
		wantContains []string
	}{
		{
			name: "when upload Ok",
			body: `{"name":"nginx.conf","content":"c2VydmVyIHsgbGlzdGVuIDgwOyB9"}`,
			setupMock: func() *mocks.MockObjectStoreManager {
				mock := mocks.NewMockObjectStoreManager(s.mockCtrl)
				mock.EXPECT().
					PutBytes(gomock.Any(), "nginx.conf", gomock.Any()).
					Return(&jetstream.ObjectInfo{
						ObjectMeta: jetstream.ObjectMeta{Name: "nginx.conf"},
						Size:       21,
					}, nil)
				return mock
			},
			wantCode:     http.StatusCreated,
			wantContains: []string{`"name":"nginx.conf"`, `"sha256"`, `"size"`},
		},
		{
			name: "when validation error",
			body: `{"name":"","content":"c2VydmVyIHsgbGlzdGVuIDgwOyB9"}`,
			setupMock: func() *mocks.MockObjectStoreManager {
				return mocks.NewMockObjectStoreManager(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{"required"},
		},
		{
			name: "when object store error",
			body: `{"name":"nginx.conf","content":"c2VydmVyIHsgbGlzdGVuIDgwOyB9"}`,
			setupMock: func() *mocks.MockObjectStoreManager {
				mock := mocks.NewMockObjectStoreManager(s.mockCtrl)
				mock.EXPECT().
					PutBytes(gomock.Any(), "nginx.conf", gomock.Any()).
					Return(nil, assert.AnError)
				return mock
			},
			wantCode:     http.StatusInternalServerError,
			wantContains: []string{"failed to store file"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			objMock := tc.setupMock()

			fileHandler := apifile.New(s.logger, objMock)
			strictHandler := gen.NewStrictHandler(fileHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPost,
				"/file",
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacUploadTestSigningKey = "test-signing-key-for-file-upload-rbac"

func (s *FileUploadPublicTestSuite) TestPostFileRBACHTTP() {
	tokenManager := authtoken.New(s.logger)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		setupMock    func() *mocks.MockObjectStoreManager
		wantCode     int
		wantContains []string
	}{
		{
			name: "when no token returns 401",
			setupAuth: func(_ *http.Request) {
				// No auth header set
			},
			setupMock: func() *mocks.MockObjectStoreManager {
				return mocks.NewMockObjectStoreManager(s.mockCtrl)
			},
			wantCode:     http.StatusUnauthorized,
			wantContains: []string{"Bearer token required"},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacUploadTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"node:read"},
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupMock: func() *mocks.MockObjectStoreManager {
				return mocks.NewMockObjectStoreManager(s.mockCtrl)
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with file:write returns 201",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacUploadTestSigningKey,
					[]string{"admin"},
					"test-user",
					[]string{"file:write"},
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupMock: func() *mocks.MockObjectStoreManager {
				mock := mocks.NewMockObjectStoreManager(s.mockCtrl)
				mock.EXPECT().
					PutBytes(gomock.Any(), "nginx.conf", gomock.Any()).
					Return(&jetstream.ObjectInfo{
						ObjectMeta: jetstream.ObjectMeta{Name: "nginx.conf"},
						Size:       21,
					}, nil)
				return mock
			},
			wantCode:     http.StatusCreated,
			wantContains: []string{`"name":"nginx.conf"`, `"sha256"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			objMock := tc.setupMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacUploadTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetFileHandler(objMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/file",
				strings.NewReader(`{"name":"nginx.conf","content":"c2VydmVyIHsgbGlzdGVuIDgwOyB9"}`),
			)
			req.Header.Set("Content-Type", "application/json")
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

func TestFileUploadPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FileUploadPublicTestSuite))
}
