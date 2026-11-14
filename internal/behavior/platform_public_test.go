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

package behavior_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

type PlatformPublicTestSuite struct {
	suite.Suite
}

func (s *PlatformPublicTestSuite) TestProvider() {
	tests := []struct {
		name     string
		declared string
		platform string
		host     any
		build    func(host.Host) (fakeProvider, error)
		want     any
		wantErr  string
	}{
		{
			name:     "when only linux is declared, a distribution without its own provider gets it",
			declared: "linux",
			platform: "rocky",
			host:     host.Host{Hostname: "web-01"},
			build: func(h host.Host) (fakeProvider, error) {
				return fakeProvider{prefix: h.Hostname}, nil
			},
			want: fakeProvider{prefix: "web-01"},
		},
		{
			name:     "when only linux is declared, macOS does not get it",
			declared: "linux",
			platform: "darwin",
			host:     host.Host{},
			build: func(host.Host) (fakeProvider, error) {
				return fakeProvider{prefix: "linux"}, nil
			},
		},
		{
			name:     "when the platform is declared, its builder gets the host",
			platform: "debian",
			host:     host.Host{Hostname: "web-01"},
			build: func(h host.Host) (fakeProvider, error) {
				return fakeProvider{prefix: h.Hostname}, nil
			},
			want: fakeProvider{prefix: "web-01"},
		},
		{
			name:     "when the platform is not declared",
			platform: "darwin",
			host:     host.Host{},
			build: func(host.Host) (fakeProvider, error) {
				return fakeProvider{}, nil
			},
		},
		{
			name:     "when the builder fails",
			platform: "debian",
			host:     host.Host{},
			build: func(host.Host) (fakeProvider, error) {
				return fakeProvider{}, errors.New("no socket")
			},
			wantErr: "sysctl provider on debian: no socket",
		},
		{
			name:     "when the host is not the type the builder takes",
			platform: "debian",
			host:     "not a host",
			build: func(host.Host) (fakeProvider, error) {
				return fakeProvider{}, nil
			},
			wantErr: "the builder takes a host.Host, not a string",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			declared := tc.declared
			if declared == "" {
				declared = "debian"
			}
			b := behavior.Define("sysctl", "Kernel parameters", behavior.Node,
				behavior.Platform(declared, tc.build),
			)

			got, err := b.Provider(tc.platform, tc.host)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

// TestProviderPrefersOwnPlatform: a distribution with a provider of its own
// gets it, not the linux one.
func (s *PlatformPublicTestSuite) TestProviderPrefersOwnPlatform() {
	tests := []struct {
		name     string
		platform string
		want     fakeProvider
	}{
		{name: "when the host is debian", platform: "debian", want: fakeProvider{prefix: "debian"}},
		{
			name:     "when the host is another distribution",
			platform: "fedora",
			want:     fakeProvider{prefix: "linux"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			b := behavior.Define("sysctl", "Kernel parameters", behavior.Node,
				behavior.Platform("linux", func(host.Host) (fakeProvider, error) {
					return fakeProvider{prefix: "linux"}, nil
				}),
				behavior.Platform("debian", func(host.Host) (fakeProvider, error) {
					return fakeProvider{prefix: "debian"}, nil
				}),
			)

			got, err := b.Provider(tc.platform, host.Host{})

			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *PlatformPublicTestSuite) TestSudo() {
	b := behavior.Define("sysctl", "Kernel parameters", behavior.Node,
		behavior.Sudo("sysctl"), behavior.Sudo("modprobe", "tee"),
	)

	s.Equal([]string{"sysctl", "modprobe", "tee"}, b.Sudo)
}

func (s *PlatformPublicTestSuite) TestRuns() {
	b := behavior.Define("neighbors", "Neighbor table", behavior.Node,
		behavior.Runs("ip"), behavior.Runs("ss", "awk"),
	)

	s.Equal([]string{"ip", "ss", "awk"}, b.Runs)
}

func TestPlatformPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PlatformPublicTestSuite))
}
