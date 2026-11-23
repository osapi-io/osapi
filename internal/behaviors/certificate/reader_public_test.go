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

package certificate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviors/certificate"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"
)

type ReaderPublicTestSuite struct {
	suite.Suite
}

// TestWrites: on macOS the reader changes nothing. The definition limits the
// writes with On, so this is the reader's own answer.
func (s *ReaderPublicTestSuite) TestWrites() {
	p, err := certificate.NewReader(host.Host{})
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = p.Deploy(ctx, "internal-ca", "internal-ca.pem", true)
	s.ErrorIs(err, behavior.ErrUnsupported)
	_, err = p.Remove(ctx, "internal-ca")
	s.ErrorIs(err, behavior.ErrUnsupported)
}

// TestReadOnly: an agent with no file state store has no managedfile, and
// SUSE and Alpine run a program that the linux provider does not declare. On
// each, the writer is the reader, and its writes are unsupported.
func (s *ReaderPublicTestSuite) TestReadOnly() {
	for _, h := range []host.Host{
		{Platform: behavior.PlatformDebian},
		{Platform: "suse", Required: map[string]any{"managedfile": managedfile.NewDebian(nil, nil, "", nil, "")}},
		{Platform: "alpine", Required: map[string]any{"managedfile": managedfile.NewDebian(nil, nil, "", nil, "")}},
	} {
		p, err := certificate.NewWriter(h)
		s.Require().NoError(err)

		_, err = p.Remove(context.Background(), "internal-ca")
		s.ErrorIs(err, behavior.ErrUnsupported, h.Platform)
	}
}

func TestReaderPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReaderPublicTestSuite))
}
