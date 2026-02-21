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

package telemetry

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MapCarrierTestSuite struct {
	suite.Suite
}

func (s *MapCarrierTestSuite) TestGet() {
	tests := []struct {
		name         string
		data         map[string]interface{}
		key          string
		validateFunc func(string)
	}{
		{
			name: "when key exists returns value",
			data: map[string]interface{}{
				"traceparent": "00-abc123-def456-01",
			},
			key: "traceparent",
			validateFunc: func(result string) {
				s.Equal("00-abc123-def456-01", result)
			},
		},
		{
			name: "when key does not exist returns empty string",
			data: map[string]interface{}{},
			key:  "traceparent",
			validateFunc: func(result string) {
				s.Empty(result)
			},
		},
		{
			name: "when value is not a string returns empty string",
			data: map[string]interface{}{
				"count": 42,
			},
			key: "count",
			validateFunc: func(result string) {
				s.Empty(result)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := mapCarrier{data: tc.data}

			tc.validateFunc(c.Get(tc.key))
		})
	}
}

func (s *MapCarrierTestSuite) TestSet() {
	tests := []struct {
		name         string
		key          string
		value        string
		validateFunc func(map[string]interface{})
	}{
		{
			name:  "when setting a value stores it in data",
			key:   "traceparent",
			value: "00-abc123-def456-01",
			validateFunc: func(data map[string]interface{}) {
				s.Equal("00-abc123-def456-01", data["traceparent"])
			},
		},
		{
			name:  "when setting overwrites existing value",
			key:   "traceparent",
			value: "00-new-value-01",
			validateFunc: func(data map[string]interface{}) {
				s.Equal("00-new-value-01", data["traceparent"])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			data := map[string]interface{}{
				"traceparent": "00-old-value-01",
			}
			c := mapCarrier{data: data}
			c.Set(tc.key, tc.value)

			tc.validateFunc(data)
		})
	}
}

func (s *MapCarrierTestSuite) TestKeys() {
	tests := []struct {
		name         string
		data         map[string]interface{}
		validateFunc func([]string)
	}{
		{
			name: "when data has entries returns all keys",
			data: map[string]interface{}{
				"traceparent":   "00-abc123-def456-01",
				"tracestate":    "vendor=opaque",
				"custom-header": "value",
			},
			validateFunc: func(keys []string) {
				s.Len(keys, 3)
				s.ElementsMatch(
					[]string{"traceparent", "tracestate", "custom-header"},
					keys,
				)
			},
		},
		{
			name: "when data is empty returns empty slice",
			data: map[string]interface{}{},
			validateFunc: func(keys []string) {
				s.Empty(keys)
			},
		},
		{
			name: "when data has single entry returns one key",
			data: map[string]interface{}{
				"traceparent": "00-abc123-def456-01",
			},
			validateFunc: func(keys []string) {
				s.Len(keys, 1)
				s.Equal("traceparent", keys[0])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := mapCarrier{data: tc.data}

			tc.validateFunc(c.Keys())
		})
	}
}

func TestMapCarrierTestSuite(t *testing.T) {
	suite.Run(t, new(MapCarrierTestSuite))
}
