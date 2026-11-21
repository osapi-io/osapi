// Copyright (c) 2024 John Dewey

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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/gohaitest"
	"github.com/osapi-io/osapi/internal/provider/node/host"
)

type HostPublicTestSuite struct {
	suite.Suite
}

func TestHostPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HostPublicTestSuite))
}

func (s *HostPublicTestSuite) TestRead() {
	g := &host.Gohai{Sources: gohaitest.Linux(s.T(), map[string]string{
		"/etc/os-release": "ID=ubuntu\nVERSION_ID=\"24.04\"\n",
		"/proc/uptime":    "3600.50 100.00\n",
		"/proc/stat":      "btime 1700000000\n",
	}, map[string]string{
		"hostname -s": "osapi-test\n",
		"hostname":    "osapi-test.invalid\n",
	})}

	name, err := g.GetHostname()
	s.Require().NoError(err)
	s.Equal("osapi-test.invalid", name)

	os, err := g.GetOSInfo()
	s.Require().NoError(err)
	s.Equal(&host.Result{Distribution: "ubuntu", Version: "24.04"}, os)

	up, err := g.GetUptime()
	s.Require().NoError(err)
	s.Equal(3600*time.Second, up)
}

func (s *HostPublicTestSuite) TestEmptyHost() {
	g := &host.Gohai{Sources: gohaitest.Linux(s.T(), nil, nil)}

	_, err := g.GetHostname()
	s.ErrorContains(err, "get hostname")

	_, err = g.GetUptime()
	s.ErrorContains(err, "get uptime")

	os, err := g.GetOSInfo()
	s.Require().NoError(err, "the platform collector reads what it can")
	s.Empty(os.Distribution)
}

func (s *HostPublicTestSuite) TestNew() {
	s.NotNil(host.New())
}
