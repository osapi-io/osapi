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

package endpointdsl

import "github.com/osapi-io/osapi/internal/controller/endpoint"

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	sampleList = ListOutput{Items: []Summary{
		{
			Name:    "agent",
			Methods: []string{"status", "list", "get", "drain", "undrain"},
			Doc:     "Manage the agents in the fleet: discovery, drain and PKI enrollment",
		},
		{
			Name:    "audit",
			Methods: []string{"list", "export", "get"},
			Doc:     "Read the audit log",
		},
	}}

	sampleGet = GetOutput{
		Name: "agent",
		Doc:  "Manage the agents in the fleet: discovery, drain and PKI enrollment",
		Methods: []Method{
			{
				Name:        "list",
				Verb:        "GET",
				Path:        "/api/agent",
				Permission:  "agent:read",
				OperationID: "GetAgents",
			},
			{
				Name:        "drain",
				Verb:        "POST",
				Path:        "/api/agent/{hostname}/drain",
				Permission:  "agent:write",
				OperationID: "DrainAgent",
			},
		},
		Types: []Type{
			{Name: "list", Input: "ListInput", Output: "ListAgentsResponse { items, total_items }"},
			{Name: "drain", Input: "DrainInput { hostname* }"},
		},
	}

	sampleMethod = MethodOutput{
		Name:        "drain",
		Doc:         "Drain an agent",
		HTTPMethod:  "POST",
		Route:       "/api/agent/{hostname}/drain",
		Permission:  "agent:write",
		AdminOnly:   true,
		OperationID: "DrainAgent",
		Command:     "osapi client agent drain",
		InputType:   "DrainInput",
		Fields: []Field{{
			Field:       "hostname",
			In:          "path",
			Type:        "string",
			Required:    true,
			Constraints: "min length 1, max length 255",
			Description: "The hostname of the agent to drain.",
		}},
		Results:  []Result{},
		Examples: []Example{{Flags: "--hostname web-01"}},
		Refuses:  []string{"a hostname over 255 characters"},
		Errors: []Error{
			{Code: 404, Reason: "not_found", Doc: "Agent not found."},
			{Code: 409, Reason: "conflict", Doc: "Agent already in requested state."},
		},
	}

	sampleOpenAPI = OpenAPIOutput{
		Name: "agent",
		YAML: "paths:\n" +
			"  /api/agent/{hostname}/drain:\n" +
			"    post:\n" +
			"      operationId: DrainAgent\n" +
			"      summary: Drain an agent\n" +
			"      tags:\n" +
			"        - agent_operations\n",
	}

	sampleCreate = CreateOutput{
		Name: "trace",
		Files: []File{
			{Path: "internal/controller/endpoints/trace/trace.go"},
			{Path: "internal/controller/endpoints/trace/deps.go"},
			{Path: "internal/controller/endpoints/trace/methods.go"},
			{Path: "internal/controller/endpoints/trace/methods_public_test.go"},
			{Path: "internal/controller/endpoints/trace/types.go"},
		},
		Next: []Step{
			{Command: "go generate ./internal/controller/endpoints/all"},
			{Command: "mise exec -- just generate"},
			{Command: "go test ./internal/controller/endpoints/trace/..."},
			{Command: "go run . controller endpoint validate"},
		},
	}

	// sampleSpec is the first part of the spec that spec get prints.
	sampleSpec = SpecOutput{Kinds: endpoint.Spec.Sections()[:1]}

	sampleValidate = ValidateOutput{Groups: 10, Valid: true}
)
