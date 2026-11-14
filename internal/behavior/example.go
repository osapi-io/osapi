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

import "strings"

// Example is one sample call of a method, used for the docs pages.
type Example struct {
	// Doc says what the example shows.
	Doc string `json:"doc"`
	// Input is a value of the method's input type.
	Input any `json:"input"`
	// Output is a value of the method's output type, as one host would
	// return it.
	Output any `json:"output,omitempty"`
	// Given is the host the example runs on, and Effects what the call
	// changes on it. The provider tests run every example against a fake host
	// built from Given and check its Output and Effects; see HostOption.
	Given   []Given  `json:"given,omitempty"`
	Effects []Effect `json:"effects,omitempty"`
}

// Refusal is a call a method fails: an input it rejects, or a host it cannot
// act on. The provider tests run it on a fake host set up as Given says and
// check the error contains Error.
type Refusal struct {
	Example
	// Error is text the error contains.
	Error string `json:"error"`
}

// Fails declares a call the method fails, with text its error contains, such
// as a key with a slash in it or a parameter osapi does not manage. host sets
// up the fake host the call runs on, as for Example. The docs page lists what
// a method refuses, and the provider tests check each.
func (m *Method) Fails(
	doc string,
	input any,
	errorText string,
	host ...HostOption,
) *Method {
	r := Refusal{Example: Example{Doc: doc, Input: input}, Error: errorText}
	for _, o := range host {
		o.applyExample(&r.Example)
	}
	m.Refusals = append(m.Refusals, r)

	return m
}

// Example adds a sample call for the docs pages.
func (m *Method) Example(
	doc string,
	input any,
	output any,
	host ...HostOption,
) *Method {
	e := Example{Doc: doc, Input: input, Output: output}
	for _, o := range host {
		o.applyExample(&e)
	}
	m.Examples = append(m.Examples, e)

	return m
}

// A HostOption describes the host an example runs on, or what the call
// changes there. The provider tests build a fake host from an example's
// options, run the method on it with each platform's provider, and check
// the example's output and effects, so the example a docs page shows is
// also the test that proves it.
type HostOption interface {
	applyExample(e *Example)
}

// GivenKind is what a Given sets up.
type GivenKind string

const (
	// GivenCall runs another of the behavior's methods first.
	GivenCall GivenKind = "call"
	// GivenOutput is what a command prints.
	GivenOutput GivenKind = "output"
	// GivenFile is a file already on the host.
	GivenFile GivenKind = "file"
)

// Given is one thing true of the host before an example runs.
type Given struct {
	Kind GivenKind `json:"kind"`
	// Input is the call's input, for GivenCall; the method is the one whose
	// input type it is.
	Input any `json:"input,omitempty"`
	// Command is the command line, for GivenOutput.
	Command string `json:"command,omitempty"`
	// Path and Content are the file, for GivenFile; Content is also the
	// command's output, for GivenOutput.
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	// After is the command line that must have run before Command prints
	// Content, for a GivenOutput declared with PrintsAfter.
	After string `json:"after,omitempty"`
}

func (g Given) applyExample(
	e *Example,
) {
	e.Given = append(e.Given, g)
}

// EffectKind is what an Effect checks.
type EffectKind string

const (
	// EffectWrote checks a file holds exactly some content.
	EffectWrote EffectKind = "wrote"
	// EffectRemoved checks a file is gone.
	EffectRemoved EffectKind = "removed"
	// EffectRan checks a command ran.
	EffectRan EffectKind = "ran"
)

// Effect is one change an example makes on the host.
type Effect struct {
	Kind EffectKind `json:"kind"`
	// Path is the file, for EffectWrote and EffectRemoved.
	Path string `json:"path,omitempty"`
	// Content is the file's whole content, for EffectWrote.
	Content string `json:"content,omitempty"`
	// Command is the command line, for EffectRan.
	Command string `json:"command,omitempty"`
}

func (f Effect) applyExample(
	e *Example,
) {
	e.Effects = append(e.Effects, f)
}

// After runs another of the behavior's methods with input before the
// example: the method whose input type input is. Use it for the state a read
// or a change needs, such as a parameter that must already be managed.
func After(
	input any,
) Given {
	return Given{Kind: GivenCall, Input: input}
}

// Prints says a command prints output when it runs.
func Prints(
	command string,
	output string,
) Given {
	return Given{Kind: GivenOutput, Command: command, Content: output}
}

// PrintsAfter says a command prints output once another command line, after,
// has run. It is for state a program keeps rather than a file: a package
// query that prints "not installed" until the install runs, and "installed"
// from then on. Declare the first answer with Prints and the later one with
// PrintsAfter.
func PrintsAfter(
	after string,
	command string,
	output string,
) Given {
	return Given{Kind: GivenOutput, Command: command, Content: output, After: after}
}

// Exists says a file holds content before the example runs.
func Exists(
	path string,
	content string,
) Given {
	return Given{Kind: GivenFile, Path: path, Content: content}
}

// Wrote says the call leaves path holding exactly content.
func Wrote(
	path string,
	content string,
) Effect {
	return Effect{Kind: EffectWrote, Path: path, Content: content}
}

// Removed says the call leaves path absent.
func Removed(
	path string,
) Effect {
	return Effect{Kind: EffectRemoved, Path: path}
}

// Ran says the call runs a command, with or without sudo.
func Ran(
	name string,
	args ...string,
) Effect {
	return Effect{Kind: EffectRan, Command: strings.Join(append([]string{name}, args...), " ")}
}
