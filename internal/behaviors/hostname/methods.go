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

package hostname

import (
	"context"
	"fmt"
	"regexp"
)

// name is a hostname as RFC 1123 allows it: labels of letters, digits and
// hyphens, with no hyphen at either end, up to 63 characters, joined by dots.
var name = regexp.MustCompile(
	`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?(\.[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$`,
)

// Get is the host's name.
func Get(
	ctx context.Context,
	p Provider,
	_ GetInput,
) (Entry, error) {
	name, err := p.Hostname(ctx)
	if err != nil {
		return Entry{}, err
	}

	return Entry{Name: name}, nil
}

// Update sets the host's name.
func Update(
	ctx context.Context,
	p Provider,
	in UpdateInput,
) (Result, error) {
	// The API checked the name, but the provider passes it to hostnamectl,
	// so it is checked again here.
	if len(in.Name) > 253 || !name.MatchString(in.Name) {
		return Result{}, fmt.Errorf("invalid hostname %q", in.Name)
	}
	changed, err := p.SetHostname(ctx, in.Name)
	if err != nil {
		return Result{}, err
	}

	return Result{Name: in.Name, Changed: changed}, nil
}
