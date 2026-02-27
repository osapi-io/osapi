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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	auditGen "github.com/retr0h/osapi/internal/api/audit/gen"
	auditstore "github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
)

type AuditExportIntegrationTestSuite struct {
	suite.Suite

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *AuditExportIntegrationTestSuite) SetupTest() {
	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *AuditExportIntegrationTestSuite) TestGetAuditExportValidation() {
	tests := []struct {
		name         string
		store        *fakeStore
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request returns entries",
			store: &fakeStore{
				listAllEntries: []auditstore.Entry{
					{
						ID:           "550e8400-e29b-41d4-a716-446655440000",
						Timestamp:    time.Now(),
						User:         "user@example.com",
						Roles:        []string{"admin"},
						Method:       "GET",
						Path:         "/node/hostname",
						SourceIP:     "127.0.0.1",
						ResponseCode: 200,
						DurationMs:   42,
					},
				},
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total_items":1`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			a := api.New(suite.appConfig, suite.logger)

			auditHandler := newTestAuditHandler(suite.logger, tc.store)
			strictHandler := auditGen.NewStrictHandler(auditHandler, nil)
			auditGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodGet,
				"/audit/export",
				nil,
			)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacAuditExportTestSigningKey = "test-signing-key-for-rbac-export"

func (suite *AuditExportIntegrationTestSuite) TestGetAuditExportRBAC() {
	tokenManager := authtoken.New(suite.logger)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		wantCode     int
		wantContains []string
	}{
		{
			name: "when no token returns 401",
			setupAuth: func(_ *http.Request) {
				// No auth header set
			},
			wantCode:     http.StatusUnauthorized,
			wantContains: []string{"Bearer token required"},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacAuditExportTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"job:read"},
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with audit:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacAuditExportTestSigningKey,
					[]string{"admin"},
					"test-user",
					nil,
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total_items":0`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			store := &fakeStore{
				listAllEntries: []auditstore.Entry{},
			}

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacAuditExportTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetAuditHandler(store)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodGet,
				"/audit/export",
				nil,
			)
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

func TestAuditExportIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuditExportIntegrationTestSuite))
}
