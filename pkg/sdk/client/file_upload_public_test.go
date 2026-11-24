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

package client_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type FilePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *FilePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *FilePublicTestSuite) TestUpload() {
	fileContent := []byte("content")
	hash := sha256.Sum256(fileContent)
	contentSHA := fmt.Sprintf("%x", hash)

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		file         io.Reader
		validateFunc func(*client.Response[client.FileUploadResponse], error)
	}{
		{
			name: "when uploading new file returns result",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"file not found"}`))
					return
				}
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(
						`{"name":"nginx.conf","sha256":"abc123","size":1024,"changed":true,"content_type":"raw"}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal("abc123", resp.Data.SHA256)
				suite.Equal(1024, resp.Data.Size)
				suite.True(resp.Data.Changed)
				suite.Equal(types.ContentRaw, resp.Data.ContentType)
			},
		},
		{
			name: "when pre-check SHA matches skips upload",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprintf(
						w,
						`{"name":"nginx.conf","sha256":"%s","size":7,"content_type":"raw"}`,
						contentSHA,
					)
					return
				}
				// POST should NOT be called — fail if it is.
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"unexpected POST"}`))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal(contentSHA, resp.Data.SHA256)
				suite.False(resp.Data.Changed)
				suite.Nil(resp.RawJSON())
			},
		},
		{
			name: "when server returns 409 returns ConflictError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"file not found"}`))
					return
				}
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error":"file already exists"}`))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.ConflictError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(http.StatusConflict, target.StatusCode)
			},
		},
		{
			name: "when server returns 400 returns ValidationError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"not found"}`))
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"name is required"}`))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.ValidationError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name: "when server returns 403 returns AuthError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Error(err)
				suite.Nil(resp)
			},
		},
		{
			name: "when server returns 201 with no JSON body returns a decode error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusCreated)
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Nil(resp)
				suite.ErrorContains(err, "POST /api/file: decode response")
			},
		},
		{
			name: "when file reader returns error",
			file: &errReader{err: errors.New("read failed")},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "read file")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			file := tc.file
			if file == nil {
				file = bytes.NewReader(fileContent)
			}

			resp, err := sut.File.Upload(
				suite.ctx,
				"nginx.conf",
				"raw",
				file,
			)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestUpdate() {
	body := func(changed bool) string {
		return fmt.Sprintf(
			`{"name":"nginx.conf","sha256":"abc","size":7,"changed":%v,"content_type":"raw"}`,
			changed,
		)
	}

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		validateFunc func(*client.Response[client.FileUploadResponse], error)
	}{
		{
			name: "when the file existed, it is replaced: 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				suite.Equal(http.MethodPut, r.Method)
				suite.Equal("/api/file/nginx.conf", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(body(true)))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Require().NoError(err)
				suite.True(resp.Data.Changed)
			},
		},
		{
			name: "when the name is new, it is created: 201",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Location", "/api/file/nginx.conf")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(body(true)))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Require().NoError(err)
				suite.Equal("nginx.conf", resp.Data.Name)
			},
		},
		{
			name: "when the server returns 403 returns AuthError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"cannot overwrite protected osapi file"}`))
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Nil(resp)
				var target *client.AuthError
				suite.Require().True(errors.As(err, &target), "err is %T: %v", err, err)
			},
		},
		{
			name:      "when the HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Nil(resp)
				suite.ErrorContains(err, "PUT /api/file/nginx.conf")
			},
		},
		{
			name: "when the answer has no body returns a decode error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.FileUploadResponse], err error) {
				suite.Nil(resp)
				suite.ErrorContains(err, "PUT /api/file/nginx.conf: decode response")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			url := tc.serverURL
			if tc.handler != nil {
				server := httptest.NewServer(tc.handler)
				defer server.Close()
				url = server.URL
			}
			sut := client.New(url, "test-token", client.WithLogger(slog.Default()))

			tc.validateFunc(sut.File.Update(
				suite.ctx, "nginx.conf", "raw", bytes.NewReader([]byte("content")),
			))
		})
	}
}

func (suite *FilePublicTestSuite) TestUpdateReadError() {
	sut := client.New("http://127.0.0.1:0", "test-token", client.WithLogger(slog.Default()))

	resp, err := sut.File.Update(
		suite.ctx,
		"nginx.conf",
		"raw",
		&errReader{err: errors.New("read failed")},
	)
	suite.Nil(resp)
	suite.ErrorContains(err, "read file")
}

func (suite *FilePublicTestSuite) TestChanged() {
	fileContent := []byte("content")
	hash := sha256.Sum256(fileContent)
	contentSHA := fmt.Sprintf("%x", hash)

	differentContent := []byte("different")
	diffHash := sha256.Sum256(differentContent)
	diffSHA := fmt.Sprintf("%x", diffHash)

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		file         io.Reader
		validateFunc func(*client.Response[client.FileChanged], error)
	}{
		{
			name: "when file does not exist returns changed true",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"file not found"}`))
			},
			validateFunc: func(resp *client.Response[client.FileChanged], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.True(resp.Data.Changed)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal(contentSHA, resp.Data.SHA256)
			},
		},
		{
			name: "when SHA matches returns changed false",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(
					w,
					`{"name":"nginx.conf","sha256":"%s","size":7,"content_type":"raw"}`,
					contentSHA,
				)
			},
			validateFunc: func(resp *client.Response[client.FileChanged], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.False(resp.Data.Changed)
				suite.Equal(contentSHA, resp.Data.SHA256)
			},
		},
		{
			name: "when SHA differs returns changed true",
			file: bytes.NewReader(differentContent),
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(
					w,
					`{"name":"nginx.conf","sha256":"%s","size":7,"content_type":"raw"}`,
					contentSHA,
				)
			},
			validateFunc: func(resp *client.Response[client.FileChanged], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.True(resp.Data.Changed)
				suite.Equal(diffSHA, resp.Data.SHA256)
			},
		},
		{
			name: "when server returns 403 returns error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.FileChanged], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "check file nginx.conf")
			},
		},
		{
			name: "when file reader returns error",
			file: &errReader{err: errors.New("read failed")},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.FileChanged], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "read file")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			file := tc.file
			if file == nil {
				file = bytes.NewReader(fileContent)
			}

			resp, err := sut.File.GetChanged(
				suite.ctx,
				"nginx.conf",
				file,
			)
			tc.validateFunc(resp, err)
		})
	}
}

type errReader struct {
	err error
}

func (r *errReader) Read(
	_ []byte,
) (int, error) {
	return 0, r.err
}

func TestFilePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(FilePublicTestSuite))
}
