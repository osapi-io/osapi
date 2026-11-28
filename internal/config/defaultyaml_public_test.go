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
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"

	"github.com/avfs/avfs"
)

type DefaultYAMLPublicTestSuite struct {
	suite.Suite
}

// TestDefaultYAML is each key at its default, which loads back as a valid
// config: the file that osapi config default get prints runs as it is.
func (s *DefaultYAMLPublicTestSuite) TestDefaultYAML() {
	out := config.DefaultYAML()
	s.Contains(out, "controller:\n  client:\n")
	s.Contains(out, "    port: 8080\n")
	s.Contains(out, "# url: (where a person reaches it; empty shows the bind address)",
		"a choice of the operator is commented out, with why")
	s.Contains(out, "back_off:\n      - 30s\n", "a list is a block under its key")

	path := filepath.Join(s.T().TempDir(), "osapi.yaml")
	s.Require().NoError(avfs.WriteFile(testFS, path, []byte(out), 0o600))
	var c config.Config
	s.Require().NoError(config.Load(viper.New(), path, &c))
	s.Equal(8080, c.Controller.API.Port)
	s.True(strings.HasPrefix(out, "# The config of osapi"))
}

func TestDefaultYAMLPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultYAMLPublicTestSuite))
}
