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

package skill

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/skills"
)

// List is each skill this binary has.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	all, err := skills.Catalog(d.Installed, d.Dev)
	if err != nil {
		return ListOutput{}, err
	}

	return ListOutput{Items: all}, nil
}

// Get is a skill's SKILL.md, or one of its reference files.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	data, err := skills.Read(d.Installed, d.Dev, in.Name, in.Reference)
	if err != nil {
		return GetOutput{}, fmt.Errorf("%w: run osapi skill list", err)
	}

	return GetOutput{Name: in.Name, Reference: in.Reference, Content: string(data)}, nil
}

// Install installs the skills of the binary for each tool, claude without one.
func Install(
	_ context.Context,
	d Deps,
	in InstallInput,
) (InstallOutput, error) {
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

	results, err := skills.Install(d.FS, root, tools, d.Installed, d.Version())
	if err != nil {
		return InstallOutput{}, err
	}
	changed := 0
	for _, r := range results {
		if r.Status != skills.Unchanged {
			changed++
		}
	}

	return InstallOutput{Version: d.Version(), Tools: names, Changed: changed, Files: results}, nil
}
