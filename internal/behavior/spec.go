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

package behavior

import (
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
)

// The kinds of term in Spec, in the order that a behavior is written.
const (
	KindDefinition = "definition"
	KindOption     = "option"
	KindMethod     = "method"
	KindRefine     = "refine"
	KindExample    = "example"
	KindCall       = "call"
	KindCompose    = "compose"
	KindBlock      = "block"
	KindTag        = "tag"
	KindTemplate   = "template"
	KindGenerated  = "generated"
	KindNotYet     = "not-yet"
)

// Spec is every term of the behavior DSL. osapi behavior spec prints it, so
// an author, or an agent, learns the DSL from the binary that it builds. A
// test fails on an exported option, method, refinement or example part of
// this package without an entry, and on a field tag that this package reads
// without one.
var Spec = guide.Spec{
	Kinds: []guide.Kind{
		{Kind: KindDefinition, Title: "Definition"},
		{Kind: KindOption, Title: "Options of Define"},
		{Kind: KindMethod, Title: "Methods"},
		{Kind: KindRefine, Title: "Refinements of a method"},
		{Kind: KindExample, Title: "Hosts and effects of an example"},
		{Kind: KindCall, Title: "Calls in a provider"},
		{Kind: KindCompose, Title: "Building on other behaviors"},
		{Kind: KindTag, Title: "Tags on input, output and settings fields"},
		{Kind: KindTemplate, Title: "What a template reads"},
		{Kind: KindGenerated, Title: "Generated, never edited"},
		{Kind: KindNotYet, Title: "Not supported yet"},
	},
	Terms: []guide.Term{
		{
			Kind: KindDefinition,
			Name: `Register(Define(name, doc, options...))`,
			Does: "Declares the behavior and adds it to the registry. The agent, the server, the CLI and the generator read the registry.",
			When: "One time, in behavior.go: var Behavior = behavior.Register(behavior.Define(...)).",
		},

		{
			Kind: KindOption,
			Name: "Node",
			Does: "Runs the methods on a host, through the job system.",
			When: "Always. It is the only placement.",
		},
		{
			Kind: KindOption,
			Name: "API",
			Does: "Serves the methods over REST, with routes, permissions and the OpenAPI spec.",
			When: "When an operator or a program calls the behavior.",
		},
		{
			Kind: KindOption,
			Name: "CLI",
			Does: "Makes osapi client commands and their docs pages. Needs API.",
			When: "When an operator calls the behavior from a shell.",
		},
		{
			Kind: KindOption,
			Name: `Under("group")`,
			Does: "Nests the behavior in a group: its routes, commands and directory.",
			When: "When the behavior belongs to a group, such as network.",
		},
		{
			Kind: KindOption,
			Name: "Internal",
			Does: "Gives the behavior no route, command or permission. Only the behaviors that require it can call it.",
			When: "For work that two behaviors share, such as writing a file.",
		},
		{
			Kind: KindOption,
			Name: "Requires(behaviors...)",
			Does: "Builds the providers of those behaviors first. The builder gets them with host.Require.",
			When: "When the behavior uses the methods of another behavior.",
		},
		{
			Kind: KindOption,
			Name: `Platform("debian", build)`,
			Does: "Builds the provider on that platform: a distribution, darwin, or linux for each Linux host without its own. On another platform, the methods report skipped.",
			When: "One time for each platform that the behavior supports.",
		},
		{
			Kind: KindOption,
			Name: "Sudo(programs...)",
			Does: "Lists the programs that the providers run with sudo. The agent preflight checks that sudo allows each.",
			When: "For each program that needs root.",
		},
		{
			Kind: KindOption,
			Name: "Runs(programs...)",
			Does: "Lists the programs that the providers run without sudo. The agent warns when one is not installed.",
			When: "For each program that does not need root.",
		},
		{
			Kind: KindOption,
			Name: `Writes("/etc/x.d/osapi-*.conf")`,
			Does: "Lists the paths that the providers write, remove or change the mode of. Each such call is a file span in a trace. \"*\" is the path that the caller names.",
			When: "For each path, also when a required behavior, such as managedfile, does the write.",
		},
		{
			Kind: KindOption,
			Name: "KeepsState",
			Does: "Says that the providers keep a record in the file-state bucket. Each read and write of it is a kv span.",
			When: "When the behavior keeps a record of what it wrote.",
		},
		{
			Kind: KindOption,
			Name: "Config(defaults)",
			Does: "Declares settings under behaviors.<name>, with a flag and an environment variable for each. A provider reads them with Settings.",
			When: "For standing config of the agent. A value for one call is an input.",
		},
		{
			Kind: KindOption,
			Name: `Validator("tag", check, schema, message)`,
			Does: "Declares a validate tag of the behavior, with the check that the server runs and the JSON Schema of the spec.",
			When: "When no standard validate tag checks a value.",
		},
		{
			Kind: KindOption,
			Name: "Next",
			Does: "Serves the behavior under /api/next and osapi client next, beside the hand-written domain that it replaces.",
			When: "Only while a behavior replaces a hand-written domain.",
		},

		{
			Kind: KindMethod,
			Name: `Read("name", fn)`,
			Does: "A method that reads the host: GET and <group>:read. Named list, get or status.",
			When: "For each read. fn is func(ctx, p Provider, in In) (Out, error).",
		},
		{
			Kind: KindMethod,
			Name: `Write("name", fn)`,
			Does: "A method that changes the host: POST, PUT or DELETE, and <group>:write. Named create, update or delete. Out has Changed bool.",
			When: "For each change. fn is func(ctx, p Provider, in In) (Out, error).",
		},

		{
			Kind: KindRefine,
			Name: `.Key("field")`,
			Does: "Puts that input field in the route and the CLI, and records it on the span of each call.",
			When: "On each method that acts on one item, such as get, update and delete of a key.",
		},
		{
			Kind: KindRefine,
			Name: `.Describe("text")`,
			Does: "The one-line description of the method, in the help, the spec and the docs.",
			When: "On each method.",
		},
		{
			Kind: KindRefine,
			Name: `.Explain("text")`,
			Does: "The longer description: what the method touches, and what a second call does.",
			When: "When one line does not say what the method changes.",
		},
		{
			Kind: KindRefine,
			Name: `.Example("doc", in, out, host...)`,
			Does: "A sample call. The docs page shows it, and the provider tests run it on a fake host.",
			When: "On each method. A CLI behavior needs one.",
		},
		{
			Kind: KindRefine,
			Name: `.Fails("doc", in, "error text", host...)`,
			Does: "A call that the method refuses. The docs page lists it, and the provider tests check the error.",
			When: "For each input or host that the method refuses.",
		},

		{
			Kind: KindExample,
			Name: "After(input)",
			Does: "Runs the method of that input before the example.",
			When: "When the example needs state, such as a parameter that is already managed.",
		},
		{
			Kind: KindExample,
			Name: `Prints("command", "output")`,
			Does: "Says what a command prints on the fake host.",
			When: "For each command that the provider reads.",
		},
		{
			Kind: KindExample,
			Name: `PrintsAfter("after", "command", "output")`,
			Does: "Says what a command prints after another command has run.",
			When: "For state that a program keeps, not a file.",
		},
		{
			Kind: KindExample,
			Name: `Exists("path", "content")`,
			Does: "Puts a file on the fake host before the call.",
			When: "When the call reads or changes a file that is there.",
		},
		{
			Kind: KindExample,
			Name: `Wrote("path", "content")`,
			Does: "Says that the call leaves the file with that content.",
			When: "For each file that the call writes.",
		},
		{
			Kind: KindExample,
			Name: `Removed("path")`,
			Does: "Says that the call removes the file.",
			When: "For each file that the call removes.",
		},
		{
			Kind: KindExample,
			Name: `Ran("program", args...)`,
			Does: "Says that the call runs the command.",
			When: "For each command that the call runs to make a change.",
		},

		{
			Kind: KindCompose,
			Name: "Use a building block",
			Does: "A building block is an internal behavior that does one generic job, such as writing a file that osapi manages. Its spans, checks and records come with it. \"Building blocks\" lists the ones in this binary.",
			When: "Before a provider does generic work itself: a managed file, a service reload, a call to an outside API. Require the block with Requires, and call it with Call.",
		},
		{
			Kind: KindCompose,
			Name: "Make a building block",
			Does: "Generic work that no block does yet becomes a new internal behavior. The behavior that needs it requires it.",
			When: "When the work is not about the one thing that the behavior manages, or a second behavior would need it.",
		},
		{
			Kind: KindCompose,
			Name: "One host, one step",
			Does: "A behavior acts on one host, in one job. Its building blocks run on the same host, in the same job.",
			When: "Work across hosts, or in ordered steps, is an osapi-orchestrator workflow, not a behavior.",
		},
		{
			Kind: KindCall,
			Name: "Call(ctx, other.Behavior, other.Method, p, in)",
			Does: "Runs a method of a required behavior in its own span, so the trace shows it as a step.",
			When: "Each time a provider uses a required behavior. A direct call fails the tests.",
		},
		{
			Kind: KindCall,
			Name: "atomicfile.Write(ctx, h.FS, path, content, mode)",
			Does: "Writes the file atomically, in a file.write span. atomicfile.Remove and atomicfile.Chmod make file.remove and file.chmod spans.",
			When: "In a building block that writes files. A behavior that manages a file requires managedfile instead. A change through h.FS fails the tests.",
		},
		{
			Kind: KindCall,
			Name: "Render(text, TemplateData{Facts, Vars, Hostname})",
			Does: "Renders a template with the facts of the agent, the vars of the request and the hostname. A key that the data does not have is an error.",
			When: "In the provider of a method with a template input. The input is checked at the API already.",
		},
		{
			Kind: KindCall,
			Name: "Settings[T](Behavior)",
			Does: "Returns the settings of the behavior, or the defaults.",
			When: "In a builder or a method that reads a setting.",
		},
		{
			Kind: KindCall,
			Name: "host.Require[P](h)",
			Does: "Returns the provider of a required behavior, in a builder.",
			When: "In the builder of a behavior that uses Requires.",
		},
		{
			Kind: KindCall,
			Name: `ValidatorSchema("tag", Pattern(re))`,
			Does: "Declares the JSON Schema of a custom validate tag. Format(f) sets a format.",
			When: "When a validate tag is declared outside Validator.",
		},

		{
			Kind: KindTag,
			Name: `json:"name"`,
			Does: "The name of the field in the request, the flag, the column and --json.",
			When: "On each field.",
		},
		{
			Kind: KindTag,
			Name: `validate:"..."`,
			Does: "Checked by the API before a job is queued. Becomes the constraints of the spec.",
			When: "On each input field that has limits.",
		},
		{
			Kind: KindTag,
			Name: `doc:"..."`,
			Does: "The description in the spec, the flag help and the docs pages.",
			When: "On each field.",
		},
		{
			Kind: KindTag,
			Name: `render:"..."`,
			Does: `How an output column shows: "-" hides it; bytes, age or status.`,
			When: "On an output field that a table must show differently.",
		},
		{
			Kind: KindTag,
			Name: `telemetry:"attr"`,
			Does: "Records the input field on the span of the method. The field that .Key names is recorded without it.",
			When: "On the field that names the item, when the method has no key, and on other fields a trace must show.",
		},
		{
			Kind: KindTag,
			Name: `template:"vars"`,
			Does: "Marks a string input as a Go template, whose .Vars are in the input field vars, a map[string]string. The API refuses a template that does not parse, that reads a field other than .Facts, .Vars and .Hostname, or that reads a var the request does not give. Each such method gets a generated 400 test.",
			When: "On an input that holds a template, such as the content of a file. Render it with Render.",
		},
		{
			Kind: KindTag,
			Name: `secret:"true"`,
			Does: "Redacted from the audit log, write-only in the spec, never recorded on a span.",
			When: "On each input that holds a password, a key or a token.",
		},
		{
			Kind: KindTag,
			Name: `mapstructure:"name"`,
			Does: "The config key of a settings field.",
			When: "On each field of the struct given to Config.",
		},

		{
			Kind: KindTemplate,
			Name: ".Vars.<name>",
			Does: "A value from the vars of the request. The API refuses a template that reads a var that the request does not give.",
			When: "For a value that the caller gives with each call.",
		},
		{
			Kind: KindTemplate,
			Name: ".Hostname",
			Does: "The hostname of the agent that renders the template.",
			When: "For a value that names the host.",
		},
		{
			Kind: KindTemplate,
			Name: ".Facts.<key>",
			Does: "A fact of the agent that renders the template. Facts are known only on each host, so a fact that a host does not have fails on that host.",
			When: "For a value that differs from host to host.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/behaviors/<name>/gen/",
			Does: "The routes, handlers, commands, spec and tests of the behavior, from go generate ./internal/behaviors/all.",
			When: "Change the definition, then generate. A generated test that fails is fixed in the definition or in internal/behavior/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/behaviors/<name>/mocks/",
			Does: "The double of the provider, from go generate ./internal/behaviors/<name>/mocks.",
			When: "Run it again after provider.go changes.",
		},
		{
			Kind: KindGenerated,
			Name: "docs/docs/sidebar/behaviors/<name>/",
			Does: "The docs page of the behavior, from go generate ./internal/behaviors/all.",
			When: "Change the definition, then generate. A generated test that fails is fixed in the definition or in internal/behavior/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "docs/docs/sidebar/usage/cli/",
			Does: "The CLI pages, from just cli-docs-generate. The help of a behavior command comes from .Describe, .Explain, .Example and .Fails.",
			When: "Change the definition, then generate. A generated test that fails is fixed in the definition or in internal/behavior/gen.",
		},
		{
			Kind: KindGenerated,
			Name: "internal/controller/api/gen/, pkg/sdk/client/gen/, ui/src/sdk/gen/, docs/docs/gen/",
			Does: "The joined spec, the SDK, the UI client and the API pages, from just generate.",
			When: "Change the definition, then generate. A generated test that fails is fixed in the definition or in internal/behavior/gen.",
		},
		{
			Kind: KindNotYet,
			Name: "Action methods",
			Does: "A method such as restart or reboot. validate refuses a write that is not create, update or delete.",
			When: "Stop, and tell the person what is missing. Do not build around it.",
		},
		{
			Kind: KindNotYet,
			Name: "A command or a shell from the caller",
			Does: "There is no execute permission.",
			When: "Stop, and tell the person what is missing. Do not build around it.",
		},
		{
			Kind: KindNotYet,
			Name: "A resource under another",
			Does: "Two keys in the path, such as the SSH keys of a user.",
			When: "Stop, and tell the person what is missing. Do not build around it.",
		},
		{
			Kind: KindNotYet,
			Name: "Streaming and uploads",
			Does: "And the object store that they would deploy from.",
			When: "Stop, and tell the person what is missing. Do not build around it.",
		},
		{
			Kind: KindNotYet,
			Name: "A step on the controller",
			Does: "Work before osapi queues the job, such as a password hash.",
			When: "Stop, and tell the person what is missing. Do not build around it.",
		},
	},
}

// CurrentSpec is Spec, with "Building blocks" after "Building on other
// behaviors": each internal behavior in the registry, with what it does and
// its methods. factKeys, the keys that .Facts has, are added to that term.
// Thus both lists agree with the binary.
func CurrentSpec(
	factKeys ...string,
) guide.Spec {
	spec := guide.Spec{Terms: append([]guide.Term(nil), Spec.Terms...)}
	for i, t := range spec.Terms {
		if t.Name == ".Facts.<key>" && len(factKeys) > 0 {
			spec.Terms[i].Does += " The keys: " + strings.Join(factKeys, ", ") + "."
		}
	}
	for _, k := range Spec.Kinds {
		spec.Kinds = append(spec.Kinds, k)
		if k.Kind == KindCompose {
			spec.Kinds = append(spec.Kinds, guide.Kind{Kind: KindBlock, Title: "Building blocks"})
		}
	}

	for _, b := range All() {
		if !b.Internal {
			continue
		}
		names := make([]string, 0, len(b.Methods))
		for _, m := range b.Methods {
			names = append(names, m.Name)
		}
		spec.Terms = append(spec.Terms, guide.Term{
			Kind: KindBlock,
			Name: b.Name,
			Does: b.Doc + ". Methods: " + strings.Join(names, ", ") + ".",
			When: "Require it, and call its methods with Call. osapi behavior get --name " +
				b.Name + " shows its inputs.",
		})
	}

	return spec
}
