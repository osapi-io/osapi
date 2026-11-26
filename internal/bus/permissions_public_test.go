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

package bus_test

import (
	"strings"
	"testing"

	natsserver "github.com/nats-io/nats-server/v2/server"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/config"
)

type PermissionsPublicTestSuite struct {
	suite.Suite
	cfg config.NATS
}

func (s *PermissionsPublicTestSuite) SetupTest() {
	s.cfg = config.NATS{
		Stream:     config.NATSStream{Name: "JOBS", Subjects: "jobs.>"},
		KV:         config.NATSKV{Bucket: "job-queue"},
		Registry:   config.NATSRegistry{Bucket: "agent-registry"},
		Facts:      config.NATSFacts{Bucket: "agent-facts"},
		State:      config.NATSState{Bucket: "agent-state"},
		FileState:  config.NATSFileState{Bucket: "file-state"},
		Enrollment: config.NATSEnrollment{Bucket: "enrollment"},
		Secrets:    config.NATSSecrets{Bucket: "secrets"},
		Objects:    config.NATSObjects{Bucket: "file-objects"},
		Audit:      config.NATSAudit{Stream: "AUDIT", Subject: "audit"},
	}
}

// allows reports whether one of patterns matches subject, as NATS matches
// them: * is one token, > is one or more at the end.
func allows(
	patterns []string,
	subject string,
) bool {
	for _, p := range patterns {
		pt, st := strings.Split(p, "."), strings.Split(subject, ".")
		ok := true
		for i, tok := range pt {
			if tok == ">" {
				ok = len(st) > i
				break
			}
			if i >= len(st) || (tok != "*" && tok != st[i]) {
				ok = false
				break
			}
			if i == len(pt)-1 && len(st) != len(pt) {
				ok = false
			}
		}
		if ok {
			return true
		}
	}

	return false
}

func (s *PermissionsPublicTestSuite) TestRolePermissions() {
	tests := []struct {
		name   string
		role   bus.Role
		pub    []string
		notPub []string
		sub    []string
		notSub []string
	}{
		{
			name: "an agent takes jobs, writes its buckets and asks to enroll",
			role: bus.RoleAgent,
			pub: []string{
				"$JS.API.INFO",
				"$JS.API.CONSUMER.DURABLE.CREATE.osapi-JOBS.agent-m1",
				"$JS.API.CONSUMER.CREATE.osapi-JOBS.agent-m1.osapi.jobs.query.host.m1",
				"$JS.API.CONSUMER.MSG.NEXT.osapi-JOBS.agent-m1",
				"$JS.ACK.osapi-JOBS.agent-m1.1.2.3.4.5",
				"$KV.osapi-agent-registry.agents.m1",
				"$KV.osapi-job-queue.status.j1",
				"$JS.API.DIRECT.GET.KV_osapi-file-state.x",
				"$JS.API.STREAM.INFO.OBJ_osapi-file-objects",
				"osapi.enroll.request",
				"osapi.join.renew",
			},
			notPub: []string{
				"$JS.API.STREAM.UPDATE.KV_osapi-file-state",
				"$JS.API.STREAM.CREATE.OBJ_osapi-file-objects",
				"$JS.API.STREAM.DELETE.osapi-JOBS",
				"$KV.osapi-secrets.x",
				"$KV.osapi-enrollment.x",
				"$O.osapi-file-objects.x",
				"osapi.jobs.modify.host.m2",
				"osapi.audit.x",
				"osapi.enroll.response.m1",
			},
			sub:    []string{"_INBOX.abc", "osapi.enroll.response.m1"},
			notSub: []string{"osapi.enroll.request", "osapi.jobs.query._all"},
		},
		{
			name: "the controller sends jobs, audits and answers enrollment",
			role: bus.RoleController,
			pub: []string{
				"osapi.jobs.modify.host.m1",
				"osapi.jobs.query._any",
				"osapi.audit.e1",
				"osapi.enroll.response.m1",
				"$KV.osapi-secrets.x",
				"$KV.osapi-enrollment.m1",
				"$O.osapi-file-objects.M.x",
				"$JS.API.STREAM.PURGE.OBJ_osapi-file-objects",
				"$JS.API.STREAM.MSG.GET.osapi-AUDIT",
				"$JS.API.STREAM.UPDATE.OBJ_osapi-file-objects",
			},
			notPub: []string{
				"$JS.API.STREAM.DELETE.osapi-JOBS",
				"$JS.API.STREAM.UPDATE.osapi-JOBS",
				"osapi.enroll.request",
			},
			sub: []string{
				"_INBOX.abc",
				"osapi.enroll.request",
				"osapi.join.request",
				"osapi.join.renew",
			},
			notSub: []string{"osapi.jobs.query._all"},
		},
		{
			name: "the join user sends the join request, and nothing else",
			role: bus.RoleJoin,
			pub:  []string{"osapi.join.request"},
			notPub: []string{
				"osapi.enroll.request",
				"$JS.API.INFO",
				"$KV.osapi-agent-registry.x",
			},
			sub:    []string{"_INBOX.abc"},
			notSub: []string{"osapi.join.request"},
		},
		{
			name: "setup makes and updates each stream, and nothing else",
			role: bus.RoleSetup,
			pub: []string{
				"$JS.API.STREAM.CREATE.osapi-JOBS",
				"$JS.API.STREAM.UPDATE.osapi-JOBS-DLQ",
				"$JS.API.STREAM.UPDATE.osapi-AUDIT",
				"$JS.API.STREAM.CREATE.KV_osapi-secrets",
				"$JS.API.STREAM.UPDATE.OBJ_osapi-file-objects",
				"$KV.osapi-agent-registry.agents.server",
			},
			notPub: []string{
				"$JS.API.STREAM.DELETE.osapi-JOBS",
				"$KV.osapi-secrets.x",
				"osapi.jobs.modify._all",
			},
			sub: []string{"_INBOX.abc"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			p, err := bus.RolePermissions(tt.role, "osapi", s.cfg)
			s.Require().NoError(err)
			for _, subj := range tt.pub {
				s.True(allows(p.Publish, subj), "may publish %s", subj)
			}
			for _, subj := range tt.notPub {
				s.False(allows(p.Publish, subj), "may not publish %s", subj)
			}
			for _, subj := range tt.sub {
				s.True(allows(p.Subscribe, subj), "may subscribe %s", subj)
			}
			for _, subj := range tt.notSub {
				s.False(allows(p.Subscribe, subj), "may not subscribe %s", subj)
			}
		})
	}
}

func (s *PermissionsPublicTestSuite) TestControllerAnswersRequests() {
	p, err := bus.RolePermissions(bus.RoleController, "osapi", s.cfg)
	s.Require().NoError(err)
	s.True(p.Responses)
	s.False(allows(p.Publish, "_INBOX.abc"), "it answers through the response permission only")

	auth := config.NATSServerAuth{Type: "user_pass", Users: []config.NATSServerUser{
		{Username: "c", Password: "p", Role: "controller"},
	}}
	opts, err := bus.ServerOptions("127.0.0.1", 4222, "", auth, config.ServerTLS{})
	s.Require().NoError(err)
	s.Require().NoError(bus.ApplyRoles(opts, auth, "osapi", s.cfg))
	s.Require().NotNil(opts.Users[0].Permissions.Response)
	s.Equal(1, opts.Users[0].Permissions.Response.MaxMsgs)
}

func (s *PermissionsPublicTestSuite) TestNothingNamed() {
	for _, role := range bus.Roles {
		p, err := bus.RolePermissions(role, "osapi", config.NATS{})
		s.Require().NoError(err)
		for _, subj := range append(p.Publish, p.Subscribe...) {
			s.NotContains(subj, "..", "%s: %s", role, subj)
			s.False(strings.HasSuffix(subj, "."), "%s: %s", role, subj)
			s.NotContains(subj, "KV_.", "%s: %s", role, subj)
			s.NotContains(subj, "OBJ_.", "%s: %s", role, subj)
			s.NotContains(subj, "osapi-.", "%s: %s", role, subj)
		}
	}
}

func (s *PermissionsPublicTestSuite) TestNoNamespace() {
	p, err := bus.RolePermissions(bus.RoleAgent, "", s.cfg)
	s.Require().NoError(err)
	s.True(allows(p.Publish, "enroll.request"))
	s.True(allows(p.Publish, "$KV.agent-registry.agents.m1"))
}

func (s *PermissionsPublicTestSuite) TestUnknownRole() {
	_, err := bus.RolePermissions("admin", "osapi", s.cfg)
	s.ErrorIs(err, bus.ErrRole)
}

// Two public nkeys of users.
const (
	agentNKey = "UDXU4RCSJNZOIQHZNWXHXORDPRTGNJAHAHFRGZNEEJCPQTT2M7NLCNF4"
	otherNKey = "UCQC4ORR5JFIBTV44AGMNY5Q7UNSLDBQIQZ3SS3JAG4JMBZTEDZHRMYX"
)

func (s *PermissionsPublicTestSuite) TestApplyRoles() {
	tests := []struct {
		name    string
		auth    config.NATSServerAuth
		wantErr string
		check   func(opts *natsserver.Options)
	}{
		{
			name: "a user with a role gets its permissions, one without keeps each subject",
			auth: config.NATSServerAuth{Type: "user_pass", Users: []config.NATSServerUser{
				{Username: "agent", Password: "a", Role: "agent"},
				{Username: "admin", Password: "b"},
			}},
			check: func(opts *natsserver.Options) {
				s.Require().NotNil(opts.Users[0].Permissions)
				s.True(allows(opts.Users[0].Permissions.Publish.Allow, "osapi.enroll.request"))
				s.False(allows(opts.Users[0].Permissions.Publish.Allow, "osapi.jobs.modify._all"))
				s.True(allows(opts.Users[0].Permissions.Subscribe.Allow, "_INBOX.x"))
				s.Nil(opts.Users[0].Permissions.Response, "an agent answers no request")
				s.Nil(opts.Users[1].Permissions)
			},
		},
		{
			name: "an nkey with a role gets its permissions, one without keeps each subject",
			auth: config.NATSServerAuth{
				Type:      "nkey",
				NKeys:     []string{agentNKey, otherNKey},
				NKeyRoles: map[string]string{agentNKey: "agent"},
			},
			check: func(opts *natsserver.Options) {
				s.Require().Len(opts.Nkeys, 2)
				s.Require().NotNil(opts.Nkeys[0].Permissions)
				s.True(allows(opts.Nkeys[0].Permissions.Publish.Allow, "osapi.enroll.request"))
				s.Nil(opts.Nkeys[1].Permissions)
			},
		},
		{
			name: "an nkey with a role that osapi does not have is an error",
			auth: config.NATSServerAuth{
				Type:      "nkey",
				NKeys:     []string{agentNKey},
				NKeyRoles: map[string]string{agentNKey: "admin"},
			},
			wantErr: "nkey " + agentNKey + ": no such role",
		},
		{
			name: "an unknown role is an error",
			auth: config.NATSServerAuth{Type: "user_pass", Users: []config.NATSServerUser{
				{Username: "x", Password: "a", Role: "admin"},
			}},
			wantErr: "user x: no such role",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts, err := bus.ServerOptions("127.0.0.1", 4222, "", tt.auth, config.ServerTLS{})
			s.Require().NoError(err)
			err = bus.ApplyRoles(opts, tt.auth, "osapi", s.cfg)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				return
			}
			s.Require().NoError(err)
			tt.check(opts)
		})
	}
}

func TestPermissionsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PermissionsPublicTestSuite))
}
