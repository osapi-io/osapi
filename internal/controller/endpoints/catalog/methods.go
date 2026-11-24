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

package catalog

import (
	"context"
	"maps"
	"slices"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/cli"
)

// Get is the catalog, from the registries of the controller: each request
// builds it, so it has each behavior that is registered when it runs.
func Get(
	_ context.Context,
	_ Deps,
	_ GetInput,
) (CatalogResponse, error) {
	out := CatalogResponse{
		Behaviors:    Behaviors(behavior.All(), behaviortest.Chains),
		Areas:        []CatalogArea{},
		Permissions:  authtoken.AllPermissions(),
		Roles:        []CatalogRole{},
		Operations:   []CatalogOperation{},
		StatusTones:  cli.StatusTones,
		CountTones:   cli.CountTones,
		HostFamilies: []string{},
	}
	for _, a := range area.All {
		out.Areas = append(out.Areas, CatalogArea{Name: string(a.Name), Label: a.Label})
	}
	for _, r := range authtoken.Roles() {
		perms, _ := authtoken.RolePermissions(r)
		out.Roles = append(out.Roles, CatalogRole{
			Name:        string(r),
			Description: authtoken.RoleDescription(r),
			Permissions: perms,
		})
	}
	ops := authtoken.Operations()
	for _, id := range slices.Sorted(maps.Keys(ops)) {
		out.Operations = append(out.Operations, CatalogOperation{
			ID:         OperationID(id),
			Permission: ops[id],
		})
	}
	for _, f := range behavior.HostFamilies {
		out.HostFamilies = append(out.HostFamilies, string(f))
	}

	return out, nil
}
