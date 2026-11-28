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

package config_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type TypesPublicTestSuite struct {
	suite.Suite
}

func (s *TypesPublicTestSuite) TestJoinUsers() {
	server := []config.NATSServerUser{
		{Username: "setup", Password: "s", Role: "setup"},
		{Username: "join", Password: "j", Role: "join"},
		{Username: "agent", Password: "a", Role: "agent"},
	}
	tests := []struct {
		name      string
		cfg       func() config.Config
		wantJoin  config.NATSUser
		wantAgent config.NATSUser
	}{
		{
			name: "the embedded server's users with the roles",
			cfg: func() config.Config {
				var c config.Config
				c.NATS.Server.Auth.Users = server

				return c
			},
			wantJoin:  config.NATSUser{Username: "join", Password: "j"},
			wantAgent: config.NATSUser{Username: "agent", Password: "a"},
		},
		{
			name: "controller.join's users first",
			cfg: func() config.Config {
				var c config.Config
				c.NATS.Server.Auth.Users = server
				c.Controller.Join.User = config.NATSUser{Username: "j2", Password: "x"}
				c.Controller.Join.Agent = config.NATSUser{Username: "a2", Password: "y"}

				return c
			},
			wantJoin:  config.NATSUser{Username: "j2", Password: "x"},
			wantAgent: config.NATSUser{Username: "a2", Password: "y"},
		},
		{
			name: "none when nothing names them",
			cfg:  func() config.Config { return config.Config{} },
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			join, agent := tt.cfg().JoinUsers()
			s.Equal(tt.wantJoin, join)
			s.Equal(tt.wantAgent, agent)
		})
	}
}

func (s *TypesPublicTestSuite) TestJoinURL() {
	var c config.Config
	c.NATS.Server.URL = "tls://hub:4222"
	s.Equal("tls://hub:4222", c.JoinURL())
	c.Controller.Join.URL = "nats://edge:4222"
	s.Equal("nats://edge:4222", c.JoinURL())
}

func TestTypesPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TypesPublicTestSuite))
}
