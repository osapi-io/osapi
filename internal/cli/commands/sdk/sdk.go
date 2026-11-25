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

// Package sdk is the sdk group: osapi sdk list, get, spec get and validate,
// which describe and check the Go SDK in this binary, as each DSL's group
// does and as osapi agent collector does for gohai's collectors.
package sdk

import (
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/sdk/sdkspec"
)

// Group is the sdk group's definition.
var Group = command.Register(command.Define(
	"sdk", "List, describe and check the Go SDK in this binary: its types and its services",
	command.Uses[Deps](),
	command.SubDoc("spec", "Print the parts of the Go SDK"),
	command.Method[ListInput, ListOutput]("list").
		Short("List the types for values and the services of the client").
		Example("").
		Prints(sampleList).
		Example("--json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Show a type with its values, or a service with its methods").
		Long("A name of a type, such as ContentType, shows its doc and its values. A\n"+
			"name of a service, such as Hostname, shows each method with the permission\n"+
			"that it needs, its route, its input and its result.").
		Example("--name Hostname").
		Prints(sampleGet).
		Example("--name ContentType --json").
		Fails("no name", GetInput{}).
		Handle(Get),
	command.Method[SpecInput, SpecOutput]("get").
		Sub("spec").
		Short("Print the Go SDK: its packages, how a call works, its types and its services").
		Long("The list comes from this binary, from types.All and client.Services. Thus it\n"+
			"agrees with the build that you run. The SDK skill is made from it.").
		Example("").
		Prints(sampleSpec).
		Example("--json").
		Handle(Spec),
	command.Method[ValidateInput, ValidateOutput]("validate").
		Short("Check the SDK against its rules").
		Long("It prints each problem with its name, rule and message, then the fix of each\n"+
			"rule, and exits 1 when a rule fails.\n\n"+
			"The rules:\n\n"+sdkspec.Checks.Help()).
		Example("").
		Prints(sampleValidate).
		Example("--json").
		Handle(Validate),
))
