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

package main

import (
	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/behavior"
)

const (
	cobraPath       = "github.com/spf13/cobra"
	commandPath     = "github.com/osapi-io/osapi/internal/behavior/command"
	commandtestPath = "github.com/osapi-io/osapi/internal/behavior/command/commandtest"
	suitePath       = "github.com/stretchr/testify/suite"
	referencePath   = "github.com/osapi-io/osapi/internal/cli/reference"
)

// allFile is the one file that imports every behavior's generated code, so
// a binary needs a single import to get every behavior's commands and
// routes.
func allFile(
	bs []*behavior.Behavior,
) *jen.File {
	f := jen.NewFilePathName(module+"/internal/behaviors/all/gen", "gen")
	f.HeaderComment(generated)
	for _, line := range wrap("Package gen imports the generated code of every behavior with an "+
		"API or a CLI: its commands and its routes.", 76) {
		f.PackageComment(line)
	}
	for _, b := range bs {
		if b.API || b.CLI {
			f.Anon(pkgPath(b) + "/gen")
		}
	}

	return f
}
