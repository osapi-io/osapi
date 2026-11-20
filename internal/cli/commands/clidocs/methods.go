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
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/cli/reference"
)

// Generate writes the page of each command of the binary into the
// directory, and removes each page there that no command has.
func Generate(
	_ context.Context,
	d Deps,
	in GenerateInput,
) (GenerateOutput, error) {
	dir := in.Dir
	if dir == "" {
		dir = "docs/docs/sidebar/usage/cli"
	}
	n, err := reference.Write(d.Root, dir)
	if err != nil {
		return GenerateOutput{}, fmt.Errorf("write the CLI reference into %s: %w", dir, err)
	}

	return GenerateOutput{Dir: dir, Pages: n}, nil
}
