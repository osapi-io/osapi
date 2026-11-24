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

package catalog_test

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
	"github.com/osapi-io/osapi/internal/controller/endpoints/catalog"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

// TestGet: the catalog is what the registries have when it runs.
func (s *MethodsPublicTestSuite) TestGet() {
	got, err := catalog.Get(context.Background(), catalog.Deps{}, catalog.GetInput{})
	s.Require().NoError(err)

	var served int
	for _, b := range behavior.All() {
		if b.API || b.Internal {
			served++
		}
	}
	s.Len(got.Behaviors, served, "each served and each internal behavior")
	s.Require().Len(got.Areas, len(area.All))
	s.Equal(string(area.All[0].Name), got.Areas[0].Name, "the areas, in order")
	s.Equal(area.All[0].Label, got.Areas[0].Label)
	s.Equal(authtoken.AllPermissions(), got.Permissions)
	s.Contains(got.Permissions, types.PermCatalogRead)

	s.Require().Len(got.Roles, len(authtoken.Roles()))
	for i, r := range authtoken.Roles() {
		perms, _ := authtoken.RolePermissions(r)
		s.Equal(string(r), got.Roles[i].Name, "the roles, in order")
		s.Equal(authtoken.RoleDescription(r), got.Roles[i].Description)
		s.Equal(perms, got.Roles[i].Permissions)
		s.Contains(got.Roles[i].Permissions, types.PermCatalogRead,
			"each built-in role reads the catalog, so each can use the UI")
	}

	ops := authtoken.Operations()
	s.Len(got.Operations, len(ops))
	for i, op := range got.Operations {
		s.Equal(ops[string(op.ID)], op.Permission, "the permission of %s", op.ID)
		if i > 0 {
			s.Less(string(got.Operations[i-1].ID), string(op.ID), "in the order of the IDs")
		}
	}
	s.Contains(got.Operations, catalog.CatalogOperation{
		ID: "GetCatalog", Permission: types.PermCatalogRead,
	})

	s.Equal(cli.StatusTones, got.StatusTones)
	s.Equal(cli.CountTones, got.CountTones)
	s.Require().Len(got.HostFamilies, len(behavior.HostFamilies))
	s.Equal(string(behavior.HostFamilies[0]), got.HostFamilies[0])
}

// TestSpecEnums: the spec's Permission is each permission, and its
// OperationID each operation that needs one, so a client checks a value
// when it compiles.
func (s *MethodsPublicTestSuite) TestSpecEnums() {
	doc, err := endpoint.OpenAPI(catalog.Group)
	s.Require().NoError(err)

	enum := func(name string) []string {
		values := doc.Components.Schemas[name].Value.Enum
		out := make([]string, 0, len(values))
		for _, v := range values {
			out = append(out, v.(string))
		}

		return out
	}
	s.Equal(authtoken.Strings(authtoken.AllPermissions()), enum("Permission"))
	ids := enum("OperationID")
	s.Len(ids, len(authtoken.Operations()))
	s.Contains(ids, "GetCatalog")
	s.True(slices.IsSorted(ids))
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
