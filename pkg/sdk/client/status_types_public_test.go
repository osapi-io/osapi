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

type StatusTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *StatusTypesPublicTestSuite) TestNodeStatusCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.NodeStatusCollectionResponse
		validateFunc func(client.Collection[client.NodeStatus])
	}{
		{
			name: "when all sub-types are populated",
			input: func() *gen.NodeStatusCollectionResponse {
				uptime := "5d 3h 22m"
				changed := false
				disks := gen.DisksResponse{
					{
						Name:  "/dev/sda1",
						Total: 500000000000,
						Used:  250000000000,
						Free:  250000000000,
					},
				}

				return &gen.NodeStatusCollectionResponse{
					JobId: &testUUID,
					Results: []gen.NodeStatusResponse{
						{
							Hostname: "web-01",
							Uptime:   &uptime,
							Changed:  &changed,
							Disks:    &disks,
							LoadAverage: &gen.LoadAverageResponse{
								N1min:  0.5,
								N5min:  1.2,
								N15min: 0.8,
							},
							Memory: &gen.MemoryResponse{
								Total: 8589934592,
								Used:  4294967296,
								Free:  4294967296,
							},
							OsInfo: &gen.OSInfoResponse{
								Distribution: "Ubuntu",
								Version:      "22.04",
							},
						},
					},
				}
			}(),
			validateFunc: func(c client.Collection[client.NodeStatus]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				ns := c.Results[0]
				suite.Equal("web-01", ns.Hostname)
				suite.Equal("5d 3h 22m", ns.Uptime)
				suite.Empty(ns.Error)
				suite.False(ns.Changed)

				suite.Require().Len(ns.Disks, 1)
				suite.Equal("/dev/sda1", ns.Disks[0].Name)
				suite.Equal(500000000000, ns.Disks[0].Total)

				suite.Require().NotNil(ns.LoadAverage)
				suite.InDelta(0.5, float64(ns.LoadAverage.OneMin), 0.001)
				suite.InDelta(1.2, float64(ns.LoadAverage.FiveMin), 0.001)
				suite.InDelta(0.8, float64(ns.LoadAverage.FifteenMin), 0.001)

				suite.Require().NotNil(ns.Memory)
				suite.Equal(8589934592, ns.Memory.Total)
				suite.Equal(4294967296, ns.Memory.Used)
				suite.Equal(4294967296, ns.Memory.Free)

				suite.Require().NotNil(ns.OSInfo)
				suite.Equal("Ubuntu", ns.OSInfo.Distribution)
				suite.Equal("22.04", ns.OSInfo.Version)
			},
		},
		{
			name: "when minimal",
			input: &gen.NodeStatusCollectionResponse{
				Results: []gen.NodeStatusResponse{
					{Hostname: "minimal-host"},
				},
			},
			validateFunc: func(c client.Collection[client.NodeStatus]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)

				ns := c.Results[0]
				suite.Equal("minimal-host", ns.Hostname)
				suite.Empty(ns.Uptime)
				suite.Empty(ns.Error)
				suite.False(ns.Changed)
				suite.Nil(ns.Disks)
				suite.Nil(ns.LoadAverage)
				suite.Nil(ns.Memory)
				suite.Nil(ns.OSInfo)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportNodeStatusCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestStatusTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTypesPublicTestSuite))
}
