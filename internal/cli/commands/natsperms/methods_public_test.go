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

package natsperms_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/cli/commands/natsperms"
	"github.com/osapi-io/osapi/internal/config"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestGet() {
	cfg := config.Config{}
	cfg.Controller.NATS.Namespace = "osapi"
	cfg.NATS.Stream = config.NATSStream{Name: "JOBS", Subjects: "jobs.>"}
	cfg.NATS.Registry.Bucket = "agent-registry"
	d := natsperms.Deps{Config: cfg}

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{name: "a role prints its subjects", role: "agent"},
		{name: "the controller answers requests", role: "controller"},
		{name: "a role that osapi does not have is the error", role: "admin", wantErr: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := natsperms.Get(context.Background(), d, natsperms.GetInput{Role: tt.role})
			if tt.wantErr {
				s.ErrorIs(err, bus.ErrRole)
				return
			}
			s.Require().NoError(err)
			s.Equal("osapi", got.Namespace)
			s.Contains(got.Subscribe, "_INBOX.>")
			s.Contains(got.NSC, "--allow-pub \"")
			s.Equal(tt.role == "controller", got.Responses)
			s.Equal(tt.role == "controller", strings.Contains(got.NSC, "--allow-pub-response 1"))
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
