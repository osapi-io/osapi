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

package validation_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/validation"
)

type HostnamePublicTestSuite struct {
	suite.Suite
}

func (s *HostnamePublicTestSuite) TestTargetHostname() {
	tests := []struct {
		name         string
		agents       []validation.AgentTarget
		hostname     string
		validateFunc func(msg string, ok bool)
	}{
		{
			name:     "when the hostname is a literal agent",
			agents:   []validation.AgentTarget{{Hostname: "server1"}},
			hostname: "server1",
			validateFunc: func(msg string, ok bool) {
				s.True(ok)
				s.Empty(msg)
			},
		},
		{
			name:     "when the hostname is the _all routing value",
			agents:   []validation.AgentTarget{{Hostname: "server1"}},
			hostname: "_all",
			validateFunc: func(msg string, ok bool) {
				s.True(ok)
				s.Empty(msg)
			},
		},
		{
			name: "when the hostname is a label selector",
			agents: []validation.AgentTarget{
				{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			},
			hostname: "group:web",
			validateFunc: func(msg string, ok bool) {
				s.True(ok)
				s.Empty(msg)
			},
		},
		{
			name:     "when the hostname is empty",
			agents:   []validation.AgentTarget{{Hostname: "server1"}},
			hostname: "",
			validateFunc: func(msg string, ok bool) {
				s.False(ok)
				s.NotEmpty(msg)
			},
		},
		{
			name:     "when no agent matches the hostname",
			agents:   []validation.AgentTarget{{Hostname: "server1"}},
			hostname: "nosuchhost",
			validateFunc: func(msg string, ok bool) {
				s.False(ok)
				s.NotEmpty(msg)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agents := tt.agents
			validation.RegisterTargetValidator(
				func(_ context.Context) ([]validation.AgentTarget, error) {
					return agents, nil
				},
			)

			tt.validateFunc(validation.TargetHostname(tt.hostname))
		})
	}
}

func TestHostnamePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HostnamePublicTestSuite))
}
