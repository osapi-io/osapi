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

// Package token is the token group of local commands: osapi token generate,
// roles list and validate, which sign and read the bearer tokens of the API
// with the signing key of the config. They run in the CLI, and call no API.
package token

import (
	"github.com/osapi-io/osapi/internal/cli/command"
)

// Group is the token group's definition. The commands, their flags, their
// help and their pages come from it. osapi cli command spec get lists every
// option.
var Group = command.Register(command.Define(
	"token", "Generate and validate API tokens",
	command.Uses[Deps](),
	command.SubDoc("roles", "Read the roles and what each grants"),
	command.Method[GenerateInput, GenerateOutput]("generate").
		Short("Generate a new token").
		Example("--subject user123 --roles=read").
		Prints(sampleGenerate).
		Example("--subject svc@example.com --roles=read \\\n    -p status:read -p health:read").
		Fails("no roles", GenerateInput{Subject: "ops"}).
		Fails("no subject", GenerateInput{Roles: []string{"read"}}).
		Fails("an unknown permission", GenerateInput{
			Roles: []string{"read"}, Subject: "ops", Permissions: []string{"root:all"},
		}).
		Handle(Generate),
	command.Method[RolesInput, RolesOutput]("list").
		Sub("roles").
		Short("List the roles and the permissions each grants").
		Long("The list has the permissions of each behavior, so it is correct for the\n"+
			"binary that you run. A custom role from the config's\n"+
			"`controller.api.security.roles` gets a column of its own. Without a config\n"+
			"file it shows the built-in roles only. A behavior's permissions are\n"+
			"`<name>:read` and `<name>:write`, or the group's when it is nested under\n"+
			"one. `admin` and `write` grant both, and `read` grants the read permission.\n"+
			"Use this list, not a table in the docs.").
		Example("").
		Prints(sampleRoles).
		Example("--json").
		Handle(Roles),
	command.Method[RoleInput, Role]("get").
		Sub("roles").
		Short("Show a role and each permission it grants").
		Example("--name admin").
		Prints(sampleRole).
		Fails("no name", RoleInput{}).
		Handle(RoleGet),
	command.Method[ValidateInput, Claims]("validate").
		Short("Validate a token for authenticity and claims").
		Example("--token eyJhbGciOiJIUzI1NiIs...").
		Prints(sampleClaims).
		Fails("no token", ValidateInput{}).
		Handle(Validate),
))
