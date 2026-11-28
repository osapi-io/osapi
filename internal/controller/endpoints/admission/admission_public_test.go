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

package admission_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	store "github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/admission"
	"github.com/osapi-io/osapi/internal/controller/endpoints/admission/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/admission/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-admission-group"}

// AdmissionPublicTestSuite runs requests through the admission group's generated
// routes, with the auth, the decoding and the status mapping of the
// endpoint runtime, as the controller serves them.
type AdmissionPublicTestSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	store *mocks.MockStore
	jobs  *mocks.MockJobs
}

func TestAdmissionPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AdmissionPublicTestSuite))
}

func (s *AdmissionPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mocks.NewMockStore(s.ctrl)
	s.jobs = mocks.NewMockJobs(s.ctrl)
	s.jobs.EXPECT().ListAgents(gomock.Any()).Return(nil, nil).AnyTimes()
	s.jobs.EXPECT().AgentsLastSeen(gomock.Any()).Return(nil, nil).AnyTimes()
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, nil)

	return t
}

func (s *AdmissionPublicTestSuite) serve(
	deps admission.Deps,
	method, path, body, tok string,
) *httptest.ResponseRecorder {
	srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, deps)
	e := echo.New()
	srv.Mount(e, gen.Routes(srv))

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

func (s *AdmissionPublicTestSuite) TestRoutes() {
	read, write := token("admission:read"), token("admission:read", "admission:write")
	agentWrite := token("agent:read", "agent:write")

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		tok    string
		setup  func()
		code   int
		want   string
	}{
		{
			name:   "when list has admission:read, it is a 200 with the records",
			method: http.MethodGet,
			path:   "/api/admission",
			tok:    read,
			code:   http.StatusOK,
			want:   `"machine_id":"m-1"`,
			setup: func() {
				s.store.EXPECT().List(gomock.Any()).
					Return([]store.Record{{MachineID: "m-1", State: types.AdmissionPending}}, nil)
			},
		},
		{
			name: "when list's state is not one of the four, it is a 400", method: http.MethodGet,
			path: "/api/admission?state=accepted", tok: read, code: http.StatusBadRequest,
		},
		{
			name: "when there is no token, it is a 401", method: http.MethodGet,
			path: "/api/admission", code: http.StatusUnauthorized,
		},
		{
			name:   "when get finds no record, it is a 404",
			method: http.MethodGet,
			path:   "/api/admission/m-x",
			tok:    read,
			code:   http.StatusNotFound,
			setup:  func() { s.store.EXPECT().Get(gomock.Any(), "m-x").Return(store.Record{}, store.ErrNotFound) },
		},
		{
			name: "when admit matches, it is a 204", method: http.MethodPost,
			path: "/api/admission/m-1/admit", body: `{"fingerprint":"SHA256:ab"}`, tok: write,
			code:  http.StatusNoContent,
			setup: func() { s.store.EXPECT().Admit(gomock.Any(), "m-1", "SHA256:ab").Return(nil) },
		},
		{
			name: "when admit has no fingerprint, it is a 400", method: http.MethodPost,
			path: "/api/admission/m-1/admit", body: `{}`, tok: write, code: http.StatusBadRequest,
		},
		{
			name: "when admit has no body, it is a 400", method: http.MethodPost,
			path: "/api/admission/m-1/admit", tok: write, code: http.StatusBadRequest,
		},
		{
			name: "when admit's fingerprint differs, it is a 409", method: http.MethodPost,
			path: "/api/admission/m-1/admit", body: `{"fingerprint":"SHA256:no"}`, tok: write,
			code: http.StatusConflict, want: "osapi pki key get --name agent",
			setup: func() {
				s.store.EXPECT().
					Admit(gomock.Any(), "m-1", "SHA256:no").
					Return(store.ErrFingerprintMismatch)
			},
		},
		{
			name: "when admit has only admission:read, it is a 403", method: http.MethodPost,
			path: "/api/admission/m-1/admit", body: `{"fingerprint":"SHA256:ab"}`, tok: read,
			code: http.StatusForbidden,
		},
		{
			name:   "when admit has agent:write but not admission:write, it is a 403",
			method: http.MethodPost,
			path:   "/api/admission/m-1/admit",
			body:   `{"fingerprint":"SHA256:ab"}`,
			tok:    agentWrite,
			code:   http.StatusForbidden,
		},
		{
			name:   "when reject has a reason, it is a 204",
			method: http.MethodPost,
			path:   "/api/admission/m-1/reject",
			body:   `{"reason":"unknown host"}`,
			tok:    write,
			code:   http.StatusNoContent,
			setup:  func() { s.store.EXPECT().Reject(gomock.Any(), "m-1", "unknown host").Return(nil) },
		},
		{
			name:   "when revoke finds no admitted machine, it is a 404",
			method: http.MethodPost,
			path:   "/api/admission/m-1/revoke",
			tok:    write,
			code:   http.StatusNotFound,
			setup:  func() { s.store.EXPECT().Revoke(gomock.Any(), "m-1").Return(store.ErrNoAcceptedAgent) },
		},
		{
			name: "when delete finds the machine, it is a 204", method: http.MethodDelete,
			path: "/api/admission/m-1", tok: write, code: http.StatusNoContent,
			setup: func() { s.store.EXPECT().Delete(gomock.Any(), "m-1").Return(nil) },
		},
		{
			name: "when delete has only admission:read, it is a 403", method: http.MethodDelete,
			path: "/api/admission/m-1", tok: read, code: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}
			deps := admission.Deps{Store: s.store, Jobs: s.jobs}

			rec := s.serve(deps, tc.method, tc.path, tc.body, tc.tok)

			s.Equal(tc.code, rec.Code)
			if tc.want != "" {
				s.Contains(rec.Body.String(), tc.want)
			}
		})
	}
}
