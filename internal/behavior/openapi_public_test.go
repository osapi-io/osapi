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

package behavior_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type OpenAPIPublicTestSuite struct {
	suite.Suite
}

func (s *OpenAPIPublicTestSuite) TestOperationID() {
	b := top()
	tests := []struct {
		method string
		want   string
	}{
		{method: "list", want: "GetNodeSysctl"},
		{method: "get", want: "GetNodeSysctlByKey"},
		{method: "update", want: "PutNodeSysctl"},
	}

	for _, tc := range tests {
		s.Run(tc.method, func() {
			s.Equal(tc.want, b.OperationID(b.Method(tc.method)))
		})
	}
}

func (s *OpenAPIPublicTestSuite) TestTag() {
	tests := []struct {
		name string
		b    *behavior.Behavior
		want string
	}{
		{
			name: "when the behavior has no group",
			b:    behavior.Define("sysctl", "Kernel", behavior.Node),
			want: "sysctl_operations",
		},
		{
			name: "when the behavior is grouped",
			b:    behavior.Define("dns", "Resolvers", behavior.Node, behavior.Under("network")),
			want: "network_dns_operations",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.b.Tag())
		})
	}
}

func TestOpenAPIPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OpenAPIPublicTestSuite))
}
