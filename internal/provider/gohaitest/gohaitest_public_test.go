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

package gohaitest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/gohaitest"
)

type GohaitestPublicTestSuite struct {
	suite.Suite
}

func TestGohaitestPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GohaitestPublicTestSuite))
}

func (s *GohaitestPublicTestSuite) TestLinux() {
	src := gohaitest.Linux(s.T(), map[string]string{"/proc/loadavg": "1 2 3"}, map[string]string{
		"sysctl -n vm.loadavg": "{ 1 2 3 }",
	})
	s.Equal("debian", src.Family)

	b, err := src.FS.ReadFile("/proc/loadavg")
	s.Require().NoError(err)
	s.Equal("1 2 3", string(b))

	out, err := src.Exec.Execute(context.Background(), "sysctl", "-n", "vm.loadavg")
	s.Require().NoError(err)
	s.Equal("{ 1 2 3 }", string(out))

	_, err = src.Exec.Execute(context.Background(), "nope")
	s.ErrorContains(err, "exec nope: not found")
}
