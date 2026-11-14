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

package host_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

type dirSettings struct {
	Dir string `mapstructure:"dir" doc:"Where it writes."`
}

type writer struct{ dir string }

type noIn struct{}

type noOut struct{}

func read(
	_ context.Context,
	_ writer,
	_ noIn,
) (noOut, error) {
	return noOut{}, nil
}

type HostPublicTestSuite struct {
	suite.Suite
}

func (s *HostPublicTestSuite) TestSettings() {
	tests := []struct {
		name string
		h    host.Host
		want dirSettings
	}{
		{
			name: "when the settings are the type asked for",
			h:    host.Host{Settings: dirSettings{Dir: "/etc/x"}},
			want: dirSettings{Dir: "/etc/x"},
		},
		{
			name: "when they are another type, the zero value",
			h:    host.Host{Settings: 3},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, host.Settings[dirSettings](tc.h))
		})
	}
}

func (s *HostPublicTestSuite) TestBuild() {
	// writerB has settings; userB requires filesB, which can fail.
	writerB := func(decoded any, buildErr error) *behavior.Behavior {
		b := behavior.Define("x", "Writes", behavior.Node,
			behavior.Config(dirSettings{Dir: "/etc/x"}),
			behavior.Platform("debian", func(h host.Host) (writer, error) {
				return writer{dir: host.Settings[dirSettings](h).Dir}, buildErr
			}),
			behavior.Read("get", read),
		)
		if decoded != nil {
			b.SetSettings(decoded)
		}

		return b
	}
	userB := func(filesErr error) *behavior.Behavior {
		filesB := behavior.Define("files", "Files", behavior.Node,
			behavior.Platform("debian", func(host.Host) (files, error) {
				return files{dir: "/srv"}, filesErr
			}),
		)

		return behavior.Define("user", "Users", behavior.Node,
			behavior.Requires(filesB),
			behavior.Platform("debian", func(h host.Host) (user, error) {
				return user{files: host.Require[files](h, filesB)}, nil
			}),
		)
	}

	tests := []struct {
		name     string
		b        *behavior.Behavior
		platform string
		want     any
		wantErr  string
	}{
		{
			name:     "when the config has been read, the builder gets the decoded settings",
			b:        writerB(dirSettings{Dir: "/srv/x"}, nil),
			platform: "debian",
			want:     writer{dir: "/srv/x"},
		},
		{
			name:     "when it has not, the builder gets the defaults",
			b:        writerB(nil, nil),
			platform: "debian",
			want:     writer{dir: "/etc/x"},
		},
		{
			name:     "when the platform has no provider",
			b:        writerB(nil, nil),
			platform: "darwin",
		},
		{
			name:     "when the builder fails",
			b:        writerB(nil, errors.New("no socket")),
			platform: "debian",
			wantErr:  "x provider on debian: no socket",
		},
		{
			name:     "when a required behavior builds, the builder gets its provider",
			b:        userB(nil),
			platform: "debian",
			want:     user{files: files{dir: "/srv"}},
		},
		{
			name:     "when a required behavior declines the host, the builder gets none",
			b:        userB(behavior.ErrUnsupported),
			platform: "debian",
			want:     user{},
		},
		{
			name:     "when a required behavior fails to build, so does this one",
			b:        userB(errors.New("no socket")),
			platform: "debian",
			wantErr:  "user requires files: files provider on debian: no socket",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := host.Build(tc.b, host.Host{Platform: tc.platform})

			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

// files is a provider another behavior requires.
type files struct{ dir string }

// user is a provider built on files.
type user struct{ files files }

func (s *HostPublicTestSuite) TestRequire() {
	b := behavior.Define("files", "Files")
	h := host.Host{Required: map[string]any{"files": files{dir: "/srv"}}}

	s.Equal(files{dir: "/srv"}, host.Require[files](h, b))
	s.Equal(
		files{},
		host.Require[files](host.Host{}, b),
		"a provider not built reads as the zero value",
	)
}

func TestHostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HostPublicTestSuite))
}
