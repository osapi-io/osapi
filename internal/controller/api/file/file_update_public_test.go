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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	apifile "github.com/osapi-io/osapi/internal/controller/api/file"
	"github.com/osapi-io/osapi/internal/controller/api/file/gen"
	"github.com/osapi-io/osapi/internal/controller/api/file/mocks"
)

// TestPutFileByName holds the replace of a file: 201 when the name is new,
// 200 when it existed, and no write when the content is the same.
func (s *FileUploadPublicTestSuite) TestPutFileByName() {
	content := []byte("server { listen 80; }")
	sum := sha256.Sum256(content)
	sameDigest := "SHA-256=" + base64.URLEncoding.EncodeToString(sum[:])

	tests := []struct {
		name         string
		fileName     string
		body         func() gen.PutFileByNameRequestObject
		setupMock    func()
		validateFunc func(gen.PutFileByNameResponseObject)
	}{
		{
			name:     "when the name is new, it is created: 201 with Location",
			fileName: "nginx.conf",
			setupMock: func() {
				s.mockObjStore.EXPECT().ObjectInfo(gomock.Any(), "objects", "nginx.conf").
					Return(nil, assert.AnError)
				s.mockObjStore.EXPECT().
					ObjectPut(gomock.Any(), "objects", gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			validateFunc: func(resp gen.PutFileByNameResponseObject) {
				created, ok := resp.(gen.PutFileByName201JSONResponse)
				s.Require().True(ok)
				s.True(created.Body.Changed)
				s.Equal("/api/file/nginx.conf", *created.Headers.Location)
			},
		},
		{
			name:     "when the file exists with other content, it is replaced: 200, changed",
			fileName: "nginx.conf",
			setupMock: func() {
				s.mockObjStore.EXPECT().ObjectInfo(gomock.Any(), "objects", "nginx.conf").
					Return(&jetstream.ObjectInfo{Digest: "SHA-256=old"}, nil)
				s.mockObjStore.EXPECT().
					ObjectPut(gomock.Any(), "objects", gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			validateFunc: func(resp gen.PutFileByNameResponseObject) {
				r, ok := resp.(gen.PutFileByName200JSONResponse)
				s.Require().True(ok)
				s.True(r.Changed)
				s.Equal("nginx.conf", r.Name)
			},
		},
		{
			name:     "when the file exists with the same content, nothing is written: 200, not changed",
			fileName: "nginx.conf",
			setupMock: func() {
				s.mockObjStore.EXPECT().ObjectInfo(gomock.Any(), "objects", "nginx.conf").
					Return(&jetstream.ObjectInfo{Digest: sameDigest}, nil)
			},
			validateFunc: func(resp gen.PutFileByNameResponseObject) {
				r, ok := resp.(gen.PutFileByName200JSONResponse)
				s.Require().True(ok)
				s.False(r.Changed)
			},
		},
		{
			name:      "when the name is an osapi file, it is a 403",
			fileName:  "osapi/cron-wrapper.tmpl",
			setupMock: func() {},
			validateFunc: func(resp gen.PutFileByNameResponseObject) {
				_, ok := resp.(gen.PutFileByName403JSONResponse)
				s.True(ok)
			},
		},
		{
			name:     "when the form has no file, it is a 400",
			fileName: "nginx.conf",
			body: func() gen.PutFileByNameRequestObject {
				return gen.PutFileByNameRequestObject{
					Name: "nginx.conf",
					Body: makeMultipartReader("", "raw", nil),
				}
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutFileByNameResponseObject) {
				_, ok := resp.(gen.PutFileByName400JSONResponse)
				s.True(ok)
			},
		},
		{
			name:     "when the store fails, it is a 500",
			fileName: "nginx.conf",
			setupMock: func() {
				s.mockObjStore.EXPECT().ObjectInfo(gomock.Any(), "objects", "nginx.conf").
					Return(nil, assert.AnError)
				s.mockObjStore.EXPECT().
					ObjectPut(gomock.Any(), "objects", gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.PutFileByNameResponseObject) {
				_, ok := resp.(gen.PutFileByName500JSONResponse)
				s.True(ok)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setupMock()
			req := gen.PutFileByNameRequestObject{
				Name: tc.fileName,
				Body: makeMultipartReader("", "raw", content),
			}
			if tc.body != nil {
				req = tc.body()
			}
			resp, err := s.handler.PutFileByName(s.ctx, req)
			s.Require().NoError(err)
			tc.validateFunc(resp)
		})
	}
}

const rbacUpdateTestSigningKey = "test-signing-key-for-file-update-rbac"

// TestPutFileByNameRBACHTTP runs the replace through its route, with its
// auth.
func (s *FileUploadPublicTestSuite) TestPutFileByNameRBACHTTP() {
	tokenManager := authtoken.New(s.logger)
	content := []byte("server { listen 80; }")
	token := func(perms ...string) string {
		t, err := tokenManager.Generate(
			rbacUpdateTestSigningKey,
			[]string{"read"},
			"test-user",
			perms,
		)
		s.Require().NoError(err)

		return t
	}

	tests := []struct {
		name      string
		token     string
		setupMock func(*mocks.MockObjectStoreManager)
		wantCode  int
	}{
		{
			name:      "when there is no token, it is a 401",
			setupMock: func(*mocks.MockObjectStoreManager) {},
			wantCode:  http.StatusUnauthorized,
		},
		{
			name:      "when the token lacks file:write, it is a 403",
			token:     token("file:read"),
			setupMock: func(*mocks.MockObjectStoreManager) {},
			wantCode:  http.StatusForbidden,
		},
		{
			name:  "when the token has file:write, the file is replaced: 200",
			token: token("file:write"),
			setupMock: func(m *mocks.MockObjectStoreManager) {
				m.EXPECT().ObjectInfo(gomock.Any(), "objects", "nginx.conf").
					Return(&jetstream.ObjectInfo{Digest: "SHA-256=old"}, nil)
				m.EXPECT().ObjectPut(gomock.Any(), "objects", gomock.Any(), gomock.Any()).
					Return(&jetstream.ObjectInfo{}, nil)
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			objMock := mocks.NewMockObjectStoreManager(s.mockCtrl)
			tc.setupMock(objMock)
			appConfig := config.Config{Controller: config.Controller{API: config.APIServer{
				Security: config.ServerSecurity{SigningKey: rbacUpdateTestSigningKey},
			}}}
			server := api.New(appConfig, s.logger)
			server.RegisterHandlers(apifile.Handler(
				s.logger, objMock, "objects", nil, "", rbacUpdateTestSigningKey, nil,
			))

			body, ct := makeMultipartBody("", "raw", content)
			req := httptest.NewRequest(http.MethodPut, "/api/file/nginx.conf", body)
			req.Header.Set("Content-Type", ct)
			if tc.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
			}
			rec := httptest.NewRecorder()
			server.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
		})
	}
}
