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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/behavior"
)

type RegistryPublicTestSuite struct {
	suite.Suite
}

func (s *RegistryPublicTestSuite) SetupTest() {
	behavior.Reset()
}

func (s *RegistryPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *RegistryPublicTestSuite) TestRegister() {
	tests := []struct {
		name          string
		register      func() *behavior.Behavior
		wantDuplicate bool
	}{
		{
			name: "when registered once it returns the behavior",
			register: func() *behavior.Behavior {
				return behavior.Register(behavior.Define("sysctl", "x", behavior.Node))
			},
		},
		{
			name: "when the same value is registered twice it is not a clash",
			register: func() *behavior.Behavior {
				b := behavior.Register(behavior.Define("sysctl", "x", behavior.Node))

				return behavior.Register(b)
			},
		},
		{
			name: "when another behavior takes the name it is a clash",
			register: func() *behavior.Behavior {
				behavior.Register(behavior.Define("sysctl", "x", behavior.Node))

				return behavior.Register(behavior.Define("sysctl", "y", behavior.Node))
			},
			wantDuplicate: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			behavior.Reset()

			b := tc.register()

			assert.Same(s.T(), b, behavior.Lookup("sysctl"))
			duplicate := false
			for _, p := range behavior.Validate() {
				if p.Rule == behavior.RuleDuplicate {
					duplicate = true
				}
			}
			assert.Equal(s.T(), tc.wantDuplicate, duplicate)
		})
	}
}

func (s *RegistryPublicTestSuite) TestAll() {
	behavior.Register(behavior.Define("sysctl", "x", behavior.Node))
	behavior.Register(behavior.Define("dns", "x", behavior.Node, behavior.Under("network")))
	behavior.Register(behavior.Define("apt", "x", behavior.Node))
	behavior.Register(behavior.Define("route", "x", behavior.Node, behavior.Under("network")))

	all := behavior.All()
	names := make([]string, 0, len(all))
	for _, b := range all {
		names = append(names, b.Group+"/"+b.Name)
	}

	assert.Equal(s.T(), []string{"/apt", "/sysctl", "network/dns", "network/route"}, names)
}

func (s *RegistryPublicTestSuite) TestLookup() {
	b := behavior.Register(behavior.Define("sysctl", "x", behavior.Node))

	tests := []struct {
		name string
		want *behavior.Behavior
	}{
		{name: "sysctl", want: b},
		{name: "nope", want: nil},
	}

	for _, tc := range tests {
		s.Run("when looking up "+tc.name, func() {
			assert.Equal(s.T(), tc.want, behavior.Lookup(tc.name))
		})
	}
}

func (s *RegistryPublicTestSuite) TestReset() {
	behavior.Register(behavior.Define("sysctl", "x", behavior.Node))

	behavior.Reset()

	assert.Empty(s.T(), behavior.All())
}

func TestRegistryPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RegistryPublicTestSuite))
}
