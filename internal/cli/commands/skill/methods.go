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
