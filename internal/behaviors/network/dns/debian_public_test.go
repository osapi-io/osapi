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

package dns_test

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
	"github.com/osapi-io/osapi/internal/behaviors/network/dns"
)

// DebianPublicTestSuite covers what no example can: a container, where the
// runtime owns resolv.conf, and a netplan read that fails.
type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) build(
	h host.Host,
	np netplan.Provider,
) dns.Provider {
	h.Required = map[string]any{netplan.Behavior.Name: np}
	p, err := dns.ExportNewDebian(h)
	s.Require().NoError(err)

	return p
}

func (s *DebianPublicTestSuite) TestContainerSkips() {
	p := s.build(
		host.Host{Container: true},
		netplanmocks.NewMockProvider(gomock.NewController(s.T())),
	)
	_, err := p.Set(
		context.Background(),
		dns.UpdateInput{InterfaceName: "eth0", Servers: []string{"9.9.9.9"}},
	)
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "eth0")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func (s *DebianPublicTestSuite) TestNetplanReadFails() {
	np := netplanmocks.NewMockProvider(gomock.NewController(s.T()))
	np.EXPECT().Interface(gomock.Any(), "eth0").Return(netplan.Interface{}, errors.New("boom"))
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/sys/class/net/eth0", 0o755))
	s.Require().NoError(fs.WriteFile("/sys/class/net/eth0/ifindex", []byte("2"), 0o644))
	p := s.build(host.Host{Platform: behavior.PlatformDebian, FS: fs}, np)
	_, err := p.Set(context.Background(), dns.UpdateInput{
		InterfaceName: "eth0", Servers: []string{"9.9.9.9"}, SearchDomains: []string{"a.example"},
	})
	s.ErrorContains(err, "boom")
}

func (s *DebianPublicTestSuite) TestReaderWritesAreUnsupported() {
	p, err := dns.ExportNewReader(host.Host{Platform: behavior.PlatformDarwin})
	s.Require().NoError(err)
	_, err = p.Set(context.Background(), dns.UpdateInput{})
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "en0")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
