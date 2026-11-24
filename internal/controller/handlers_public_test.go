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
	"log/slog"
	"testing"

	"github.com/osapi-io/osapi/internal/telemetry/traces"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	auditmocks "github.com/osapi-io/osapi/internal/audit/mocks"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller"
	"github.com/osapi-io/osapi/internal/controller/mocks"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type HandlersPublicTestSuite struct {
	suite.Suite
}

func TestHandlersPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HandlersPublicTestSuite))
}

func (s *HandlersPublicTestSuite) TestTraceReader() {
	s.Nil(controller.ExportTraceReader(config.TracingConfig{}), "no URL is no backend")
	s.IsType(
		&traces.Jaeger{},
		controller.ExportTraceReader(config.TracingConfig{URL: "http://jaeger:16686"}),
	)
}

func (s *HandlersPublicTestSuite) TestHandlers() {
	ctrl := gomock.NewController(s.T())
	var cfg config.Config
	cfg.Controller.API.Security.SigningKeys = []config.SigningKey{{ID: "test", Key: "test-key"}}
	cfg.Controller.API.Security.Roles = map[client.Role]config.CustomRole{
		"reader": {Permissions: []client.Permission{"status:read"}},
	}
	base := len(controller.ExportHandlers(slog.Default(), cfg, nil, nil, ""))

	tests := []struct {
		name  string
		cfg   func() config.Config
		audit bool
		files bool
		want  int
	}{
		{
			name:  "when the audit store is there the endpoint groups serve it, with no handler of its own",
			cfg:   func() config.Config { return cfg },
			audit: true,
			want:  base,
		},
		{
			name:  "when the Object Store is there the endpoint groups serve it, with no handler of its own",
			cfg:   func() config.Config { return cfg },
			files: true,
			want:  base,
		},
		{
			name: "when the UI is off leaves the UI handler out",
			cfg: func() config.Config {
				c := cfg
				enabled := false
				c.Controller.UI.Enabled = &enabled
				return c
			},
			want: base - 1,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var store *auditmocks.MockStore
			if tc.audit {
				store = auditmocks.NewMockStore(ctrl)
			}
			var objStore *mocks.MockFileBuckets
			if tc.files {
				objStore = mocks.NewMockFileBuckets(ctrl)
			}

			var got int
			switch {
			case tc.audit:
				got = len(controller.ExportHandlers(slog.Default(), tc.cfg(), store, nil, ""))
			case tc.files:
				got = len(
					controller.ExportHandlers(slog.Default(), tc.cfg(), nil, objStore, "objects"),
				)
			default:
				got = len(controller.ExportHandlers(slog.Default(), tc.cfg(), nil, nil, ""))
			}

			s.Equal(tc.want, got)
		})
	}
}
