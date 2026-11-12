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

package behavior

import (
	"fmt"
	"reflect"
	"regexp"
)

// Problem is one rule a behavior breaks.
type Problem struct {
	Behavior string `json:"behavior"`
	Method   string `json:"method,omitempty"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}

// The rules Validate checks, by the name a Problem reports.
const (
	RuleName        = "name"
	RulePlacement   = "placement"
	RuleDoc         = "doc"
	RuleMethods     = "methods"
	RuleDuplicate   = "duplicate"
	RuleKey         = "key"
	RuleSchema      = "schema"
	RuleChanged     = "changed"
	RuleOutputShape = "output"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// reservedFields are what the platform puts in every per-host result, so an
// output type that declares one would collide with it.
var reservedFields = map[string]bool{
	"hostname": true,
	"status":   true,
	"error":    true,
}

// Validate checks every registered behavior against the rules the platform
// relies on, and returns what each breaks. No problems means every route,
// command and schema can be built.
func Validate() []Problem {
	all := All()
	problems := make([]Problem, 0, len(all))

	mu.RLock()
	for _, name := range duplicates {
		problems = append(problems, Problem{
			Behavior: name,
			Rule:     RuleDuplicate,
			Message:  "registered more than once",
		})
	}
	mu.RUnlock()

	for _, b := range all {
		problems = append(problems, ValidateBehavior(b)...)
	}

	return problems
}

// ValidateBehavior checks one behavior.
func ValidateBehavior(
	b *Behavior,
) []Problem {
	var problems []Problem
	add := func(method, rule, format string, args ...any) {
		problems = append(problems, Problem{
			Behavior: b.Name,
			Method:   method,
			Rule:     rule,
			Message:  fmt.Sprintf(format, args...),
		})
	}

	if !namePattern.MatchString(b.Name) {
		add("", RuleName, "name %q must be lowercase letters, digits and dashes", b.Name)
	}
	if b.Group != "" && !namePattern.MatchString(b.Group) {
		add("", RuleName, "group %q must be lowercase letters, digits and dashes", b.Group)
	}
	if b.Placement == "" {
		add("", RulePlacement, "no placement; add behavior.Node")
	}
	if b.Doc == "" {
		add("", RuleDoc, "no description")
	}
	if len(b.Methods) == 0 {
		add("", RuleMethods, "no methods")
	}

	seen := map[string]bool{}
	for _, m := range b.Methods {
		if seen[m.Name] {
			add(m.Name, RuleDuplicate, "method declared twice")
		}
		seen[m.Name] = true

		if !namePattern.MatchString(m.Name) {
			add(m.Name, RuleName, "method name must be lowercase letters, digits and dashes")
		}
		if m.Doc == "" {
			add(m.Name, RuleDoc, "no description; add .Describe(...)")
		}

		in, err := SchemaFor(m.In)
		if err != nil {
			add(m.Name, RuleSchema, "input: %v", err)
		} else if m.KeyField != "" {
			if _, ok := in.Properties[m.KeyField]; !ok {
				add(m.Name, RuleKey, "key %q is not a field of the input", m.KeyField)
			}
		}

		out, err := SchemaFor(m.Out)
		if err != nil {
			add(m.Name, RuleSchema, "output: %v", err)

			continue
		}

		elem := m.Out
		for elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Slice {
			elem = elem.Elem()
			out = out.Items
		}
		if elem.Kind() != reflect.Struct {
			add(m.Name, RuleOutputShape,
				"output must be a struct or a slice of structs, got %s", elem.Kind())

			continue
		}
		for name := range out.Properties {
			if reservedFields[name] {
				add(
					m.Name,
					RuleOutputShape,
					"output field %q collides with the per-host result",
					name,
				)
			}
		}

		changed, ok := out.Properties["changed"]
		if m.Verb == VerbWrite && (!ok || changed.Type != "boolean") {
			add(m.Name, RuleChanged, `a write's output needs Changed bool `+"`json:\"changed\"`")
		}
	}

	return problems
}
