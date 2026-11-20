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

// Package token is the token group of local commands: osapi token generate,
// roles list and validate, which sign and read the bearer tokens of the API
// with the signing key of the config. They run in the CLI, and call no API.
package initcmd

import (
	"context"

	"github.com/osapi-io/osapi/internal/skills"
)

// Init installs the skills of the binary for each tool, claude without one.
func Init(
	_ context.Context,
	d Deps,
	in Input,
) (Output, error) {
	names := in.Tools
	if len(names) == 0 {
		names = []string{"claude"}
	}
	root := in.Path
	if root == "" {
		root = "."
	}
	tools := make([]skills.Tool, 0, len(names))
	for _, n := range names {
		tools = append(tools, skills.Tools[n])
	}

	results, err := skills.Install(d.FS, root, tools, d.Source, d.Version())
	if err != nil {
		return Output{}, err
	}
	changed := 0
	for _, r := range results {
		if r.Status != skills.Unchanged {
			changed++
		}
	}

	return Output{Version: d.Version(), Tools: names, Changed: changed, Files: results}, nil
}
