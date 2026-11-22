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

package dsl_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/dsl"
)

type DSLPublicTestSuite struct {
	suite.Suite
}

func (s *DSLPublicTestSuite) TestRegister() {
	b := dsl.Register(dsl.DSL{Name: "zzb", Permissions: func() []dsl.Declared { return nil }})
	dsl.Register(dsl.DSL{Name: "zza"})
	dsl.Register(dsl.DSL{Name: "zza", Skill: "the same name sorts beside the first"})

	all := dsl.All()
	names := make([]string, 0, len(all))
	for _, d := range all {
		names = append(names, d.Name)
	}
	s.Equal([]string{"zza", "zza", "zzb"}, names, "by name")
	s.True(b.ServesAPI(), "a DSL with permissions serves the API")
	s.False(dsl.All()[0].ServesAPI())
}

func (s *DSLPublicTestSuite) TestCommands() {
	s.Equal([]string{
		"list", "get", "method get", "spec get", "validate", "create", "openapi get",
	}, dsl.Commands())
}

func TestDSLPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DSLPublicTestSuite))
}
