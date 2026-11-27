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

	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	netplanmocks "github.com/osapi-io/osapi/internal/behaviors/netplan/mocks"
	"go.uber.org/mock/gomock"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/netplan"
	"github.com/osapi-io/osapi/internal/behaviors/network/route"
)

type DebianPublicTestSuite struct {
	suite.Suite
}

// TestUnsupported: a container and a platform with no writer read the
// routes, and refuse each write.
func (s *DebianPublicTestSuite) TestUnsupported() {
	ctx := context.Background()
	in := route.CreateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "192.0.2.1"}
	del := route.DeleteInput{Name: "osapi0", Destination: "198.51.100.0/24"}
	inContainer, err := route.ExportNewDebian(host.Host{Container: true})
	s.Require().NoError(err)
	reader, err := route.ExportNewReader(host.Host{})
	s.Require().NoError(err)
	for _, p := range []route.Provider{inContainer, reader} {
		_, err = route.Create(ctx, p, in)
		s.ErrorIs(err, behavior.ErrUnsupported)
		_, err = route.Delete(ctx, p, del)
		s.ErrorIs(err, behavior.ErrUnsupported)
	}
}

// TestDamagedFile: a routes file that osapi cannot read is an error, not
// routes that osapi drops; a file with no definition has no routes.
func (s *DebianPublicTestSuite) TestDamagedFile() {
	file := netplan.Dir + "/osapi-osapi0-routes.yaml"
	for content, want := range map[string]string{
		"network: [":                 "read " + file,
		"network:\n  ethernets: 5\n": "read " + file,
		"network:\n  version: 2\n":   "not found",
	} {
		fs := memfs.New()
		s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
		s.Require().NoError(fs.WriteFile(file, []byte(content), 0o600))
		p, err := route.ExportNewDebian(host.Host{FS: fs})
		s.Require().NoError(err)

		_, err = route.Update(context.Background(), p,
			route.UpdateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "192.0.2.1"})
		s.ErrorContains(err, want, content)
		_, err = route.Delete(context.Background(), p,
			route.DeleteInput{Name: "osapi0", Destination: "198.51.100.0/24"})
		if want != "not found" {
			s.ErrorContains(err, want, content)
		}
	}
}

// TestNetplanFails: when netplan cannot say where the interface is, the
// route is not written.
func (s *DebianPublicTestSuite) TestNetplanFails() {
	ctrl := gomock.NewController(s.T())
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/sys/class/net/osapi0", 0o755))
	s.Require().NoError(fs.WriteFile("/sys/class/net/osapi0/ifindex", []byte("3"), 0o644))
	s.Require().NoError(fs.WriteFile("/sys/class/net/osapi0/flags", []byte("0x82"), 0o644))
	exec := hostmocks.NewMockExec(ctrl)
	behaviortest.ReadsThroughRunCmd(exec)
	exec.EXPECT().RunCmd(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	np := netplanmocks.NewMockProvider(ctrl)
	np.EXPECT().
		Interface(gomock.Any(), "osapi0").
		Return(netplan.Interface{}, errors.New("netplan broke"))
	p, err := route.ExportNewDebian(host.Host{
		FS: fs, Exec: exec, Platform: "debian",
		Required: map[string]any{netplan.Behavior.Name: np},
	})
	s.Require().NoError(err)

	_, err = route.Create(context.Background(), p,
		route.CreateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "192.0.2.1"})
	s.ErrorContains(err, "netplan broke")
}

// TestUnreadableFile: a routes file that osapi cannot read, such as one it
// has no permission for, is an error, not a file with no routes that a write
// would replace.
func (s *DebianPublicTestSuite) TestUnreadableFile() {
	fs := memfs.New()
	// A directory where the file goes fails the read, as permission denied
	// does, and is not "does not exist".
	s.Require().NoError(fs.MkdirAll(netplan.Dir+"/osapi-osapi0-routes.yaml", 0o755))
	p, err := route.ExportNewDebian(host.Host{FS: fs})
	s.Require().NoError(err)

	_, err = route.Create(context.Background(), p,
		route.CreateInput{Name: "osapi0", Destination: "198.51.100.0/24", Via: "192.0.2.1"})
	s.ErrorContains(err, "read "+netplan.Dir+"/osapi-osapi0-routes.yaml")
	_, err = route.Delete(context.Background(), p,
		route.DeleteInput{Name: "osapi0", Destination: "198.51.100.0/24"})
	s.ErrorContains(err, "read "+netplan.Dir+"/osapi-osapi0-routes.yaml")
}

// TestSpellings: a destination is one network in each of its spellings, so
// a delete finds the route that a different spelling wrote, and a create of
// another spelling is the route that exists.
func (s *DebianPublicTestSuite) TestSpellings() {
	file := netplan.Dir + "/osapi-osapi0-routes.yaml"
	tests := []struct {
		name   string
		stored string
		delete string
	}{
		{
			name:   "a host address in the network",
			stored: "198.51.100.0/24",
			delete: "198.51.100.7/24",
		},
		{name: "a file with a host address", stored: "198.51.100.7/24", delete: "198.51.100.0/24"},
		{name: "IPv6 in upper case", stored: "2001:db8::/32", delete: "2001:DB8:0::/32"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctrl := gomock.NewController(s.T())
			fs := memfs.New()
			s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
			s.Require().NoError(fs.WriteFile(file, []byte("network:\n  version: 2\n  ethernets:\n"+
				"    osapi0:\n      routes:\n        - to: "+tc.stored+"\n          via: 192.0.2.1\n"), 0o600))
			np := netplanmocks.NewMockProvider(ctrl)
			np.EXPECT().Remove(gomock.Any(), file).Return(true, nil)
			p, err := route.ExportNewDebian(host.Host{
				FS: fs, Required: map[string]any{netplan.Behavior.Name: np},
			})
			s.Require().NoError(err)

			result, err := route.Delete(context.Background(), p,
				route.DeleteInput{Name: "osapi0", Destination: tc.delete})
			s.Require().NoError(err)
			s.True(result.Changed)
		})
	}
}

// TestDestinationThatDoesNotParse: a delete of a destination that is not a
// network finds no route and changes nothing.
func (s *DebianPublicTestSuite) TestDestinationThatDoesNotParse() {
	file := netplan.Dir + "/osapi-osapi0-routes.yaml"
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll(netplan.Dir, 0o755))
	s.Require().NoError(fs.WriteFile(file, []byte("network:\n  version: 2\n  ethernets:\n"+
		"    osapi0:\n      routes:\n        - to: 198.51.100.0/24\n          via: 192.0.2.1\n"), 0o600))
	p, err := route.ExportNewDebian(host.Host{FS: fs})
	s.Require().NoError(err)

	result, err := route.Delete(context.Background(), p,
		route.DeleteInput{Name: "osapi0", Destination: "not-a-network"})
	s.Require().NoError(err)
	s.False(result.Changed)
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
