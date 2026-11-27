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
	"slices"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

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
	// Platforms limits the example to these platforms, as Platform names
	// them; OnPlatform sets it. An example with none runs on each platform
	// the method serves.
	Platforms []PlatformName `json:"platforms,omitempty"`
	// NotLive is why the example runs only on the fake host, not on the
	// live host of the integration test; NotLive sets it. Empty runs it on
	// both.
	NotLive string `json:"not_live,omitempty"`
	// ReadBacks are the reads that show what a write example changed, which
	// Reads and ReadsNone add. The integration test runs each one on the
	// live host after the example.
	ReadBacks []ReadBack `json:"read_backs,omitempty"`
	// NoReadBack is why no read shows what a write example changed, which
	// NoReadBack sets.
	NoReadBack string `json:"no_read_back,omitempty"`
	// Live are the fields of Input that the integration test fills with a
	// fact of the live target, which Live sets.
	Live []LiveField `json:"live,omitempty"`
}

// Fact is a value of the live target that an example's input can take in
// the integration test, in place of the value that it names for the fake
// host.
type Fact string

const (
	// FactInterface is the target's primary network interface, such as
	// enp0s1, as agent get reports it.
	FactInterface Fact = "interface"
	// FactAdminGroup is the group of the administrator: root on Linux,
	// wheel on macOS.
	FactAdminGroup Fact = "admin_group"
	// FactUser is a user that each host has: root.
	FactUser Fact = "user"
	// FactPID is a process that each host runs: 1, the init process.
	FactPID Fact = "pid"
	// FactSSHService is the SSH server's service: ssh on the Debian
	// family, sshd on the others.
	FactSSHService Fact = "ssh_service"
	// FactSSHUnit is FactSSHService's systemd unit, such as ssh.service.
	FactSSHUnit Fact = "ssh_unit"
)

// LiveField is a field of an example's input, by its JSON name, and the
// fact of the live target that the integration test puts in it.
type LiveField struct {
	Field string `json:"field"`
	Fact  Fact   `json:"fact"`
}

// Live says that on the live target of the integration test, the input's
// field, by its JSON name, takes fact in place of the value that the
// example names for the fake host: an eth0 that the target does not have
// becomes its own interface. The provider tests keep the example's value.
func Live(
	field string,
	fact Fact,
) HostOption {
	return exampleLive(LiveField{Field: field, Fact: fact})
}

// exampleLive is the HostOption of Live.
type exampleLive LiveField

func (l exampleLive) applyExample(
	e *Example,
) {
	e.Live = append(e.Live, LiveField(l))
}

// ReadBack is a read of the behavior, or of its parent, that shows what a
// write example changed: the read whose input type Input is, run with
// Input, and the fields of Want that it must answer. With None, no row
// answers them: the example removed the item.
type ReadBack struct {
	// Input is a value of the read's input type.
	Input any `json:"input"`
	// Want is a value of the read's output type, or of its element for a
	// read that answers a list.
	Want any `json:"want"`
	// Fields are the JSON names of the fields of Want to compare. With
	// none, each field of Want that is not its zero value.
	Fields []string `json:"fields,omitempty"`
	// None is a read that no row of answers Want: the item is gone, or the
	// read reports that it is not found.
	None bool `json:"none,omitempty"`
}

// Reads says what a read answers on the live host after a write example:
// the read whose input type input is, run with input, answers a row with
// the fields of want that fields names, or with none, each field of want
// that is not its zero value. Name a field to compare its zero value, such
// as managed for false. A list matches a list that has each of its values. The read can be the parent's, as a key's read for a
// user. The integration test runs it, so a write that reports changed but
// does nothing fails.
func Reads(
	input any,
	want any,
	fields ...string,
) HostOption {
	return exampleReadBack(ReadBack{Input: input, Want: want, Fields: fields})
}

// ReadsNone says that after a write example no row of the read whose input
// type input is answers want, or the read reports the item not found: the
// example removed it. fields names the fields of want to compare, as for
// Reads.
func ReadsNone(
	input any,
	want any,
	fields ...string,
) HostOption {
	return exampleReadBack(ReadBack{Input: input, Want: want, Fields: fields, None: true})
}

// exampleReadBack is the HostOption of Reads and ReadsNone.
type exampleReadBack ReadBack

func (r exampleReadBack) applyExample(
	e *Example,
) {
	e.ReadBacks = append(e.ReadBacks, ReadBack(r))
}

// NoReadBack says why no read of the behavior shows what a write example
// changed, such as a command's run, whose output is its result. The gate
// takes it in place of a Reads.
func NoReadBack(
	why string,
) HostOption {
	return exampleNoReadBack(why)
}

// exampleNoReadBack is the HostOption of NoReadBack.
type exampleNoReadBack string

func (n exampleNoReadBack) applyExample(
	e *Example,
) {
	e.NoReadBack = string(n)
}

// NotLive keeps a write example off the live host of the integration test,
// for one whose state a real host cannot be given through the behavior's
// own methods, such as the process that a signal stops: After can only call
// the behavior's methods, and none of process's starts a process. why says
// so, and the docs page shows it. The provider tests still run it on the
// fake host.
func NotLive(
	why string,
) HostOption {
	return exampleNotLive(why)
}

// exampleNotLive is the HostOption of NotLive.
type exampleNotLive string

func (n exampleNotLive) applyExample(
	e *Example,
) {
	e.NotLive = string(n)
}

// OnPlatform limits an example to platforms, for a read whose answer is
// the platform's own, such as the name of the operating system: an Ubuntu
// host reads ubuntu, and a Mac reads darwin. The provider tests run the
// example only with these platforms' providers, and each platform the
// method serves must still have an example.
func OnPlatform(
	platforms ...PlatformName,
) HostOption {
	return examplePlatforms(platforms)
}

// examplePlatforms is the HostOption of OnPlatform.
type examplePlatforms []PlatformName

func (p examplePlatforms) applyExample(
	e *Example,
) {
	e.Platforms = append(e.Platforms, p...)
}

// RunsOn reports whether the example runs on platform.
func (e Example) RunsOn(
	platform PlatformName,
) bool {
	return len(e.Platforms) == 0 || slices.Contains(e.Platforms, platform)
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
	// GivenLink is a symbolic link already on the host.
	GivenLink GivenKind = "link"
	// GivenObject is an object in the object store of the fleet.
	GivenObject GivenKind = "object"
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
	// Content, for a GivenOutput declared with PrintsAfter, or before the
	// file holds Content, for a GivenFile declared with ExistsAfter.
	After string `json:"after,omitempty"`
	// ExitCode is the command's exit code, for GivenOutput declared with
	// Exits.
	ExitCode int `json:"exit_code,omitempty"`
	// ContentType is the object's content type, for GivenObject, whose name
	// is Path.
	ContentType types.ContentType `json:"content_type,omitempty"`
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
// or a change needs, such as a parameter that must already be managed. The
// method can be the parent's, for a behavior that is Of one: the keys of a
// user need the user.
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

// Exits says a command exits with code when it runs, as RunCmdFull returns
// it. A command with no Exits exits 0.
func Exits(
	command string,
	code int,
) Given {
	return Given{Kind: GivenOutput, Command: command, ExitCode: code}
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

// Stored says the object store of the fleet has the object name, with
// content and the content type its uploader gave it, raw or template, as
// osapi client file upload stores it.
func Stored(
	name string,
	content string,
	contentType types.ContentType,
) Given {
	return Given{Kind: GivenObject, Path: name, Content: content, ContentType: contentType}
}

// Exists says a file holds content before the example runs.
func Exists(
	path string,
	content string,
) Given {
	return Given{Kind: GivenFile, Path: path, Content: content}
}

// ExistsAfter says a file holds content once a command line, after, has
// run: a file that a program changes, such as /etc/group after groupadd.
// Declare the file before with Exists and its later content with
// ExistsAfter, as PrintsAfter does for a command's output.
func ExistsAfter(
	after string,
	path string,
	content string,
) Given {
	return Given{Kind: GivenFile, Path: path, Content: content, After: after}
}

// Links gives the host a symbolic link at path to target before the call,
// such as /etc/localtime to the zone file of the host's timezone.
func Links(
	path string,
	target string,
) Given {
	return Given{Kind: GivenLink, Path: path, Content: target}
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
