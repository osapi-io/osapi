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

// Package behavior defines what osapi can do to a host. A behavior is one
// directory: a definition, typed inputs and outputs, and the logic for each
// platform. The API routes, the agent's dispatch, the CLI commands, the
// OpenAPI document, the config keys and the docs pages are all built from the
// definition.
//
// # Starting one
//
// osapi behavior new --name <name> --description "..." writes the standard
// layout under internal/behaviors/<name>/, a working behavior that, once go
// generate has written its mock and gen/, builds, validates and passes every
// test. internal/behaviors/sysctl is the reference. Every file with code has a test
// named for it.
//
//	behavior.go      the definition, with Define and Register
//	types.go         the input and output types, with their tags
//	methods.go       one function per method, calling the provider
//	provider.go      the Provider interface each platform implements
//	<platform>.go    a platform's provider and its builder, such as debian.go
//	settings.go      optional: the behavior's own config, see Config
//	mocks/           the Provider's double, by mockgen
//	gen/             written by go generate ./internal/behaviors/all: api.go, the
//	                 route table; api_<method>.go, each method's handler;
//	                 cli.go, the commands; api.yaml, the spec; and a test
//	                 for each Go file
//
// Nothing outside the directory changes but its import in internal/behaviors/all,
// which osapi behavior new adds.
//
// # The definition
//
// Define takes a name, a one-line description and options:
//
//	Node          the work runs on a host, through the job system; required
//	API           serve it over REST and put it in the OpenAPI spec
//	CLI           generate its commands and docs pages; needs API
//	Under(group)  nest it: /api/node/{hostname}/<group>/<name>
//	Config(v)     declare settings, with v as their defaults
//	Platform(p, build)
//	              build the provider on platform p, from a host.Host: debian,
//	              darwin, or linux for every Linux host without its own; on
//	              a platform not declared, the methods report skipped
//	Sudo(progs)   programs the providers run with sudo; preflight checks them
//	Runs(progs)   programs they run without sudo; the agent warns if one is missing
//	Validator(tag, check, schema, message)
//	              a validate tag of the behavior's own
//	Requires(bs)  behaviors it is built on; their providers reach its builder
//	              through host.Require, and it calls their methods as Go calls
//	Internal      only other behaviors use it: no route, command or permission
//	Read(n, fn)   a method that reads the host: GET, <name>:read
//	Write(n, fn)  a method that changes it: POST, or PUT for update and
//	              DELETE for delete, <name>:write
//
// A method is refined with Key (the input field in the path), Describe (one
// line), Explain (what it touches and what running it twice does) and Example
// (a sample input and output, rendered onto its docs page).
//
// An example also describes the host it runs on and what it changes there,
// with After (call another method first), Prints (a command's output),
// PrintsAfter (its output once another command has run, for state a program
// keeps), Exists (a file already there), Wrote, Removed and Ran. Those are the provider's
// tests: see Testing.
//
// Fails declares a call the method refuses, with text its error contains:
// bad input, or a host it cannot act on.
//
// A method function has the shape
//
//	func(ctx context.Context, p Provider, in In) (Out, error)
//
// The agent finds p by its type among the providers it builds. A write's Out
// has a changed bool.
//
// # Tags on input and output types
//
//	json:"key"        the name in the request, the flag, the column and --json
//	validate:"..."    checked before a job is queued; becomes spec constraints
//	doc:"..."         the description in the spec, the flag help and the docs
//	render:"..."      on an output: "-" hides the column; bytes, age, status
//	telemetry:"attr"  on an input: recorded on the method's span
//	secret:"true"     on an input: redacted from the audit log, write-only in
//	                  the spec, never on a span
//
// A custom validate rule needs ValidatorSchema, so the spec checks what the
// server does.
//
// # Settings
//
// Config declares a struct of standing agent config under behaviors.<name>,
// with the same mapstructure, doc and validate tags as osapi's config. Each
// field gets an OSAPI_BEHAVIORS_<NAME>_<FIELD> variable and an osapi agent
// start flag, --<name>-<field>. A method reads them with Settings. A value that
// changes per call belongs in the input, not here.
//
// # Testing
//
// A behavior's tests above its provider are inferred from the definition and
// run for every registered behavior, in internal/behaviors/all:
//
//   - every example runs on a fake host with each platform's provider, and
//     must produce its output and its declared effects; a write must declare
//     effects, and run twice must report changed: false the second time;
//     each filesystem, command and store call failing in turn, and each store
//     read returning a damaged record, must not panic; each refusal must fail
//     with its text (package behaviortest);
//   - every example is sent through the real API handler and middleware on an
//     HTTP server, and the responses, broadcast, 401, 403, 404, 500 and each
//     missing required field are checked against the generated spec;
//   - every example's flags, given to its generated command, send its input.
//
// A behavior's own test calls behaviortest.Run, which osapi behavior new
// writes; the examples and refusals reach the methods through the provider. A
// test by hand is only for what an example cannot show.
//
// # Checking one
//
//	go generate ./internal/behaviors/all   write gen/, the spec and the docs pages
//	osapi behavior validate          the rules, see Validate
//	osapi behavior get --name <name> what was registered: routes, schemas
//	just test                        including behavior-gen-check
package behavior
