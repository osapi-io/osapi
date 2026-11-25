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

// Package sdktypes lists the closed sets of pkg/sdk/types in a spec, from
// the SDK's own description of them (types.All), as package collectors
// lists gohai's collectors. A field of a DSL that takes one of a closed set
// of values has the type, with validate:"enum", and never a string with
// oneof, which validate refuses.
package sdktypes

import (
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Kind is the kind of the spec's type terms.
const Kind = "type"

// Append is s with a section of the SDK's closed sets at the end.
func Append(
	s guide.Spec,
) guide.Spec {
	s.Kinds = append(
		append([]guide.Kind(nil), s.Kinds...),
		guide.Kind{Kind: Kind, Title: "Types for values"},
	)
	s.Terms = append(append([]guide.Term(nil), s.Terms...), Terms(types.All())...)

	return s
}

// Terms are ts as terms of a spec: what each is and its values, and how a
// field uses it.
func Terms(
	ts []types.TypeSpec,
) []guide.Term {
	out := make([]guide.Term, 0, len(ts))
	for _, t := range ts {
		out = append(out, guide.Term{
			Kind: Kind,
			Name: "types." + t.Name,
			Does: t.Doc + " Values: " + strings.Join(t.Values, ", ") + ".",
			When: "A field that takes one of these values: types." + t.Name +
				` with validate:"enum". A closed set that pkg/sdk/types does not have is a ` +
				"named type with Values and Valid in the package's types.go; validate refuses oneof.",
		})
	}

	return out
}
