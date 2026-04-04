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

package route_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	execmocks "github.com/retr0h/osapi/internal/exec/mocks"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	netinfomocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	"github.com/retr0h/osapi/internal/provider/network/netplan/route"
)

const testHostname = "test-host"

type RoutePublicTestSuite struct {
	suite.Suite

	ctrl        *gomock.Controller
	ctx         context.Context
	logger      *slog.Logger
	memFs       avfs.VFS
	mockStateKV *jobmocks.MockKeyValue
	mockExec    *execmocks.MockManager
	mockNetinfo *netinfomocks.MockProvider
	provider    *route.Debian
}

func (suite *RoutePublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.ctx = context.Background()
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.memFs = memfs.New()
	suite.mockStateKV = jobmocks.NewMockKeyValue(suite.ctrl)
	suite.mockExec = execmocks.NewMockManager(suite.ctrl)
	suite.mockNetinfo = netinfomocks.NewMockProvider(suite.ctrl)

	_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)

	suite.provider = route.NewDebianProvider(
		suite.logger,
		suite.memFs,
		suite.mockStateKV,
		suite.mockExec,
		testHostname,
		suite.mockNetinfo,
	)
}

func (suite *RoutePublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *RoutePublicTestSuite) TearDownSubTest() {
	route.ResetMarshalJSON()
}

func (suite *RoutePublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func([]route.ListEntry, error)
	}{
		{
			name: "when routes exist",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetRoutes().
					Return([]netinfo.RouteResult{
						{
							Destination: "10.1.0.0",
							Gateway:     "10.0.0.1",
							Interface:   "eth0",
							Mask:        "255.255.0.0",
							Metric:      100,
							Flags:       "UG",
						},
						{
							Destination: "192.168.1.0",
							Gateway:     "192.168.0.1",
							Interface:   "eth1",
						},
					}, nil)
			},
			validateFunc: func(result []route.ListEntry, err error) {
				suite.Require().NoError(err)
				suite.Require().Len(result, 2)

				suite.Equal("10.1.0.0", result[0].Destination)
				suite.Equal("10.0.0.1", result[0].Gateway)
				suite.Equal("eth0", result[0].Interface)
				suite.Equal("255.255.0.0", result[0].Mask)
				suite.Equal(100, result[0].Metric)
				suite.Equal("UG", result[0].Flags)

				suite.Equal("192.168.1.0", result[1].Destination)
				suite.Equal("192.168.0.1", result[1].Gateway)
				suite.Equal("eth1", result[1].Interface)
			},
		},
		{
			name: "when no routes exist",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetRoutes().
					Return([]netinfo.RouteResult{}, nil)
			},
			validateFunc: func(result []route.ListEntry, err error) {
				suite.Require().NoError(err)
				suite.Empty(result)
			},
		},
		{
			name: "when GetRoutes fails",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetRoutes().
					Return(nil, errors.New("netinfo error"))
			},
			validateFunc: func(result []route.ListEntry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan route list:")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.List(suite.ctx)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *RoutePublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		interfaceNm  string
		setup        func()
		validateFunc func(*route.Entry, error)
	}{
		{
			name:        "when routes found in state KV",
			interfaceNm: "eth0",
			setup: func() {
				routes := []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1", Metric: 100},
				}
				routesJSON, _ := json.Marshal(routes)
				state := map[string]interface{}{
					"path":        "/etc/netplan/osapi-eth0-routes.yaml",
					"sha256":      "abc123",
					"deployed_at": "2026-01-01T00:00:00Z",
					"metadata": map[string]string{
						"interface": "eth0",
						"routes":    string(routesJSON),
					},
				}
				stateBytes, _ := json.Marshal(state)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Interface)
				suite.Require().Len(result.Routes, 1)
				suite.Equal("10.1.0.0/16", result.Routes[0].To)
				suite.Equal("10.0.0.1", result.Routes[0].Via)
				suite.Equal(100, result.Routes[0].Metric)
			},
		},
		{
			name:        "when state KV entry not found",
			interfaceNm: "eth0",
			setup: func() {
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "not found")
			},
		},
		{
			name:        "when state entry is undeployed",
			interfaceNm: "eth0",
			setup: func() {
				state := map[string]interface{}{
					"path":          "/etc/netplan/osapi-eth0-routes.yaml",
					"sha256":        "abc123",
					"deployed_at":   "2026-01-01T00:00:00Z",
					"undeployed_at": "2026-01-02T00:00:00Z",
					"metadata": map[string]string{
						"interface": "eth0",
						"routes":    `[{"to":"10.1.0.0/16","via":"10.0.0.1"}]`,
					},
				}
				stateBytes, _ := json.Marshal(state)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "not found")
			},
		},
		{
			name:        "when metadata has no routes key",
			interfaceNm: "eth0",
			setup: func() {
				state := map[string]interface{}{
					"path":        "/etc/netplan/osapi-eth0-routes.yaml",
					"sha256":      "abc123",
					"deployed_at": "2026-01-01T00:00:00Z",
					"metadata": map[string]string{
						"interface": "eth0",
					},
				}
				stateBytes, _ := json.Marshal(state)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "no route metadata")
			},
		},
		{
			name:        "when routes JSON is invalid",
			interfaceNm: "eth0",
			setup: func() {
				state := map[string]interface{}{
					"path":        "/etc/netplan/osapi-eth0-routes.yaml",
					"sha256":      "abc123",
					"deployed_at": "2026-01-01T00:00:00Z",
					"metadata": map[string]string{
						"interface": "eth0",
						"routes":    "not-valid-json",
					},
				}
				stateBytes, _ := json.Marshal(state)

				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return(stateBytes)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "unmarshal routes")
			},
		},
		{
			name:        "when state JSON is invalid",
			interfaceNm: "eth0",
			setup: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(suite.ctrl)
				mockEntry.EXPECT().Value().Return([]byte("not-json"))

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(mockEntry, nil)
			},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "unmarshal state")
			},
		},
		{
			name:        "when interface name is empty",
			interfaceNm: "",
			setup:       func() {},
			validateFunc: func(result *route.Entry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "interface name must not be empty")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Get(suite.ctx, tc.interfaceNm)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *RoutePublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		entry        route.Entry
		setup        func()
		validateFunc func(*route.Result, error)
	}{
		{
			name: "when new routes deploy successfully",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1", Metric: 100},
					{To: "172.16.0.0/12", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				// KV Get returns not found (new file).
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))

				// netplan generate succeeds.
				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"generate"}).
					Return("", nil)

				// netplan apply succeeds.
				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"apply"}).
					Return("", nil)

				// KV Put succeeds.
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Interface)
				suite.True(result.Changed)

				// Verify file was written.
				data, readErr := suite.memFs.ReadFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
				)
				suite.Require().NoError(readErr)
				suite.Contains(string(data), "eth0:")
				suite.Contains(string(data), "to: 10.1.0.0/16")
				suite.Contains(string(data), "via: 10.0.0.1")
				suite.Contains(string(data), "metric: 100")
				suite.Contains(string(data), "to: 172.16.0.0/12")
			},
		},
		{
			name: "when route file already exists",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				// File already exists on disk.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
					[]byte("existing"),
					0o644,
				)
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "already managed")
			},
		},
		{
			name: "when routes contain default route IPv4",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "0.0.0.0/0", Via: "10.0.0.1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "default route")
			},
		},
		{
			name: "when routes contain default route IPv6",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "::/0", Via: "fe80::1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "default route")
			},
		},
		{
			name: "when routes contain default keyword",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "default", Via: "10.0.0.1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "default route")
			},
		},
		{
			name: "when interface name is empty",
			entry: route.Entry{
				Interface: "",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "name must not be empty")
			},
		},
		{
			name: "when interface name has invalid characters",
			entry: route.Entry{
				Interface: "eth 0!",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid characters")
			},
		},
		{
			name: "when ApplyConfig fails",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				// KV Get returns not found.
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))

				// netplan generate fails.
				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"generate"}).
					Return("", errors.New("invalid YAML"))
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan route create:")
			},
		},
		{
			name: "when buildRouteMetadata fails",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				route.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan route create:")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Create(suite.ctx, tc.entry)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *RoutePublicTestSuite) TestUpdate() {
	tests := []struct {
		name         string
		entry        route.Entry
		setup        func()
		validateFunc func(*route.Result, error)
	}{
		{
			name: "when update succeeds",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.2.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				// Managed file exists on disk.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
					[]byte("old content"),
					0o644,
				)

				// KV Get returns not found (different SHA).
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))

				// netplan generate succeeds.
				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"generate"}).
					Return("", nil)

				// netplan apply succeeds.
				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"apply"}).
					Return("", nil)

				// KV Put succeeds.
				suite.mockStateKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Interface)
				suite.True(result.Changed)
			},
		},
		{
			name: "when route file not managed",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				// No file on disk.
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "not managed")
			},
		},
		{
			name: "when routes contain default route",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "0.0.0.0/0", Via: "10.0.0.1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "default route")
			},
		},
		{
			name: "when interface name is empty",
			entry: route.Entry{
				Interface: "",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "name must not be empty")
			},
		},
		{
			name: "when ApplyConfig fails",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
					[]byte("old"),
					0o644,
				)

				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))

				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"generate"}).
					Return("", errors.New("validation error"))
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan route update:")
			},
		},
		{
			name: "when buildRouteMetadata fails",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
					[]byte("old"),
					0o644,
				)

				route.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan route update:")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Update(suite.ctx, tc.entry)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *RoutePublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		interfaceNm  string
		setup        func()
		validateFunc func(*route.Result, error)
	}{
		{
			name:        "when file exists and removal succeeds",
			interfaceNm: "eth0",
			setup: func() {
				// Create the managed file on disk.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
					[]byte("network:\n"),
					0o644,
				)

				// netplan apply succeeds.
				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"apply"}).
					Return("", nil)

				// KV state exists for undeploy marking.
				suite.mockStateKV.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Interface)
				suite.True(result.Changed)

				// Verify file was removed.
				_, statErr := suite.memFs.Stat(
					"/etc/netplan/osapi-eth0-routes.yaml",
				)
				suite.Error(statErr)
			},
		},
		{
			name:        "when file does not exist",
			interfaceNm: "eth0",
			setup:       func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Interface)
				suite.False(result.Changed)
			},
		},
		{
			name:        "when interface name is empty",
			interfaceNm: "",
			setup:       func() {},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "interface name must not be empty")
			},
		},
		{
			name:        "when RemoveConfig fails",
			interfaceNm: "eth0",
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0-routes.yaml",
					[]byte("network:\n"),
					0o644,
				)

				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"apply"}).
					Return("", errors.New("apply failed"))
			},
			validateFunc: func(result *route.Result, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan route delete:")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := suite.provider.Delete(suite.ctx, tc.interfaceNm)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *RoutePublicTestSuite) TestGenerateRouteYAML() {
	tests := []struct {
		name         string
		entry        route.Entry
		validateFunc func(string)
	}{
		{
			name: "when single route with metric",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1", Metric: 100},
				},
			},
			validateFunc: func(result string) {
				suite.Contains(result, "network:")
				suite.Contains(result, "  version: 2")
				suite.Contains(result, "  ethernets:")
				suite.Contains(result, "    eth0:")
				suite.Contains(result, "      routes:")
				suite.Contains(result, "        - to: 10.1.0.0/16")
				suite.Contains(result, "          via: 10.0.0.1")
				suite.Contains(result, "          metric: 100")
			},
		},
		{
			name: "when multiple routes without metric",
			entry: route.Entry{
				Interface: "ens3",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
					{To: "172.16.0.0/12", Via: "10.0.0.1"},
				},
			},
			validateFunc: func(result string) {
				suite.Contains(result, "    ens3:")
				suite.Contains(result, "        - to: 10.1.0.0/16")
				suite.Contains(result, "          via: 10.0.0.1")
				suite.Contains(result, "        - to: 172.16.0.0/12")
				suite.NotContains(result, "metric:")
			},
		},
		{
			name: "when route with zero metric omits metric",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1", Metric: 0},
				},
			},
			validateFunc: func(result string) {
				suite.Contains(result, "        - to: 10.1.0.0/16")
				suite.Contains(result, "          via: 10.0.0.1")
				suite.NotContains(result, "metric:")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := route.GenerateRouteYAML(tc.entry)

			tc.validateFunc(string(result))
		})
	}
}

func (suite *RoutePublicTestSuite) TestContainsDefaultRoute() {
	tests := []struct {
		name         string
		routes       []route.Route
		validateFunc func(bool)
	}{
		{
			name: "when contains IPv4 default route",
			routes: []route.Route{
				{To: "10.1.0.0/16", Via: "10.0.0.1"},
				{To: "0.0.0.0/0", Via: "10.0.0.1"},
			},
			validateFunc: func(result bool) {
				suite.True(result)
			},
		},
		{
			name: "when contains IPv6 default route",
			routes: []route.Route{
				{To: "::/0", Via: "fe80::1"},
			},
			validateFunc: func(result bool) {
				suite.True(result)
			},
		},
		{
			name: "when contains default keyword",
			routes: []route.Route{
				{To: "default", Via: "10.0.0.1"},
			},
			validateFunc: func(result bool) {
				suite.True(result)
			},
		},
		{
			name: "when no default routes",
			routes: []route.Route{
				{To: "10.1.0.0/16", Via: "10.0.0.1"},
				{To: "172.16.0.0/12", Via: "10.0.0.1"},
			},
			validateFunc: func(result bool) {
				suite.False(result)
			},
		},
		{
			name:   "when empty routes",
			routes: []route.Route{},
			validateFunc: func(result bool) {
				suite.False(result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := route.ContainsDefaultRoute(tc.routes)

			tc.validateFunc(result)
		})
	}
}

func (suite *RoutePublicTestSuite) TestBuildRouteMetadata() {
	tests := []struct {
		name         string
		entry        route.Entry
		setup        func()
		validateFunc func(map[string]string, error)
	}{
		{
			name: "when routes serialize successfully",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1", Metric: 100},
				},
			},
			setup: func() {},
			validateFunc: func(result map[string]string, err error) {
				suite.Require().NoError(err)
				suite.Equal("eth0", result["interface"])

				var routes []route.Route
				unmarshalErr := json.Unmarshal([]byte(result["routes"]), &routes)
				suite.Require().NoError(unmarshalErr)
				suite.Require().Len(routes, 1)
				suite.Equal("10.1.0.0/16", routes[0].To)
				suite.Equal("10.0.0.1", routes[0].Via)
				suite.Equal(100, routes[0].Metric)
			},
		},
		{
			name: "when marshalJSON fails",
			entry: route.Entry{
				Interface: "eth0",
				Routes: []route.Route{
					{To: "10.1.0.0/16", Via: "10.0.0.1"},
				},
			},
			setup: func() {
				route.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
			},
			validateFunc: func(result map[string]string, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "marshal routes")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result, err := route.BuildRouteMetadata(tc.entry)

			tc.validateFunc(result, err)
		})
	}
}

func (suite *RoutePublicTestSuite) TestRouteFilePath() {
	tests := []struct {
		name         string
		interfaceNm  string
		validateFunc func(string)
	}{
		{
			name:        "when standard interface name",
			interfaceNm: "eth0",
			validateFunc: func(result string) {
				suite.Equal("/etc/netplan/osapi-eth0-routes.yaml", result)
			},
		},
		{
			name:        "when interface with dashes",
			interfaceNm: "ens3-bridge",
			validateFunc: func(result string) {
				suite.Equal("/etc/netplan/osapi-ens3-bridge-routes.yaml", result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := route.RouteFilePath(tc.interfaceNm)

			tc.validateFunc(result)
		})
	}
}

func TestRoutePublicTestSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(RoutePublicTestSuite))
}
