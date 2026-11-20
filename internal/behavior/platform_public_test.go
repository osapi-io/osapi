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

func (s *PlatformPublicTestSuite) TestOn() {
	linux := func(host.Host) (fakeProvider, error) { return fakeProvider{}, nil }
	tests := []struct {
		name       string
		on         []string
		platforms  []string
		host       string
		wantRuns   bool
		serves     string
		wantServes bool
	}{
		{
			name: "when the method names no platform", platforms: []string{"linux"},
			host: "fedora", wantRuns: true, serves: "linux", wantServes: true,
		},
		{
			name: "when the host is the platform On names", on: []string{"debian"},
			platforms: []string{"linux"}, host: "debian", wantRuns: true,
			serves: "linux", wantServes: true,
		},
		{
			name: "when the host is another distribution", on: []string{"debian"},
			platforms: []string{"linux"}, host: "fedora", wantRuns: false,
			serves: "darwin", wantServes: false,
		},
		{
			name: "when On names linux and the host is a distribution", on: []string{"linux"},
			platforms: []string{"linux", "darwin"}, host: "fedora", wantRuns: true,
			serves: "darwin", wantServes: false,
		},
		{
			name: "when On names linux and the host is macOS", on: []string{"linux"},
			platforms: []string{"linux"}, host: "darwin", wantRuns: false,
			serves: "linux", wantServes: true,
		},
		{
			name: "when On names a platform with no provider", on: []string{"darwin"},
			platforms: []string{"linux"}, host: "darwin", wantRuns: true,
			serves: "linux", wantServes: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			opts := []behavior.Option{behavior.Node, behavior.Read("get", get).On(tc.on...)}
			for _, p := range tc.platforms {
				opts = append(opts, behavior.Platform(p, linux))
			}
			b := behavior.Define("sysctl", "Kernel parameters", opts...)

			s.Equal(tc.wantRuns, b.Methods[0].RunsOn(tc.host))
			s.Equal(tc.wantServes, b.Serves(b.Methods[0], tc.serves))
		})
	}
}

func (s *PlatformPublicTestSuite) TestPrograms() {
	b := behavior.Define("hostname", "Hostname", behavior.Node,
		behavior.Sudo("tee"), behavior.Runs("sysctl"),
		behavior.Read("get", get).Runs("cat"),
		behavior.Write("update", get).On("debian").Sudo("hostnamectl", "tee"),
	)
	tests := []struct {
		name     string
		platform string
		wantSudo []string
		wantRuns []string
	}{
		{
			name: "when the method runs on the platform", platform: "debian",
			wantSudo: []string{"tee", "hostnamectl"}, wantRuns: []string{"sysctl", "cat"},
		},
		{
			name: "when the method does not run on the platform", platform: "darwin",
			wantSudo: []string{"tee"}, wantRuns: []string{"sysctl", "cat"},
		},
		{
			name: "when the platform is empty", platform: "",
			wantSudo: []string{"tee", "hostnamectl"}, wantRuns: []string{"sysctl", "cat"},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			sudo, runs := b.Programs(tc.platform)
			s.Equal(tc.wantSudo, sudo)
			s.Equal(tc.wantRuns, runs)
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
