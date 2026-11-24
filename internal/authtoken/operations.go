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

package authtoken

import (
	"fmt"
	"slices"

	"github.com/osapi-io/osapi/internal/dsl"

	"gopkg.in/yaml.v3"
)

// methods are the keys of a path item that are operations.
var methods = []string{"get", "put", "post", "delete", "patch", "head", "options"}

// Operations is the permission of each operation that the API serves, by
// its operationId, such as GetNodeStatus. The endpoint groups and the
// behaviors give theirs from their definitions. The full spec is output, so
// nothing here reads it. A public method needs no permission and is not in
// it. The catalog endpoint serves it, so the UI asks it before a call.
func Operations() map[string]Permission {
	out := map[string]Permission{}
	for _, d := range dsl.All() {
		if !d.ServesAPI() {
			continue
		}
		for _, op := range d.Operations() {
			out[op.ID] = op.Permission
		}
	}

	return out
}

// OperationPermissions is the permission of each operation of spec, such as
// the full spec, by its operationId. An operation with no scope, such as a
// health probe, needs no permission and is not in it. A test holds
// Operations to the full spec with it.
func OperationPermissions(
	spec []byte,
) (map[string]Permission, error) {
	var doc struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		return nil, fmt.Errorf("read the spec: %w", err)
	}
	out := map[string]Permission{}
	for _, item := range doc.Paths {
		for method, node := range item {
			// A path item also has keys that are not operations, such as
			// parameters.
			if !slices.Contains(methods, method) {
				continue
			}
			var op struct {
				OperationID string                `yaml:"operationId"`
				Security    []map[string][]string `yaml:"security"`
			}
			if err := node.Decode(&op); err != nil {
				return nil, fmt.Errorf("read the spec: %w", err)
			}
			for _, req := range op.Security {
				for _, scopes := range req {
					if len(scopes) > 0 && op.OperationID != "" {
						out[op.OperationID] = Permission(scopes[0])
					}
				}
			}
		}
	}

	return out, nil
}
