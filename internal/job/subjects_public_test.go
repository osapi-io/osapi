// Copyright (c) 2025 John Dewey

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

package job_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
)

type SubjectsPublicTestSuite struct {
	suite.Suite
}

func (suite *SubjectsPublicTestSuite) SetupTest() {}

func (suite *SubjectsPublicTestSuite) SetupSubTest() {}

func (suite *SubjectsPublicTestSuite) TearDownTest() {}

func (suite *SubjectsPublicTestSuite) TestBuildQuerySubject() {
	tests := []struct {
		name      string
		hostname  string
		category  string
		operation string
		want      string
	}{
		{
			name:      "when building system status query subject",
			hostname:  "server-01",
			category:  job.SubjectCategorySystem,
			operation: job.SystemOperationStatus,
			want:      "jobs.query.server-01.system.status",
		},
		{
			name:      "when building network DNS query subject",
			hostname:  "web-server",
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationDNS,
			want:      "jobs.query.web-server.network.dns",
		},
		{
			name:      "when building with wildcard hostname",
			hostname:  job.AllHosts,
			category:  job.SubjectCategorySystem,
			operation: job.SystemOperationHostname,
			want:      "jobs.query.*.system.hostname",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildQuerySubject(tt.hostname, tt.category, tt.operation)
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestBuildModifySubject() {
	tests := []struct {
		name      string
		hostname  string
		category  string
		operation string
		want      string
	}{
		{
			name:      "when building network DNS modify subject",
			hostname:  "server-01",
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationDNS,
			want:      "jobs.modify.server-01.network.dns",
		},
		{
			name:      "when building network ping modify subject",
			hostname:  "db-server",
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationPing,
			want:      "jobs.modify.db-server.network.ping",
		},
		{
			name:      "when building with wildcard hostname",
			hostname:  job.AllHosts,
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationDNS,
			want:      "jobs.modify.*.network.dns",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildModifySubject(tt.hostname, tt.category, tt.operation)
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestBuildQuerySubjectForAllHosts() {
	tests := []struct {
		name      string
		category  string
		operation string
		want      string
	}{
		{
			name:      "when building system status query for all hosts",
			category:  job.SubjectCategorySystem,
			operation: job.SystemOperationStatus,
			want:      "jobs.query.*.system.status",
		},
		{
			name:      "when building network DNS query for all hosts",
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationDNS,
			want:      "jobs.query.*.network.dns",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildQuerySubjectForAllHosts(tt.category, tt.operation)
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestBuildModifySubjectForAllHosts() {
	tests := []struct {
		name      string
		category  string
		operation string
		want      string
	}{
		{
			name:      "when building network DNS modify for all hosts",
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationDNS,
			want:      "jobs.modify.*.network.dns",
		},
		{
			name:      "when building network ping modify for all hosts",
			category:  job.SubjectCategoryNetwork,
			operation: job.NetworkOperationPing,
			want:      "jobs.modify.*.network.ping",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildModifySubjectForAllHosts(tt.category, tt.operation)
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestParseSubject() {
	tests := []struct {
		name          string
		subject       string
		wantPrefix    string
		wantHostname  string
		wantCategory  string
		wantOperation string
		wantErr       bool
	}{
		{
			name:          "when parsing valid query subject",
			subject:       "jobs.query.server-01.system.status",
			wantPrefix:    "jobs.query",
			wantHostname:  "server-01",
			wantCategory:  "system",
			wantOperation: "status",
			wantErr:       false,
		},
		{
			name:          "when parsing valid modify subject",
			subject:       "jobs.modify.web-01.network.dns",
			wantPrefix:    "jobs.modify",
			wantHostname:  "web-01",
			wantCategory:  "network",
			wantOperation: "dns",
			wantErr:       false,
		},
		{
			name:          "when parsing subject with wildcard hostname",
			subject:       "jobs.query.*.system.hostname",
			wantPrefix:    "jobs.query",
			wantHostname:  "*",
			wantCategory:  "system",
			wantOperation: "hostname",
			wantErr:       false,
		},
		{
			name:    "when parsing invalid subject with too few parts",
			subject: "jobs.query.server-01",
			wantErr: true,
		},
		{
			name:          "when parsing valid dotted operation subject",
			subject:       "jobs.query.server-01.system.hostname.get",
			wantPrefix:    "jobs.query",
			wantHostname:  "server-01",
			wantCategory:  "system",
			wantOperation: "hostname.get",
			wantErr:       false,
		},
		{
			name:    "when parsing empty subject",
			subject: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			gotPrefix, gotHostname, gotCategory, gotOperation, err := job.ParseSubject(tt.subject)

			if tt.wantErr {
				suite.Error(err)
				return
			}

			suite.NoError(err)
			suite.Equal(tt.wantPrefix, gotPrefix)
			suite.Equal(tt.wantHostname, gotHostname)
			suite.Equal(tt.wantCategory, gotCategory)
			suite.Equal(tt.wantOperation, gotOperation)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestGetLocalHostname() {
	// This test just ensures the function doesn't error
	hostname, err := job.GetLocalHostname()
	suite.NoError(err)
	suite.NotEmpty(hostname)
}

func (suite *SubjectsPublicTestSuite) TestSanitizeHostname() {
	tests := []struct {
		name     string
		hostname string
		want     string
	}{
		{
			name:     "when hostname has no special characters",
			hostname: "server01",
			want:     "server01",
		},
		{
			name:     "when hostname has hyphens",
			hostname: "web-server-01",
			want:     "web_server_01",
		},
		{
			name:     "when hostname has dots",
			hostname: "server.example.com",
			want:     "server_example_com",
		},
		{
			name:     "when hostname has hyphens and dots",
			hostname: "Johns-MacBook-Pro-2.local",
			want:     "Johns_MacBook_Pro_2_local",
		},
		{
			name:     "when hostname has mixed special characters",
			hostname: "test@host#123.domain!",
			want:     "test_host_123_domain_",
		},
		{
			name:     "when hostname has underscores (should be preserved)",
			hostname: "test_server_01",
			want:     "test_server_01",
		},
		{
			name:     "when hostname has numbers",
			hostname: "server123",
			want:     "server123",
		},
		{
			name:     "when hostname is empty",
			hostname: "",
			want:     "",
		},
		{
			name:     "when hostname has spaces",
			hostname: "my server name",
			want:     "my_server_name",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.SanitizeHostname(tt.hostname)
			suite.Equal(tt.want, got)
		})
	}
}

func TestSubjectsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SubjectsPublicTestSuite))
}
