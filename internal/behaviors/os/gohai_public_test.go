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

package os_test

import (
	"errors"
	"testing"

	"github.com/osapi-io/gohai/pkg/gohai/collectors/platform"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behaviors/os"
)

type GohaiPublicTestSuite struct {
	suite.Suite
}

// TestRelease: a platform result is its release, and a collector error or
// another result is an error, not a dereference of a nil result.
func (s *GohaiPublicTestSuite) TestRelease() {
	failed := errors.New("collector failed")
	tests := []struct {
		name         string
		got          any
		err          error
		validateFunc func(os.Release, error)
	}{
		{
			name: "a platform result",
			got:  &platform.Info{Name: "ubuntu", Version: "24.04"},
			validateFunc: func(r os.Release, err error) {
				s.Require().NoError(err)
				s.Equal(os.Release{Distribution: "ubuntu", Version: "24.04"}, r)
			},
		},
		{
			name: "a collector error",
			err:  failed,
			validateFunc: func(_ os.Release, err error) {
				s.ErrorIs(err, failed)
				s.ErrorContains(err, "read the platform")
			},
		},
		{
			name: "no result",
			validateFunc: func(_ os.Release, err error) {
				s.ErrorContains(err, "unexpected result")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(os.ExportRelease(tc.got, tc.err))
		})
	}
}

func TestGohaiPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(GohaiPublicTestSuite))
}
