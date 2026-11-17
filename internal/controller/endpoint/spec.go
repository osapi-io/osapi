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

package endpoint

import "github.com/osapi-io/osapi/internal/guide"

// The kinds of term in Spec, in the order that a group is written.
const (
	KindDefinition = "definition"
	KindOption     = "option"
	KindMethod     = "method"
	KindRefine     = "refine"
	KindError      = "error"
	KindTag        = "tag"
	KindCheck      = "check"
	KindNotYet     = "not-yet"
)

// Spec is every term of the endpoint DSL. osapi controller endpoint spec
// prints it, so an author, or an agent, learns the DSL from the binary that
// it builds. A test fails on an exported option, method or refinement of
// this package without an entry.
var Spec = guide.Spec{
	Kinds: []guide.Kind{
		{Kind: KindDefinition, Title: "Definition"},
		{Kind: KindOption, Title: "Options of Define"},
		{Kind: KindMethod, Title: "Methods"},
		{Kind: KindRefine, Title: "Refinements of a method"},
		{Kind: KindError, Title: "Errors a method maps"},
		{Kind: KindTag, Title: "Tags on input and output fields"},
		{Kind: KindCheck, Title: "What holds a group to its spec"},
		{Kind: KindNotYet, Title: "Not supported yet"},
	},
	Terms: []guide.Term{
		{
			Kind: KindDefinition,
			Name: `Register(Define(name, doc, options...))`,
			Does: "Declares the group and adds it to the registry. The commands and the tests read the registry. Each path of the group starts /api/<name>.",
			When: "One time, in <group>.go under internal/controller/endpoints/<group>: var Group = endpoint.Register(endpoint.Define(...)). Add its import to internal/controller/endpoints/all.",
		},

		{
			Kind: KindOption,
			Name: `Title("title")`,
			Does: "Sets the title of the group's spec. The joined spec puts it in front of each tag, and the UI's client files get their names from the result.",
			When: "When a group moves from a hand-written spec: copy its info.title, so no client file is renamed.",
		},
		{
			Kind: KindOption,
			Name: "Uses[Deps]()",
			Does: "Names the struct of what the group's handlers use, such as the job client. The controller gives one at startup, and a test gives one with fakes. It is the place of the host in a behavior.",
			When: "In each group that is served. Keep it to interfaces, so a test can fake each one.",
		},
		{
			Kind: KindOption,
			Name: `Doc[T]("description")`,
			Does: "Describes T, a named type, in its schema. Go has no tag on a type, so a group says it here.",
			When: "For each type whose schema has a description, such as one that another spec defines too: the join fails when two specs define one schema in different words.",
		},
		{
			Kind: KindOption,
			Name: `Tag("name", "display", "doc")`,
			Does: "Sets the tag of each operation, its name in the docs and its description.",
			When: "When a group moves: copy its tag. A new group can leave it; the tag is then the group's name.",
		},

		{
			Kind: KindMethod,
			Name: `Get[In, Out]("name")`,
			Does: "A GET that takes In and answers with Out. Its permission is <group>:read.",
			When: "For a read.",
		},
		{
			Kind: KindMethod,
			Name: `Post[In, Out]("name")`,
			Does: "A POST that takes In and answers with Out. Its permission is <group>:write.",
			When: "For a create, or an action such as drain.",
		},
		{
			Kind: KindMethod,
			Name: `Put[In, Out]("name")`,
			Does: "A PUT that takes In and answers with Out. Its permission is <group>:write.",
			When: "For an update of the thing the key names.",
		},
		{
			Kind: KindMethod,
			Name: `Delete[In, Out]("name")`,
			Does: "A DELETE that takes In and answers with Out. Its permission is <group>:write.",
			When: "For a removal.",
		},

		{
			Kind: KindRefine,
			Name: ".Handle(fn)",
			Does: "Sets the function that runs the method: func(ctx context.Context, d Deps, in In) (Out, error). The generated route decodes and validates In, calls it, and maps its error with Errors.",
			When: "On each method of a group that Uses a type. Validate checks the signature.",
		},
		{
			Kind: KindRefine,
			Name: `.Sub("segment")`,
			Does: "Puts a fixed segment after the group's name, before the key: /api/<group>/<segment>[/{key}].",
			When: "For a path beside the key, such as /api/agent/pending, or before it, such as /api/telemetry/trace/{id}.",
		},
		{
			Kind: KindRefine,
			Name: `.Action("segment")`,
			Does: "Puts a segment after the key: /api/<group>/{key}/<segment>.",
			When: "For an action on one thing, such as /api/agent/{hostname}/drain. Needs a key.",
		},
		{
			Kind: KindRefine,
			Name: `.ID("operationId")`,
			Does: "Sets the operationId. The Go SDK and the UI client get their names from it. Without it, the ID is the verb, the group, the Sub and the name.",
			When: "When a group moves: keep the ID of the hand-written spec, so no name that a program calls changes.",
		},
		{
			Kind: KindRefine,
			Name: `.Describe("summary")`,
			Does: "Sets the summary: what the method does, in one line.",
			When: "Always.",
		},
		{
			Kind: KindRefine,
			Name: `.Explain("text")`,
			Does: "Sets the description: the rest of what a caller needs to know.",
			When: "When the summary is not enough, such as what occurs when two agents share a hostname.",
		},
		{
			Kind: KindRefine,
			Name: `.Returns("doc")`,
			Does: "Describes the response of a success. Without it, the summary describes it.",
			When: "When the response needs its own words, such as \"List of active agents.\".",
		},
		{
			Kind: KindRefine,
			Name: `.Command("short", "long")`,
			Does: "Sets the CLI command's own help: its one line and the rest. Without it, the command has the summary and the description of the API. The one line, and the doc of each input field that becomes a flag's help, start with a capital letter and have no period at the end; the long help is sentences. A test of every command holds this.",
			When: "When the CLI needs words for a person at a shell, such as what the command reads and what to run next.",
		},
		{
			Kind: KindRefine,
			Name: `.Example("--flag value")`,
			Does: "Adds an example of the CLI command: the flags after osapi client <group> <method>. The help and the CLI page show each one.",
			When: "At least one for each method: the call a person makes most.",
		},
		{
			Kind: KindRefine,
			Name: `.Notes("text")`,
			Does: "Adds a Notes section to the CLI command's page.",
			When: "For what a person needs that the help does not say, such as what --json has that the table does not.",
		},
		{
			Kind: KindRefine,
			Name: `.Result("text")`,
			Does: "Adds a Result section to the CLI command's page.",
			When: "For what the command does in a case that needs words, such as when two agents share a hostname.",
		},
		{
			Kind: KindRefine,
			Name: `.Permission("group:verb")`,
			Does: "Names the permission the method needs, in place of <group>:read or <group>:write.",
			When: "When a different permission covers it, such as agent:read for facts, or trace:read for telemetry.",
		},
		{
			Kind: KindRefine,
			Name: ".Public()",
			Does: "Makes the method need no token. It has no 401 or 403.",
			When: "Only for the health probes that a load balancer calls.",
		},
		{
			Kind: KindRefine,
			Name: ".Status(code)",
			Does: "Sets the status of a success. 204 has no body.",
			When: "For 201 after a create or a retry, or 204 after a delete with nothing to return.",
		},
		{
			Kind: KindRefine,
			Name: ".Errors(errors...)",
			Does: "Maps the errors a method returns to status codes. An error that is not in the list is a 500. Each method also gets 400 when it has an input, 401 and 403 when it is not public, and 500.",
			When: "For each answer a caller acts on, such as 404 or 409. Never map a failure of the controller to 409.",
		},

		{
			Kind: KindError,
			Name: `Errorf(sentinel, "format", args...)`,
			Does: "Is an error whose text is the message, and that errors.Is matches to sentinel. The answer has the status that Errors maps sentinel to, and the message as its error.",
			When: "In a handler, to answer 404 or 409 with a message that names what it was called with, such as no pending agent with hostname \"web-01\".",
		},
		{
			Kind: KindError,
			Name: `NotFound(err, "doc")`,
			Does: "Maps err to 404.",
			When: "When the thing the key names is not there.",
		},
		{
			Kind: KindError,
			Name: `Conflict(err, "doc")`,
			Does: "Maps err to 409.",
			When: "When the state of the thing does not allow the call, such as an agent that is already drained.",
		},
		{
			Kind: KindError,
			Name: `Unavailable(err, "doc")`,
			Does: "Maps err to 503.",
			When: "When a backend the method needs is not set, such as the trace backend.",
		},

		{
			Kind: KindTag,
			Name: `in:"path"`,
			Does: "Makes the input field the key: {name} in the path. It is required. A group has at most one.",
			When: "For the field that names the thing, such as hostname or id.",
		},
		{
			Kind: KindTag,
			Name: `in:"query"`,
			Does: "Makes the input field a query parameter.",
			When: "For a filter, a page or an option, such as status, limit or force.",
		},
		{
			Kind: KindTag,
			Name: `in:"body"`,
			Does: "Makes the field's type the whole request body. Without it, the input's other fields are the body, as one object.",
			When: "When the body is a named schema, such as a retry request.",
		},
		{
			Kind: KindTag,
			Name: `json:"name"`,
			Does: "Names the field in JSON, the parameter and the schema property. In an output type, a field without ,omitempty or ,omitzero is required.",
			When: "On each field.",
		},
		{
			Kind: KindTag,
			Name: `doc:"..."`,
			Does: "Describes the field in the spec.",
			When: "On each field a caller reads or sets.",
		},
		{
			Kind: KindTag,
			Name: `example:"..."`,
			Does: "Gives the field an example in the spec and the API docs: a number for an integer or a number, true or false for a boolean, and the text for the rest.",
			When: "On each field a caller reads, so the docs show a value.",
		},
		{
			Kind: KindTag,
			Name: `pattern:"^...$"`,
			Does: "Gives a string field a regular expression, in the spec, and the generated route answers 400 for a value that does not match.",
			When: "For an ID of a fixed form, such as a trace ID of 32 hexadecimal characters.",
		},
		{
			Kind: KindTag,
			Name: `validate:"..."`,
			Does: "In an input, required makes the field required. oneof gives an enum, and min and max give the length of a string.",
			When: "On each input field with a limit. On an output field, only for oneof.",
		},
		{
			Kind: KindTag,
			Name: "time.Time",
			Does: "Is a string with the date-time format.",
			When: "For a timestamp. Tag an optional one ,omitzero: ,omitempty does not leave out a zero time.",
		},
		{
			Kind: KindTag,
			Name: "a named struct, or a struct with no name",
			Does: "A named struct is a schema in the components, with its name. A struct with no name, or an alias of one, is in place.",
			When: "Give a type the name its schema has. Use an alias of a struct with no name, such as Message, for a schema that the spec has in place.",
		},

		{
			Kind: KindCheck,
			Name: "TestEndpointParity",
			Does: "Compares the contract of a group's spec with its hand-written spec, line by line: tags, operations, parameters, auth, bodies, responses and schemas. Prose is not compared.",
			When: "Add a row to moving in internal/controller/endpoints/all/parity_public_test.go for a group that moves. A difference on purpose is accepted with its reason and an issue.",
		},
		{
			Kind: KindCheck,
			Name: "TestJoinedSpec",
			Does: "Holds the joined spec to its contract, internal/apicontract/testdata/joined.golden.",
			When: "When you change the API on purpose, run it with -update and keep the diff.",
		},
		{
			Kind: KindCheck,
			Name: "osapi controller endpoint get --name <group> --openapi",
			Does: "Prints the group's spec, as the definition makes it.",
			When: "To read the prose beside the hand-written spec.",
		},

		{
			Kind: KindCheck,
			Name: "go generate ./internal/controller/endpoints/all",
			Does: "Writes, for each group that Uses a type, gen/api.yaml, the routes in gen/api.go, the decoders in gen/decode.go and the CLI commands in gen/cli.go, one for each method under osapi client <group>, and internal/controller/endpoints/all/gen, which registers every group's routes and commands. just generate then joins gen/api.yaml into the main spec.",
			When: "After each change to a group. go-generate-check fails when a gen/ file is stale. Never edit a gen/ file.",
		},
		{
			Kind: KindCheck,
			Name: "command.Render(Group, \"method\", fn)",
			Does: "Sets how a method's answer prints without --json: fn gets the input the command sent and the output the API answered. A method without a renderer prints its answer as JSON. The generated command, its flags and its help come from the definition.",
			When: "In cmd/client_<group>.go, the file of the group's command, for each method a person reads at a shell, such as a table for a list or a detail view for a get.",
		},
		{
			Kind: KindCheck,
			Name: "osapi controller endpoint new --name <group> --description \"...\"",
			Does: "Writes a new group under internal/controller/endpoints/<group>: a list and a get, an empty Deps, its types and the tests of its methods, and adds its import to internal/controller/endpoints/all. It builds, validates and passes its tests; endpoint-scaffold-check in just test holds it to that.",
			When: "To start a new group. Then run go generate ./internal/controller/endpoints/all and just generate, and change it.",
		},
		{
			Kind: KindCheck,
			Name: "Deps in internal/controller/handlers.go",
			Does: "The controller gives each group its Deps, such as agentDeps. An empty Deps needs no wiring. A group whose Deps has a field and is not given stops the controller at startup.",
			When: "When a group's Deps gets its first field: add a line there, and a nil pointer must be a nil interface.",
		},
		{
			Kind: KindNotYet,
			Name: "Multipart uploads",
			Does: "A body of multipart/form-data, as POST /api/file has.",
			When: "Comes when the file group moves.",
		},
	},
}
