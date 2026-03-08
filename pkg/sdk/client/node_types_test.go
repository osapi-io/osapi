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

package client

import (
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

type NodeTypesTestSuite struct {
	suite.Suite
}

func (suite *NodeTypesTestSuite) TestLoadAverageFromGen() {
	tests := []struct {
		name         string
		input        *gen.LoadAverageResponse
		validateFunc func(*LoadAverage)
	}{
		{
			name: "when populated",
			input: &gen.LoadAverageResponse{
				N1min:  0.5,
				N5min:  1.2,
				N15min: 0.8,
			},
			validateFunc: func(la *LoadAverage) {
				suite.Require().NotNil(la)
				suite.InDelta(0.5, float64(la.OneMin), 0.001)
				suite.InDelta(1.2, float64(la.FiveMin), 0.001)
				suite.InDelta(0.8, float64(la.FifteenMin), 0.001)
			},
		},
		{
			name:  "when nil",
			input: nil,
			validateFunc: func(la *LoadAverage) {
				suite.Nil(la)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := loadAverageFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestMemoryFromGen() {
	tests := []struct {
		name         string
		input        *gen.MemoryResponse
		validateFunc func(*Memory)
	}{
		{
			name: "when populated",
			input: &gen.MemoryResponse{
				Total: 8589934592,
				Used:  4294967296,
				Free:  4294967296,
			},
			validateFunc: func(m *Memory) {
				suite.Require().NotNil(m)
				suite.Equal(8589934592, m.Total)
				suite.Equal(4294967296, m.Used)
				suite.Equal(4294967296, m.Free)
			},
		},
		{
			name:  "when nil",
			input: nil,
			validateFunc: func(m *Memory) {
				suite.Nil(m)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := memoryFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestOSInfoFromGen() {
	tests := []struct {
		name         string
		input        *gen.OSInfoResponse
		validateFunc func(*OSInfo)
	}{
		{
			name: "when populated",
			input: &gen.OSInfoResponse{
				Distribution: "Ubuntu",
				Version:      "22.04",
			},
			validateFunc: func(oi *OSInfo) {
				suite.Require().NotNil(oi)
				suite.Equal("Ubuntu", oi.Distribution)
				suite.Equal("22.04", oi.Version)
			},
		},
		{
			name:  "when nil",
			input: nil,
			validateFunc: func(oi *OSInfo) {
				suite.Nil(oi)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := osInfoFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestDisksFromGen() {
	tests := []struct {
		name         string
		input        *gen.DisksResponse
		validateFunc func([]Disk)
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
			validateFunc: func(disks []Disk) {
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
			validateFunc: func(disks []Disk) {
				suite.Nil(disks)
			},
		},
		{
			name: "when empty",
			input: func() *gen.DisksResponse {
				d := gen.DisksResponse{}

				return &d
			}(),
			validateFunc: func(disks []Disk) {
				suite.NotNil(disks)
				suite.Empty(disks)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := disksFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestHostnameCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x00,
	}

	tests := []struct {
		name         string
		input        *gen.HostnameCollectionResponse
		validateFunc func(Collection[HostnameResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.HostnameCollectionResponse {
				labels := map[string]string{"group": "web", "env": "prod"}
				errMsg := "timeout"
				changed := false

				return &gen.HostnameCollectionResponse{
					JobId: &testUUID,
					Results: []gen.HostnameResponse{
						{
							Hostname: "web-01",
							Labels:   &labels,
							Changed:  &changed,
						},
						{
							Hostname: "web-02",
							Error:    &errMsg,
							Changed:  &changed,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[HostnameResult]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 2)

				suite.Equal("web-01", c.Results[0].Hostname)
				suite.Equal(map[string]string{"group": "web", "env": "prod"}, c.Results[0].Labels)
				suite.Empty(c.Results[0].Error)
				suite.False(c.Results[0].Changed)

				suite.Equal("web-02", c.Results[1].Hostname)
				suite.Equal("timeout", c.Results[1].Error)
				suite.Nil(c.Results[1].Labels)
				suite.False(c.Results[1].Changed)
			},
		},
		{
			name: "when minimal",
			input: &gen.HostnameCollectionResponse{
				Results: []gen.HostnameResponse{
					{Hostname: "minimal-host"},
				},
			},
			validateFunc: func(c Collection[HostnameResult]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("minimal-host", c.Results[0].Hostname)
				suite.Empty(c.Results[0].Error)
				suite.Nil(c.Results[0].Labels)
				suite.False(c.Results[0].Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := hostnameCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestNodeStatusCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x00,
	}

	tests := []struct {
		name         string
		input        *gen.NodeStatusCollectionResponse
		validateFunc func(Collection[NodeStatus])
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
			validateFunc: func(c Collection[NodeStatus]) {
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
			validateFunc: func(c Collection[NodeStatus]) {
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
			result := nodeStatusCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestDiskCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DiskCollectionResponse
		validateFunc func(Collection[DiskResult])
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
			validateFunc: func(c Collection[DiskResult]) {
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
			validateFunc: func(c Collection[DiskResult]) {
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
			result := diskCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestCommandCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x00,
	}

	tests := []struct {
		name         string
		input        *gen.CommandResultCollectionResponse
		validateFunc func(Collection[CommandResult])
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
			validateFunc: func(c Collection[CommandResult]) {
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
			validateFunc: func(c Collection[CommandResult]) {
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
			result := commandCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestDNSConfigCollectionFromGen() {
	testUUID := openapi_types.UUID{
		0x55,
		0x0e,
		0x84,
		0x00,
		0xe2,
		0x9b,
		0x41,
		0xd4,
		0xa7,
		0x16,
		0x44,
		0x66,
		0x55,
		0x44,
		0x00,
		0x00,
	}

	tests := []struct {
		name         string
		input        *gen.DNSConfigCollectionResponse
		validateFunc func(Collection[DNSConfig])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DNSConfigCollectionResponse {
				servers := []string{"8.8.8.8", "8.8.4.4"}
				searchDomains := []string{"example.com", "local"}
				changed := false

				return &gen.DNSConfigCollectionResponse{
					JobId: &testUUID,
					Results: []gen.DNSConfigResponse{
						{
							Hostname:      "web-01",
							Changed:       &changed,
							Servers:       &servers,
							SearchDomains: &searchDomains,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DNSConfig]) {
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", c.JobID)
				suite.Require().Len(c.Results, 1)

				dc := c.Results[0]
				suite.Equal("web-01", dc.Hostname)
				suite.Empty(dc.Error)
				suite.False(dc.Changed)
				suite.Equal([]string{"8.8.8.8", "8.8.4.4"}, dc.Servers)
				suite.Equal([]string{"example.com", "local"}, dc.SearchDomains)
			},
		},
		{
			name: "when minimal",
			input: &gen.DNSConfigCollectionResponse{
				Results: []gen.DNSConfigResponse{
					{Hostname: "web-01"},
				},
			},
			validateFunc: func(c Collection[DNSConfig]) {
				suite.Empty(c.JobID)
				suite.Require().Len(c.Results, 1)
				suite.Equal("web-01", c.Results[0].Hostname)
				suite.False(c.Results[0].Changed)
				suite.Nil(c.Results[0].Servers)
				suite.Nil(c.Results[0].SearchDomains)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dnsConfigCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestDNSUpdateCollectionFromGen() {
	tests := []struct {
		name         string
		input        *gen.DNSUpdateCollectionResponse
		validateFunc func(Collection[DNSUpdateResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.DNSUpdateCollectionResponse {
				changed := true

				return &gen.DNSUpdateCollectionResponse{
					Results: []gen.DNSUpdateResultItem{
						{
							Hostname: "web-01",
							Status:   gen.DNSUpdateResultItemStatus("applied"),
							Changed:  &changed,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[DNSUpdateResult]) {
				suite.Require().Len(c.Results, 1)

				dr := c.Results[0]
				suite.Equal("web-01", dr.Hostname)
				suite.Equal("applied", dr.Status)
				suite.True(dr.Changed)
				suite.Empty(dr.Error)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := dnsUpdateCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestPingCollectionFromGen() {
	tests := []struct {
		name         string
		input        *gen.PingCollectionResponse
		validateFunc func(Collection[PingResult])
	}{
		{
			name: "when all fields are populated",
			input: func() *gen.PingCollectionResponse {
				packetsSent := 5
				packetsReceived := 5
				packetLoss := 0.0
				minRtt := "1.234ms"
				avgRtt := "2.345ms"
				maxRtt := "3.456ms"
				changed := false

				return &gen.PingCollectionResponse{
					Results: []gen.PingResponse{
						{
							Hostname:        "web-01",
							Changed:         &changed,
							PacketsSent:     &packetsSent,
							PacketsReceived: &packetsReceived,
							PacketLoss:      &packetLoss,
							MinRtt:          &minRtt,
							AvgRtt:          &avgRtt,
							MaxRtt:          &maxRtt,
						},
					},
				}
			}(),
			validateFunc: func(c Collection[PingResult]) {
				suite.Require().Len(c.Results, 1)

				pr := c.Results[0]
				suite.Equal("web-01", pr.Hostname)
				suite.Equal(5, pr.PacketsSent)
				suite.Equal(5, pr.PacketsReceived)
				suite.InDelta(0.0, pr.PacketLoss, 0.001)
				suite.Equal("1.234ms", pr.MinRtt)
				suite.Equal("2.345ms", pr.AvgRtt)
				suite.Equal("3.456ms", pr.MaxRtt)
				suite.Empty(pr.Error)
				suite.False(pr.Changed)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := pingCollectionFromGen(tc.input)
			tc.validateFunc(result)
		})
	}
}

func (suite *NodeTypesTestSuite) TestDerefString() {
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
			tc.validateFunc(derefString(tc.input))
		})
	}
}

func (suite *NodeTypesTestSuite) TestDerefInt() {
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
			tc.validateFunc(derefInt(tc.input))
		})
	}
}

func (suite *NodeTypesTestSuite) TestDerefInt64() {
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
			tc.validateFunc(derefInt64(tc.input))
		})
	}
}

func (suite *NodeTypesTestSuite) TestDerefFloat64() {
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
			tc.validateFunc(derefFloat64(tc.input))
		})
	}
}

func (suite *NodeTypesTestSuite) TestDerefBool() {
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
			tc.validateFunc(derefBool(tc.input))
		})
	}
}

func (suite *NodeTypesTestSuite) TestJobIDFromGen() {
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
			tc.validateFunc(jobIDFromGen(tc.input))
		})
	}
}

func TestNodeTypesTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTypesTestSuite))
}
