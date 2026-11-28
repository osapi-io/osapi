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

package packages_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/packages"
)

type ReaderPublicTestSuite struct {
	suite.Suite
}

// TestWritesAreUnsupported covers the reader's writes, which no host calls:
// install and remove run on the Debian family only.
func (s *ReaderPublicTestSuite) TestWritesAreUnsupported() {
	p, err := packages.ExportNewReader(host.Host{Platform: behavior.PlatformLinux})
	s.Require().NoError(err)

	_, err = p.Install(context.Background(), "curl", "")
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(context.Background(), "curl")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

func TestReaderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReaderPublicTestSuite))
}
