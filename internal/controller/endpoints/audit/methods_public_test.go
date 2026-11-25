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
package audit_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	auditstore "github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/audit"
	"github.com/osapi-io/osapi/internal/controller/endpoints/audit/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/audit/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

const (
	entryID = "550e8400-e29b-41d4-a716-446655440000"
)

// signingKey signs every token a case sends.
var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-audit-group"}

var when = time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)

// full is an entry with each field set, and bare one with only those the
// store always has.
var (
	full = auditstore.Entry{
		ID: entryID, Timestamp: when, User: "ops@example.com",
		Roles: []types.Role{types.RoleAdmin}, Method: "POST", Path: "/api/node/web-01/command/exec",
		OperationID: "PostNodeCommandExec", SourceIP: "192.168.1.100",
		ResponseCode: 202, DurationMs: 42, TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		JobID: "7c9e6679-7425-40de-944b-e07fc1f90ae7", RequestSummary: `{"command":"uptime"}`,
		Steps: []auditstore.HostSteps{{
			Hostname: "web-01",
			Steps:    []types.Step{{Behavior: "managedfile", Method: "update", Error: "denied"}},
		}},
	}
	bare = auditstore.Entry{
		ID: entryID, Timestamp: when, User: "ops@example.com", Roles: []types.Role{types.RoleRead},
		Method: "GET", Path: "/api/agent", SourceIP: "10.0.0.1", ResponseCode: 200,
	}
)

// scope is a target scope, and web, all and short are entries of requests
// to a target in it, to _all, and to the bare path of a node.
var (
	scope = []string{"group:web"}
	web   = auditstore.Entry{
		ID: "web", Method: "GET", Path: "/api/node/group:web.dev/hostname",
	}
	all = auditstore.Entry{
		ID: "all", Method: "GET", Path: "/api/node/_all/hostname",
	}
	short = auditstore.Entry{ID: "short", Method: "GET", Path: "/api/node/group:web"}
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	store *mocks.MockStore
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mocks.NewMockStore(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) deps(
	configured bool,
) audit.Deps {
	if !configured {
		return audit.Deps{}
	}

	return audit.Deps{Store: s.store}
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		in           audit.ListInput
		scope        []string
		noStore      bool
		setup        func()
		validateFunc func(audit.ListAuditResponse, error)
	}{
		{
			name: "when the caller gives no limit, it reads a page of 20 and maps each field",
			setup: func() {
				s.store.EXPECT().List(gomock.Any(), 20, 0).
					Return([]auditstore.Entry{full, bare}, 150, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(150, got.TotalItems)
				s.Require().Len(got.Items, 2)
				s.Equal(audit.AuditEntry{
					ID:             entryID,
					Timestamp:      when,
					User:           "ops@example.com",
					Roles:          []types.Role{types.RoleAdmin},
					Method:         "POST",
					Path:           "/api/node/web-01/command/exec",
					OperationID:    "PostNodeCommandExec",
					SourceIP:       "192.168.1.100",
					ResponseCode:   202,
					DurationMs:     42,
					TraceID:        "4bf92f3577b34da6a3ce929d0e0e4736",
					JobID:          "7c9e6679-7425-40de-944b-e07fc1f90ae7",
					RequestSummary: `{"command":"uptime"}`,
					Steps: []audit.HostSteps{{
						Hostname: "web-01",
						Steps: []types.Step{
							{Behavior: "managedfile", Method: "update", Error: "denied"},
						},
					}},
					Chain: []string{"web-01: managedfile update (failed)"},
				}, got.Items[0])
				s.Empty(got.Items[1].Steps)
				s.Empty(got.Items[1].Chain)
				s.Empty(got.Items[1].OperationID)
				s.Empty(got.Items[1].TraceID)
				s.Empty(got.Items[1].JobID)
				s.Empty(got.Items[1].RequestSummary)
			},
		},
		{
			name: "when the caller gives a limit and an offset, it reads that page",
			in:   audit.ListInput{Limit: 10, Offset: 20},
			setup: func() {
				s.store.EXPECT().List(gomock.Any(), 10, 20).Return(nil, 0, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(audit.ListAuditResponse{TotalItems: 0, Items: []audit.AuditEntry{}}, got)
			},
		},
		{
			name: "when a filter is set, it reads the whole log and pages the entries that match",
			in:   audit.ListInput{Writes: true, Limit: 1},
			setup: func() {
				other := full
				other.ID = "other"
				s.store.EXPECT().ListAll(gomock.Any()).
					Return([]auditstore.Entry{bare, full, other}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems, "the GET is not a write")
				s.Require().Len(got.Items, 1)
				s.Equal(entryID, got.Items[0].ID)
			},
		},
		{
			name: "when each filter is set, an entry must match each",
			in: audit.ListInput{
				User: "ops@example.com", Operation: "EXEC", Status: 202, Offset: 0,
			},
			setup: func() {
				otherUser, otherStatus, otherPath := full, full, full
				otherUser.User = "dev@example.com"
				otherStatus.ResponseCode = 403
				otherPath.OperationID, otherPath.Path = "GetNodeHostname", "/api/node/web-01/hostname"
				byPath := full
				byPath.OperationID = ""
				s.store.EXPECT().ListAll(gomock.Any()).
					Return([]auditstore.Entry{full, otherUser, otherStatus, otherPath, byPath}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems, "the operation ID or the path, in any case")
			},
		},
		{
			name: "when the offset is past the entries that match, the page is empty",
			in:   audit.ListInput{Status: 500, Offset: 5},
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return([]auditstore.Entry{full}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(audit.ListAuditResponse{TotalItems: 0, Items: []audit.AuditEntry{}}, got)
			},
		},
		{
			name: "when a filter is set and the store fails, the caller gets a message without the store's error",
			in:   audit.ListInput{User: "x"},
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(_ audit.ListAuditResponse, err error) {
				s.EqualError(err, "failed to list audit entries")
			},
		},
		{
			name:    "when the audit log is not configured, it is ErrNotConfigured and names the setting",
			noStore: true,
			setup:   func() {},
			validateFunc: func(_ audit.ListAuditResponse, err error) {
				s.ErrorIs(err, audit.ErrNotConfigured)
				s.ErrorContains(err, "nats.audit.stream")
			},
		},
		{
			name: "when the store fails, the caller gets a message without the store's error",
			setup: func() {
				s.store.EXPECT().
					List(gomock.Any(), 20, 0).
					Return(nil, 0, errors.New("nats: timeout"))
			},
			validateFunc: func(_ audit.ListAuditResponse, err error) {
				s.EqualError(err, "failed to list audit entries")
				s.ErrorContains(errors.Unwrap(err), "nats: timeout")
			},
		},
		{
			name:  "when the token is scoped, it reads the whole log and counts only the entries in its scope",
			in:    audit.ListInput{Limit: 1},
			scope: scope,
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).
					Return([]auditstore.Entry{full, web, bare, all, short}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems, "only the entries for group:web and its children")
				s.Require().Len(got.Items, 1)
				s.Equal("web", got.Items[0].ID)
			},
		},
		{
			name:  "when the token is scoped and the caller filters, both apply",
			in:    audit.ListInput{Operation: "hostname"},
			scope: scope,
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).
					Return([]auditstore.Entry{web, all, short}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(1, got.TotalItems)
				s.Equal("web", got.Items[0].ID)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(audit.List(ctx, s.deps(!tc.noStore), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestExport() {
	tests := []struct {
		name         string
		scope        []string
		noStore      bool
		setup        func()
		validateFunc func(audit.ListAuditResponse, error)
	}{
		{
			name: "when the store has entries, it answers each and their count",
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return([]auditstore.Entry{full, bare}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems)
				s.Len(got.Items, 2)
				s.Equal("PostNodeCommandExec", got.Items[0].OperationID)
			},
		},
		{
			name: "when the store is empty, it answers an empty list",
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return(nil, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(audit.ListAuditResponse{TotalItems: 0, Items: []audit.AuditEntry{}}, got)
			},
		},
		{
			name:    "when the audit log is not configured, it is ErrNotConfigured",
			noStore: true,
			setup:   func() {},
			validateFunc: func(_ audit.ListAuditResponse, err error) {
				s.ErrorIs(err, audit.ErrNotConfigured)
			},
		},
		{
			name: "when the store fails, the caller gets a message without the store's error",
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(_ audit.ListAuditResponse, err error) {
				s.EqualError(err, "failed to list all audit entries")
			},
		},
		{
			name:  "when the token is scoped, it answers only the entries in its scope",
			scope: scope,
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).
					Return([]auditstore.Entry{full, web, bare, all, short}, nil)
			},
			validateFunc: func(got audit.ListAuditResponse, err error) {
				s.Require().NoError(err)
				s.Equal(2, got.TotalItems)
				s.Require().Len(got.Items, 2)
				s.Equal("web", got.Items[0].ID)
				s.Equal("short", got.Items[1].ID)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(
				audit.Export(ctx, s.deps(!tc.noStore), audit.ExportInput{}),
			)
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		scope        []string
		noStore      bool
		setup        func()
		validateFunc func(audit.AuditEntryResponse, error)
	}{
		{
			name: "when the store has the entry, it answers it",
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(&full, nil)
			},
			validateFunc: func(got audit.AuditEntryResponse, err error) {
				s.Require().NoError(err)
				s.Equal(entryID, got.Entry.ID)
				s.Equal("7c9e6679-7425-40de-944b-e07fc1f90ae7", got.Entry.JobID)
				s.Equal([]string{"web-01: managedfile update (failed)"}, got.Entry.Chain)
			},
		},
		{
			name: "when the entry has no roles, it answers an empty list, never null",
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(&auditstore.Entry{
					ID:     entryID,
					Actor:  auditstore.ActorLocal,
					User:   "ops",
					Action: "jointoken.create",
				}, nil)
			},
			validateFunc: func(got audit.AuditEntryResponse, err error) {
				s.Require().NoError(err)
				s.NotNil(got.Entry.Roles)
				s.Empty(got.Entry.Roles)
				s.Equal("local", got.Entry.Actor)
			},
		},
		{
			name: "when the store does not have the entry, it is ErrNotFound and names the ID",
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(nil, auditstore.ErrNotFound)
			},
			validateFunc: func(_ audit.AuditEntryResponse, err error) {
				s.ErrorIs(err, audit.ErrNotFound)
				s.ErrorContains(err, entryID)
			},
		},
		{
			name:    "when the audit log is not configured, it is ErrNotConfigured",
			noStore: true,
			setup:   func() {},
			validateFunc: func(_ audit.AuditEntryResponse, err error) {
				s.ErrorIs(err, audit.ErrNotConfigured)
			},
		},
		{
			name: "when the store fails, the caller gets a message without the store's error",
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(nil, errors.New("nats: timeout"))
			},
			validateFunc: func(_ audit.AuditEntryResponse, err error) {
				s.EqualError(err, "failed to get audit entry")
			},
		},
		{
			name:  "when a scoped token gets an entry in its scope, it answers it",
			scope: scope,
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(&web, nil)
			},
			validateFunc: func(got audit.AuditEntryResponse, err error) {
				s.Require().NoError(err)
				s.Equal("web", got.Entry.ID)
			},
		},
		{
			name:  "when a scoped token gets an entry for _all, it is ErrNotFound",
			scope: scope,
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(&all, nil)
			},
			validateFunc: func(_ audit.AuditEntryResponse, err error) {
				s.ErrorIs(err, audit.ErrNotFound)
			},
		},
		{
			name:  "when a scoped token gets an entry with no target, it is ErrNotFound",
			scope: scope,
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(&bare, nil)
			},
			validateFunc: func(_ audit.AuditEntryResponse, err error) {
				s.ErrorIs(err, audit.ErrNotFound)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(
				audit.Get(ctx, s.deps(!tc.noStore), audit.GetInput{ID: entryID}),
			)
		})
	}
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, nil)

	return t
}

// TestRoutes runs requests through the generated routes, with the auth, the
// decoding, the validation and the status mapping of the endpoint runtime.
func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name     string
		path     string
		token    string
		noStore  bool
		setup    func()
		wantCode int
		wantBody string
	}{
		{
			name:  "when the token has audit:read, list is a 200",
			path:  "/api/audit",
			token: token("audit:read"),
			setup: func() {
				s.store.EXPECT().List(gomock.Any(), 20, 0).Return(nil, 0, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"total_items":0,"items":[]}`,
		},
		{
			name:  "when the caller gives a limit and an offset, list reads that page",
			path:  "/api/audit?limit=10&offset=20",
			token: token("audit:read"),
			setup: func() {
				s.store.EXPECT().List(gomock.Any(), 10, 20).Return(nil, 0, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "when the limit is over 100, it is a 400",
			path:     "/api/audit?limit=101",
			token:    token("audit:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when the offset is negative, it is a 400",
			path:     "/api/audit?offset=-1",
			token:    token("audit:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "when there is no token, it is a 401",
			path:     "/api/audit",
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks audit:read, it is a 403",
			path:     "/api/audit/export",
			token:    token("agent:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:  "when the token has audit:read, export is a 200",
			path:  "/api/audit/export",
			token: token("audit:read"),
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return([]auditstore.Entry{bare}, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:  "when the entry is there, get is a 200",
			path:  "/api/audit/" + entryID,
			token: token("audit:read"),
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(&bare, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "when the ID is not a UUID, it is a 400",
			path:     "/api/audit/not-a-uuid",
			token:    token("audit:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:  "when the entry is not there, it is a 404",
			path:  "/api/audit/" + entryID,
			token: token("audit:read"),
			setup: func() {
				s.store.EXPECT().Get(gomock.Any(), entryID).Return(nil, auditstore.ErrNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "when the audit log is not configured, it is a 503",
			path:     "/api/audit",
			token:    token("audit:read"),
			noStore:  true,
			setup:    func() {},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name:  "when the store fails, it is a 500 that does not name the store's error",
			path:  "/api/audit/export",
			token: token("audit:read"),
			setup: func() {
				s.store.EXPECT().ListAll(gomock.Any()).Return(nil, errors.New("nats: timeout"))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"failed to list all audit entries","reason":"server_error"}`,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, s.deps(!tc.noStore))
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
		})
	}
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
