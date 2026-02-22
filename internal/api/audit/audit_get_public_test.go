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

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"

	auditapi "github.com/retr0h/osapi/internal/api/audit"
	"github.com/retr0h/osapi/internal/api/audit/gen"
	auditstore "github.com/retr0h/osapi/internal/audit"
)

type AuditGetPublicTestSuite struct {
	suite.Suite

	handler *auditapi.Audit
	store   *fakeStore
	ctx     context.Context
}

func (s *AuditGetPublicTestSuite) SetupTest() {
	s.store = &fakeStore{}
	s.handler = auditapi.New(slog.Default(), s.store)
	s.ctx = context.Background()
}

func (s *AuditGetPublicTestSuite) TestGetAuditLogByID() {
	testID := uuid.New()

	tests := []struct {
		name         string
		id           openapi_types.UUID
		setupStore   func()
		validateFunc func(resp gen.GetAuditLogByIDResponseObject)
	}{
		{
			name: "returns entry successfully",
			id:   testID,
			setupStore: func() {
				s.store.getEntry = &auditstore.Entry{
					ID:           testID.String(),
					Timestamp:    time.Now(),
					User:         "user@example.com",
					Roles:        []string{"admin"},
					Method:       "GET",
					Path:         "/system/hostname",
					SourceIP:     "127.0.0.1",
					ResponseCode: 200,
					DurationMs:   42,
				}
			},
			validateFunc: func(resp gen.GetAuditLogByIDResponseObject) {
				r, ok := resp.(gen.GetAuditLogByID200JSONResponse)
				s.True(ok)
				s.Equal("user@example.com", r.Entry.User)
			},
		},
		{
			name: "returns 404 when not found",
			id:   testID,
			setupStore: func() {
				s.store.getErr = fmt.Errorf("get audit entry: not found")
			},
			validateFunc: func(resp gen.GetAuditLogByIDResponseObject) {
				_, ok := resp.(gen.GetAuditLogByID404JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "returns 500 on store error",
			id:   testID,
			setupStore: func() {
				s.store.getErr = fmt.Errorf("connection error")
			},
			validateFunc: func(resp gen.GetAuditLogByIDResponseObject) {
				_, ok := resp.(gen.GetAuditLogByID500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.store.reset()
			tt.setupStore()
			resp, err := s.handler.GetAuditLogByID(s.ctx, gen.GetAuditLogByIDRequestObject{
				Id: tt.id,
			})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestAuditGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AuditGetPublicTestSuite))
}
