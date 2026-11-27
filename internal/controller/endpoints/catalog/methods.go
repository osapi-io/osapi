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
	"sync"

	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/cli"
)

// chains runs the examples of a behavior, for the call tree of the catalog.
var chains Runs = behaviortest.Chains

var (
	builtMu  sync.Mutex
	builtFor []*behavior.Behavior
	built    []Behavior
)

// registered is the behaviors of the catalog. Their examples run only when
// the registry differs from the one of the last build, so a request with
// the same registry runs none, and a behavior registered later is in the
// next catalog.
func registered() []Behavior {
	all := behavior.All()
	builtMu.Lock()
	defer builtMu.Unlock()
	if built == nil || !slices.Equal(all, builtFor) {
		built = Behaviors(all, chains)
		builtFor = all
	}

	return built
}

// Get is the catalog, from the registries of the controller when it runs.
// Its behaviors are built again only when the registry changed.
func Get(
	_ context.Context,
	_ Deps,
	_ GetInput,
) (Response, error) {
	out := Response{
		Behaviors:    registered(),
		Areas:        []Area{},
		Permissions:  authtoken.AllPermissions(),
		Roles:        []Role{},
		Operations:   []Operation{},
		StatusTones:  cli.StatusTones,
		CountTones:   cli.CountTones,
		HostFamilies: []string{},
	}
	for _, a := range area.All {
		out.Areas = append(out.Areas, Area{Name: string(a.Name), Label: a.Label})
	}
	for _, r := range authtoken.Roles() {
		perms, _ := authtoken.RolePermissions(r)
		out.Roles = append(out.Roles, Role{
			Name:        r,
			Description: authtoken.RoleDescription(r),
			Permissions: perms,
		})
	}
	ops := authtoken.Operations()
	for _, id := range slices.Sorted(maps.Keys(ops)) {
		out.Operations = append(out.Operations, Operation{
			ID:         OperationID(id),
			Permission: ops[id],
		})
	}
	for _, f := range behavior.HostFamilies {
		out.HostFamilies = append(out.HostFamilies, string(f))
	}
	out.Targets = []Target{
		{
			Value: types.TargetAll,
			Label: "All agents",
			Doc:   "Each agent answers, one row for each host.",
		},
		{Value: types.TargetAny, Label: "Any agent", Doc: "One agent, any that is free, answers."},
	}
	out.Controller = controllerMethods(endpoint.All(), ops)
	out.RenderTags = clicommand.RenderTags()
	out.Outcomes = []string{
		string(types.RowStatusOK), "changed", string(types.RowStatusSkipped),
		string(types.RowStatusTimeout), string(types.RowStatusFailed),
	}

	return out, nil
}

// controllerMethods are the methods of the endpoint groups of the Object
// Store area, in the order of the groups: the calls of the controller that
// take no target.
func controllerMethods(
	groups []*endpoint.Group,
	ops map[string]types.Permission,
) []ControllerMethod {
	out := []ControllerMethod{}
	for _, g := range groups {
		if g.Area != area.ObjectStore {
			continue
		}
		for _, m := range g.Methods {
			id := m.OperationID(g)
			out = append(out, ControllerMethod{
				Group:      g.Name,
				Method:     m.Name,
				Doc:        m.CommandShort(),
				Operation:  OperationID(id),
				Permission: ops[id],
			})
		}
	}

	return out
}
