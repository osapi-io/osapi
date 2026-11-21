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

import (
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// The kinds of term in Spec, in the order that a group is written.
const (
	KindDefinition = "definition"
	KindOption     = "option"
	KindMethod     = "method"
	KindRefine     = "refine"
	KindError      = "error"
	KindDeps       = "deps"
	KindTag        = "tag"
	KindCheck      = "check"
	KindGenerated  = "generated"
)

// Spec is every term of the endpoint DSL. osapi controller endpoint spec get
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
		{Kind: KindDeps, Title: "What a Deps takes"},
		{Kind: KindTag, Title: "Tags on input and output fields"},
		{Kind: KindCheck, Title: "What holds a group to its spec"},
		{Kind: KindGenerated, Title: "Generated, never edited"},
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
			Name: `Example[T]("field", "example")`,
			Does: "Gives the field of T with that json name its example in the schema.",
			When: "For an example with a double quote, such as a JSON document. A struct tag cannot keep the backslash of a quote, because the formatters drop it.",
		},
		{
			Kind: KindOption,
			Name: `Tag("name", "display", "doc")`,
			Does: "Sets the tag of each operation, its name in the docs and its description.",
			When: "When a group moves: copy its tag. A new group can leave it; the tag is then the group's name.",
		},
		{
			Kind: KindOption,
			Name: `CommandLong("text")`,
			Does: "Is the rest of the help of the group's command, after its one line.",
			When: "When the group's command needs more than its one line, such as what most of its commands read.",
		},
		{
			Kind: KindOption,
			Name: `SubDoc("segment", "short")`,
			Does: "Is the one line of help of the command that a Sub segment makes, such as trace in osapi client telemetry trace. Without it, the command says <Segment> commands.",
			When: "For each Sub segment of the group, so its command says what it holds.",
		},
		{
			Kind: KindOption,
			Name: `CommandNotes("text")`,
			Does: "Adds a Notes section to the page of the group's own command, such as a link to a page that tells more about the group.",
			When: "When the group's command has a page of its own to point at, as each behavior's does.",
		},
		{
			Kind: KindOption,
			Name: `Prefix("next/node/{hostname}")`,
			Does: "Puts each path of the group under a prefix, between /api/ and the group's name. Each {field} of the prefix is an in:\"path\" field of each method's input, which is not the method's key.",
			When: "For the groups that the behavior generator writes, whose paths start with the target host. A group of the controller has none. The fixed segments are the command's path too: next node sysctl.",
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
			Does: "Sets the function that runs the method: func(ctx context.Context, d Deps, in In) (Out, error). The generated route decodes and validates In, calls it, and maps its error with Errors. Only the route calls it; other code that needs the work goes through the API.",
			When: "On each method of a group that Uses a type. Validate checks the signature.",
		},
		{
			Kind: KindRefine,
			Name: `.Sub("segment")`,
			Does: "Puts a fixed segment after the group's name, before the key: /api/<group>/<segment>[/{key}]. When the segment is not the method's name, the CLI command is under it too: osapi client <group> <segment> <method>.",
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
			Name: `.Fails("doc", in)`,
			Does: "Declares an input that the API refuses with 400: the method's input, with the other fields valid and one rule broken. A test sends each one through the routes. " + rule.FailsRequires,
			When: rule.FailsWhen,
		},
		{
			Kind: KindRefine,
			Name: `.Example("--flag value", "doc")`,
			Does: "Adds an example of the CLI command: the flags after osapi client <group> <method>, and, when given, what it shows, as a # line above it. The help and the CLI page show each one.",
			When: "At least one for each method: the call a person makes most.",
		},
		{
			Kind: KindRefine,
			Name: `.Use("name")`,
			Does: "Names the CLI command apart from the method. A method whose Sub differs from its command name goes under a command of the Sub.",
			When: "When the command's verb is a method name the group has already, such as list under pending beside the group's own list.",
		},
		{
			Kind: KindRefine,
			Name: `.Location("field")`,
			Does: "Gives a 201 the Location header: the group's path and the value of the output's field with that json name.",
			When: "For each method that creates a resource and answers 201, such as job retry with job_id.",
		},
		{
			Kind: KindRefine,
			Name: `.Answers(code, "doc")`,
			Does: "Adds a status that the method answers with its own output, not an ErrorResponse. The method returns its output with endpoint.Answer(code).",
			When: "For a probe whose failure has a body that says what failed, such as the 503 of health ready.",
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
			Name: `.Says("state")`,
			Does: "Sets what the CLI prints for a method that answers with no body: the key of the input, then the text, such as web-01 draining. Without it, the CLI prints the past tense of the method, such as m-1 admitted for admit.",
			When: "On a method that answers 204 and whose result is a state, not the past tense of the method, such as drain.",
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
			Name: `Internal("msg", cause)`,
			Does: "Is a failure of the controller whose cause stays out of the answer: the caller gets msg and a 500, and the log gets the cause.",
			When: "When the cause has what a caller must not see, such as the address or the error of a backend.",
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
			Name: `Answer(code)`,
			Does: "Returned with the method's output, sends that output with code: a status that .Answers added.",
			When: "When a probe fails and its body says what failed, such as health ready's 503.",
		},
		{
			Kind: KindError,
			Name: `Forbidden(err, "doc")`,
			Does: "Maps err to 403. Its doc is the 403's description, so say the permission failure too.",
			When: "When the token may make the request, but not on this thing, such as a write to a file that osapi owns.",
		},
		{
			Kind: KindError,
			Name: `Invalid(err, "doc")`,
			Does: "Maps err to 400. Its doc is the 400's description, so say the decode failures too.",
			When: "When the request is well formed but the method cannot do it, such as a retry of a job with no operation data.",
		},
		{
			Kind: KindError,
			Name: `Unavailable(err, "doc")`,
			Does: "Maps err to 503.",
			When: "When a backend the method needs is not set, such as the trace backend.",
		},

		{
			Kind: KindDeps,
			Name: "an interface of the group, in deps.go",
			Does: "Each field of Deps is an interface that the group declares in its deps.go, with only the methods that its handlers call. internal/controller/handlers.go gives the real service of the controller. The tests give the mockgen mock from the group's mocks/, so a handler test sets what each call returns and asserts each call it makes.",
			When: "For each service that a handler uses. Declare a narrow interface, not the concrete type, and never write a fake by hand. A group whose service the controller may not have, such as a store with no bucket, checks for nil.",
		},
		{
			Kind: KindDeps,
			Name: "Jobs",
			Does: "The job client of the controller: it lists and reads the agents and the jobs, sets the drain flag, writes the timeline of an agent, and deletes and retries jobs. handlers.go gives d.jobs.",
			When: "For a group that reads or changes the agents or the jobs, such as agent, job and admission. A request that runs on a host is a behavior, not a Jobs call of an endpoint.",
		},
		{
			Kind: KindDeps,
			Name: "Store",
			Does: "A store of the controller in its own KV bucket: the admission records of the agents, or the entries of the audit log.",
			When: "For the group that owns the records. A Store can be nil when the controller has no bucket for it; the group answers that the feature is off.",
		},
		{
			Kind: KindDeps,
			Name: "Objects, States",
			Does: "The Object Store, which holds the files that osapi client file upload stores, and the file-state KV, whose keys are the files that agents deployed. fileDeps in handlers.go opens both.",
			When: "For the file group. An agent reads an object through the object behavior, not through an endpoint.",
		},
		{
			Kind: KindDeps,
			Name: "Health",
			Does: "The health.Reporter of the controller: the state of NATS, the controller and its parts, for the health, controller, nats and agent groups.",
			When: "For a group that reports whether a part of osapi is up.",
		},
		{
			Kind: KindDeps,
			Name: "Traces",
			Does: "The TraceReader: it reads one trace from the tracing backend that the config names, for osapi client telemetry trace get.",
			When: "For the telemetry group.",
		},
		{
			Kind: KindTag,
			Name: `param:"Name"`,
			Does: "Makes a path field a shared parameter: the spec has it once, under components.parameters.Name, and each operation refers to it.",
			When: "On a path field that each method of a group has, such as the hostname of a behavior's target.",
		},
		{
			Kind: KindTag,
			Name: `in:"path"`,
			Does: "Makes the input field the key: {name} in the path. It is required. A method has at most one, besides the fields that the group's Prefix names.",
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
			Name: `in:"form"`,
			Does: "Makes the field a part of a multipart/form-data body. A []byte field is the file, a string of format binary.",
			When: "For an upload of a file, such as POST /api/file.",
		},
		{
			Kind: KindTag,
			Name: `in:"body"`,
			Does: "Makes the field's type the whole request body. Without it, the input's other fields are the body, as one object. The body is optional unless the field is validate:\"required\".",
			When: "When the body is a named schema, such as a retry request.",
		},
		{
			Kind: KindTag,
			Name: `flag:"name"`,
			Does: "Names the field's CLI flag. Without it, the flag is the json name with _ as -. A field of an in:\"body\" struct is a flag too.",
			When: "When the flag has a name that the CLI uses elsewhere, such as --target for target_hostname.",
		},
		{
			Kind: KindTag,
			Name: `short:"T"`,
			Does: "Gives the field's CLI flag a one-letter shorthand.",
			When: "When the CLI uses that shorthand elsewhere, such as -T for a target.",
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
			Name: `default:"20"`,
			Does: "Gives a field its default in the spec, and its CLI flag the same value. The method still sees the zero value when the caller sends none, and applies the default itself.",
			When: "For an optional query parameter, such as limit.",
		},
		{
			Kind: KindTag,
			Name: `pattern:"^...$"`,
			Does: "Gives a string field a regular expression, in the spec, and the generated route answers 400 for a value that does not match.",
			When: "For an ID of a fixed form, such as a trace ID of 32 hexadecimal characters.",
		},
		{
			Kind: KindTag,
			Name: rule.TagName,
			Does: rule.TagDoes + " In an endpoint, the generated route checks them before the handler runs, and a handler runs only from its route: TestHandlersRunFromRoutes, in internal/controller/endpoints/all, fails a call of it from other code. The author writes no check.",
			When: rule.TagWhen,
		},
		{
			Kind: KindTag,
			Name: rule.RegisterName,
			Does: rule.RegisterDoes,
			When: rule.RegisterWhen,
		},
		{
			Kind: KindTag,
			Name: "time.Time",
			Does: "Is a string with the date-time format. An int64 is an integer of format int64 and the other integers have none; a float64 is a number of format double and a float32 has none; as oapi-codegen maps them back.",
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
			Name: "SentEmpty(fields)",
			Does: "The optional rule: each optional field of a request body, with omitempty in its validate tag and none in its json tag, which a client would send as \"\" or 0. controller endpoint validate checks each write with it, and behavior validate checks each behavior write's body with the same func.",
			When: "Add ,omitempty to the json tag of each field it names.",
		},
		{
			Kind: KindCheck,
			Name: "TestJoinedSpec",
			Does: "Holds the joined spec to its contract, internal/apicontract/testdata/joined.golden.",
			When: "When you change the API on purpose, run it with -update and keep the diff.",
		},
		{
			Kind: KindCheck,
			Name: "osapi controller endpoint openapi get --name <group>",
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
			Name: "osapi controller endpoint create --name <group> --description \"...\"",
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
			Kind: KindGenerated,
			Name: "internal/controller/endpoints/<group>/gen/",
			Does: "The routes, decoders, CLI commands and spec of the group, from go generate ./internal/controller/endpoints/all.",
			When: "Change the definition, then generate. A generated file that is wrong is fixed in the definition or in internal/controller/endpoint/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/controller/endpoints/<group>/mocks/",
			Does: "The doubles of the group's Deps, from go generate ./internal/controller/endpoints/<group>/mocks.",
			When: "Run it again after deps.go changes.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/controller/endpoints/all/gen/",
			Does: "The import of each group's gen package, so the controller registers its routes and commands, from go generate ./internal/controller/endpoints/all.",
			When: "Change the definition, then generate. A generated file that is wrong is fixed in the definition or in internal/controller/endpoint/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "docs/docs/sidebar/usage/cli/",
			Does: "The CLI pages, from just cli-docs-generate. The help of an endpoint command comes from .Command, .Example, .Notes, .Result and each input field's doc.",
			When: "Change the definition, then generate. A generated file that is wrong is fixed in the definition or in internal/controller/endpoint/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/controller/api/gen/, pkg/sdk/client/gen/, ui/src/sdk/gen/, docs/docs/gen/",
			Does: "The joined spec, the SDK, the UI client and the API pages, from just generate.",
			When: "Change the definition, then generate. A generated file that is wrong is fixed in the definition or in internal/controller/endpoint/gen.",
		},
	},
}
