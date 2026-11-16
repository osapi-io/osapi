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

package controller_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/controller/mocks"
)

type AuditPublicTestSuite struct {
	suite.Suite
}

func TestAuditPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AuditPublicTestSuite))
}

func (s *AuditPublicTestSuite) TestNewAuditStore() {
	tests := []struct {
		name      string
		stream    string
		setup     func(*mocks.MockAuditStreams, *gomock.Controller)
		wantStore bool
		wantErr   string
	}{
		{
			name: "when the config names no stream returns no store",
		},
		{
			name:   "when the stream opens returns a store",
			stream: "AUDIT",
			setup: func(nc *mocks.MockAuditStreams, _ *gomock.Controller) {
				nc.EXPECT().StreamCreateOrUpdate(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStore: true,
		},
		{
			name:   "when the stream is not made returns the error",
			stream: "AUDIT",
			setup: func(nc *mocks.MockAuditStreams, _ *gomock.Controller) {
				nc.EXPECT().StreamCreateOrUpdate(gomock.Any(), gomock.Any()).Return(errBoom)
			},
			wantErr: "create audit stream: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			nc := mocks.NewMockAuditStreams(ctrl)
			if tc.setup != nil {
				tc.setup(nc, ctrl)
			}

			store, err := controller.ExportNewAuditStore(
				context.Background(),
				slog.Default(),
				nc,
				"ns",
				config.NATSAudit{Stream: tc.stream, Subject: "audit"},
			)

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.wantStore, store != nil)
		})
	}
}
