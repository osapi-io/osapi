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

import "github.com/osapi-io/osapi/internal/guide"

// Checks are the rules that Validate applies, in the order that a behavior is
// usually written. behavior validate --help prints each one, and each problem
// that Validate reports carries the fix of its rule. Thus an author, or an
// agent, can read the rules before it breaks one. Each Rule constant has one
// entry; a test fails on a rule without one.
var Checks = guide.Checks{
	{
		Rule:     RuleName,
		Requires: "The behavior, its group and each method have a name of lowercase letters, digits and dashes. No behavior or group is named all.",
		Fix:      "Rename it, such as sysctl or ip-forward.",
	},
	{
		Rule:     RulePlacement,
		Requires: "The behavior has a placement.",
		Fix:      "Add behavior.Node to Define.",
	},
	{
		Rule:     RuleDoc,
		Requires: "The behavior and each method have a description.",
		Fix:      "Give Define a one-line description, and each method .Describe(\"...\").",
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
		Requires: "A read is named list, get or status. A write is named create, update or delete.",
		Fix:      "Rename the method to the standard verb that the message names.",
	},
	{
		Rule:     RuleRoute,
		Requires: "No two methods have the same route.",
		Fix:      "Give one of them a key with .Key(\"field\").",
	},
	{
		Rule:     RuleKey,
		Requires: "The key of a method is a field of its input.",
		Fix:      "Give .Key the JSON name of an input field, or add the field to the input.",
	},
	{
		Rule:     RuleSchema,
		Requires: "Each input is a struct, and each input and output type has a JSON Schema.",
		Fix:      "Use structs of strings, numbers, bools, lists and structs. Declare a custom validate tag with behavior.Validator.",
	},
	{
		Rule:     RuleOutputShape,
		Requires: "Each output is a struct or a list of structs, with no field named hostname, status or error.",
		Fix:      "Put the value in a struct, and rename a field that has one of those names.",
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
		Rule:     RuleSecret,
		Requires: "No input field is both secret:\"true\" and telemetry:\"attr\".",
		Fix:      "Remove telemetry:\"attr\" from the secret field. A secret is never recorded on a span.",
	},
	{
		Rule:     RuleSurface,
		Requires: "A behavior with CLI also has API. An internal behavior has no API or CLI. Each input field of a CLI behavior is a string, bool or number.",
		Fix:      "Add behavior.API, remove API and CLI from the internal behavior, or give the field a type that a flag can hold.",
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
		Rule:     RuleRequires,
		Requires: "Each required behavior is registered, and no behavior requires itself through others.",
		Fix:      "Import the required behavior in internal/behaviors/all, or remove the requirement that makes the cycle.",
	},
	{
		Rule:     RuleInternal,
		Requires: "Another behavior requires each internal behavior. Nothing else can call it, because it has no API and no CLI.",
		Fix:      "Add behavior.Requires(it) to the behavior that uses it, or remove behavior.Internal if an operator calls it.",
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
	TestWrites     = "writes"
	TestFileSpan   = "file-span"
	TestKeepsState = "keeps-state"
	TestCall       = "call"
)

// TestChecks are what the tests of every behavior check, when go test runs
// its examples and refusals on a fake host. behavior validate --help prints
// each one, and the message of each failure ends with its fix.
var TestChecks = guide.Checks{
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
