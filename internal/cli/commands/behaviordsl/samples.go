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

package behaviordsl

import (
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/guide"
)

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	sampleList = ListOutput{Items: []Summary{
		{
			Name:    "hostname",
			Methods: []string{"get", "update"},
			Doc:     "The hostname of the host",
		},
		{
			Name:     "managedfile",
			Methods:  []string{"get", "update", "delete"},
			Doc:      "The files that osapi writes and tracks",
			Internal: true,
		},
		{
			Name:    "dns",
			Group:   "network",
			Methods: []string{"get", "update", "delete"},
			Doc:     "The DNS resolvers of an interface",
		},
		{
			Name:    "sysctl",
			Methods: []string{"list", "get", "create", "update", "delete"},
			Doc:     "Kernel parameters, written to sysctl.d files and applied",
		},
		{
			Name:    "user",
			Methods: []string{"list", "get", "create", "update", "delete", "passwd"},
			Doc:     "The users of the host",
		},
		{
			Name:    "sshkey",
			Of:      "user",
			Methods: []string{"list", "create", "delete"},
			Doc:     "The authorized SSH keys of a user",
		},
	}}

	sampleGet = GetOutput{
		Name:      "sysctl",
		Doc:       "Kernel parameters, written to sysctl.d files and applied",
		Placement: behavior.NodePlacement,
		API:       true,
		CLI:       true,
		Platforms: []string{"debian"},
		Sudo:      []string{"sysctl"},
		Writes:    []string{"/etc/sysctl.d/osapi-*.conf"},
		Uses:      []string{"managedfile"},
		Methods: []Method{
			{
				Name:       "get",
				Verb:       behavior.VerbRead,
				HTTPMethod: "GET",
				Route:      "/api/node/{hostname}/sysctl/{key}",
				Permission: "sysctl:read",
				Command:    "osapi client node sysctl get",
				Doc:        "Get a kernel parameter's current value, and whether osapi manages it",
			},
			{
				Name:       "update",
				Verb:       behavior.VerbWrite,
				HTTPMethod: "PUT",
				Route:      "/api/node/{hostname}/sysctl/{key}",
				Permission: "sysctl:write",
				Command:    "osapi client node sysctl update",
				Doc:        "Set a kernel parameter, manage it, and apply it",
			},
		},
		Types: []Type{
			{Name: "get", Input: "GetInput { key* }", Output: "Entry { key, value, managed }"},
			{
				Name:   "update",
				Input:  "UpdateInput { key*, value* }",
				Output: "MutationResult { key, changed }",
			},
		},
		Settings: []Setting{
			{
				Key:     "behaviors.sysctl.dir",
				Flag:    "sysctl-dir",
				Type:    "string",
				Default: "/etc/sysctl.d",
			},
		},
		Telemetry: Telemetry{
			Span:       "behavior.sysctl.<method>",
			ChildSpans: []string{"behavior.managedfile.<method>", "exec sysctl"},
			Metrics:    []string{"osapi_behavior_calls_total", "osapi_behavior_duration_seconds"},
			Attributes: []string{"osapi.behavior.name", "osapi.behavior.method"},
			Recorded:   "key (get, update)",
			Inputs:     map[string][]string{"get": {"key"}, "update": {"key"}},
		},
	}

	sampleMethod = MethodOutput{
		Name:       "update",
		Verb:       behavior.VerbWrite,
		Doc:        "Set a kernel parameter, manage it, and apply it",
		HTTPMethod: "PUT",
		Route:      "/api/node/{hostname}/sysctl/{key}",
		Permission: "sysctl:write",
		Command:    "osapi client node sysctl update",
		InputType:  "UpdateInput",
		OutputType: "MutationResult",
		Fields: []Field{
			{
				Field:       "key",
				Type:        "string",
				Required:    true,
				Constraints: "pattern ^[a-zA-Z0-9][a-zA-Z0-9._-]*$, min length 1",
				Description: "Parameter name, such as net.ipv4.ip_forward",
			},
			{
				Field:       "value",
				Type:        "string",
				Required:    true,
				Constraints: `pattern ^[^\r\n]*$`,
				Description: "Value to set",
			},
		},
		Results: []Result{
			{Field: "key", Type: "string", Description: "Parameter name"},
			{Field: "changed", Type: "boolean", Description: "Whether the host changed"},
		},
		Examples: []ExampleRow{{
			Example: "Lower swappiness",
			Input:   "key=vm.swappiness, value=10",
			Output:  "key=vm.swappiness, changed=true",
		}},
	}

	sampleOpenAPI = OpenAPIOutput{
		Name: "sysctl",
		YAML: "paths:\n" +
			"  /api/node/{hostname}/sysctl/{key}:\n" +
			"    put:\n" +
			"      operationId: PutNodeSysctl\n" +
			"      summary: Set a kernel parameter, manage it, and apply it\n" +
			"      tags:\n" +
			"        - sysctl_operations\n",
	}

	sampleCreate = CreateOutput{
		Name: "motd",
		Files: []File{
			{Path: "internal/behaviors/motd/behavior.go"},
			{Path: "internal/behaviors/motd/behavior_public_test.go"},
			{Path: "internal/behaviors/motd/debian.go"},
			{Path: "internal/behaviors/motd/methods.go"},
			{Path: "internal/behaviors/motd/provider.go"},
			{Path: "internal/behaviors/all/all.go"},
		},
		Next: []Step{
			{Step: "go generate ./internal/behaviors/motd/mocks ./internal/behaviors/all"},
			{Step: "change the types, the provider and the examples in internal/behaviors/motd"},
			{Step: "go generate ./internal/behaviors/all"},
			{Step: "osapi dsl behavior validate"},
			{Step: "just test"},
		},
	}

	// sampleSpec is the first term of the spec that spec get prints.
	sampleSpec = SpecOutput{Kinds: []guide.Section{{
		Kind:  "definition",
		Title: "Definition",
		Terms: []guide.Term{{
			Kind: "definition",
			Name: "Register(Define(name, doc, options...))",
			Does: "Declares the behavior and adds it to the registry. The agent, the " +
				"server, the CLI and the generator read the registry.",
			When: "One time, in behavior.go: var Behavior = " +
				"behavior.Register(behavior.Define(...)).",
		}},
	}}}

	sampleValidate = ValidateOutput{Behaviors: 33, Valid: true}
)
