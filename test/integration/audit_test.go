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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuditSmokeSuite struct {
	suite.Suite
}

func (s *AuditSmokeSuite) TestAuditList() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns audit entries list",
			args: []string{"client", "audit", "list", "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				s.Require().NoError(parseJSON(stdout, &result))
				s.Contains(result, "items")
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

func (s *AuditSmokeSuite) TestAuditExport() {
	exportPath := filepath.Join(tempDir, "audit-export.json")

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "exports audit log to file",
			args: []string{"client", "audit", "export", "--output", exportPath, "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				info, err := os.Stat(exportPath)
				s.Require().NoError(err)
				s.Greater(info.Size(), int64(0))
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

func (s *AuditSmokeSuite) TestAuditGet() {
	listOut, _, listCode := runCLI("client", "audit", "list", "--json")
	s.Require().Equal(0, listCode)

	var listResp struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	s.Require().NoError(parseJSON(listOut, &listResp))
	s.Require().NotEmpty(listResp.Items, "audit list must contain at least one entry")

	auditID := listResp.Items[0].ID

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns audit entry details for known id",
			args: []string{"client", "audit", "get", "--audit-id", auditID, "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result struct {
					Entry struct {
						ID   string `json:"id"`
						User string `json:"user"`
					} `json:"entry"`
				}
				s.Require().NoError(parseJSON(stdout, &result))
				s.Equal(auditID, result.Entry.ID)
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

func TestAuditSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(AuditSmokeSuite))
}
