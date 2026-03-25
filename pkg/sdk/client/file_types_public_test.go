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

func (suite *FileTypesPublicTestSuite) TestFileDeployCollectionFromGen() {
	trueVal := true
	falseVal := false
	errMsg := "deploy failed"

	tests := []struct {
		name         string
		input        *gen.FileDeployCollectionResponse
		validateFunc func(client.Collection[client.FileDeployResult])
	}{
		{
			name: "when results present returns collection with results",
			input: &gen.FileDeployCollectionResponse{
				Results: []gen.FileDeployResult{
					{Hostname: "web-01", Changed: &trueVal},
					{Hostname: "web-02", Changed: &falseVal, Error: &errMsg},
				},
			},
			validateFunc: func(result client.Collection[client.FileDeployResult]) {
				suite.Len(result.Results, 2)
				suite.Equal("web-01", result.Results[0].Hostname)
				suite.True(result.Results[0].Changed)
				suite.Empty(result.Results[0].Error)
				suite.Equal("web-02", result.Results[1].Hostname)
				suite.False(result.Results[1].Changed)
				suite.Equal("deploy failed", result.Results[1].Error)
			},
		},
		{
			name: "when empty results returns empty collection",
			input: &gen.FileDeployCollectionResponse{
				Results: []gen.FileDeployResult{},
			},
			validateFunc: func(result client.Collection[client.FileDeployResult]) {
				suite.Empty(result.Results)
				suite.Empty(result.JobID)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileDeployCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileUndeployCollectionFromGen() {
	trueVal := true
	errMsg := "undeploy failed"

	tests := []struct {
		name         string
		input        *gen.FileUndeployCollectionResponse
		validateFunc func(client.Collection[client.FileUndeployResult])
	}{
		{
			name: "when results present returns collection with results",
			input: &gen.FileUndeployCollectionResponse{
				Results: []gen.FileUndeployResult{
					{Hostname: "web-01", Changed: &trueVal},
					{Hostname: "web-02", Error: &errMsg},
				},
			},
			validateFunc: func(result client.Collection[client.FileUndeployResult]) {
				suite.Len(result.Results, 2)
				suite.Equal("web-01", result.Results[0].Hostname)
				suite.True(result.Results[0].Changed)
				suite.Equal("web-02", result.Results[1].Hostname)
				suite.Equal("undeploy failed", result.Results[1].Error)
			},
		},
		{
			name: "when empty results returns empty collection",
			input: &gen.FileUndeployCollectionResponse{
				Results: []gen.FileUndeployResult{},
			},
			validateFunc: func(result client.Collection[client.FileUndeployResult]) {
				suite.Empty(result.Results)
				suite.Empty(result.JobID)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileUndeployCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *FileTypesPublicTestSuite) TestFileStatusCollectionFromGen() {
	sha := "abc123"
	changed := false
	errMsg := "status failed"
	path := "/etc/nginx/nginx.conf"
	status := "in-sync"
	missingPath := "/etc/missing.conf"
	missingStatus := "missing"

	tests := []struct {
		name         string
		input        *gen.FileStatusCollectionResponse
		validateFunc func(client.Collection[client.FileStatusResult])
	}{
		{
			name: "when all fields populated returns FileStatusResult",
			input: &gen.FileStatusCollectionResponse{
				Results: []gen.FileStatusResult{
					{
						Hostname: "web-03",
						Path:     &path,
						Status:   &status,
						Sha256:   &sha,
						Changed:  &changed,
						Error:    &errMsg,
					},
				},
			},
			validateFunc: func(result client.Collection[client.FileStatusResult]) {
				suite.Len(result.Results, 1)
				r := result.Results[0]
				suite.Equal("web-03", r.Hostname)
				suite.Equal("/etc/nginx/nginx.conf", r.Path)
				suite.Equal("in-sync", r.Status)
				suite.Equal("abc123", r.SHA256)
				suite.False(r.Changed)
				suite.Equal("status failed", r.Error)
			},
		},
		{
			name: "when sha256 is nil returns empty string",
			input: &gen.FileStatusCollectionResponse{
				Results: []gen.FileStatusResult{
					{
						Hostname: "web-04",
						Path:     &missingPath,
						Status:   &missingStatus,
						Sha256:   nil,
					},
				},
			},
			validateFunc: func(result client.Collection[client.FileStatusResult]) {
				suite.Len(result.Results, 1)
				r := result.Results[0]
				suite.Equal("missing", r.Status)
				suite.Empty(r.SHA256)
				suite.False(r.Changed)
				suite.Empty(r.Error)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportFileStatusCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestFileTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FileTypesPublicTestSuite))
}
