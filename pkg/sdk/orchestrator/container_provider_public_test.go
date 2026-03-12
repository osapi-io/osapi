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

package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

// mockTarget implements RuntimeTarget for testing.
type mockTarget struct {
	responses map[string][]byte
	errors    map[string]error
	lastData  []byte
}

func (m *mockTarget) Name() string {
	return "test-container"
}

func (m *mockTarget) Runtime() string {
	return "docker"
}

func (m *mockTarget) ExecProvider(
	_ context.Context,
	provider string,
	operation string,
	data []byte,
) ([]byte, error) {
	m.lastData = data
	key := provider + "/" + operation

	if err, ok := m.errors[key]; ok {
		return nil, err
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	return nil, fmt.Errorf("unexpected call: %s", key)
}

type ContainerProviderPublicTestSuite struct {
	suite.Suite
}

func TestContainerProviderPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerProviderPublicTestSuite))
}

func (suite *ContainerProviderPublicTestSuite) TestGetHostname() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result string, err error)
	}{
		{
			name:     "returns hostname",
			response: []byte(`"my-container"`),
			validateFunc: func(result string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "my-container", result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "connection refused")
			},
		},
		{
			name:     "unmarshal error",
			response: []byte(`not-json`),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "unmarshal")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-hostname": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-hostname"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetHostname(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetOSInfo() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result *orchestrator.HostOSInfo, err error)
	}{
		{
			name:     "returns os info",
			response: []byte(`{"Distribution":"ubuntu","Version":"24.04"}`),
			validateFunc: func(result *orchestrator.HostOSInfo, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "ubuntu", result.Distribution)
				assert.Equal(suite.T(), "24.04", result.Version)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("timeout"),
			validateFunc: func(_ *orchestrator.HostOSInfo, err error) {
				assert.Error(suite.T(), err)
			},
		},
		{
			name:     "unmarshal error",
			response: []byte(`{bad`),
			validateFunc: func(_ *orchestrator.HostOSInfo, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "unmarshal")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-os-info": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-os-info"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetOSInfo(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetMemStats() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result *orchestrator.MemStats, err error)
	}{
		{
			name:     "parses memory stats",
			response: []byte(`{"Total":8192000000,"Available":4096000000,"Free":2048000000,"Cached":1024000000}`),
			validateFunc: func(result *orchestrator.MemStats, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), uint64(8192000000), result.Total)
				assert.Equal(suite.T(), uint64(4096000000), result.Available)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"mem/get-stats": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["mem/get-stats"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetMemStats(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetLoadStats() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result *orchestrator.LoadStats, err error)
	}{
		{
			name:     "parses load stats",
			response: []byte(`{"Load1":0.5,"Load5":0.75,"Load15":1.0}`),
			validateFunc: func(result *orchestrator.LoadStats, err error) {
				assert.NoError(suite.T(), err)
				assert.InDelta(suite.T(), float32(0.5), result.Load1, 0.01)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"load/get-average-stats": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["load/get-average-stats"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetLoadStats(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestExec() {
	tests := []struct {
		name         string
		params       orchestrator.ExecParams
		response     []byte
		err          error
		validateFunc func(result *orchestrator.CommandResult, err error, lastData []byte)
	}{
		{
			name: "runs command with changed true",
			params: orchestrator.ExecParams{
				Command: "uname",
				Args:    []string{"-a"},
			},
			response: []byte(`{"stdout":"Linux container 5.15.0\n","stderr":"","exit_code":0,"duration_ms":5,"changed":true}`),
			validateFunc: func(result *orchestrator.CommandResult, err error, lastData []byte) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "Linux container 5.15.0\n", result.Stdout)
				assert.True(suite.T(), result.Changed)

				var sent orchestrator.ExecParams
				_ = json.Unmarshal(lastData, &sent)
				assert.Equal(suite.T(), "uname", sent.Command)
				assert.Equal(suite.T(), []string{"-a"}, sent.Args)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"command/exec": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["command/exec"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.Exec(context.Background(), tc.params)
			tc.validateFunc(result, err, target.lastData)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestShell() {
	tests := []struct {
		name         string
		command      string
		response     []byte
		err          error
		validateFunc func(result *orchestrator.CommandResult, err error)
	}{
		{
			name:     "runs shell command",
			command:  "echo hello",
			response: []byte(`{"stdout":"hello\n","stderr":"","exit_code":0,"duration_ms":2}`),
			validateFunc: func(result *orchestrator.CommandResult, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "hello\n", result.Stdout)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"command/shell": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["command/shell"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.Shell(context.Background(), orchestrator.ShellParams{
				Command: tc.command,
			})
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestPing() {
	tests := []struct {
		name         string
		address      string
		response     []byte
		err          error
		validateFunc func(result *orchestrator.PingResult, err error, lastData []byte)
	}{
		{
			name:     "pings address",
			address:  "8.8.8.8",
			response: []byte(`{"PacketsSent":3,"PacketsReceived":3,"PacketLoss":0}`),
			validateFunc: func(result *orchestrator.PingResult, err error, lastData []byte) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), 3, result.PacketsSent)

				var sent orchestrator.PingParams
				_ = json.Unmarshal(lastData, &sent)
				assert.Equal(suite.T(), "8.8.8.8", sent.Address)
			},
		},
		{
			name:    "exec error",
			address: "8.8.8.8",
			err:     fmt.Errorf("timeout"),
			validateFunc: func(_ *orchestrator.PingResult, err error, _ []byte) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"ping/do": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["ping/do"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.Ping(context.Background(), tc.address)
			tc.validateFunc(result, err, target.lastData)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetArchitecture() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result string, err error)
	}{
		{
			name:     "returns architecture",
			response: []byte(`"x86_64"`),
			validateFunc: func(result string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "x86_64", result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-architecture": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-architecture"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetArchitecture(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetKernelVersion() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result string, err error)
	}{
		{
			name:     "returns kernel version",
			response: []byte(`"5.15.0-91-generic"`),
			validateFunc: func(result string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "5.15.0-91-generic", result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-kernel-version": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-kernel-version"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetKernelVersion(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetUptime() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result time.Duration, err error)
	}{
		{
			name:     "returns uptime duration",
			response: []byte(`3600000000000`),
			validateFunc: func(result time.Duration, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), time.Hour, result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ time.Duration, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-uptime": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-uptime"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetUptime(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetFQDN() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result string, err error)
	}{
		{
			name:     "returns fqdn",
			response: []byte(`"web-01.example.com"`),
			validateFunc: func(result string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "web-01.example.com", result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-fqdn": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-fqdn"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetFQDN(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetCPUCount() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result int, err error)
	}{
		{
			name:     "returns cpu count",
			response: []byte(`4`),
			validateFunc: func(result int, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), 4, result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ int, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-cpu-count": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-cpu-count"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetCPUCount(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetServiceManager() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result string, err error)
	}{
		{
			name:     "returns service manager",
			response: []byte(`"systemd"`),
			validateFunc: func(result string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "systemd", result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-service-manager": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-service-manager"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetServiceManager(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetPackageManager() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result string, err error)
	}{
		{
			name:     "returns package manager",
			response: []byte(`"apt"`),
			validateFunc: func(result string, err error) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), "apt", result)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ string, err error) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"host/get-package-manager": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["host/get-package-manager"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetPackageManager(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetDiskUsage() {
	tests := []struct {
		name         string
		response     []byte
		err          error
		validateFunc func(result []orchestrator.DiskUsage, err error)
	}{
		{
			name:     "returns disk usage",
			response: []byte(`[{"Name":"/","Total":50000000000,"Used":25000000000,"Free":25000000000}]`),
			validateFunc: func(result []orchestrator.DiskUsage, err error) {
				assert.NoError(suite.T(), err)
				assert.Len(suite.T(), result, 1)
				assert.Equal(suite.T(), "/", result[0].Name)
				assert.Equal(suite.T(), uint64(50000000000), result[0].Total)
			},
		},
		{
			name: "exec error",
			err:  fmt.Errorf("connection refused"),
			validateFunc: func(_ []orchestrator.DiskUsage, err error) {
				assert.Error(suite.T(), err)
			},
		},
		{
			name:     "unmarshal error",
			response: []byte(`not-json`),
			validateFunc: func(_ []orchestrator.DiskUsage, err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "unmarshal")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"disk/get-local-usage-stats": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["disk/get-local-usage-stats"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetDiskUsage(context.Background())
			tc.validateFunc(result, err)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestGetDNS() {
	tests := []struct {
		name         string
		iface        string
		response     []byte
		err          error
		validateFunc func(result *orchestrator.DNSGetResult, err error, lastData []byte)
	}{
		{
			name:     "returns dns config",
			iface:    "eth0",
			response: []byte(`{"DNSServers":["8.8.8.8","8.8.4.4"],"SearchDomains":["example.com"]}`),
			validateFunc: func(result *orchestrator.DNSGetResult, err error, lastData []byte) {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), []string{"8.8.8.8", "8.8.4.4"}, result.DNSServers)
				assert.Equal(suite.T(), []string{"example.com"}, result.SearchDomains)

				var sent orchestrator.DNSGetParams
				_ = json.Unmarshal(lastData, &sent)
				assert.Equal(suite.T(), "eth0", sent.InterfaceName)
			},
		},
		{
			name:  "exec error",
			iface: "eth0",
			err:   fmt.Errorf("connection refused"),
			validateFunc: func(_ *orchestrator.DNSGetResult, err error, _ []byte) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"dns/get-resolv-conf": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["dns/get-resolv-conf"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.GetDNS(context.Background(), tc.iface)
			tc.validateFunc(result, err, target.lastData)
		})
	}
}

func (suite *ContainerProviderPublicTestSuite) TestUpdateDNS() {
	tests := []struct {
		name         string
		params       orchestrator.DNSUpdateParams
		response     []byte
		err          error
		validateFunc func(result *orchestrator.DNSUpdateResult, err error, lastData []byte)
	}{
		{
			name: "updates dns config",
			params: orchestrator.DNSUpdateParams{
				Servers:       []string{"8.8.8.8"},
				SearchDomains: []string{"example.com"},
				InterfaceName: "eth0",
			},
			response: []byte(`{"changed":true}`),
			validateFunc: func(result *orchestrator.DNSUpdateResult, err error, lastData []byte) {
				assert.NoError(suite.T(), err)
				assert.True(suite.T(), result.Changed)

				var sent orchestrator.DNSUpdateParams
				_ = json.Unmarshal(lastData, &sent)
				assert.Equal(suite.T(), []string{"8.8.8.8"}, sent.Servers)
				assert.Equal(suite.T(), "eth0", sent.InterfaceName)
			},
		},
		{
			name: "exec error",
			params: orchestrator.DNSUpdateParams{
				Servers:       []string{"8.8.8.8"},
				InterfaceName: "eth0",
			},
			err: fmt.Errorf("connection refused"),
			validateFunc: func(_ *orchestrator.DNSUpdateResult, err error, _ []byte) {
				assert.Error(suite.T(), err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &mockTarget{
				responses: map[string][]byte{"dns/update-resolv-conf": tc.response},
				errors:    map[string]error{},
			}
			if tc.err != nil {
				target.errors["dns/update-resolv-conf"] = tc.err
			}

			cp := orchestrator.NewContainerProvider(target)
			result, err := cp.UpdateDNS(context.Background(), tc.params)
			tc.validateFunc(result, err, target.lastData)
		})
	}
}
