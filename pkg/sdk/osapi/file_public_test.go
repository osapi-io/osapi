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

package osapi_test

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

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/osapi"
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
		opts         []osapi.UploadOption
		validateFunc func(*osapi.Response[osapi.FileUpload], error)
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
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal("abc123", resp.Data.SHA256)
				suite.Equal(1024, resp.Data.Size)
				suite.True(resp.Data.Changed)
				suite.Equal("raw", resp.Data.ContentType)
			},
		},
		{
			name: "when pre-check SHA matches skips upload",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprintf(w,
						`{"name":"nginx.conf","sha256":"%s","size":7,"content_type":"raw"}`,
						contentSHA,
					)
					return
				}
				// POST should NOT be called — fail if it is.
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"unexpected POST"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal(contentSHA, resp.Data.SHA256)
				suite.False(resp.Data.Changed)
				suite.Nil(resp.RawJSON())
			},
		},
		{
			name: "when force skips pre-check and uploads",
			opts: []osapi.UploadOption{osapi.WithForce()},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet {
					// GET should NOT be called — fail if it is.
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"unexpected GET"}`))
					return
				}
				suite.Contains(r.URL.RawQuery, "force=true")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(
						`{"name":"nginx.conf","sha256":"abc123","size":7,"changed":true,"content_type":"raw"}`,
					),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.True(resp.Data.Changed)
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
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ConflictError
				suite.True(errors.As(err, &target))
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
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
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
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)
			},
		},
		{
			name: "when server returns 201 with no JSON body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusCreated)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusCreated, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
		{
			name: "when file reader returns error",
			file: &errReader{err: errors.New("read failed")},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
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

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
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
				tc.opts...,
			)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestList() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		validateFunc func(*osapi.Response[osapi.FileList], error)
	}{
		{
			name: "when listing files returns results",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"files":[{"name":"file1.txt","sha256":"aaa","size":100,"content_type":"raw"}],"total":1}`,
					),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Files, 1)
				suite.Equal(1, resp.Data.Total)
				suite.Equal("file1.txt", resp.Data.Files[0].Name)
				suite.Equal("raw", resp.Data.Files[0].ContentType)
			},
		},
		{
			name: "when server returns 403 returns AuthError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "list files")
			},
		},
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.List(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		fileName     string
		validateFunc func(*osapi.Response[osapi.FileMetadata], error)
	}{
		{
			name:     "when getting file returns metadata",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"name":"nginx.conf","sha256":"def456","size":512,"content_type":"raw"}`,
					),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal("def456", resp.Data.SHA256)
				suite.Equal(512, resp.Data.Size)
				suite.Equal("raw", resp.Data.ContentType)
			},
		},
		{
			name:     "when server returns 400 returns ValidationError",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid file name"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:     "when server returns 404 returns NotFoundError",
			fileName: "missing.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"file not found"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
		{
			name:     "when server returns 403 returns AuthError",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			fileName:  "nginx.conf",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get file nginx.conf")
			},
		},
		{
			name:     "when server returns 200 with no JSON body returns UnexpectedStatusError",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.Get(suite.ctx, tc.fileName)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		fileName     string
		validateFunc func(*osapi.Response[osapi.FileDelete], error)
	}{
		{
			name:     "when deleting file returns result",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"name":"old.conf","deleted":true}`),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("old.conf", resp.Data.Name)
				suite.True(resp.Data.Deleted)
			},
		},
		{
			name:     "when server returns 400 returns ValidationError",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid file name"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:     "when server returns 404 returns NotFoundError",
			fileName: "missing.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"file not found"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
		{
			name:     "when server returns 403 returns AuthError",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			fileName:  "old.conf",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "delete file old.conf")
			},
		},
		{
			name:     "when server returns 200 with no JSON body returns UnexpectedStatusError",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.Delete(suite.ctx, tc.fileName)
			tc.validateFunc(resp, err)
		})
	}
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
		validateFunc func(*osapi.Response[osapi.FileChanged], error)
	}{
		{
			name: "when file does not exist returns changed true",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"file not found"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileChanged], err error) {
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
				_, _ = fmt.Fprintf(w,
					`{"name":"nginx.conf","sha256":"%s","size":7,"content_type":"raw"}`,
					contentSHA,
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileChanged], err error) {
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
				_, _ = fmt.Fprintf(w,
					`{"name":"nginx.conf","sha256":"%s","size":7,"content_type":"raw"}`,
					contentSHA,
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileChanged], err error) {
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
			validateFunc: func(resp *osapi.Response[osapi.FileChanged], err error) {
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
			validateFunc: func(resp *osapi.Response[osapi.FileChanged], err error) {
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

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			file := tc.file
			if file == nil {
				file = bytes.NewReader(fileContent)
			}

			resp, err := sut.File.Changed(
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

func TestFilePublicTestSuite(t *testing.T) {
	suite.Run(t, new(FilePublicTestSuite))
}
