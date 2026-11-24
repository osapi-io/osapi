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

// Package token is the token group of local commands: osapi controller
// token generate and validate, which sign and read the bearer tokens of the
// API with the signing key of the controller's config. They run in the CLI,
// and call no API.
package token

import (
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Group is the token group's definition. The commands, their flags, their
// help and their pages come from it. osapi dsl command spec get lists every
// option.
var Group = command.Register(command.Define(
	"token", "Generate and validate API tokens",
	command.Uses[Deps](),
	command.Under(command.ParentController),
	command.Method[GenerateInput, GenerateOutput]("generate").
		Short("Generate a new token").
		Example("--subject user123 --roles=read").
		Prints(sampleGenerate).
		Example("--subject svc@example.com --roles=read \\\n    -p status:read -p health:read").
		Example("--subject team-web --roles=operator \\\n    -t group:web -t web-01").
		Fails("no roles", GenerateInput{Subject: "ops"}).
		Fails("no subject", GenerateInput{Roles: []types.Role{types.RoleRead}}).
		Fails("an unknown permission", GenerateInput{
			Roles: []types.Role{types.RoleRead}, Subject: "ops", Permissions: []string{"root:all"},
		}).
		Fails("a target scope of _all", GenerateInput{
			Roles: []types.Role{types.RoleRead}, Subject: "ops", Targets: []string{types.TargetAll},
		}).
		Handle(Generate),
	command.Method[ValidateInput, Claims]("validate").
		Short("Validate a token for authenticity and claims").
		Example("--token -").
		Prints(sampleClaims).
		Fails("no token", ValidateInput{}).
		Handle(Validate),
))
