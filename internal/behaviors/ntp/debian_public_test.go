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

package ntp_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/ntp"
)

type DebianPublicTestSuite struct {
	suite.Suite
}

func (s *DebianPublicTestSuite) TestParseTracking() {
	tests := []struct {
		name   string
		output string
		want   ntp.Status
	}{
		{
			name: "when a line has no colon or its fields are short, it is skipped",
			output: "no colon here\n" +
				"System time     : 0.1\n" +
				"Reference ID    : 7F7F0101\n" +
				"Stratum         : not a number\n",
			want: ntp.Status{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, ntp.ParseTracking(tc.output))
		})
	}
}

func (s *DebianPublicTestSuite) TestParseSources() {
	tests := []struct {
		name   string
		output string
		want   []string
	}{
		{
			name:   "when a line has fewer than three fields or no name, it is skipped",
			output: "^,*\n\n^,*, ,3\n^,*,time.cloudflare.com,3\n",
			want:   []string{"time.cloudflare.com"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, ntp.ParseSources(tc.output))
		})
	}
}

func (s *DebianPublicTestSuite) TestNewDebian() {
	// Without the managedfile provider that the agent builds first, the
	// methods report skipped.
	_, err := ntp.NewDebian(host.Host{})
	s.ErrorIs(err, behavior.ErrUnsupported)
	// A container's clock is its host's, so it has nothing to set.
	_, err = ntp.NewDebian(host.Host{Container: true})
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func TestDebianPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DebianPublicTestSuite))
}
