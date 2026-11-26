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

// Package catalog is the catalog endpoint group: what the UI reads from the
// controller after sign-in. The behaviors with their methods, the areas,
// the permissions and the roles, the permission of each operation, the
// tones of the statuses and the counts, and the host families. The
// controller builds it from its registries on each request, so a page
// shows each behavior that the controller has, and nobody edits the UI.
package catalog

import (
	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Group is the catalog group's definition. The routes, the spec, the CLI
// commands and their pages come from it. osapi dsl endpoint spec get
// lists every option.
var Group = endpoint.Register(endpoint.Define(
	"catalog",
	"The catalog of what the controller serves",
	endpoint.Area(area.Fleet),
	endpoint.Title("Catalog API"),
	endpoint.Tag("catalog", "Catalog",
		"The behaviors, the permissions and the roles that the controller serves, for the UI."),
	endpoint.Uses[Deps](),
	endpoint.Doc[types.Permission]("A permission that osapi knows, as resource:verb."),
	endpoint.Doc[OperationID]("An operationId of an operation that needs a permission."),
	endpoint.Doc[types.MethodKind](
		"What a method does to a host, by the verb of its permission: read, write or action.",
	),
	endpoint.Get[GetInput, Response]("get").
		Handle(Get).
		ID("GetCatalog").
		Describe("Get the catalog").
		Explain("Returns each behavior with its methods, the areas, the permissions, "+
			"the built-in roles, the permission of each operation, the tones of the "+
			"statuses and the counts, and the host families. The UI reads it after "+
			"sign-in. Each built-in role has catalog:read.").
		Returns("The catalog.").
		Command("Get the catalog of the controller",
			"The catalog is what the UI reads after sign-in: each behavior with its\n"+
				"methods, the areas, the permissions, the built-in roles, the permission\n"+
				"of each operation, the tones of the statuses and the counts, and the\n"+
				"host families. The controller builds it from its registries.").
		Notes("The table leaves out the permissions, the tones and the doc of each\n"+
			"behavior, and the permissions of each role. --json prints the whole\n"+
			"catalog, with the methods of each behavior. osapi controller role get\n"+
			"--name <role> shows the permissions of a role.").
		Example("").
		Prints(sampleCatalog).
		Example("--json"),
))
