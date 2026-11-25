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

package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/sdk/sdkspec"
)

// List is each type for values, then each service of the client.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []Item{}}
	for _, t := range d.Types() {
		out.Items = append(
			out.Items,
			Item{Kind: KindType, Name: t.Name, Count: len(t.Values), Doc: t.Doc},
		)
	}
	for _, s := range d.Services() {
		out.Items = append(
			out.Items,
			Item{Kind: KindService, Name: s.Name, Count: len(s.Methods), Doc: s.Doc},
		)
	}

	return out, nil
}

// Get is the type or the service of in.Name. A type may be named with its
// package, such as types.ContentType.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	name := strings.TrimPrefix(in.Name, "types.")
	for _, t := range d.Types() {
		if t.Name == name {
			return GetOutput{Type: &t}, nil
		}
	}
	for _, s := range d.Services() {
		if s.Name == in.Name {
			return GetOutput{Service: &s}, nil
		}
	}

	return GetOutput{}, fmt.Errorf(
		"no type or service named %q: osapi sdk list shows them",
		in.Name,
	)
}

// Spec is each part of the SDK.
func Spec(
	_ context.Context,
	d Deps,
	_ SpecInput,
) (SpecOutput, error) {
	return SpecOutput{Kinds: sdkspec.Spec(d.Types(), d.Services()).Sections()}, nil
}

// Validate is each problem of the SDK against sdkspec.Checks.
func Validate(
	_ context.Context,
	d Deps,
	_ ValidateInput,
) (ValidateOutput, error) {
	ts, ss := d.Types(), d.Services()
	out := ValidateOutput{Types: len(ts), Services: len(ss), Problems: sdkspec.Validate(ts, ss)}
	out.Valid = len(out.Problems) == 0
	if out.Valid {
		return out, nil
	}
	out.Fixes = map[string]any{}
	for _, p := range out.Problems {
		out.Fixes[p.Rule] = p.Fix
	}

	return out, command.Fail()
}
