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

package endpoint

import (
	"github.com/osapi-io/osapi/internal/dsl"
)

func init() {
	dsl.Register(dsl.DSL{
		Name:        "endpoint",
		Command:     []string{"controller", "endpoint"},
		Skill:       "build-an-endpoint",
		Page:        "endpoint.md",
		Permissions: declaredPermissions,
		Operations:  servedOperations,
		Without:     map[string]string{},
	})
}

// declaredPermissions are the permissions of each method of each registered
// group that needs one.
func declaredPermissions() []dsl.Declared {
	var out []dsl.Declared
	for _, g := range All() {
		for _, d := range g.Declared() {
			out = append(out, dsl.Declared{
				Permission: d.Permission,
				AdminOnly:  d.AdminOnly,
				Named:      d.Named,
				Where:      "endpoint " + d.Where,
			})
		}
	}

	return out
}

// servedOperations are the methods of each registered group that need a
// permission, by operationId.
func servedOperations() []dsl.Operation {
	var out []dsl.Operation
	for _, g := range All() {
		for _, m := range g.Methods {
			if perm := g.Permission(m); perm != "" {
				out = append(out, dsl.Operation{ID: m.OperationID(g), Permission: perm})
			}
		}
	}

	return out
}
