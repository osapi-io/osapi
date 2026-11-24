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

// Package endpointdsl is the endpoint group under osapi controller: the
// commands that list, describe, scaffold, teach and check the endpoint
// groups of the controller. The package is not named endpoint, which is
// the DSL itself.
package endpointdsl

import (
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint"

	// Each endpoint group registers itself, so these commands read them.
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
)

// Group is the endpoint group's definition.
var Group = command.Register(command.Define(
	"endpoint",
	"List, describe and validate the controller's own endpoints, and print their DSL",
	command.Under(command.ParentDSL),
	command.Long("An endpoint is one operation of the controller itself, such as\n"+
		"`GET /api/agent`. It runs on the controller, not on a host. Work on a host\n"+
		"is a behavior: see `osapi dsl behavior`. These commands read the definitions in\n"+
		"the binary. Thus they need no config file and no running controller."),
	command.Uses[Deps](),
	command.SubDoc(
		"spec",
		"Print the endpoint DSL: each option, method, refinement, error and field tag",
	),
	command.SubDoc("openapi", "Print the OpenAPI spec of an endpoint group"),
	command.SubDoc("method", "Show one method of an endpoint group in full"),
	command.Method[ListInput, ListOutput]("list").
		Short("List the endpoint groups this binary has").
		Example("").
		Prints(sampleList).
		Example("--json").
		Handle(List),
	command.Method[GetInput, GetOutput]("get").
		Short("Describe one endpoint group: its routes, permissions and operation IDs").
		Long("`osapi dsl endpoint openapi get` prints the group's part of the\n"+
			"OpenAPI spec.").
		Example("--name agent").
		Prints(sampleGet).
		Example("--name agent --json").
		Fails("no name", GetInput{}).
		Handle(Get),
	command.Method[MethodInput, MethodOutput]("get").
		Sub("method").
		Short("Show one method of an endpoint group: its route, fields, output, examples and errors").
		Long("Each input field shows where the request carries it: the path, the query,\n"+
			"a form or the body. `--json` has the JSON Schema of the input and the\n"+
			"output. `osapi dsl behavior method get` shows a method of a behavior the same\n"+
			"way.").
		Example("--name agent --method drain").
		Prints(sampleMethod).
		Example("--name agent --method drain --json").
		Fails("no name", MethodInput{Method: "list"}).
		Fails("no method", MethodInput{Name: "agent"}).
		Handle(MethodGet),
	command.Method[GetInput, OpenAPIOutput]("get").
		Sub("openapi").
		Short("Print the group's part of the OpenAPI spec, as YAML").
		Long("It prints the spec as the definition makes it. Read it beside the\n"+
			"hand-written spec of the group to compare the prose. `TestEndpointParity`\n"+
			"compares the rest.").
		Example("--name agent").
		Prints(sampleOpenAPI).
		Fails("no name", GetInput{}).
		Handle(OpenAPIGet),
	command.Method[CreateInput, CreateOutput]("create").
		Short("Write a new endpoint group that builds and passes its tests").
		Long("It writes `internal/controller/endpoints/<name>/` in the osapi checkout in\n"+
			"the current directory, and adds its import to\n"+
			"`internal/controller/endpoints/all`. The group has a list and a get, an\n"+
			"empty `Deps` and the tests of its methods. Then run the steps that it\n"+
			"prints. While its `Deps` is empty, the controller needs no wiring.\n"+
			"`osapi dsl endpoint spec get` lists each part of the DSL.").
		Example(`--name trace --description "The traces of the controller"`).
		Prints(sampleCreate).
		Fails("no name", CreateInput{Description: "x"}).
		Fails("no description", CreateInput{Name: "x"}).
		Handle(Create),
	command.Method[SpecInput, SpecOutput]("get").
		Sub("spec").
		Short("Print the endpoint DSL: each option, method, refinement, error and field tag").
		Long("Print each part of the endpoint DSL, what it makes osapi do and when to use\n"+
			"it, and what it cannot do yet. The list comes from this binary. Thus it\n"+
			"agrees with the build that you run. Read it before you write a group, then\n"+
			"check the result with `osapi dsl endpoint validate`.").
		Example("").
		Prints(sampleSpec).
		Example("--json").
		Handle(Spec),
	command.Method[ValidateInput, ValidateOutput]("validate").
		Short("Check every endpoint group against the rules of a definition").
		Long("It prints each problem with its group, method, rule and message, then the\n"+
			"fix of each rule, and exits 1 when a rule fails. `osapi dsl endpoint\n"+
			"spec get` describes each part of a group.\n\n"+
			"The rules:\n\n"+endpoint.Checks.Help()).
		Example("").
		Prints(sampleValidate).
		Example("--json").
		Handle(Validate),
))
