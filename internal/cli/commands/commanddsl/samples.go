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

package commanddsl

import "github.com/osapi-io/osapi/internal/cli/command"

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	sampleList = ListOutput{
		Items: []Summary{
			{
				Command: "osapi controller token generate",
				Layer:   "command",
				Group:   "controller token",
				Doc:     "Generate a new token",
			},
			{
				Command: "osapi controller token validate",
				Layer:   "command",
				Group:   "controller token",
				Doc:     "Validate a token for authenticity and claims",
			},
			{
				Command: "osapi client agent list",
				Layer:   "endpoint",
				Group:   "agent",
				Doc:     "List active agents",
			},
			{
				Command: "osapi client node hostname get",
				Layer:   "behavior",
				Group:   "hostname",
				Doc:     "Get the hostname",
			},
		},
		Command:  2,
		Endpoint: 1,
		Behavior: 1,
	}

	sampleFlagRoles = Flag{
		Flag:     "--roles",
		Short:    "r",
		Type:     "[]string",
		Validate: "required",
		Doc:      "Roles for the token: built-in, or a custom role of the config",
	}
	sampleFlagSubject = Flag{
		Flag:     "--subject",
		Short:    "u",
		Type:     "string",
		Validate: "required",
		Doc:      "Subject for the token (e.g., user ID or unique identifier)",
	}

	sampleGet = GetOutput{
		Name: "controller token",
		Doc:  "Generate and validate API tokens",
		Commands: []Command{
			{
				Command: "osapi controller token generate",
				Method:  "generate",
				Short:   "Generate a new token",
				Fails:   []string{"no roles", "no subject"},
				Flags:   []Flag{sampleFlagRoles, sampleFlagSubject},
			},
			{
				Command: "osapi controller role list",
				Method:  "roles list",
				Short:   "List the roles and the permissions each grants",
			},
		},
	}

	sampleMethod = MethodOutput{
		Command:    "osapi controller token generate",
		Method:     "generate",
		Doc:        "Generate a new token",
		Flags:      []Flag{sampleFlagRoles, sampleFlagSubject},
		OutputType: "GenerateOutput",
		Results: []Result{
			{Field: "token", Type: "string"},
			{Field: "subject", Type: "string"},
			{Field: "roles", Type: "array"},
		},
		Examples: []Example{{Flags: "--subject user123 --roles=read"}},
		Refuses:  []string{"no roles", "no subject"},
	}

	sampleCreate = CreateOutput{
		Name: "widget",
		Files: []File{
			{Path: "internal/cli/commands/widget/widget.go"},
			{Path: "internal/cli/commands/widget/deps.go"},
			{Path: "internal/cli/commands/widget/methods.go"},
			{Path: "internal/cli/commands/widget/methods_public_test.go"},
			{Path: "internal/cli/commands/widget/types.go"},
			{Path: "internal/cli/commands/all/all.go"},
		},
		Next: []Step{
			{Command: "go generate ./internal/cli/commands/all"},
			{Command: "go test ./internal/cli/commands/widget/..."},
			{Command: "go run . dsl command validate"},
			{Command: "mise exec -- just cli-docs-generate"},
		},
	}

	// sampleSpec is the first part of the spec that spec get prints.
	sampleSpec = SpecOutput{Kinds: command.Spec.Sections()[:1]}

	sampleValidate = ValidateOutput{Local: 17, Groups: 57, Valid: true}
)
