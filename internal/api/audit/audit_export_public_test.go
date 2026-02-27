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
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	auditapi "github.com/retr0h/osapi/internal/api/audit"
	"github.com/retr0h/osapi/internal/api/audit/gen"
	auditstore "github.com/retr0h/osapi/internal/audit"
)

type AuditExportPublicTestSuite struct {
	suite.Suite

	handler *auditapi.Audit
	store   *fakeStore
	ctx     context.Context
}

func (s *AuditExportPublicTestSuite) SetupTest() {
	s.store = &fakeStore{}
	s.handler = auditapi.New(slog.Default(), s.store)
	s.ctx = context.Background()
}

func (s *AuditExportPublicTestSuite) TestGetAuditExport() {
	tests := []struct {
		name         string
		setupStore   func()
		validateFunc func(resp gen.GetAuditExportResponseObject)
	}{
		{
			name: "returns entries successfully",
			setupStore: func() {
				s.store.listAllEntries = []auditstore.Entry{
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
				}
			},
			validateFunc: func(resp gen.GetAuditExportResponseObject) {
				r, ok := resp.(gen.GetAuditExport200JSONResponse)
				s.True(ok)
				s.Equal(1, r.TotalItems)
				s.Len(r.Items, 1)
				s.Equal("user@example.com", r.Items[0].User)
			},
		},
		{
			name: "returns empty list",
			setupStore: func() {
				s.store.listAllEntries = []auditstore.Entry{}
			},
			validateFunc: func(resp gen.GetAuditExportResponseObject) {
				r, ok := resp.(gen.GetAuditExport200JSONResponse)
				s.True(ok)
				s.Equal(0, r.TotalItems)
				s.Empty(r.Items)
			},
		},
		{
			name: "returns entry with operation ID",
			setupStore: func() {
				s.store.listAllEntries = []auditstore.Entry{
					{
						ID:           "550e8400-e29b-41d4-a716-446655440000",
						Timestamp:    time.Now(),
						User:         "user@example.com",
						Roles:        []string{"admin"},
						Method:       "GET",
						Path:         "/node/hostname",
						OperationID:  "getSystemHostname",
						SourceIP:     "127.0.0.1",
						ResponseCode: 200,
						DurationMs:   42,
					},
				}
			},
			validateFunc: func(resp gen.GetAuditExportResponseObject) {
				r, ok := resp.(gen.GetAuditExport200JSONResponse)
				s.True(ok)
				s.Len(r.Items, 1)
				s.Require().NotNil(r.Items[0].OperationId)
				s.Equal("getSystemHostname", *r.Items[0].OperationId)
			},
		},
		{
			name: "returns 500 on store error",
			setupStore: func() {
				s.store.listAllErr = fmt.Errorf("store error")
			},
			validateFunc: func(resp gen.GetAuditExportResponseObject) {
				_, ok := resp.(gen.GetAuditExport500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.store.reset()
			tt.setupStore()
			resp, err := s.handler.GetAuditExport(s.ctx, gen.GetAuditExportRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestAuditExportPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AuditExportPublicTestSuite))
}
