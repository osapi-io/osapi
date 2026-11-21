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

import (
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/guide"
)

// Deps is what the methods use: the root command, whose tree generate
// documents, and the DSLs, whose specs dsl generate writes as pages.
type Deps struct {
	Root *cobra.Command
	DSLs func() []DSL
	// Pages are the pages of the reference: each is the output of the list
	// or get commands that show it.
	Pages func() ([]Page, error)
}

// Page is one page of the reference, at Path from the repository root. A
// partial, such as _catalog.mdx, is a part that a landing page imports: it
// has no front matter and no title. Each part is the output of a command of
// the group Group, rendered as the CLI prints it.
type Page struct {
	Path    string
	Title   string
	Intro   string
	Partial bool
	Parts   []Part
}

// Part is the output of one command, and the group whose output it is.
type Part struct {
	Group string
	Out   any
	// Heading, when set, heads the part in place of its output's title.
	Heading string
	// Link, when set, is the page that a cell of a table links to, by the
	// header of its column and its text.
	Link func(header string, cell string) string
}

// DSL is one DSL of the binary as a page of the docs site: its file name,
// title, intro, place in the sidebar, and its spec, the same that its spec get
// command prints.
type DSL struct {
	Slug     string
	Title    string
	Intro    string
	Position int
	Spec     guide.Spec
}
