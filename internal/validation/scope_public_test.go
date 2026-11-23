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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/validation"
)

type ScopePublicTestSuite struct {
	suite.Suite
}

func (s *ScopePublicTestSuite) TestScopeEntry() {
	tests := []struct {
		name  string
		entry string
		want  bool
	}{
		{name: "when the entry is a hostname", entry: "web-01", want: true},
		{name: "when the entry is a label", entry: "group:web", want: true},
		{name: "when the entry is a nested label", entry: "group:web.dev", want: true},
		{name: "when the entry is _all", entry: "_all", want: false},
		{name: "when the entry is _any", entry: "_any", want: false},
		{name: "when the entry is _local", entry: "_local", want: false},
		{name: "when the entry is empty", entry: "", want: false},
		{name: "when the label has no value", entry: "group:", want: false},
		{name: "when the label has no key", entry: ":web", want: false},
		{name: "when the label value is not subject safe", entry: "group:w*b", want: false},
		{name: "when the hostname starts with a dash", entry: "-web", want: false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, validation.ScopeEntry(tt.entry))
		})
	}
}

func (s *ScopePublicTestSuite) TestInScope() {
	agents := []validation.AgentTarget{
		{
			MachineID: "m-web1", Hostname: "web-01", Verified: true,
			Labels: map[string]string{"group": "web.dev"},
		},
		{
			MachineID: "m-db1", Hostname: "db-01", Verified: true,
			Labels: map[string]string{"group": "db"},
		},
	}
	tests := []struct {
		name    string
		target  string
		scope   []string
		listErr error
		want    bool
	}{
		{name: "when the scope is empty", target: "_all", want: true},
		{name: "when a scoped token targets _all", target: "_all", scope: []string{"group:web"}},
		{name: "when a scoped token targets _any", target: "_any", scope: []string{"group:web"}},
		{
			name:   "when the label target is the scope's label",
			target: "group:web", scope: []string{"group:web"}, want: true,
		},
		{
			name:   "when the label target is under the scope's label",
			target: "group:web.dev", scope: []string{"group:web"}, want: true,
		},
		{
			name:   "when the label target is wider than the scope's label",
			target: "group:web", scope: []string{"group:web.dev"},
		},
		{
			name:   "when the label target only shares a prefix of the value",
			target: "group:webapp", scope: []string{"group:web"},
		},
		{
			name:   "when the label target has another key",
			target: "env:web", scope: []string{"group:web"},
		},
		{
			name:   "when the label target is not in a hostname scope",
			target: "group:web", scope: []string{"web-01"},
		},
		{
			name:   "when the scope names the host",
			target: "web-01", scope: []string{"web-01"}, want: true,
		},
		{
			name:   "when the scope names the host and the target is its machine ID",
			target: "m-web1", scope: []string{"web-01"}, want: true,
		},
		{
			name:   "when the host has a label the scope covers",
			target: "web-01", scope: []string{"group:web"}, want: true,
		},
		{
			name:   "when the host has no label the scope covers",
			target: "db-01", scope: []string{"group:web"},
		},
		{
			name:   "when the host is another host of a hostname scope",
			target: "db-01", scope: []string{"web-01"},
		},
		{
			name:   "when the host is not registered",
			target: "gone-01", scope: []string{"group:web"},
		},
		{
			name:   "when the registry cannot be read",
			target: "web-01", scope: []string{"group:web"}, listErr: errors.New("nats down"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			validation.RegisterTargetValidator(
				func(_ context.Context) ([]validation.AgentTarget, error) {
					return agents, tt.listErr
				},
			)

			s.Equal(tt.want, validation.InScope(context.Background(), tt.target, tt.scope))
		})
	}
}

func TestScopePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ScopePublicTestSuite))
}
