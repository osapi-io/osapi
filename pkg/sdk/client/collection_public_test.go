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

package client_test

import (
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

type CollectionPublicTestSuite struct {
	suite.Suite
}

func (suite *CollectionPublicTestSuite) TestDerefString() {
	s := "hello"

	tests := []struct {
		name         string
		input        *string
		validateFunc func(string)
	}{
		{
			name:  "when pointer is non-nil",
			input: &s,
			validateFunc: func(result string) {
				suite.Equal("hello", result)
			},
		},
		{
			name:  "when pointer is nil",
			input: nil,
			validateFunc: func(result string) {
				suite.Empty(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(client.ExportDerefString(tc.input))
		})
	}
}

func (suite *CollectionPublicTestSuite) TestDerefInt() {
	i := 42

	tests := []struct {
		name         string
		input        *int
		validateFunc func(int)
	}{
		{
			name:  "when pointer is non-nil",
			input: &i,
			validateFunc: func(result int) {
				suite.Equal(42, result)
			},
		},
		{
			name:  "when pointer is nil",
			input: nil,
			validateFunc: func(result int) {
				suite.Zero(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(client.ExportDerefInt(tc.input))
		})
	}
}

func (suite *CollectionPublicTestSuite) TestDerefInt64() {
	i := int64(42)

	tests := []struct {
		name         string
		input        *int64
		validateFunc func(int64)
	}{
		{
			name:  "when pointer is non-nil",
			input: &i,
			validateFunc: func(result int64) {
				suite.Equal(int64(42), result)
			},
		},
		{
			name:  "when pointer is nil",
			input: nil,
			validateFunc: func(result int64) {
				suite.Zero(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(client.ExportDerefInt64(tc.input))
		})
	}
}

func (suite *CollectionPublicTestSuite) TestDerefFloat64() {
	f := 3.14

	tests := []struct {
		name         string
		input        *float64
		validateFunc func(float64)
	}{
		{
			name:  "when pointer is non-nil",
			input: &f,
			validateFunc: func(result float64) {
				suite.InDelta(3.14, result, 0.001)
			},
		},
		{
			name:  "when pointer is nil",
			input: nil,
			validateFunc: func(result float64) {
				suite.InDelta(0.0, result, 0.001)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(client.ExportDerefFloat64(tc.input))
		})
	}
}

func (suite *CollectionPublicTestSuite) TestDerefBool() {
	b := true

	tests := []struct {
		name         string
		input        *bool
		validateFunc func(bool)
	}{
		{
			name:  "when pointer is non-nil",
			input: &b,
			validateFunc: func(result bool) {
				suite.True(result)
			},
		},
		{
			name:  "when pointer is nil",
			input: nil,
			validateFunc: func(result bool) {
				suite.False(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(client.ExportDerefBool(tc.input))
		})
	}
}

func (suite *CollectionPublicTestSuite) TestJobIDFromGen() {
	id := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *openapi_types.UUID
		validateFunc func(string)
	}{
		{
			name:  "when pointer is non-nil",
			input: &id,
			validateFunc: func(result string) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", result)
			},
		},
		{
			name:  "when pointer is nil",
			input: nil,
			validateFunc: func(result string) {
				suite.Empty(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.validateFunc(client.ExportJobIDFromGen(tc.input))
		})
	}
}

func (suite *CollectionPublicTestSuite) TestCollectionFirst() {
	tests := []struct {
		name         string
		col          client.Collection[client.HostnameResult]
		validateFunc func(client.HostnameResult, bool)
	}{
		{
			name: "returns first result and true",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{
					{Hostname: "web-01"},
					{Hostname: "web-02"},
				},
				JobID: "job-1",
			},
			validateFunc: func(
				r client.HostnameResult,
				ok bool,
			) {
				suite.True(ok)
				suite.Equal("web-01", r.Hostname)
			},
		},
		{
			name: "returns zero value and false when empty",
			col: client.Collection[client.HostnameResult]{
				Results: []client.HostnameResult{},
			},
			validateFunc: func(
				r client.HostnameResult,
				ok bool,
			) {
				suite.False(ok)
				suite.Equal("", r.Hostname)
			},
		},
		{
			name: "returns zero value and false when nil",
			col:  client.Collection[client.HostnameResult]{},
			validateFunc: func(
				r client.HostnameResult,
				ok bool,
			) {
				suite.False(ok)
				suite.Equal("", r.Hostname)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			r, ok := tt.col.First()
			tt.validateFunc(r, ok)
		})
	}
}

func TestCollectionPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CollectionPublicTestSuite))
}
