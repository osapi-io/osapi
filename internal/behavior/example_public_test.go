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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type ExamplePublicTestSuite struct {
	suite.Suite
}

func (s *ExamplePublicTestSuite) TestExample() {
	m := behavior.Write("create", set).
		Example("One", getIn{Key: "a"}, result{Key: "a"},
			behavior.After(getIn{Key: "b"}),
			behavior.Wrote("/etc/a", "a = 1\n"),
		).
		Example("Two", getIn{Key: "c"}, result{Key: "c"})

	s.Equal([]behavior.Example{
		{
			Doc: "One", Input: getIn{Key: "a"}, Output: result{Key: "a"},
			Given: []behavior.Given{{Kind: behavior.GivenCall, Input: getIn{Key: "b"}}},
			Effects: []behavior.Effect{
				{Kind: behavior.EffectWrote, Path: "/etc/a", Content: "a = 1\n"},
			},
		},
		{Doc: "Two", Input: getIn{Key: "c"}, Output: result{Key: "c"}},
	}, m.Examples)
}

func (s *ExamplePublicTestSuite) TestFails() {
	m := behavior.Read("get", get).
		Fails("An unmanaged key", getIn{Key: "a"}, "not managed", behavior.After(getIn{Key: "b"}))

	s.Equal([]behavior.Refusal{{
		Example: behavior.Example{
			Doc:   "An unmanaged key",
			Input: getIn{Key: "a"},
			Given: []behavior.Given{{Kind: behavior.GivenCall, Input: getIn{Key: "b"}}},
		},
		Error: "not managed",
	}}, m.Refusals)
}

func (s *ExamplePublicTestSuite) TestHostOptions() {
	tests := []struct {
		name   string
		option behavior.HostOption
		want   behavior.Example
	}{
		{
			name:   "when After names a call",
			option: behavior.After(getIn{Key: "b"}),
			want: behavior.Example{
				Given: []behavior.Given{{Kind: behavior.GivenCall, Input: getIn{Key: "b"}}},
			},
		},
		{
			name:   "when Prints gives a command's output",
			option: behavior.Prints("sysctl -n a", "1"),
			want: behavior.Example{Given: []behavior.Given{
				{Kind: behavior.GivenOutput, Command: "sysctl -n a", Content: "1"},
			}},
		},
		{
			name:   "when PrintsAfter gives a command's output once another has run",
			option: behavior.PrintsAfter("apt-get install tree", "dpkg-query -W tree", "installed"),
			want: behavior.Example{Given: []behavior.Given{
				{
					Kind:    behavior.GivenOutput,
					Command: "dpkg-query -W tree",
					Content: "installed",
					After:   "apt-get install tree",
				},
			}},
		},
		{
			name:   "when Exists puts a file there first",
			option: behavior.Exists("/etc/a", "x"),
			want: behavior.Example{Given: []behavior.Given{
				{Kind: behavior.GivenFile, Path: "/etc/a", Content: "x"},
			}},
		},
		{
			name:   "when Wrote says a file's content after",
			option: behavior.Wrote("/etc/a", "a = 1\n"),
			want: behavior.Example{Effects: []behavior.Effect{
				{Kind: behavior.EffectWrote, Path: "/etc/a", Content: "a = 1\n"},
			}},
		},
		{
			name:   "when Removed says a file is gone",
			option: behavior.Removed("/etc/b"),
			want: behavior.Example{
				Effects: []behavior.Effect{{Kind: behavior.EffectRemoved, Path: "/etc/b"}},
			},
		},
		{
			name:   "when Ran names a command line",
			option: behavior.Ran("sysctl", "-p", "/etc/a"),
			want: behavior.Example{Effects: []behavior.Effect{
				{Kind: behavior.EffectRan, Command: "sysctl -p /etc/a"},
			}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := behavior.Read("get", get).Example("", nil, nil, tc.option)

			s.Equal(tc.want, m.Examples[0])
		})
	}
}

func TestExamplePublicTestSuite(t *testing.T) {
	suite.Run(t, new(ExamplePublicTestSuite))
}
