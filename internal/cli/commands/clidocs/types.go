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

package clidocs

// GenerateInput is the root of the repository whose pages to write.
type GenerateInput struct {
	Root string `json:"root" validate:"omitempty" doc:"Root of the osapi repository" default:"."`
}

// GenerateOutput is how many pages of each kind generate wrote.
type GenerateOutput struct {
	_         struct{} `title:"Docs"`
	CLI       int      `             json:"cli"       label:"CLI pages"`
	DSL       int      `             json:"dsl"       label:"DSL pages"`
	Reference int      `             json:"reference" label:"Reference pages"`
}

// CLIInput is where to write the pages.
type CLIInput struct {
	Dir string `json:"dir" validate:"omitempty" doc:"Directory of the CLI reference pages" default:"docs/docs/sidebar/usage/cli"`
}

// CLIOutput is where the pages are, and how many there are.
type CLIOutput struct {
	_     struct{} `title:"CLI reference"`
	Dir   string   `                      json:"dir"`
	Pages int      `                      json:"pages"`
}

// DSLInput is where to write the DSL pages.
type DSLInput struct {
	Dir string `json:"dir" validate:"omitempty" doc:"Directory of the DSL pages" default:"docs/docs/sidebar/development/dsl"`
}

// DSLOutput is where the DSL pages are, and how many there are.
type DSLOutput struct {
	_     struct{} `title:"DSL reference"`
	Dir   string   `                      json:"dir"`
	Pages int      `                      json:"pages"`
}

// ReferenceInput is the root of the repository whose reference to write.
type ReferenceInput struct {
	Root string `json:"root" validate:"omitempty" doc:"Root of the osapi repository" default:"."`
}

// ReferenceOutput is how many pages the reference has.
type ReferenceOutput struct {
	_     struct{} `title:"Reference"`
	Pages int      `                  json:"pages"`
}
