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

package load_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/gohaitest"
	"github.com/osapi-io/osapi/internal/provider/node/load"
)

type LoadPublicTestSuite struct {
	suite.Suite
}

func TestLoadPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LoadPublicTestSuite))
}

func (s *LoadPublicTestSuite) TestGetAverageStats() {
	g := &load.Gohai{Sources: gohaitest.Linux(s.T(), map[string]string{
		"/proc/loadavg": "1.00 0.50 0.25 1/100 1234\n",
	}, nil)}
	got, err := g.GetAverageStats()
	s.Require().NoError(err)
	s.Equal(&load.Result{Load1: 1, Load5: 0.5, Load15: 0.25}, got)

	_, err = (&load.Gohai{Sources: gohaitest.Linux(s.T(), nil, nil)}).GetAverageStats()
	s.ErrorContains(err, "get load")

	s.NotNil(load.New())
}
