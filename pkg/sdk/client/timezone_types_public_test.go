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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// strPtr returns a pointer to a string value.
func strPtr(
	s string,
) *string {
	return &s
}

type TimezoneTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *TimezoneTypesPublicTestSuite) TestTimezoneCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.TimezoneCollectionResponse
		validateFunc func(client.Collection[client.TimezoneResult])
	}{
		{
			name: "converts full response",
			input: &gen.TimezoneCollectionResponse{
				JobId: &testUUID,
				Results: []gen.TimezoneEntry{
					{
						Hostname:  "agent1",
						Status:    gen.TimezoneEntryStatusOk,
						Timezone:  strPtr("America/New_York"),
						UtcOffset: strPtr("-05:00"),
					},
				},
			},
			validateFunc: func(c client.Collection[client.TimezoneResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("agent1", c.Results[0].Hostname)
				suite.Equal("ok", c.Results[0].Status)
				suite.Equal("America/New_York", c.Results[0].Timezone)
				suite.Equal("-05:00", c.Results[0].UTCOffset)
			},
		},
		{
			name: "converts response with nil optional fields",
			input: &gen.TimezoneCollectionResponse{
				Results: []gen.TimezoneEntry{
					{
						Hostname: "agent1",
						Status:   gen.TimezoneEntryStatusSkipped,
						Error:    strPtr("unsupported"),
					},
				},
			},
			validateFunc: func(c client.Collection[client.TimezoneResult]) {
				suite.Equal("", c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("agent1", c.Results[0].Hostname)
				suite.Equal("skipped", c.Results[0].Status)
				suite.Equal("", c.Results[0].Timezone)
				suite.Equal("unsupported", c.Results[0].Error)
			},
		},
		{
			name: "converts empty results",
			input: &gen.TimezoneCollectionResponse{
				Results: []gen.TimezoneEntry{},
			},
			validateFunc: func(c client.Collection[client.TimezoneResult]) {
				suite.Empty(c.Results)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Use the exported constructor to test via SDK client
			result := client.ExportTimezoneCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *TimezoneTypesPublicTestSuite) TestTimezoneMutationCollectionFromUpdate() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66,
		0x55, 0x44, 0x00, 0x00,
	}
	changedTrue := true

	tests := []struct {
		name         string
		input        *gen.TimezoneUpdateResponse
		validateFunc func(client.Collection[client.TimezoneMutationResult])
	}{
		{
			name: "converts full response",
			input: &gen.TimezoneUpdateResponse{
				JobId: &testUUID,
				Results: []gen.TimezoneMutationResult{
					{
						Hostname: "agent1",
						Status:   gen.TimezoneMutationResultStatusOk,
						Timezone: strPtr("America/New_York"),
						Changed:  &changedTrue,
					},
				},
			},
			validateFunc: func(c client.Collection[client.TimezoneMutationResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("agent1", c.Results[0].Hostname)
				suite.Equal("ok", c.Results[0].Status)
				suite.Equal("America/New_York", c.Results[0].Timezone)
				suite.True(c.Results[0].Changed)
			},
		},
		{
			name: "converts response with nil optional fields",
			input: &gen.TimezoneUpdateResponse{
				Results: []gen.TimezoneMutationResult{
					{
						Hostname: "agent1",
						Status:   gen.TimezoneMutationResultStatusSkipped,
						Error:    strPtr("unsupported"),
					},
				},
			},
			validateFunc: func(c client.Collection[client.TimezoneMutationResult]) {
				suite.Require().Len(c.Results, 1)
				suite.Equal("skipped", c.Results[0].Status)
				suite.False(c.Results[0].Changed)
				suite.Equal("unsupported", c.Results[0].Error)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportTimezoneMutationCollectionFromUpdate(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestTimezoneTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TimezoneTypesPublicTestSuite))
}
