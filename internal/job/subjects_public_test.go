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
		name     string
		hostname string
		want     string
	}{
		{
			name:     "when building query subject for specific server",
			hostname: "server-01",
			want:     "jobs.query.server-01",
		},
		{
			name:     "when building query subject for web server",
			hostname: "web-server",
			want:     "jobs.query.web-server",
		},
		{
			name:     "when building with wildcard hostname",
			hostname: job.AllHosts,
			want:     "jobs.query.*",
		},
		{
			name:     "when building with any hostname",
			hostname: job.AnyHost,
			want:     "jobs.query._any",
		},
		{
			name:     "when building query subject for all hosts",
			hostname: "",
			want:     "jobs.query.*",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var got string
			if tt.hostname == "" {
				got = job.BuildQuerySubjectForAllHosts()
			} else {
				got = job.BuildQuerySubject(tt.hostname)
			}
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestBuildModifySubject() {
	tests := []struct {
		name     string
		hostname string
		want     string
	}{
		{
			name:     "when building modify subject for specific server",
			hostname: "server-01",
			want:     "jobs.modify.server-01",
		},
		{
			name:     "when building modify subject for db server",
			hostname: "db-server",
			want:     "jobs.modify.db-server",
		},
		{
			name:     "when building with wildcard hostname",
			hostname: job.AllHosts,
			want:     "jobs.modify.*",
		},
		{
			name:     "when building with any hostname",
			hostname: job.AnyHost,
			want:     "jobs.modify._any",
		},
		{
			name:     "when building modify subject for all hosts",
			hostname: "",
			want:     "jobs.modify.*",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var got string
			if tt.hostname == "" {
				got = job.BuildModifySubjectForAllHosts()
			} else {
				got = job.BuildModifySubject(tt.hostname)
			}
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestParseSubject() {
	tests := []struct {
		name         string
		subject      string
		wantPrefix   string
		wantHostname string
		wantErr      bool
	}{
		{
			name:         "when parsing valid query subject",
			subject:      "jobs.query.server-01",
			wantPrefix:   "jobs.query",
			wantHostname: "server-01",
			wantErr:      false,
		},
		{
			name:         "when parsing valid modify subject",
			subject:      "jobs.modify.web-01",
			wantPrefix:   "jobs.modify",
			wantHostname: "web-01",
			wantErr:      false,
		},
		{
			name:         "when parsing subject with wildcard hostname",
			subject:      "jobs.query.*",
			wantPrefix:   "jobs.query",
			wantHostname: "*",
			wantErr:      false,
		},
		{
			name:         "when parsing subject with any hostname",
			subject:      "jobs.modify._any",
			wantPrefix:   "jobs.modify",
			wantHostname: "_any",
			wantErr:      false,
		},
		{
			name:    "when parsing invalid subject with too few parts",
			subject: "jobs.query",
			wantErr: true,
		},
		{
			name:    "when parsing invalid subject with too many parts",
			subject: "jobs.query.server-01.extra.part",
			wantErr: true,
		},
		{
			name:    "when parsing empty subject",
			subject: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			gotPrefix, gotHostname, err := job.ParseSubject(tt.subject)

			if tt.wantErr {
				suite.Error(err)
				return
			}

			suite.NoError(err)
			suite.Equal(tt.wantPrefix, gotPrefix)
			suite.Equal(tt.wantHostname, gotHostname)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestGetLocalHostname() {
	// This test uses the real system hostname
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

func (suite *SubjectsPublicTestSuite) TestBuildWorkerSubscriptionPattern() {
	tests := []struct {
		name     string
		hostname string
		want     []string
	}{
		{
			name:     "when building subscription pattern for specific hostname",
			hostname: "web-server-01",
			want: []string{
				"jobs.*.web-server-01",
				"jobs.*._any",
				"jobs.*._all",
			},
		},
		{
			name:     "when building subscription pattern for localhost",
			hostname: "localhost",
			want: []string{
				"jobs.*.localhost",
				"jobs.*._any",
				"jobs.*._all",
			},
		},
		{
			name:     "when building subscription pattern with complex hostname",
			hostname: "api.example.com",
			want: []string{
				"jobs.*.api.example.com",
				"jobs.*._any",
				"jobs.*._all",
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildWorkerSubscriptionPattern(tt.hostname)
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestBuildWorkerQueueGroup() {
	tests := []struct {
		name     string
		category string
		want     string
	}{
		{
			name:     "when building queue group for system category",
			category: "system",
			want:     "workers.system",
		},
		{
			name:     "when building queue group for network category",
			category: "network",
			want:     "workers.network",
		},
		{
			name:     "when building queue group for jobs category",
			category: "jobs",
			want:     "workers.jobs",
		},
		{
			name:     "when building queue group with empty category",
			category: "",
			want:     "workers.",
		},
		{
			name:     "when building queue group with complex category",
			category: "custom-service",
			want:     "workers.custom-service",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildWorkerQueueGroup(tt.category)
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestIsSpecialHostname() {
	tests := []struct {
		name     string
		hostname string
		want     bool
	}{
		{
			name:     "when hostname is AllHosts wildcard",
			hostname: job.AllHosts,
			want:     true,
		},
		{
			name:     "when hostname is AnyHost",
			hostname: job.AnyHost,
			want:     true,
		},
		{
			name:     "when hostname is LocalHost",
			hostname: job.LocalHost,
			want:     true,
		},
		{
			name:     "when hostname is BroadcastHost",
			hostname: job.BroadcastHost,
			want:     true,
		},
		{
			name:     "when hostname is regular server name",
			hostname: "web-server-01",
			want:     false,
		},
		{
			name:     "when hostname is localhost",
			hostname: "localhost",
			want:     false,
		},
		{
			name:     "when hostname is FQDN",
			hostname: "api.example.com",
			want:     false,
		},
		{
			name:     "when hostname is empty",
			hostname: "",
			want:     false,
		},
		{
			name:     "when hostname looks like special but isn't exact",
			hostname: "_any_server",
			want:     false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.IsSpecialHostname(tt.hostname)
			suite.Equal(tt.want, got)
		})
	}
}

func TestSubjectsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SubjectsPublicTestSuite))
}
