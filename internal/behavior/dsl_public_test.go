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
	"github.com/osapi-io/osapi/internal/dsl"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type DSLPublicTestSuite struct {
	suite.Suite
}

func (s *DSLPublicTestSuite) SetupTest() {
	behavior.Reset()
}

func (s *DSLPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

// lookup is the registered DSL with the name.
func (s *DSLPublicTestSuite) lookup(
	name string,
) dsl.DSL {
	for _, d := range dsl.All() {
		if d.Name == name {
			return d
		}
	}
	s.FailNow("no DSL named " + name)

	return dsl.DSL{}
}

// TestBehavior gives the permissions and the operations of each behavior
// that the API serves, and none of an internal one.
func (s *DSLPublicTestSuite) TestBehavior() {
	behavior.Register(behavior.Define("zzserved", "Served", behavior.Node, behavior.API,
		behavior.Read("get", get),
		behavior.Write("update", set).AdminOnly(),
	))
	behavior.Register(behavior.Define("zzinner", "Inner", behavior.Node, behavior.Internal,
		behavior.Read("get", get),
	))

	d := s.lookup("behavior")
	s.True(d.ServesAPI())
	s.Equal([]string{"dsl", "behavior"}, d.Command)
	s.Equal([]dsl.Declared{
		{Permission: client.Permission("zzserved:read"), Where: "behavior zzserved get"},
		{
			Permission: client.Permission("zzserved:write"),
			AdminOnly:  true,
			Where:      "behavior zzserved update",
		},
	}, d.Permissions())
	ops := d.Operations()
	s.Require().Len(ops, 2)
	s.Equal(client.Permission("zzserved:write"), ops[1].Permission)
	s.NotEmpty(ops[0].ID)
}

// TestChain serves no API, and says why it has no create, method get or
// openapi get.
func (s *DSLPublicTestSuite) TestChain() {
	d := s.lookup("chain")
	s.False(d.ServesAPI())
	s.Equal([]string{"dsl", "chain"}, d.Command)
	s.Len(d.Without, 3)
	for _, c := range []string{dsl.CommandCreate, dsl.CommandMethod, dsl.CommandOpenAPI} {
		s.NotEmpty(d.Without[c])
	}
}

func TestDSLPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DSLPublicTestSuite))
}
