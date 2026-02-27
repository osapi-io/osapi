// Copyright (c) 2025 John Dewey

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

package client

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AgentTestSuite struct {
	suite.Suite
}

func (s *AgentTestSuite) TestSanitizeKeyForNATS() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid characters only",
			input:    "validKey123",
			expected: "validKey123",
		},
		{
			name:     "alphanumeric with underscores and hyphens",
			input:    "valid_key-123",
			expected: "valid_key-123",
		},
		{
			name:     "hostname with dots",
			input:    "server.example.com",
			expected: "server_example_com",
		},
		{
			name:     "hostname with special characters",
			input:    "agent.host-name@domain.com",
			expected: "agent_host-name_domain_com",
		},
		{
			name:     "email-like string",
			input:    "user@domain.com",
			expected: "user_domain_com",
		},
		{
			name:     "string with spaces",
			input:    "agent node 1",
			expected: "agent_node_1",
		},
		{
			name:     "string with mixed special characters",
			input:    "agent#1!@#$%^&*()",
			expected: "agent_1__________",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "__________",
		},
		{
			name:     "path-like string",
			input:    "/path/to/resource",
			expected: "_path_to_resource",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := sanitizeKeyForNATS(tt.input)
			s.Equal(tt.expected, got)
		})
	}
}

func TestAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AgentTestSuite))
}
