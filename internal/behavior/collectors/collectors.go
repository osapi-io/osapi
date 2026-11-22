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

// Package collectors lists gohai's collectors in the behavior spec, from
// gohai's own description of them (gohai.Collectors). A provider reads the
// host through a collector, built with h.Sources(), and an example's fake
// host gives each file and command the collector reads. A fact that no
// collector has is a change to gohai first.
package collectors

import (
	"path"
	"strings"

	"github.com/osapi-io/gohai/pkg/gohai"

	"github.com/osapi-io/osapi/internal/guide"
)

// Kind is the kind of the spec's collector terms.
const Kind = "collector"

// Append is s with a section of gohai's collectors at the end.
func Append(
	s guide.Spec,
) guide.Spec {
	s.Kinds = append(
		append([]guide.Kind(nil), s.Kinds...),
		guide.Kind{Kind: Kind, Title: "gohai collectors"},
	)
	s.Terms = append(append([]guide.Term(nil), s.Terms...), Terms(gohai.Collectors())...)

	return s
}

// Terms are cs as terms of the spec: what each returns and how a provider
// builds it, and what it reads on each platform.
func Terms(
	cs []gohai.CollectorSpec,
) []guide.Term {
	out := make([]guide.Term, 0, len(cs))
	for _, c := range cs {
		fields := make([]string, 0, len(c.Fields))
		for _, f := range c.Fields {
			fields = append(fields, f.Name)
		}
		does := "Returns " + list(fields, "no fields") + ". Build it with " + constructor(c) + "."
		out = append(out, guide.Term{Kind: Kind, Name: c.Name, Does: does, When: reads(c)})
	}

	return out
}

// constructor is how a provider builds c: with the host's sources, or, for
// a collector that reads the network, with none.
func constructor(
	c gohai.CollectorSpec,
) string {
	if c.Network != "" {
		return c.Constructor + "; it reads " + c.Network + ", not the host"
	}

	return strings.TrimSuffix(c.Constructor, "(src)") + "(h.Sources())"
}

// reads is what c reads on each platform, for the examples of a behavior
// whose provider uses it.
func reads(
	c gohai.CollectorSpec,
) string {
	if c.Network != "" {
		return "Not on a host: an example cannot give it."
	}
	var parts []string
	for _, r := range c.Reads {
		var what []string
		if len(r.Files) > 0 {
			what = append(what, "Exists for "+strings.Join(r.Files, ", "))
		}
		if len(r.Commands) > 0 {
			what = append(what, "Prints for "+strings.Join(r.Commands, "; "))
		}
		parts = append(parts, r.Platform+": "+list(what, "nothing"))
	}

	return "An example's fake host gives, on " + strings.Join(parts, ". On ") +
		". With content, it reads more under those paths. gohai collector get --name " +
		c.Name + " shows it in full; " + path.Base(c.Package) + " is its package."
}

// list is items joined, or none when there are none.
func list(
	items []string,
	none string,
) string {
	if len(items) == 0 {
		return none
	}

	return strings.Join(items, ", ")
}
