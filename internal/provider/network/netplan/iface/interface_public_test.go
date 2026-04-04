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

package iface_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	execmocks "github.com/retr0h/osapi/internal/exec/mocks"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/netinfo"
	netinfomocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	"github.com/retr0h/osapi/internal/provider/network/netplan/iface"
)

const testHostname = "test-host"

type InterfacePublicTestSuite struct {
	suite.Suite

	ctrl        *gomock.Controller
	ctx         context.Context
	logger      *slog.Logger
	memFs       avfs.VFS
	mockStateKV *jobmocks.MockKeyValue
	mockExec    *execmocks.MockManager
	mockNetinfo *netinfomocks.MockProvider
	provider    *iface.Debian
}

func (suite *InterfacePublicTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.ctx = context.Background()
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	suite.memFs = memfs.New()
	suite.mockStateKV = jobmocks.NewMockKeyValue(suite.ctrl)
	suite.mockExec = execmocks.NewMockManager(suite.ctrl)
	suite.mockNetinfo = netinfomocks.NewMockProvider(suite.ctrl)

	_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)

	suite.provider = iface.NewDebianProvider(
		suite.logger,
		suite.memFs,
		suite.mockStateKV,
		suite.mockExec,
		testHostname,
		suite.mockNetinfo,
	)
}

func (suite *InterfacePublicTestSuite) SetupSubTest() {
	suite.SetupTest()
}

func (suite *InterfacePublicTestSuite) TearDownSubTest() {}

func (suite *InterfacePublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func([]iface.InterfaceEntry, error)
	}{
		{
			name: "when interfaces exist with managed files",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return([]netinfo.InterfaceResult{
						{Name: "eth0", IPv4: "10.0.0.5"},
						{Name: "eth1", IPv4: "10.0.1.5"},
					}, nil)

				// Create a managed file for eth0 only.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
					[]byte("network:\n"),
					0o644,
				)
			},
			validateFunc: func(result []iface.InterfaceEntry, err error) {
				suite.Require().NoError(err)
				suite.Require().Len(result, 2)

				suite.Equal("eth0", result[0].Name)
				suite.True(result[0].Managed)

				suite.Equal("eth1", result[1].Name)
				suite.False(result[1].Managed)
			},
		},
		{
			name: "when no managed files exist",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return([]netinfo.InterfaceResult{
						{Name: "eth0"},
					}, nil)
			},
			validateFunc: func(result []iface.InterfaceEntry, err error) {
				suite.Require().NoError(err)
				suite.Require().Len(result, 1)

				suite.Equal("eth0", result[0].Name)
				suite.False(result[0].Managed)
			},
		},
		{
			name: "when no interfaces exist",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return([]netinfo.InterfaceResult{}, nil)
			},
			validateFunc: func(result []iface.InterfaceEntry, err error) {
				suite.Require().NoError(err)
				suite.Empty(result)
			},
		},
		{
			name: "when GetInterfaces fails",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return(nil, errors.New("netinfo error"))
			},
			validateFunc: func(result []iface.InterfaceEntry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan interface list:")
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

func (suite *InterfacePublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		interfaceNm  string
		setup        func()
		validateFunc func(*iface.InterfaceEntry, error)
	}{
		{
			name:        "when interface found and managed",
			interfaceNm: "eth0",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return([]netinfo.InterfaceResult{
						{Name: "eth0", IPv4: "10.0.0.5"},
					}, nil)

				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
					[]byte("network:\n"),
					0o644,
				)
			},
			validateFunc: func(result *iface.InterfaceEntry, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Name)
				suite.True(result.Managed)
			},
		},
		{
			name:        "when interface found and not managed",
			interfaceNm: "eth0",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return([]netinfo.InterfaceResult{
						{Name: "eth0"},
					}, nil)
			},
			validateFunc: func(result *iface.InterfaceEntry, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Name)
				suite.False(result.Managed)
			},
		},
		{
			name:        "when interface not found",
			interfaceNm: "eth99",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return([]netinfo.InterfaceResult{
						{Name: "eth0"},
					}, nil)
			},
			validateFunc: func(result *iface.InterfaceEntry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "not found")
			},
		},
		{
			name:        "when name is empty",
			interfaceNm: "",
			setup:       func() {},
			validateFunc: func(result *iface.InterfaceEntry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "name must not be empty")
			},
		},
		{
			name:        "when GetInterfaces fails",
			interfaceNm: "eth0",
			setup: func() {
				suite.mockNetinfo.EXPECT().
					GetInterfaces().
					Return(nil, errors.New("netinfo error"))
			},
			validateFunc: func(result *iface.InterfaceEntry, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan interface get:")
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

func (suite *InterfacePublicTestSuite) TestCreate() {
	dhcp4True := true

	tests := []struct {
		name         string
		entry        iface.InterfaceEntry
		setup        func()
		validateFunc func(*iface.InterfaceResult, error)
	}{
		{
			name: "when new interface deploys successfully",
			entry: iface.InterfaceEntry{
				Name:      "eth0",
				DHCP4:     &dhcp4True,
				Addresses: []string{"10.0.0.5/24"},
				Gateway4:  "10.0.0.1",
				MTU:       1500,
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
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Name)
				suite.True(result.Changed)

				// Verify file was written.
				data, readErr := suite.memFs.ReadFile(
					"/etc/netplan/osapi-eth0.yaml",
				)
				suite.Require().NoError(readErr)
				suite.Contains(string(data), "eth0:")
				suite.Contains(string(data), "dhcp4: true")
				suite.Contains(string(data), "10.0.0.5/24")
				suite.Contains(string(data), "gateway4: 10.0.0.1")
				suite.Contains(string(data), "mtu: 1500")
			},
		},
		{
			name: "when interface already managed",
			entry: iface.InterfaceEntry{
				Name:  "eth0",
				DHCP4: &dhcp4True,
			},
			setup: func() {
				// File already exists on disk.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
					[]byte("existing"),
					0o644,
				)
			},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "already managed")
			},
		},
		{
			name: "when name is empty",
			entry: iface.InterfaceEntry{
				Name: "",
			},
			setup: func() {},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "name must not be empty")
			},
		},
		{
			name: "when name has invalid characters",
			entry: iface.InterfaceEntry{
				Name: "eth 0!",
			},
			setup: func() {},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "invalid characters")
			},
		},
		{
			name: "when ApplyConfig fails",
			entry: iface.InterfaceEntry{
				Name:  "eth0",
				DHCP4: &dhcp4True,
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
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan interface create:")
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

func (suite *InterfacePublicTestSuite) TestUpdate() {
	dhcp4False := false

	tests := []struct {
		name         string
		entry        iface.InterfaceEntry
		setup        func()
		validateFunc func(*iface.InterfaceResult, error)
	}{
		{
			name: "when update succeeds",
			entry: iface.InterfaceEntry{
				Name:      "eth0",
				DHCP4:     &dhcp4False,
				Addresses: []string{"10.0.0.10/24"},
			},
			setup: func() {
				// Managed file exists on disk.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
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
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Name)
				suite.True(result.Changed)
			},
		},
		{
			name: "when interface not managed",
			entry: iface.InterfaceEntry{
				Name:  "eth0",
				DHCP4: &dhcp4False,
			},
			setup: func() {
				// No file on disk.
			},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "not managed")
			},
		},
		{
			name: "when name is empty",
			entry: iface.InterfaceEntry{
				Name: "",
			},
			setup: func() {},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "name must not be empty")
			},
		},
		{
			name: "when ApplyConfig fails",
			entry: iface.InterfaceEntry{
				Name:  "eth0",
				DHCP4: &dhcp4False,
			},
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
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
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan interface update:")
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

func (suite *InterfacePublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		interfaceNm  string
		setup        func()
		validateFunc func(*iface.InterfaceResult, error)
	}{
		{
			name:        "when file exists and removal succeeds",
			interfaceNm: "eth0",
			setup: func() {
				// Create the managed file on disk.
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
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
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Name)
				suite.True(result.Changed)

				// Verify file was removed.
				_, statErr := suite.memFs.Stat(
					"/etc/netplan/osapi-eth0.yaml",
				)
				suite.Error(statErr)
			},
		},
		{
			name:        "when file does not exist",
			interfaceNm: "eth0",
			setup:       func() {},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().NoError(err)
				suite.Require().NotNil(result)
				suite.Equal("eth0", result.Name)
				suite.False(result.Changed)
			},
		},
		{
			name:        "when name is empty",
			interfaceNm: "",
			setup:       func() {},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "name must not be empty")
			},
		},
		{
			name:        "when RemoveConfig fails",
			interfaceNm: "eth0",
			setup: func() {
				_ = suite.memFs.WriteFile(
					"/etc/netplan/osapi-eth0.yaml",
					[]byte("network:\n"),
					0o644,
				)

				suite.mockExec.EXPECT().
					RunPrivilegedCmd("netplan", []string{"apply"}).
					Return("", errors.New("apply failed"))
			},
			validateFunc: func(result *iface.InterfaceResult, err error) {
				suite.Require().Error(err)
				suite.Nil(result)
				suite.Contains(err.Error(), "netplan interface delete:")
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

func (suite *InterfacePublicTestSuite) TestGenerateInterfaceYAML() {
	dhcp4True := true
	dhcp4False := false
	dhcp6True := true
	wolTrue := true

	tests := []struct {
		name         string
		entry        iface.InterfaceEntry
		validateFunc func(string)
	}{
		{
			name: "when all fields set",
			entry: iface.InterfaceEntry{
				Name:       "eth0",
				DHCP4:      &dhcp4False,
				DHCP6:      &dhcp6True,
				Addresses:  []string{"10.0.0.5/24", "10.0.0.6/24"},
				Gateway4:   "10.0.0.1",
				Gateway6:   "fe80::1",
				MTU:        9000,
				MACAddress: "aa:bb:cc:dd:ee:ff",
				WakeOnLAN:  &wolTrue,
			},
			validateFunc: func(result string) {
				suite.Contains(result, "network:")
				suite.Contains(result, "  version: 2")
				suite.Contains(result, "  ethernets:")
				suite.Contains(result, "    eth0:")
				suite.Contains(result, "      dhcp4: false")
				suite.Contains(result, "      dhcp6: true")
				suite.Contains(result, "      addresses:")
				suite.Contains(result, "        - 10.0.0.5/24")
				suite.Contains(result, "        - 10.0.0.6/24")
				suite.Contains(result, "      gateway4: 10.0.0.1")
				suite.Contains(result, "      gateway6: fe80::1")
				suite.Contains(result, "      mtu: 9000")
				suite.Contains(result, "      macaddress: aa:bb:cc:dd:ee:ff")
				suite.Contains(result, "      wakeonlan: true")
			},
		},
		{
			name: "when only DHCP4 set",
			entry: iface.InterfaceEntry{
				Name:  "eth0",
				DHCP4: &dhcp4True,
			},
			validateFunc: func(result string) {
				suite.Contains(result, "network:")
				suite.Contains(result, "    eth0:")
				suite.Contains(result, "      dhcp4: true")
				suite.NotContains(result, "addresses:")
				suite.NotContains(result, "gateway4:")
				suite.NotContains(result, "gateway6:")
				suite.NotContains(result, "mtu:")
				suite.NotContains(result, "macaddress:")
				suite.NotContains(result, "wakeonlan:")
			},
		},
		{
			name: "when only addresses set",
			entry: iface.InterfaceEntry{
				Name:      "ens3",
				Addresses: []string{"192.168.1.10/24"},
			},
			validateFunc: func(result string) {
				suite.Contains(result, "    ens3:")
				suite.Contains(result, "      addresses:")
				suite.Contains(result, "        - 192.168.1.10/24")
				suite.NotContains(result, "dhcp4:")
				suite.NotContains(result, "gateway4:")
				suite.NotContains(result, "mtu:")
			},
		},
		{
			name: "when IPv6 gateway set",
			entry: iface.InterfaceEntry{
				Name:      "eth0",
				Addresses: []string{"fd00::5/64"},
				Gateway6:  "fd00::1",
			},
			validateFunc: func(result string) {
				suite.Contains(result, "      gateway6: fd00::1")
				suite.Contains(result, "        - fd00::5/64")
				suite.NotContains(result, "gateway4:")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			result := iface.GenerateInterfaceYAML(tc.entry)

			tc.validateFunc(string(result))
		})
	}
}

func (suite *InterfacePublicTestSuite) TestDetectDHCP() {
	tests := []struct {
		name         string
		setup        func()
		ifaceName    string
		validateFunc func(*bool)
	}{
		{
			name:      "when wifi interface has dhcp4 true",
			ifaceName: "wlp0s20f3",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/00-installer-config-wifi.yaml",
					[]byte("# This is the network config written by 'subiquity'\nnetwork:\n  version: 2\n  wifis:\n    wlp0s20f3:\n      access-points:\n        Yikes:\n          password: foo\n      dhcp4: true\n"),
					0o644,
				)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/00-installer-config.yaml",
					[]byte("# This is the network config written by 'subiquity'\nnetwork:\n  ethernets:\n    eno1:\n      dhcp4: true\n  version: 2\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Require().NotNil(result)
				suite.True(*result)
			},
		},
		{
			name:      "when ethernet interface has dhcp4 true",
			ifaceName: "eno1",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/00-installer-config.yaml",
					[]byte("network:\n  ethernets:\n    eno1:\n      dhcp4: true\n  version: 2\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Require().NotNil(result)
				suite.True(*result)
			},
		},
		{
			name:      "when interface has static config no dhcp4",
			ifaceName: "eth0",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/01-config.yaml",
					[]byte("network:\n  ethernets:\n    eth0:\n      addresses:\n        - 10.0.0.5/24\n      gateway4: 10.0.0.1\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Require().NotNil(result)
				suite.False(*result)
			},
		},
		{
			name:      "when interface not found in any file",
			ifaceName: "eth99",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/01-config.yaml",
					[]byte("network:\n  ethernets:\n    eth0:\n      dhcp4: true\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Nil(result)
			},
		},
		{
			name:      "when netplan dir does not exist",
			ifaceName: "eth0",
			setup:     func() {},
			validateFunc: func(result *bool) {
				suite.Nil(result)
			},
		},
		{
			name:      "when dhcp4 yes is used",
			ifaceName: "eth0",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/01-config.yaml",
					[]byte("network:\n  ethernets:\n    eth0:\n      dhcp4: yes\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Require().NotNil(result)
				suite.True(*result)
			},
		},
		{
			name:      "when yml extension is supported",
			ifaceName: "eth0",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/01-config.yml",
					[]byte("network:\n  ethernets:\n    eth0:\n      dhcp4: true\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Require().NotNil(result)
				suite.True(*result)
			},
		},
		{
			name:      "when file read fails skips to next",
			ifaceName: "eth0",
			setup: func() {
				base := memfs.New()
				_ = base.MkdirAll("/etc/netplan", 0o755)
				_ = base.WriteFile(
					"/etc/netplan/01-bad.yaml",
					[]byte("network:\n  ethernets:\n    eth0:\n      dhcp4: true\n"),
					0o644,
				)

				ffs := failfs.New(base)
				_ = ffs.SetFailFunc(func(
					_ avfs.VFSBase,
					fn avfs.FnVFS,
					_ *failfs.FailParam,
				) error {
					if fn == avfs.FnReadFile {
						return errors.New("read error")
					}

					return nil
				})

				suite.memFs = ffs
			},
			validateFunc: func(result *bool) {
				suite.Nil(result)
			},
		},
		{
			name:      "when multiple files exist finds correct one",
			ifaceName: "eth1",
			setup: func() {
				_ = suite.memFs.MkdirAll("/etc/netplan", 0o755)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/01-eth0.yaml",
					[]byte("network:\n  ethernets:\n    eth0:\n      dhcp4: true\n"),
					0o644,
				)
				_ = suite.memFs.WriteFile(
					"/etc/netplan/02-eth1.yaml",
					[]byte("network:\n  ethernets:\n    eth1:\n      addresses:\n        - 10.0.1.5/24\n"),
					0o644,
				)
			},
			validateFunc: func(result *bool) {
				suite.Require().NotNil(result)
				suite.False(*result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			tc.setup()

			result := iface.DetectDHCP(suite.memFs, tc.ifaceName)

			tc.validateFunc(result)
		})
	}
}

func TestInterfacePublicTestSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(InterfacePublicTestSuite))
}
