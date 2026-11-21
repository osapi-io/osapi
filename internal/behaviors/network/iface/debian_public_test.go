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
	"testing"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	netplanmocks "github.com/osapi-io/osapi/internal/behaviors/netplan/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/network/iface"
)

// DebianPublicTestSuite covers what no example can: a container, where the
// runtime owns the interfaces, and a platform with no writer.
type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) TestContainerSkips() {
	p, err := iface.ExportNewDebian(host.Host{
		Container: true,
		Required: map[string]any{
			netplan.Behavior.Name: netplanmocks.NewMockProvider(gomock.NewController(s.T())),
		},
	})
	s.Require().NoError(err)
	_, err = p.Create(context.Background(), iface.Config{Name: "eth1"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Update(context.Background(), iface.Config{Name: "eth1"})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "eth1")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *DebianPublicTestSuite) TestReaderWritesAreUnsupported() {
	p, err := iface.ExportNewReader(host.Host{Platform: behavior.PlatformDarwin})
	s.Require().NoError(err)
	_, err = p.Create(context.Background(), iface.Config{})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Update(context.Background(), iface.Config{})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "en0")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *DebianPublicTestSuite) TestNetplanReadFails() {
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/sys/class/net/osapi0", 0o755))
	s.Require().NoError(fs.WriteFile("/sys/class/net/osapi0/ifindex", []byte("3"), 0o644))
	np := netplanmocks.NewMockProvider(gomock.NewController(s.T()))
	np.EXPECT().Interface(gomock.Any(), "osapi0").Return(netplan.Interface{}, errors.New("boom"))
	p, err := iface.ExportNewDebian(host.Host{
		Platform: behavior.PlatformDebian,
		FS:       fs,
		Required: map[string]any{netplan.Behavior.Name: np},
	})
	s.Require().NoError(err)
	_, err = p.Create(context.Background(), iface.Config{Name: "osapi0", DHCP4: true})
	s.ErrorContains(err, "boom")
}

// TestGateways checks the default routes that the gateways make, which no
// example sets: the examples also run on a live host.
func (s *DebianPublicTestSuite) TestGateways() {
	s.Equal(
		"network:\n  version: 2\n  ethernets:\n    eth1:\n"+
			"      dhcp4: false\n      dhcp6: false\n"+
			"      routes:\n"+
			"        - to: default\n          via: 10.0.1.1\n"+
			"        - to: default\n          via: 2001:db8::1\n",
		iface.ExportYAML("eth1", "ethernets",
			iface.Config{Name: "eth1", Gateway4: "10.0.1.1", Gateway6: "2001:db8::1"}),
	)
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
