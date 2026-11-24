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

package area_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/area"
)

type AreaPublicTestSuite struct {
	suite.Suite
}

func TestAreaPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AreaPublicTestSuite))
}

func (s *AreaPublicTestSuite) TestValid() {
	tests := []struct {
		name string
		area area.Name
		want bool
	}{
		{name: "when the area is one of All", area: area.Fleet, want: true},
		{name: "when the area is not one of All", area: "kitchen", want: false},
		{name: "when the area is empty", area: "", want: false},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, area.Valid(tt.area))
		})
	}
}

func (s *AreaPublicTestSuite) TestAll() {
	tests := []struct {
		name string
	}{
		{name: "when each area has a label and a name once"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			seen := map[area.Name]bool{}
			for _, a := range area.All {
				s.NotEmpty(a.Label, a.Name)
				s.False(seen[a.Name], a.Name)
				seen[a.Name] = true
			}
		})
	}
}

func (s *AreaPublicTestSuite) TestList() {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "when it lists the areas in order",
			want: "services, software, config, node, logs, networking, security, containers, objectstore, secrets, command, hardware, audit, fleet",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, area.List())
		})
	}
}
