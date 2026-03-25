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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

type FileTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *FileTypesPublicTestSuite) TestFileUploadFromGen() {
	tests := []struct {
		name         string
		input        *gen.FileUploadResponse
		validateFunc func(client.FileUpload)
	}{
		{
			name: "when all fields populated returns FileUpload",
			input: &gen.FileUploadResponse{
				Name:        "nginx.conf",
				Sha256:      "abc123",
				Size:        1024,
				Changed:     true,
				ContentType: "raw",
			},
			validateFunc: func(result client.FileUpload) {
				suite.Equal("nginx.conf", result.Name)
				suite.Equal("abc123", result.SHA256)
				suite.Equal(1024, result.Size)
				suite.True(result.Changed)
				suite.Equal("raw", result.ContentType)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileUploadFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileListFromGen() {
	tests := []struct {
		name         string
		input        *gen.FileListResponse
		validateFunc func(client.FileList)
	}{
		{
			name: "when files exist returns FileList with items",
			input: &gen.FileListResponse{
				Files: []gen.FileInfo{
					{Name: "file1.txt", Sha256: "aaa", Size: 100, ContentType: "raw"},
					{Name: "file2.txt", Sha256: "bbb", Size: 200, ContentType: "template"},
				},
				Total: 2,
			},
			validateFunc: func(result client.FileList) {
				suite.Len(result.Files, 2)
				suite.Equal(2, result.Total)
				suite.Equal("file1.txt", result.Files[0].Name)
				suite.Equal("aaa", result.Files[0].SHA256)
				suite.Equal(100, result.Files[0].Size)
				suite.Equal("raw", result.Files[0].ContentType)
				suite.Equal("file2.txt", result.Files[1].Name)
				suite.Equal("template", result.Files[1].ContentType)
			},
		},
		{
			name: "when no files returns empty FileList",
			input: &gen.FileListResponse{
				Files: []gen.FileInfo{},
				Total: 0,
			},
			validateFunc: func(result client.FileList) {
				suite.Empty(result.Files)
				suite.Equal(0, result.Total)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileListFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileMetadataFromGen() {
	tests := []struct {
		name         string
		input        *gen.FileInfoResponse
		validateFunc func(client.FileMetadata)
	}{
		{
			name: "when all fields populated returns FileMetadata",
			input: &gen.FileInfoResponse{
				Name:        "config.yaml",
				Sha256:      "def456",
				Size:        512,
				ContentType: "template",
			},
			validateFunc: func(result client.FileMetadata) {
				suite.Equal("config.yaml", result.Name)
				suite.Equal("def456", result.SHA256)
				suite.Equal(512, result.Size)
				suite.Equal("template", result.ContentType)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileMetadataFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileDeleteFromGen() {
	tests := []struct {
		name         string
		input        *gen.FileDeleteResponse
		validateFunc func(client.FileDelete)
	}{
		{
			name: "when deleted returns FileDelete with true",
			input: &gen.FileDeleteResponse{
				Name:    "old.conf",
				Deleted: true,
			},
			validateFunc: func(result client.FileDelete) {
				suite.Equal("old.conf", result.Name)
				suite.True(result.Deleted)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileDeleteFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileDeployResultFromGen() {
	tests := []struct {
		name         string
		input        *gen.FileDeployResponse
		validateFunc func(client.FileDeployResult)
	}{
		{
			name: "when all fields populated returns FileDeployResult",
			input: &gen.FileDeployResponse{
				JobId:    "job-123",
				Hostname: "web-01",
				Changed:  true,
			},
			validateFunc: func(result client.FileDeployResult) {
				suite.Equal("job-123", result.JobID)
				suite.Equal("web-01", result.Hostname)
				suite.True(result.Changed)
			},
		},
		{
			name: "when not changed returns false",
			input: &gen.FileDeployResponse{
				JobId:    "job-456",
				Hostname: "web-02",
				Changed:  false,
			},
			validateFunc: func(result client.FileDeployResult) {
				suite.False(result.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileDeployResultFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileStatusResultFromGen() {
	sha := "abc123"
	changed := false
	errMsg := "deploy failed"

	tests := []struct {
		name         string
		input        *gen.FileStatusResponse
		validateFunc func(client.FileStatusResult)
	}{
		{
			name: "when all fields populated returns FileStatusResult",
			input: &gen.FileStatusResponse{
				JobId:    "job-789",
				Hostname: "web-03",
				Path:     "/etc/nginx/nginx.conf",
				Status:   "in-sync",
				Sha256:   &sha,
				Changed:  &changed,
				Error:    &errMsg,
			},
			validateFunc: func(result client.FileStatusResult) {
				suite.Equal("job-789", result.JobID)
				suite.Equal("web-03", result.Hostname)
				suite.Equal("/etc/nginx/nginx.conf", result.Path)
				suite.Equal("in-sync", result.Status)
				suite.Equal("abc123", result.SHA256)
				suite.False(result.Changed)
				suite.Equal("deploy failed", result.Error)
			},
		},
		{
			name: "when sha256 is nil returns empty string",
			input: &gen.FileStatusResponse{
				JobId:    "job-000",
				Hostname: "web-04",
				Path:     "/etc/missing.conf",
				Status:   "missing",
				Sha256:   nil,
			},
			validateFunc: func(result client.FileStatusResult) {
				suite.Equal("missing", result.Status)
				suite.Empty(result.SHA256)
				suite.False(result.Changed)
				suite.Empty(result.Error)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileStatusResultFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestFileTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FileTypesPublicTestSuite))
}
