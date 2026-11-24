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

// Package role is the role group of local commands: osapi controller role
// list and get, which read the built-in roles and the custom roles of the
// controller's config, with what each grants. They run in the CLI, and call
// no API.
package role

import (
	"github.com/osapi-io/osapi/internal/cli/command"
)

// Group is the role group's definition. The commands, their flags, their
// help and their pages come from it. osapi dsl command spec get lists every
// option.
var Group = command.Register(command.Define(
	"role", "Read the roles and what each grants",
	command.Under(command.ParentController),
	command.Uses[Deps](),
	command.Method[ListInput, ListOutput]("list").
		Short("List the roles and the permissions each grants").
		Long("The list has the permissions of each behavior, so it is correct for the\n"+
			"binary that you run. A custom role from the config's\n"+
			"`controller.api.security.roles` gets a column of its own. Without a config\n"+
			"file it shows the built-in roles only. A behavior's permissions are\n"+
			"`<name>:<verb>`, such as `dns:read` and `dns:write`. A rule gives each\n"+
			"built-in role its permissions. Use this list, not a table in the docs.").
		Example("").
		Prints(sampleRoles).
		Example("--json").
		Handle(List),
	command.Method[GetInput, Role]("get").
		Short("Show a role and each permission it grants").
		Example("--name admin").
		Prints(sampleRole).
		Fails("no name", GetInput{}).
		Handle(Get),
))
