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

type DiskTypesPublicTestSuite struct {
	suite.Suite
}

func (suite *DiskTypesPublicTestSuite) TestDisksFromGen() {
	tests := []struct {
		name         string
		input        *gen.DisksResponse
		validateFunc func([]client.Disk)
	}{
		{
			name: "when populated",
			input: func() *gen.DisksResponse {
				d := gen.DisksResponse{
					{
						Name:  "/dev/sda1",
						Total: 500000000000,
						Used:  250000000000,
						Free:  250000000000,
					},
					{
						Name:  "/dev/sdb1",
						Total: 1000000000000,
						Used:  100000000000,
						Free:  900000000000,
					},
				}

				return &d
			}(),
			validateFunc: func(disks []client.Disk) {
				suite.Require().Len(disks, 2)
				suite.Equal("/dev/sda1", disks[0].Name)
				suite.Equal(500000000000, disks[0].Total)
				suite.Equal(250000000000, disks[0].Used)
				suite.Equal(250000000000, disks[0].Free)
				suite.Equal("/dev/sdb1", disks[1].Name)
			},
		},
		{
			name:  "when nil",
			input: nil,
			validateFunc: func(disks []client.Disk) {
				suite.Nil(disks)
			},
		},
		{
			name: "when empty",
			input: func() *gen.DisksResponse {
				d := gen.DisksResponse{}

				return &d
			}(),
			validateFunc: func(disks []client.Disk) {
				suite.NotNil(disks)
				suite.Empty(disks)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportDisksFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *DiskTypesPublicTestSuite) TestDiskCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55, 0x0e, 0x84, 0x00, 0xe2, 0x9b, 0x41, 0xd4,
		0xa7, 0x16, 0x44, 0x66, 0x55, 0x44, 0x00, 0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DiskCollectionResponse
		validateFunc func(client.Collection[client.DiskResult])
	}{
		{
			name: "when disks are populated",
			input: func() *gen.DiskCollectionResponse {
				changed := false
				disks := gen.DisksResponse{
					{
						Name:  "/dev/sda1",
						Total: 500000000000,
						Used:  250000000000,
						Free:  250000000000,
					},
					{
						Name:  "/dev/sdb1",
						Total: 1000000000000,
						Used:  100000000000,
						Free:  900000000000,
					},
				}

				return &gen.DiskCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DiskResultItem{
						{
							Hostname: "web-01",
							Changed:  &changed,
							Disks:    &disks,
						},
					},
				}
			}(),
			validateFunc: func(c client.Collection[client.DiskResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				dr := c.Results[0]
				suite.Equal("web-01", dr.Hostname)
				suite.Empty(dr.Error)
				suite.False(dr.Changed)
				suite.Require().Len(dr.Disks, 2)
				suite.Equal("/dev/sda1", dr.Disks[0].Name)
				suite.Equal(500000000000, dr.Disks[0].Total)
				suite.Equal(250000000000, dr.Disks[0].Used)
				suite.Equal(250000000000, dr.Disks[0].Free)
				suite.Equal("/dev/sdb1", dr.Disks[1].Name)
			},
		},
		{
			name: "when empty",
			input: &gen.DiskCollectionResponse{
				Results: []gen.DiskResultItem{
					{Hostname: "web-01"},
				},
			},
			validateFunc: func(c client.Collection[client.DiskResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("web-01", c.Results[0].Hostname)
				suite.False(c.Results[0].Changed)
				suite.Nil(c.Results[0].Disks)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := client.ExportDiskCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func TestDiskTypesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DiskTypesPublicTestSuite))
}
