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
	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// Checks are the rules that Validate applies, in the order that a behavior is
// usually written. behavior validate --help prints each one, and each problem
// that Validate reports carries the fix of its rule. Thus an author, or an
// agent, can read the rules before it breaks one. Each Rule constant has one
// entry; a test fails on a rule without one.
var Checks = guide.Checks{
	{
		Rule:     RuleName,
		Requires: "The behavior, its group and each method have a name of lowercase letters and digits, starting with a letter: the generators make Go names of it, so a dash does not compile. No behavior or group is named all.",
		Fix:      "Rename it, such as sysctl or updatepackages. A name that is a Go keyword, such as package, takes osapi behavior create --package for its Go package.",
	},
	{
		Rule:     RulePlacement,
		Requires: "The behavior has a placement.",
		Fix:      "Add behavior.Node to Define.",
	},
	{
		Rule:     RuleDoc,
		Requires: "The behavior and each method have a description. No description or .Explain text has < > { or }: the docs site reads them as markup, and the page fails to build.",
		Fix:      "Give Define a one-line description, and each method .Describe(\"...\"). Name a placeholder with an example, such as osapi-dns-eth0.yaml for eth0, not osapi-dns-<interface>.yaml.",
	},
	{
		Rule:     RuleMethods,
		Requires: "The behavior has one or more methods.",
		Fix:      "Add behavior.Read or behavior.Write to Define.",
	},
	{
		Rule:     RuleDuplicate,
		Requires: "Each behavior is registered one time, and each method name is used one time.",
		Fix:      "Rename or remove the second one.",
	},
	{
		Rule:     RuleVerb,
		Requires: "A read is named list, get or status. A write is named create, update or delete. An action, declared with Action, is named for what it does with one of the CLI's actions, such as restart, and not with a standard verb.",
		Fix:      "Rename the method to the standard verb that the message names, or declare a write named for what it does with behavior.Action and a name from osapi cli command spec get.",
	},
	{
		Rule:     RuleRoute,
		Requires: "No two methods have the same route.",
		Fix:      "Give one of them a key with .Key(\"field\").",
	},
	{
		Rule:     RuleKey,
		Requires: "The key of a method is a field of its input, and is required: it is a path segment, which is never empty. Its name in the spec, the TypeName and its title, is not also the name of a row type of the behavior.",
		Fix:      "Give .Key the JSON name of an input field, or add the field to the input, and tag it validate:\"required,...\". When the message names a type, rename the type, such as Part to PartEntry.",
	},
	{
		Rule:     RuleParent,
		Requires: "A behavior declared with Of is under a registered behavior whose items have a key, such as name for a user, and is not also under a group: the parent's path is its group.",
		Fix:      "Give Of a behavior that has a keyed method, such as user, and drop Under; the child takes the parent's path, command and permissions.",
	},
	{
		Rule:     RuleSchema,
		Requires: "Each input is a struct, and each input and output type has a JSON Schema. A struct that an input or output holds is named in one spec for every behavior, so no two behaviors have a different type of the same name.",
		Fix:      "Use structs of strings, numbers, bools, lists and structs. Declare a custom validate tag with behavior.Validator. Rename a held struct that another behavior's type already names, such as Usage to MemoryUsage, or hold that behavior's type.",
	},
	{
		Rule:     RuleOutputShape,
		Requires: "Each output is a struct or a list of structs, with no field named hostname, status or error. Each field has a json name, no field is an array, and each render tag is one the output knows (osapi cli command spec get lists them).",
		Fix:      "Put the value in a struct, rename a field that has one of those names, tag each field json:\"name\", use a slice for an array, and use a render value the output knows.",
	},
	{
		Rule:     RuleChanged,
		Requires: "The output of a write has Changed bool with json:\"changed\".",
		Fix:      "Add Changed bool `json:\"changed\"` to the output, and set it to true only when the call changed the host.",
	},
	{
		Rule:     RuleSpan,
		Requires: "A write records on its span the item that it acts on, so that a trace shows what each call changed.",
		Fix:      "Name the item with .Key(\"field\"). A method without a key, such as a create, tags that field telemetry:\"attr\".",
	},
	{
		Rule:     RuleTemplate,
		Requires: "A field tagged template:\"<vars>\" is a string, and <vars> is the JSON name of an input field of type map[string]string.",
		Fix:      "Tag the string field that holds the template with template:\"vars\", and add Vars map[string]string `json:\"vars\"` to the input.",
	},
	{
		Rule:     RuleValidate,
		Requires: rule.Requires,
		Fix:      rule.Fix,
	},
	{
		Rule:     RuleOptional,
		Requires: "An optional field of a write's request body, with omitempty in its validate tag, has omitempty in its json tag too, so the client leaves it out when it is empty. The endpoint DSL owns the rule; controller endpoint validate checks it the same way.",
		Fix:      "Add ,omitempty to the field's json tag, such as json:\"version,omitempty\".",
	},
	{
		Rule:     RuleInput,
		Requires: "Each input field has a json name and a doc, has no default, which the server does not apply, is not an array, and is not named hostname or target: every request has the target host as its hostname path segment and its --target flag.",
		Fix:      "Tag the field json:\"name\" and doc:\"...\", drop the default tag and make the provider treat the empty value as the default, use a slice for an array, and rename a field called hostname or target.",
	},
	{
		Rule:     RuleFails,
		Requires: rule.FailsRequires,
		Fix:      rule.FailsFix,
	},
	{
		Rule:     RuleSecret,
		Requires: "No input field is both secret:\"true\" and telemetry:\"attr\".",
		Fix:      "Remove telemetry:\"attr\" from the secret field. A secret is never recorded on a span.",
	},
	{
		Rule:     RuleSurface,
		Requires: "A behavior with CLI also has API, and a behavior with API has an Area of the UI. An internal behavior has no API or CLI. Each input field of a CLI behavior has a type a flag holds: string, bool, int, []string, map[string]string or []byte (a file). The CLI owns the list; its generator builds each.",
		Fix:      "Add behavior.API, add behavior.Area(...) to a served behavior, remove API and CLI from the internal behavior, or give the field a type that a flag can hold.",
	},
	{
		Rule:     RuleExample,
		Requires: "Each method of a CLI behavior has an example. Each example and refusal has a description, and the input and output types of its method. A refusal has error text and no effects.",
		Fix:      "Add .Example(\"...\", In{...}, Out{...}) to the method, and correct the example or refusal that the message names.",
	},
	{
		Rule:     RuleSettings,
		Requires: "Settings are a struct. Each field has a mapstructure name and a doc.",
		Fix:      "Add mapstructure:\"name\" and doc:\"...\" to each settings field.",
	},
	{
		Rule:     RulePlatform,
		Requires: "The behavior has a provider for one or more platforms, each platform one time, and each provider implements the provider type of each method.",
		Fix:      "Add behavior.Platform(\"debian\", newDebian), remove the second declaration, or add the missing method to the provider.",
	},
	{
		Rule: RuleRequires,
		Requires: ChainChecks.Requires(
			RuleRequires,
		) + " The chain layer owns this rule: osapi behavior chain validate.",
		Fix: ChainChecks.Fix(RuleRequires),
	},
	{
		Rule: RuleInternal,
		Requires: ChainChecks.Requires(
			RuleInternal,
		) + " The chain layer owns this rule: osapi behavior chain validate.",
		Fix: ChainChecks.Fix(RuleInternal),
	},
	{
		Rule:     RuleWrites,
		Requires: "Each path given to Writes is \"*\" or an absolute path.Match pattern.",
		Fix:      "Use a pattern such as \"/etc/sysctl.d/osapi-*.conf\".",
	},
}

// withFixes sets the fix of each problem from its rule.
func withFixes(
	problems []Problem,
) []Problem {
	for i := range problems {
		problems[i].Fix = Checks.Fix(problems[i].Rule)
	}

	return problems
}

// The checks that behaviortest.Run makes when go test runs a behavior's
// examples on a fake host, by the name in TestChecks.
const (
	TestOutput     = "output"
	TestEffects    = "effects"
	TestIdempotent = "idempotent"
	TestErrors     = "errors"
	TestState      = "state"
	TestRefusal    = "refusal"
	TestTraversal  = "traversal"
	TestNewLine    = "newline"
	TestOption     = "option"
	TestShell      = "shell"
	TestRuns       = "runs"
	TestGohai      = "gohai"
	TestWrites     = "writes"
	TestFileSpan   = "file-span"
	TestKeepsState = "keeps-state"
	TestCall       = "call"
	TestPlatforms  = "platforms"
)

// TestChecks are what the tests of every behavior check, when go test runs
// its examples and refusals on a fake host. behavior validate --help prints
// each one, and the message of each failure ends with its fix.
var TestChecks = guide.Checks{
	{
		Rule:     TestPlatforms,
		Requires: "Each platform that a method serves runs at least one of its examples.",
		Fix:      "Add an example for the platform, or name it in an example's OnPlatform.",
	},
	{
		Rule:     TestOutput,
		Requires: "Each example runs without an error, and returns the output that it declares.",
		Fix:      "Correct the output of the example, or the provider that returns a different one.",
	},
	{
		Rule:     TestEffects,
		Requires: "An example of a write that changes the host declares its effects, and the host has each one after the call.",
		Fix:      "Add Wrote, Removed or Ran to the example, or correct the effect that the provider did not make.",
	},
	{
		Rule:     TestIdempotent,
		Requires: "A second call with the same input reports changed: false.",
		Fix:      "Compare the host with the input before the provider acts, and change nothing when they agree.",
	},
	{
		Rule:     TestErrors,
		Requires: "The provider does not panic when a file, command or store call fails.",
		Fix:      "Check each error, and return it. Do not use the zero value of a call that failed.",
	},
	{
		Rule:     TestState,
		Requires: "The provider does not panic when a record in the file-state bucket is damaged.",
		Fix:      "Check the record before the provider uses it, and return an error when it is not complete.",
	},
	{
		Rule:     TestRefusal,
		Requires: "Each refusal declared with .Fails fails, with an error that contains its text.",
		Fix:      "Refuse the input in the provider or with a validate tag, or correct the error text of the refusal.",
	},
	{
		Rule:     TestTraversal,
		Requires: "No string input writes a file outside the directories that the example writes.",
		Fix:      "Add a validate pattern to the field, or check it in the provider and declare the refusal with .Fails.",
	},
	{
		Rule:     TestNewLine,
		Requires: "No string input adds a line to a file.",
		Fix:      "Refuse \r and \n in the field, with the validate tag no_linebreak.",
	},
	{
		Rule:     TestOption,
		Requires: "No string input reaches a command as an option.",
		Fix:      "Refuse a leading - in the field, or pass -- before it.",
	},
	{
		Rule:     TestShell,
		Requires: "No input reaches a shell script.",
		Fix:      "Run the program with its arguments, not through sh -c.",
	},
	{
		Rule:     TestRuns,
		Requires: "Each program that the provider runs is declared.",
		Fix:      "Add behavior.Runs(\"program\"), or behavior.Sudo(\"program\") for a program run as root.",
	},
	{
		Rule:     TestGohai,
		Requires: "A read runs each command through gohai: it reads the facts of the host with a gohai collector and h.Sources(), not with h.Exec.",
		Fix:      "Read the fact with the gohai collector that has it, such as services.NewWith(h.Sources()). A fact that gohai does not read is a change to gohai first. A read of what is not a fact of the host, such as the updates of the package index, declares why with .NotAFact(\"...\").",
	},
	{
		Rule:     TestWrites,
		Requires: "Each path that the provider writes, removes or changes the mode of is declared, also through a required behavior.",
		Fix:      "Add a pattern that covers the path to behavior.Writes.",
	},
	{
		Rule:     TestFileSpan,
		Requires: "Each change to a file goes through the atomicfile package, which writes it atomically and makes its file.write, file.remove or file.chmod span.",
		Fix:      "For a file that osapi manages, require managedfile and call it with behavior.Call. Otherwise use atomicfile.Write, atomicfile.Remove or atomicfile.Chmod with the ctx of the method, not h.FS.",
	},
	{
		Rule:     TestKeepsState,
		Requires: "A behavior that uses the file-state bucket declares it.",
		Fix:      "Add behavior.KeepsState, or require the behavior that keeps the record.",
	},
	{
		Rule:     TestCall,
		Requires: "A provider calls the methods of a required behavior through behavior.Call, so that the trace shows each call as a step.",
		Fix:      "Call it with behavior.Call(ctx, other.Behavior, other.Method, provider, in).",
	},
}

// TestFailure is message, followed by the fix of the test check rule.
func TestFailure(
	rule string,
	message string,
) string {
	return message + ". Fix: " + TestChecks.Fix(rule)
}
