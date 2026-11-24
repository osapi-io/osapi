// Copyright (c) 2024 John Dewey

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
package authtoken_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
)

type ScopePublicTestSuite struct {
	suite.Suite
}

func (s *ScopePublicTestSuite) TestInScope() {
	tests := []struct {
		name   string
		scope  []string
		target string
		want   bool
	}{
		{name: "when the context has no scope, each target is in it", target: "_all", want: true},
		{
			name:  "when the label target is under the scope",
			scope: []string{"group:web"}, target: "group:web.dev", want: true,
		},
		{name: "when a scoped token targets _all", scope: []string{"group:web"}, target: "_all"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()
			if tt.scope != nil {
				ctx = authtoken.WithTargets(ctx, tt.scope)
			}

			s.Equal(tt.scope, authtoken.TargetsFrom(ctx))
			s.Equal(tt.want, authtoken.InScope(ctx, tt.target))
		})
	}
}

func (s *ScopePublicTestSuite) TestSubject() {
	s.Empty(authtoken.SubjectFrom(context.Background()), "a request with no token")
	s.Equal("ops@example.com",
		authtoken.SubjectFrom(authtoken.WithSubject(context.Background(), "ops@example.com")))
}

func (s *ScopePublicTestSuite) TestSees() {
	tests := []struct {
		name   string
		scope  []string
		target string
		want   bool
	}{
		{name: "when the context has no scope, a record with no target is seen", want: true},
		{
			name:   "when the context has no scope, a record for _all is seen",
			target: "_all",
			want:   true,
		},
		{
			name:  "when the record's label target is under the scope",
			scope: []string{"group:web"}, target: "group:web.dev", want: true,
		},
		{name: "when a scoped token reads a record with no target", scope: []string{"group:web"}},
		{
			name:   "when a scoped token reads a record for _any",
			scope:  []string{"group:web"},
			target: "_any",
		},
		{
			name:  "when the record's label target is outside the scope",
			scope: []string{"group:web"}, target: "group:db",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()
			if tt.scope != nil {
				ctx = authtoken.WithTargets(ctx, tt.scope)
			}

			s.Equal(tt.want, authtoken.Sees(ctx, tt.target))
		})
	}
}

func TestScopePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ScopePublicTestSuite))
}
