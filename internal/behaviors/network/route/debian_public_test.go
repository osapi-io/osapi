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
	"errors"
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	netplanmocks "github.com/osapi-io/osapi/internal/behaviors/netplan/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/network/route"
)

// DebianPublicTestSuite covers what no example can: a container, a platform
// with no writer, a netplan read that fails, and each form a route string
// takes.
type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) TestContainerSkips() {
	p, err := route.ExportNewDebian(host.Host{
		Container: true,
		Required: map[string]any{
			netplan.Behavior.Name: netplanmocks.NewMockProvider(gomock.NewController(s.T())),
		},
	})
	s.Require().NoError(err)
	_, err = p.Create(context.Background(), route.Config{InterfaceName: "osapi0"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Update(context.Background(), route.Config{InterfaceName: "osapi0"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "osapi0")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *DebianPublicTestSuite) TestReaderWritesAreUnsupported() {
	p, err := route.ExportNewReader(host.Host{Platform: behavior.PlatformDarwin})
	s.Require().NoError(err)
	_, err = p.Create(context.Background(), route.Config{})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Update(context.Background(), route.Config{})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "en0")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *DebianPublicTestSuite) TestReadFails() {
	p, err := route.ExportNewReader(host.Host{Platform: behavior.PlatformDebian, FS: memfs.New()})
	s.Require().NoError(err)
	_, err = p.Routes(context.Background())
	s.ErrorContains(err, "read the routes")
	_, err = p.InterfaceRoutes(context.Background(), "eth0")
	s.ErrorContains(err, "read the routes")
}

func (s *DebianPublicTestSuite) TestNetplanReadFails() {
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/sys/class/net/osapi0", 0o755))
	s.Require().NoError(fs.WriteFile("/sys/class/net/osapi0/ifindex", []byte("3"), 0o644))
	np := netplanmocks.NewMockProvider(gomock.NewController(s.T()))
	np.EXPECT().Interface(gomock.Any(), "osapi0").Return(netplan.Interface{}, errors.New("boom"))
	p, err := route.ExportNewDebian(host.Host{
		Platform: behavior.PlatformDebian,
		FS:       fs,
		Required: map[string]any{netplan.Behavior.Name: np},
	})
	s.Require().NoError(err)
	_, err = p.Create(context.Background(), route.Config{
		InterfaceName: "osapi0", Routes: []string{"198.51.100.0/24 via 192.0.2.1"},
	})
	s.ErrorContains(err, "boom")
}

func (s *DebianPublicTestSuite) TestParse() {
	tests := []struct {
		in      string
		wantErr string
	}{
		{in: "10.0.0.0/8 via 192.168.1.1"},
		{in: "2001:db8::/32 via 2001:db8::1 metric 4294967295"},
		{in: "10.0.0.0/8 via 192.168.1.1 metric 4294967296", wantErr: "the metric"},
		{in: "10.0.0.0/8 via 192.168.1.1 weight 1", wantErr: "is not <destination>"},
		{in: "10.0.0.0/8 via", wantErr: "is not <destination>"},
		{in: "bogus via 192.168.1.1", wantErr: "CIDR"},
		{in: "10.0.0.0/8 via bogus", wantErr: "the gateway"},
	}
	for _, tc := range tests {
		s.Run(tc.in, func() {
			_, err := route.ExportParse(tc.in)
			if tc.wantErr == "" {
				s.NoError(err)

				return
			}
			s.ErrorContains(err, tc.wantErr)
		})
	}
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
