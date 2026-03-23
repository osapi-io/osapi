// Copyright (c) 2025 John Dewey

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

package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/suite"

	dockerProv "github.com/retr0h/osapi/internal/provider/docker"
	"github.com/retr0h/osapi/pkg/sdk/platform"
)

// testDockerClient is a minimal mock that satisfies dockerclient.APIClient
// by embedding the interface and overriding only Ping.
type testDockerClient struct {
	dockerclient.APIClient
	pingErr error
}

func (c *testDockerClient) Ping(
	_ context.Context,
) (types.Ping, error) {
	if c.pingErr != nil {
		return types.Ping{}, c.pingErr
	}
	return types.Ping{}, nil
}

type FactoryTestSuite struct {
	suite.Suite
}

func (s *FactoryTestSuite) TestCreateProviders() {
	tests := []struct {
		name          string
		setupMock     func() func() (*host.InfoStat, error)
		setupDocker   func()
		wantContainer bool
	}{
		{
			name: "creates ubuntu providers when platform is ubuntu",
			setupMock: func() func() (*host.InfoStat, error) {
				return func() (*host.InfoStat, error) {
					return &host.InfoStat{
						Platform: "Ubuntu",
					}, nil
				}
			},
		},
		{
			name: "creates darwin providers when platform is empty and OS is darwin",
			setupMock: func() func() (*host.InfoStat, error) {
				return func() (*host.InfoStat, error) {
					return &host.InfoStat{
						Platform: "",
						OS:       "darwin",
					}, nil
				}
			},
		},
		{
			name: "creates linux providers for unknown platform",
			setupMock: func() func() (*host.InfoStat, error) {
				return func() (*host.InfoStat, error) {
					return &host.InfoStat{
						Platform: "centos",
					}, nil
				}
			},
		},
		{
			name: "docker New error disables container provider",
			setupMock: func() func() (*host.InfoStat, error) {
				return func() (*host.InfoStat, error) {
					return &host.InfoStat{Platform: "Ubuntu"}, nil
				}
			},
			setupDocker: func() {
				factoryDockerNewFn = func() (*dockerProv.Client, error) {
					return nil, fmt.Errorf("docker not installed")
				}
			},
			wantContainer: false,
		},
		{
			name: "docker Ping error disables container provider",
			setupMock: func() func() (*host.InfoStat, error) {
				return func() (*host.InfoStat, error) {
					return &host.InfoStat{Platform: "Ubuntu"}, nil
				}
			},
			setupDocker: func() {
				factoryDockerNewFn = func() (*dockerProv.Client, error) {
					return dockerProv.NewWithClient(&testDockerClient{
						pingErr: errors.New("connection refused"),
					}), nil
				}
			},
			wantContainer: false,
		},
		{
			name: "docker available enables container provider",
			setupMock: func() func() (*host.InfoStat, error) {
				return func() (*host.InfoStat, error) {
					return &host.InfoStat{Platform: "Ubuntu"}, nil
				}
			},
			setupDocker: func() {
				factoryDockerNewFn = func() (*dockerProv.Client, error) {
					return dockerProv.NewWithClient(&testDockerClient{}), nil
				}
			},
			wantContainer: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			originalHost := platform.HostInfoFn
			originalDocker := factoryDockerNewFn
			defer func() {
				platform.HostInfoFn = originalHost
				factoryDockerNewFn = originalDocker
			}()

			platform.HostInfoFn = tt.setupMock()
			if tt.setupDocker != nil {
				tt.setupDocker()
			}

			factory := NewProviderFactory(slog.Default(), nil)
			hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider, netinfoProvider, commandProvider, dockerProvider, cronProvider := factory.CreateProviders()

			s.NotNil(hostProvider)
			s.NotNil(diskProvider)
			s.NotNil(memProvider)
			s.NotNil(loadProvider)
			s.NotNil(dnsProvider)
			s.NotNil(pingProvider)
			s.NotNil(netinfoProvider)
			s.NotNil(commandProvider)

			if tt.wantContainer {
				s.NotNil(dockerProvider)
			}
			s.NotNil(cronProvider)
		})
	}
}

func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}
