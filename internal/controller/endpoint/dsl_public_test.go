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

package endpoint_test

import (
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/dsl"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// TestDSL gives the permission of each method of each registered group, and
// the operation of each method that needs one: a public method needs none.
func (s *EndpointPublicTestSuite) TestDSL() {
	endpoint.Register(endpoint.Define("zzdsl", "A test group",
		endpoint.Get[noInput, Node]("list"),
		endpoint.Post[noInput, Node]("drain").AdminOnly(),
		endpoint.Get[noInput, Node]("probe").Sub("probe").Public(),
		endpoint.SubDoc("probe", "Probes"),
	))
	var d dsl.DSL
	for _, candidate := range dsl.All() {
		if candidate.Name == "endpoint" {
			d = candidate
		}
	}
	s.Require().True(d.ServesAPI())
	s.Equal([]string{"dsl", "endpoint"}, d.Command)

	var perms []dsl.Declared
	for _, p := range d.Permissions() {
		if p.Where == "endpoint zzdsl list" || p.Where == "endpoint zzdsl drain" {
			perms = append(perms, p)
		}
	}
	s.Equal([]dsl.Declared{
		{Permission: types.Permission("zzdsl:read"), Where: "endpoint zzdsl list"},
		{
			Permission: types.Permission("zzdsl:write"),
			AdminOnly:  true,
			Where:      "endpoint zzdsl drain",
		},
	}, perms)

	ids := map[string]types.Permission{}
	for _, op := range d.Operations() {
		ids[op.ID] = op.Permission
	}
	s.Contains(ids, endpoint.Get[noInput, Node]("list").OperationID(endpoint.Lookup("zzdsl")))
	s.NotContains(ids, "GetZzdslProbe", "a public method needs no permission")
}
