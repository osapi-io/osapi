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
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// OperationsFile is where the UI keeps the permission of each operation,
// from the root of the repository.
const OperationsFile = "ui/src/lib/operations.gen.ts"

// methods are the keys of a path item that are operations.
var methods = []string{"get", "put", "post", "delete", "patch", "head", "options"}

// HandWrittenSpecs are the specs of the hand-written domains under
// internal/controller/api, from the root of the repository. Such a domain is
// spec first: its spec is its only definition. Each goes when its domain
// becomes a behavior.
const HandWrittenSpecs = "internal/controller/api/node/*/gen/api.yaml"

// Operations is the permission of each operation that the API serves, by
// the name that the UI's generated client gives its function: the
// operationId with a small first letter, such as getNodeStatus. The
// endpoint groups and the behaviors give theirs from their definitions, and
// the hand-written domains from their own specs under root. The joined spec
// is output, so nothing here reads it. A name that two sources give
// different permissions is an error.
func Operations(
	root string,
) (map[string]string, error) {
	out := RegistryOperations()
	paths, err := filepath.Glob(filepath.Join(root, HandWrittenSpecs))
	if err != nil {
		return nil, fmt.Errorf("find the hand-written specs: %w", err)
	}
	for _, path := range paths {
		spec, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		ops, err := OperationPermissions(spec)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		for name, perm := range ops {
			if have, ok := out[name]; ok && have != perm {
				return nil, fmt.Errorf(
					"%s: operation %s needs %s, and a definition gives it %s",
					path, name, perm, have)
			}
			out[name] = perm
		}
	}

	return out, nil
}

// RegistryOperations is the permission of each operation of the registered
// endpoint groups and of each behavior that the API serves, from their
// definitions. A public method needs no permission and is not in it.
func RegistryOperations() map[string]string {
	out := map[string]string{}
	for _, g := range endpoint.All() {
		for _, m := range g.Methods {
			if perm := g.Permission(m); perm != "" {
				out[uiName(m.OperationID(g))] = string(perm)
			}
		}
	}
	for _, b := range behavior.All() {
		if !b.API {
			continue
		}
		for _, m := range b.Methods {
			out[uiName(b.OperationID(m))] = string(b.Permission(m.Verb))
		}
	}

	return out
}

// uiName is id as the UI's generated client names its function: with a
// small first letter.
func uiName(
	id string,
) string {
	return strings.ToLower(id[:1]) + id[1:]
}

// OperationPermissions is the permission of each operation of spec, the
// spec of one hand-written domain, by the name that the UI's generated
// client gives its function. An operation with no scope, such as a health
// probe, needs no permission and is not in it.
func OperationPermissions(
	spec []byte,
) (map[string]string, error) {
	var doc struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		return nil, fmt.Errorf("read the spec: %w", err)
	}
	out := map[string]string{}
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
						out[uiName(op.OperationID)] = scopes[0]
					}
				}
			}
		}
	}

	return out, nil
}

// OperationsTypeScript is ops as the UI's OPERATION_PERMISSIONS: the
// permission that each operation needs. A block of the UI names the
// operation that it calls, and reads its permission from here, so a block
// cannot check another permission than its operation's.
func OperationsTypeScript(
	ops map[string]string,
) string {
	var sb strings.Builder
	sb.WriteString(
		"// Code generated by go generate ./internal/authtoken/gen/operations. DO NOT EDIT.\n",
	)
	sb.WriteString(
		"// From the endpoint groups, the behaviors and the hand-written domains' specs.\n\n",
	)
	sb.WriteString("import type { Permission } from \"./permissions.gen\";\n\n")
	sb.WriteString("export const OPERATION_PERMISSIONS = {\n")
	names := make([]string, 0, len(ops))
	for name := range ops {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		fmt.Fprintf(&sb, "  %s: %q,\n", name, ops[name])
	}
	sb.WriteString("} as const satisfies Record<string, Permission>;\n\n")
	sb.WriteString("export type OperationId = keyof typeof OPERATION_PERMISSIONS;\n")

	return sb.String()
}
