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

package endpoint_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

type SettingsPublicTestSuite struct {
	suite.Suite
}

func TestSettingsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SettingsPublicTestSuite))
}

type limits struct {
	Max int `mapstructure:"max" doc:"The most it lists."`
}

func (s *SettingsPublicTestSuite) TestSettings() {
	s.Run("a group's settings are its defaults until the config sets them", func() {
		g := endpoint.Define("node", "The nodes", endpoint.Config(limits{Max: 20}))
		s.Equal("controller.endpoints.node", g.SettingsKey())
		s.Equal("limits", g.SettingsType().Name())
		s.Equal(20, endpoint.Settings[limits](g).Max)

		g.SetSettings(limits{Max: 50})
		s.Equal(50, endpoint.Settings[limits](g).Max)
	})
	s.Run("a group with no settings has none", func() {
		g := endpoint.Define("node", "The nodes")
		s.Nil(g.SettingsType())
		s.Nil(g.CurrentSettings())
		s.Zero(endpoint.Settings[limits](g))
	})
}
