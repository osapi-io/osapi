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

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
)

type SubjectsPublicTestSuite struct {
	suite.Suite
}

func (suite *SubjectsPublicTestSuite) SetupTest() {
	// Reset namespace to default before each test
	job.Init("")
}

func (suite *SubjectsPublicTestSuite) SetupSubTest() {}

func (suite *SubjectsPublicTestSuite) TearDownTest() {}

func (suite *SubjectsPublicTestSuite) TestParseSubject() {
	tests := []struct {
		name         string
		subject      string
		validateFunc func(string, string, error)
	}{
		{
			name:    "when parsing valid query subject",
			subject: "jobs.query.host.server-01",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("jobs.query", gotPrefix)
				suite.Equal("server-01", gotHostname)
			},
		},
		{
			name:    "when parsing valid modify subject",
			subject: "jobs.modify.host.web-01",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("jobs.modify", gotPrefix)
				suite.Equal("web-01", gotHostname)
			},
		},
		{
			name:    "when parsing a subject with a bare token that is not _any or _all",
			subject: "jobs.query.*",
			validateFunc: func(_ string, _ string, err error) {
				suite.ErrorContains(err, "invalid subject format")
			},
		},
		{
			name:    "when parsing subject with any hostname",
			subject: "jobs.modify._any",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("jobs.modify", gotPrefix)
				suite.Equal("_any", gotHostname)
			},
		},
		{
			name:    "when parsing invalid subject with too few parts",
			subject: "jobs.query",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if true {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("", gotPrefix)
				suite.Equal("", gotHostname)
			},
		},
		{
			name:    "when parsing host subject",
			subject: "jobs.query.host.server-01",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("jobs.query", gotPrefix)
				suite.Equal("server-01", gotHostname)
			},
		},
		{
			name:    "when parsing label subject",
			subject: "jobs.query.label.group.web",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("jobs.query", gotPrefix)
				suite.Equal("group:web", gotHostname)
			},
		},
		{
			name:    "when parsing hierarchical label subject",
			subject: "jobs.query.label.group.web.dev.us-east",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("jobs.query", gotPrefix)
				suite.Equal("group:web.dev.us-east", gotHostname)
			},
		},
		{
			name:    "when parsing invalid 4-part subject without host prefix",
			subject: "jobs.query.invalid.server1",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if true {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("", gotPrefix)
				suite.Equal("", gotHostname)
			},
		},
		{
			name:    "when parsing empty subject",
			subject: "",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if true {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("", gotPrefix)
				suite.Equal("", gotHostname)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.validateFunc(job.ParseSubject(tt.subject))
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestSanitizeHostname() {
	tests := []struct {
		name         string
		hostname     string
		validateFunc func(string)
	}{
		{
			name:     "when hostname has no special characters",
			hostname: "server01",
			validateFunc: func(got string) {
				suite.Equal("server01", got)
			},
		},
		{
			name:     "when hostname has hyphens",
			hostname: "web-server-01",
			validateFunc: func(got string) {
				suite.Equal("web_server_01", got)
			},
		},
		{
			name:     "when hostname has dots",
			hostname: "server.example.com",
			validateFunc: func(got string) {
				suite.Equal("server_example_com", got)
			},
		},
		{
			name:     "when hostname has hyphens and dots",
			hostname: "Johns-MacBook-Pro-2.local",
			validateFunc: func(got string) {
				suite.Equal("Johns_MacBook_Pro_2_local", got)
			},
		},
		{
			name:     "when hostname has mixed special characters",
			hostname: "test@host#123.domain!",
			validateFunc: func(got string) {
				suite.Equal("test_host_123_domain_", got)
			},
		},
		{
			name:     "when hostname has underscores (should be preserved)",
			hostname: "test_server_01",
			validateFunc: func(got string) {
				suite.Equal("test_server_01", got)
			},
		},
		{
			name:     "when hostname has numbers",
			hostname: "server123",
			validateFunc: func(got string) {
				suite.Equal("server123", got)
			},
		},
		{
			name:     "when hostname is empty",
			hostname: "",
			validateFunc: func(got string) {
				suite.Equal("", got)
			},
		},
		{
			name:     "when hostname has spaces",
			hostname: "my server name",
			validateFunc: func(got string) {
				suite.Equal("my_server_name", got)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.SanitizeHostname(tt.hostname)
			tt.validateFunc(got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestParseTarget() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(string, string, string)
	}{
		{
			name:   "when target is _any",
			target: "_any",
			validateFunc: func(rt string, key string, value string) {
				suite.Equal("_any", rt)
				suite.Equal("", key)
				suite.Equal("", value)
			},
		},
		{
			name:   "when target is _all",
			target: "_all",
			validateFunc: func(rt string, key string, value string) {
				suite.Equal("_all", rt)
				suite.Equal("", key)
				suite.Equal("", value)
			},
		},
		{
			name:   "when target is a hostname",
			target: "server1",
			validateFunc: func(rt string, key string, value string) {
				suite.Equal("host", rt)
				suite.Equal("server1", key)
				suite.Equal("", value)
			},
		},
		{
			name:   "when target is a flat label",
			target: "role:web",
			validateFunc: func(rt string, key string, value string) {
				suite.Equal("label", rt)
				suite.Equal("role", key)
				suite.Equal("web", value)
			},
		},
		{
			name:   "when target is a hierarchical label",
			target: "group:web.dev.us-east",
			validateFunc: func(rt string, key string, value string) {
				suite.Equal("label", rt)
				suite.Equal("group", key)
				suite.Equal("web.dev.us-east", value)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.validateFunc(job.ParseTarget(tt.target))
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestBuildSubjectFromTarget() {
	tests := []struct {
		name         string
		prefix       string
		target       string
		validateFunc func(string)
	}{
		{
			name:   "when target is _any",
			prefix: "jobs.query",
			target: "_any",
			validateFunc: func(got string) {
				suite.Equal("jobs.query._any", got)
			},
		},
		{
			name:   "when target is _all",
			prefix: "jobs.modify",
			target: "_all",
			validateFunc: func(got string) {
				suite.Equal("jobs.modify._all", got)
			},
		},
		{
			name:   "when target is a hostname",
			prefix: "jobs.query",
			target: "server1",
			validateFunc: func(got string) {
				suite.Equal("jobs.query.host.server1", got)
			},
		},
		{
			name:   "when target is a dotted hostname",
			prefix: "jobs.query",
			target: "my-server.local",
			validateFunc: func(got string) {
				suite.Equal("jobs.query.host.my_server_local", got)
			},
		},
		{
			name:   "when target is a flat label",
			prefix: "jobs.query",
			target: "role:web",
			validateFunc: func(got string) {
				suite.Equal("jobs.query.label.role.web", got)
			},
		},
		{
			name:   "when target is a hierarchical label",
			prefix: "jobs.query",
			target: "group:web.dev.us-east",
			validateFunc: func(got string) {
				suite.Equal("jobs.query.label.group.web.dev.us-east", got)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.BuildSubjectFromTarget(tt.prefix, tt.target)
			tt.validateFunc(got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestIsBroadcastTarget() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(bool)
	}{
		{
			name:   "when target is _all",
			target: "_all",
			validateFunc: func(got bool) {
				suite.Equal(true, got)
			},
		},
		{
			name:   "when target is a label",
			target: "role:web",
			validateFunc: func(got bool) {
				suite.Equal(true, got)
			},
		},
		{
			name:   "when target is a hierarchical label",
			target: "group:web.dev",
			validateFunc: func(got bool) {
				suite.Equal(true, got)
			},
		},
		{
			name:   "when target is _any",
			target: "_any",
			validateFunc: func(got bool) {
				suite.Equal(false, got)
			},
		},
		{
			name:   "when target is a hostname",
			target: "server1",
			validateFunc: func(got bool) {
				suite.Equal(false, got)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.IsBroadcastTarget(tt.target)
			tt.validateFunc(got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestInit() {
	tests := []struct {
		name             string
		namespace        string
		wantQueryPrefix  string
		wantModifyPrefix string
		wantBuildQuery   string
	}{
		{
			name:             "when namespace is empty",
			namespace:        "",
			wantQueryPrefix:  "jobs.query",
			wantModifyPrefix: "jobs.modify",
			wantBuildQuery:   "jobs.query._any",
		},
		{
			name:             "when namespace is set",
			namespace:        "osapi",
			wantQueryPrefix:  "osapi.jobs.query",
			wantModifyPrefix: "osapi.jobs.modify",
			wantBuildQuery:   "osapi.jobs.query._any",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			job.Init(tt.namespace)
			defer job.Init("")

			suite.Equal(tt.wantQueryPrefix, job.JobsQueryPrefix)
			suite.Equal(tt.wantModifyPrefix, job.JobsModifyPrefix)
			suite.Equal(tt.wantBuildQuery, job.BuildSubjectFromTarget(job.JobsQueryPrefix, "_any"))
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestParseSubjectWithNamespace() {
	tests := []struct {
		name         string
		namespace    string
		subject      string
		validateFunc func(string, string, error)
	}{
		{
			name:      "when parsing namespaced query subject",
			namespace: "osapi",
			subject:   "osapi.jobs.query.host.server-01",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("osapi.jobs.query", gotPrefix)
				suite.Equal("server-01", gotHostname)
			},
		},
		{
			name:      "when parsing namespaced modify subject",
			namespace: "osapi",
			subject:   "osapi.jobs.modify._any",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("osapi.jobs.modify", gotPrefix)
				suite.Equal("_any", gotHostname)
			},
		},
		{
			name:      "when parsing namespaced label subject",
			namespace: "osapi",
			subject:   "osapi.jobs.query.label.group.web.dev",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if false {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("osapi.jobs.query", gotPrefix)
				suite.Equal("group:web.dev", gotHostname)
			},
		},
		{
			name:      "when the namespace holds a jobs token of its own",
			namespace: "team.jobs",
			subject:   "team.jobs.jobs.query._any",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				suite.NoError(err)
				suite.Equal("team.jobs.jobs.query", gotPrefix)
				suite.Equal("_any", gotHostname)
			},
		},
		{
			name:      "when parsing invalid namespaced subject with too few parts",
			namespace: "osapi",
			subject:   "osapi.jobs",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if true {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("", gotPrefix)
				suite.Equal("", gotHostname)
			},
		},
		{
			name:      "when parsing namespaced subject without jobs token",
			namespace: "osapi",
			subject:   "osapi.other.query._any",
			validateFunc: func(gotPrefix string, gotHostname string, err error) {
				if true {
					suite.Error(err)
					return
				}

				suite.NoError(err)
				suite.Equal("", gotPrefix)
				suite.Equal("", gotHostname)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			job.Init(tt.namespace)
			defer job.Init("")

			tt.validateFunc(job.ParseSubject(tt.subject))
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestApplyNamespaceToInfraName() {
	tests := []struct {
		name         string
		namespace    string
		infraName    string
		validateFunc func(string)
	}{
		{
			name:      "when namespace is empty",
			namespace: "",
			infraName: "JOBS",
			validateFunc: func(got string) {
				suite.Equal("JOBS", got)
			},
		},
		{
			name:      "when namespace is set",
			namespace: "osapi",
			infraName: "JOBS",
			validateFunc: func(got string) {
				suite.Equal("osapi-JOBS", got)
			},
		},
		{
			name:      "when namespace applied to KV bucket",
			namespace: "osapi",
			infraName: "job-queue",
			validateFunc: func(got string) {
				suite.Equal("osapi-job-queue", got)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.ApplyNamespaceToInfraName(tt.namespace, tt.infraName)
			tt.validateFunc(got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestApplyNamespaceToSubjects() {
	tests := []struct {
		name         string
		namespace    string
		subjects     string
		validateFunc func(string)
	}{
		{
			name:      "when namespace is empty",
			namespace: "",
			subjects:  "jobs.>",
			validateFunc: func(got string) {
				suite.Equal("jobs.>", got)
			},
		},
		{
			name:      "when namespace is set",
			namespace: "osapi",
			subjects:  "jobs.>",
			validateFunc: func(got string) {
				suite.Equal("osapi.jobs.>", got)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.ApplyNamespaceToSubjects(tt.namespace, tt.subjects)
			tt.validateFunc(got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestRecipients() {
	agents := []job.AgentInfo{
		{MachineID: "m-web1", Hostname: "web-01", Labels: map[string]string{"group": "web.prod"}},
		{
			MachineID: "m-web2",
			Hostname:  "web-02",
			Labels:    map[string]string{"group": "web"},
			State:     types.AgentCordoned,
		},
		{MachineID: "m-db", Hostname: "db-01", Labels: map[string]string{"group": "db"}},
	}
	tests := []struct {
		name     string
		target   string
		resolved string
		want     []string
	}{
		{
			name:   "when the target is _all, it is each agent that takes jobs",
			target: "_all",
			want:   []string{"m-web1", "m-db"},
		},
		{
			name:   "when the target is _any, it is each agent that takes jobs",
			target: "_any",
			want:   []string{"m-web1", "m-db"},
		},
		{
			name:   "when the target is a label, it is each agent under it",
			target: "group:web",
			want:   []string{"m-web1"},
		},
		{
			name:     "when the target is a host, it is the agent of its machine ID",
			target:   "web-02",
			resolved: "m-web2",
			want:     []string{"m-web2"},
		},
		{
			name:     "when the target is a hostname that did not resolve, it is that agent",
			target:   "db-01",
			resolved: "db-01",
			want:     []string{"m-db"},
		},
		{name: "when no agent matches, it is none", target: "other", resolved: "other"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var got []string
			for _, a := range job.Recipients(agents, tt.target, tt.resolved) {
				got = append(got, a.MachineID)
			}
			suite.Equal(tt.want, got)
		})
	}
}

func (suite *SubjectsPublicTestSuite) TestExpectedAgentHostnames() {
	agents := []job.AgentInfo{
		{Hostname: "web-01", Labels: map[string]string{"group": "web.dev.us-east"}},
		{Hostname: "web-02", Labels: map[string]string{"group": "web.dev.us-west"}},
		{Hostname: "db-01", Labels: map[string]string{"group": "db.prod"}},
		{Hostname: "plain-01"},
	}

	tests := []struct {
		name         string
		agents       []job.AgentInfo
		target       string
		validateFunc func([]string)
	}{
		{
			name:   "when target is _all returns all hostnames",
			agents: agents,
			target: "_all",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-01", "web-02", "db-01", "plain-01"}, got)
			},
		},
		{
			name:   "when label exact match returns matching hostnames",
			agents: agents,
			target: "group:web.dev.us-east",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-01"}, got)
			},
		},
		{
			name:   "when label prefix match returns matching hostnames",
			agents: agents,
			target: "group:web",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-01", "web-02"}, got)
			},
		},
		{
			name:   "when no agents match label returns nil",
			agents: agents,
			target: "group:staging",
			validateFunc: func(got []string) {
				suite.Equal([]string(nil), got)
			},
		},
		{
			name:   "when agent list is empty returns nil",
			agents: []job.AgentInfo{},
			target: "_all",
			validateFunc: func(got []string) {
				suite.Equal([]string(nil), got)
			},
		},
		{
			name:   "when target is a hostname returns nil",
			agents: agents,
			target: "web-01",
			validateFunc: func(got []string) {
				suite.Equal([]string(nil), got)
			},
		},
		{
			name:   "when target is _any returns nil",
			agents: agents,
			target: "_any",
			validateFunc: func(got []string) {
				suite.Equal([]string(nil), got)
			},
		},
		{
			name: "when _all excludes cordoned agents",
			agents: []job.AgentInfo{
				{Hostname: "web-01"},
				{Hostname: "web-02", State: types.AgentCordoned},
				{Hostname: "web-03"},
			},
			target: "_all",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-01", "web-03"}, got)
			},
		},
		{
			name: "when _all excludes draining agents",
			agents: []job.AgentInfo{
				{Hostname: "web-01"},
				{Hostname: "web-02", State: types.AgentDraining},
			},
			target: "_all",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-01"}, got)
			},
		},
		{
			name: "when _all excludes pending agents",
			agents: []job.AgentInfo{
				{Hostname: "web-01"},
				{Hostname: "web-02", State: types.AgentPending},
			},
			target: "_all",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-01"}, got)
			},
		},
		{
			name: "when label match excludes cordoned agents",
			agents: []job.AgentInfo{
				{
					Hostname: "web-01",
					Labels:   map[string]string{"group": "web.dev"},
					State:    types.AgentCordoned,
				},
				{Hostname: "web-02", Labels: map[string]string{"group": "web.dev"}},
			},
			target: "group:web",
			validateFunc: func(got []string) {
				suite.Equal([]string{"web-02"}, got)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := job.ExpectedAgentHostnames(tt.agents, tt.target)
			tt.validateFunc(got)
		})
	}
}

func TestSubjectsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SubjectsPublicTestSuite))
}
