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

type CommandTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *CommandTypesPublicTestSuite) TestCommandCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.CommandResultCollectionResponse
		validateFunc func(client.Collection[client.CommandResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.CommandResultCollectionResponse {
				stdout := "hello world\n"
				stderr := "warning: something\n"
				exitCode := 0
				changed := true
				durationMs := int64(150)

				return &gen.CommandResultCollectionResponse{
					JobId: &testUUID,
					Results: []gen.CommandResultItem{
						{
							Hostname:   "web-01",
							Stdout:     &stdout,
							Stderr:     &stderr,
							ExitCode:   &exitCode,
							Changed:    &changed,
							DurationMs: &durationMs,
						},
					},
				}
			}(),
			validateFunc: func(c client.Collection[client.CommandResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				cr := c.Results[0]
				suite.Equal("web-01", cr.Hostname)
				suite.Equal("hello world\n", cr.Stdout)
				suite.Equal("warning: something\n", cr.Stderr)
				suite.Empty(cr.Error)
				suite.Equal(0, cr.ExitCode)
				suite.True(cr.Changed)
				suite.Equal(int64(150), cr.DurationMs)
			},
		},
		{
			name: "when minimal with error",
			input: func() *gen.CommandResultCollectionResponse {
				errMsg := "command not found"
				exitCode := 127

				return &gen.CommandResultCollectionResponse{
					Results: []gen.CommandResultItem{
						{
							Hostname: "web-01",
							Error:    &errMsg,
							ExitCode: &exitCode,
						},
					},
				}
			}(),
			validateFunc: func(c client.Collection[client.CommandResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				cr := c.Results[0]
				suite.Equal("web-01", cr.Hostname)
				suite.Equal("command not found", cr.Error)
				suite.Equal(127, cr.ExitCode)
				suite.Empty(cr.Stdout)
				suite.Empty(cr.Stderr)
				suite.False(cr.Changed)
				suite.Zero(cr.DurationMs)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportCommandCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestCommandTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CommandTypesPublicTestSuite))
}
