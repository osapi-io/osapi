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

package endpoint

import (
	"fmt"
	"regexp"
)

// name is a group, a method, a sub-path or an action: lowercase letters and
// digits, starting with a letter.
var name = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Problem is a rule a definition breaks, and how to fix it.
type Problem struct {
	// Group is the group it is about.
	Group string `json:"group"`
	// Method is the method it is about, or "" for the group.
	Method string `json:"method,omitempty"`
	// Rule says what is wrong.
	Rule string `json:"rule"`
	// Fix says what to change.
	Fix string `json:"fix"`
}

func (p Problem) String() string {
	return p.Rule + " Fix: " + p.Fix
}

// Validate is each rule g breaks. A group with no problems makes a spec
// that loads and routes that do not collide.
func Validate(
	g *Group,
) []Problem {
	var out []Problem
	add := func(method, rule, fix string) {
		out = append(out, Problem{Group: g.Name, Method: method, Rule: rule, Fix: fix})
	}

	if !name.MatchString(g.Name) {
		add("", fmt.Sprintf("group name %q is not lowercase letters and digits.", g.Name),
			"Name the group as its path segment, such as agent.")
	}
	if len(g.Methods) == 0 {
		add("", "the group has no methods.", "Add a method with Get, Post, Put or Delete.")
	}

	names := map[string]bool{}
	ids := map[string]string{}
	routes := map[string]string{}
	for _, m := range g.Methods {
		if !name.MatchString(m.Name) {
			add(m.Name, fmt.Sprintf("method name %q is not lowercase letters and digits.", m.Name),
				"Name the method for what it does, such as list or drain.")
		}
		if names[m.Name] {
			add(m.Name, "two methods have the name "+m.Name+".", "Give each method its own name.")
		}
		names[m.Name] = true

		for _, seg := range []string{m.sub, m.action} {
			if seg != "" && !name.MatchString(seg) {
				add(
					m.Name,
					fmt.Sprintf("path segment %q is not lowercase letters and digits.", seg),
					"Use one word, such as pending or drain.",
				)
			}
		}
		if m.action != "" && m.Key() == "" {
			add(m.Name, "the method has an action but no key.",
				`Tag the field the action is on in:"path", or use Sub for a fixed path.`)
		}

		keys := 0
		for _, f := range fields(m.In) {
			if f.Tag.Get("in") == "path" {
				keys++
			}
		}
		if keys > 1 {
			add(m.Name, "the input has more than one path field.",
				`Tag one field in:"path"; a group has one key.`)
		}

		id := m.OperationID(g)
		if other, ok := ids[id]; ok {
			add(m.Name, "operationId "+id+" is also the ID of "+other+".",
				"Give the method its own ID.")
		}
		ids[id] = m.Name

		route := m.Verb + " " + g.Path(m)
		if other, ok := routes[route]; ok {
			add(m.Name, route+" is also the route of "+other+".",
				"Change the verb, the sub-path or the action.")
		}
		routes[route] = m.Name
	}

	if _, err := OpenAPI(g, ErrorRef); err != nil {
		add("", "the spec does not build: "+err.Error()+".",
			"Use types that JSON can encode: structs, slices, maps, strings, numbers and bools.")
	}

	return out
}
