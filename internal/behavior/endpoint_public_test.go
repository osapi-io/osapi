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
package behavior_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

type EndpointPublicTestSuite struct {
	suite.Suite
}

// TestRegisterGroup records a behavior's endpoint group, which its spec comes
// from, and has none for a behavior whose gen package was not imported.
func (s *EndpointPublicTestSuite) TestRegisterGroup() {
	b := behavior.Define("registered", "x")
	s.Nil(behavior.GroupOf(b))

	g := endpoint.Define("registered", "x")
	behavior.RegisterGroup("registered", g)
	s.Same(g, behavior.GroupOf(b))
}

func TestEndpointPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(EndpointPublicTestSuite))
}
