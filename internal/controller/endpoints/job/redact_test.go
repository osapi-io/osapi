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

package job

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RedactTestSuite struct {
	suite.Suite
}

func (s *RedactTestSuite) TestRedactOperation() {
	tests := []struct {
		name         string
		operation    map[string]interface{}
		validateFunc func(map[string]interface{})
	}{
		{
			name:      "nil operation stays nil",
			operation: nil,
			validateFunc: func(got map[string]interface{}) {
				s.Nil(got)
			},
		},
		{
			name: "operation with no sensitive fields is unchanged",
			operation: map[string]interface{}{
				"type": "network.dns.get",
			},
			validateFunc: func(got map[string]interface{}) {
				s.Equal("network.dns.get", got["type"])
			},
		},
		{
			name: "masks password_hash nested under data",
			operation: map[string]interface{}{
				"type": "user.password",
				"data": map[string]interface{}{
					"name":          "john",
					"password_hash": "$6$abcd$deadbeef",
				},
			},
			validateFunc: func(got map[string]interface{}) {
				data, ok := got["data"].(map[string]interface{})
				s.Require().True(ok)
				s.Equal("john", data["name"])
				s.Equal(redactedPlaceholder, data["password_hash"])
			},
		},
		{
			name: "masks legacy plaintext password field",
			operation: map[string]interface{}{
				"type": "user.password",
				"data": map[string]interface{}{
					"name":     "john",
					"password": "hunter2",
				},
			},
			validateFunc: func(got map[string]interface{}) {
				data, ok := got["data"].(map[string]interface{})
				s.Require().True(ok)
				s.Equal(redactedPlaceholder, data["password"])
				s.NotEqual("hunter2", data["password"])
			},
		},
		{
			name: "masks password_hash inside a slice",
			operation: map[string]interface{}{
				"type": "user.create",
				"data": []interface{}{
					map[string]interface{}{
						"name":          "jane",
						"password_hash": "$6$zzzz$cafebabe",
					},
				},
			},
			validateFunc: func(got map[string]interface{}) {
				items, ok := got["data"].([]interface{})
				s.Require().True(ok)
				s.Require().Len(items, 1)
				entry, ok := items[0].(map[string]interface{})
				s.Require().True(ok)
				s.Equal(redactedPlaceholder, entry["password_hash"])
			},
		},
		{
			name: "does not mutate the input map",
			operation: map[string]interface{}{
				"data": map[string]interface{}{
					"password_hash": "$6$abcd$deadbeef",
				},
			},
			validateFunc: func(got map[string]interface{}) {
				// The returned map is redacted, but the caller's original
				// nested map must be left alone.
				_ = got
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			original := tc.operation

			got := redactOperation(tc.operation)
			tc.validateFunc(got)

			if original != nil {
				if data, ok := original["data"].(map[string]interface{}); ok {
					if hash, ok := data["password_hash"].(string); ok {
						s.NotEqual(redactedPlaceholder, hash, "original map must not be mutated")
					}
				}
			}
		})
	}
}

func TestRedactTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RedactTestSuite))
}
