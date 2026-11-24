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

package behavior

import (
	"github.com/osapi-io/osapi/internal/dsl"
)

func init() {
	dsl.Register(dsl.DSL{
		Name:        "behavior",
		Command:     []string{"dsl", "behavior"},
		Skill:       "build-a-behavior",
		Page:        "behavior.md",
		Permissions: declaredPermissions,
		Operations:  servedOperations,
		Without:     map[string]string{},
	})
	dsl.Register(dsl.DSL{
		Name:    "chain",
		Command: []string{"dsl", "chain"},
		Skill:   "build-a-behavior",
		Page:    "chain.md",
		Without: map[string]string{
			dsl.CommandCreate:  "a chain is made by Requires on each behavior, not on its own",
			dsl.CommandMethod:  "a chain has no methods; its behaviors have them",
			dsl.CommandOpenAPI: "a chain serves no route; its behavior at the top does",
		},
	})
}

// declaredPermissions are the permissions of each method of each registered
// behavior that the API serves.
func declaredPermissions() []dsl.Declared {
	var out []dsl.Declared
	for _, b := range All() {
		// Only a behavior served over the API has methods a token can call.
		if !b.API {
			continue
		}
		for _, m := range b.Methods {
			out = append(out, dsl.Declared{
				Permission: b.Permission(m.PermissionVerb()),
				AdminOnly:  m.Admin,
				Where:      "behavior " + b.Name + " " + m.Name,
			})
		}
	}

	return out
}

// servedOperations are the methods of each registered behavior that the API
// serves, by operationId.
func servedOperations() []dsl.Operation {
	var out []dsl.Operation
	for _, b := range All() {
		if !b.API {
			continue
		}
		for _, m := range b.Methods {
			out = append(
				out,
				dsl.Operation{ID: b.OperationID(m), Permission: b.Permission(m.PermissionVerb())},
			)
		}
	}

	return out
}
