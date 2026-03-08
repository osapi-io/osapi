//go:build integration

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

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FileSmokeSuite struct {
	suite.Suite
}

func (s *FileSmokeSuite) TestFileList() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns file list with total",
			args: []string{"client", "file", "list", "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.Contains(result, "files")
				s.Contains(result, "total")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *FileSmokeSuite) TestFileUploadGetDelete() {
	skipWriteOp(s.T(), "FILE_UPLOAD")

	filePath := writeTempFile(s.T(), "integration-test-content\n")

	// Upload
	uploadOut, _, uploadCode := runCLI(
		"client", "file", "upload",
		"--name", "test-int.conf",
		"--file", filePath,
		"--json",
	)
	s.Require().Equal(0, uploadCode)

	var uploadResp struct {
		Name    string `json:"name"`
		SHA256  string `json:"sha256"`
		Size    int    `json:"size"`
		Changed bool   `json:"changed"`
	}
	s.Require().NoError(parseJSON(uploadOut, &uploadResp))
	s.Equal("test-int.conf", uploadResp.Name)
	s.NotEmpty(uploadResp.SHA256)
	s.Greater(uploadResp.Size, 0)

	// Get
	getOut, _, getCode := runCLI(
		"client", "file", "get",
		"--name", "test-int.conf",
		"--json",
	)
	s.Require().Equal(0, getCode)

	var getResp struct {
		Name string `json:"name"`
	}
	s.Require().NoError(parseJSON(getOut, &getResp))
	s.Equal("test-int.conf", getResp.Name)

	// Delete
	deleteOut, _, deleteCode := runCLI(
		"client", "file", "delete",
		"--name", "test-int.conf",
		"--json",
	)
	s.Require().Equal(0, deleteCode)

	var deleteResp struct {
		Name    string `json:"name"`
		Deleted bool   `json:"deleted"`
	}
	s.Require().NoError(parseJSON(deleteOut, &deleteResp))
	s.Equal("test-int.conf", deleteResp.Name)
	s.True(deleteResp.Deleted)
}

func (s *FileSmokeSuite) TestFileDeployStatus() {
	skipWriteOp(s.T(), "FILE_DEPLOY")

	filePath := writeTempFile(s.T(), "deploy-test-content\n")

	// Upload the file first so it exists in Object Store.
	uploadOut, _, uploadCode := runCLI(
		"client", "file", "upload",
		"--name", "test-deploy.conf",
		"--file", filePath,
		"--json",
	)
	s.Require().Equal(0, uploadCode)

	var uploadResp struct {
		Name string `json:"name"`
	}
	s.Require().NoError(parseJSON(uploadOut, &uploadResp))
	s.Equal("test-deploy.conf", uploadResp.Name)

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "deploys a file to the host",
			args: []string{
				"client", "node", "file", "deploy",
				"--object", "test-deploy.conf",
				"--path", "/tmp/osapi-deploy-test.conf",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.NotEmpty(result["job_id"])
			},
		},
		{
			name: "checks file deployment status",
			args: []string{
				"client", "node", "file", "status",
				"--path", "/tmp/osapi-deploy-test.conf",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.NotEmpty(result["job_id"])
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}

	// Clean up the uploaded file.
	_, _, deleteCode := runCLI(
		"client", "file", "delete",
		"--name", "test-deploy.conf",
		"--json",
	)
	s.Require().Equal(0, deleteCode)
}

func TestFileSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(FileSmokeSuite))
}
