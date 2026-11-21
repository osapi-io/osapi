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

package behaviortest

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
)

// ChainRun is one example of a method, run on a fake host, with each Call
// that it made: the root step is the method, and its steps are the calls.
type ChainRun struct {
	Method   string         `json:"method"`
	Example  string         `json:"example"`
	Platform string         `json:"platform"`
	Root     *behavior.Step `json:"root"`
}

// Chains runs each example of each method of b once, on the first platform
// that serves it, as Check does, and records its calls. An example that the
// fake host cannot set up is left out: Check reports it.
func Chains(
	b *behavior.Behavior,
) []ChainRun {
	var runs []ChainRun
	for _, m := range b.Methods {
		for _, ex := range m.Examples {
			platform := ""
			for _, pp := range b.Platforms {
				if b.Serves(m, pp.Name) && ex.RunsOn(pp.Name) {
					platform = pp.Name

					break
				}
			}
			if platform == "" {
				continue
			}
			_, p, problem := prepare(b, platform, ex)
			if problem != "" {
				continue
			}
			root := &behavior.Step{
				Behavior: b.Name,
				Method:   m.Name,
				Input:    behavior.Redact(ex.Input),
			}
			data, _ := json.Marshal(ex.Input)
			out, err := m.Invoke(behavior.Record(context.Background(), root), p, data)
			if err != nil {
				root.Error = err.Error()
			} else {
				root.Output = out
			}
			runs = append(
				runs,
				ChainRun{Method: m.Name, Example: ex.Doc, Platform: platform, Root: root},
			)
		}
	}

	return runs
}

// ChainProblems are the problems of the chain rules that need a run of the
// examples of b: a call to a behavior that the caller does not require, and
// a requirement of b that no example calls. osapi behavior chain validate
// reports them, and Run fails the tests of b on them.
func ChainProblems(
	b *behavior.Behavior,
) []behavior.Problem {
	var problems []behavior.Problem
	add := func(method string, rule string, message string) {
		problems = append(problems, behavior.Problem{
			Behavior: b.Name, Method: method, Rule: rule, Message: message,
			Fix: behavior.ChainChecks.Fix(rule),
		})
	}
	called := map[string]bool{}
	var walk func(s *behavior.Step, run ChainRun)
	walk = func(s *behavior.Step, run ChainRun) {
		caller := behavior.Lookup(s.Behavior)
		for _, c := range s.Steps {
			if s.Behavior == b.Name {
				called[c.Behavior] = true
			}
			if caller != nil &&
				!slices.ContainsFunc(caller.Requires, func(r *behavior.Behavior) bool {
					return r.Name == c.Behavior
				}) {
				add(run.Method, behavior.RuleCalled, fmt.Sprintf(
					"example %q: %s %s calls %s %s, which %s does not require",
					run.Example, s.Behavior, s.Method, c.Behavior, c.Method, s.Behavior))
			}
			walk(c, run)
		}
	}
	for _, r := range Chains(b) {
		walk(r.Root, r)
	}
	for _, r := range b.Requires {
		if r != nil && !called[r.Name] {
			add(
				"",
				behavior.RuleUsed,
				fmt.Sprintf("%s requires %s, and no example calls it", b.Name, r.Name),
			)
		}
	}

	return problems
}

// ChainCall is one call of a run as a row: the call, indented under the call
// that made it, and the fields that it sends and gets.
type ChainCall struct {
	Call  string `json:"call"`
	Sends string `json:"sends"`
	Gets  string `json:"gets"`
}

// Calls are the calls of r, the method first and each call under the one
// that made it. A table trims leading spaces, so a deeper call is marked with
// a dot for each level above it. Each value is cut to width runes; 0 cuts
// nothing.
func (r ChainRun) Calls(
	width int,
) []ChainCall {
	var out []ChainCall
	var walk func(s *behavior.Step, depth int)
	walk = func(s *behavior.Step, depth int) {
		call := s.Behavior + " " + s.Method
		if depth > 0 {
			call = strings.Repeat("· ", depth-1) + "└─ " + call
		}
		gets := fields(s.Output, width)
		if s.Error != "" {
			gets = "error: " + s.Error
		}
		out = append(out, ChainCall{Call: call, Sends: fields(s.Input, width), Gets: gets})
		for _, c := range s.Steps {
			walk(c, depth+1)
		}
	}
	walk(r.Root, 0)

	return out
}

// fields is v as the fields that have a value, key=value in key order: a
// text of several lines is its count of lines, and the whole is cut to width.
func fields(
	v any,
	width int,
) string {
	data, _ := json.Marshal(v)
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return cut(string(data), width)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if text := value(m[k]); text != "" {
			parts = append(parts, k+"="+text)
		}
	}

	return cut(strings.Join(parts, " "), width)
}

// value is one field's value as fields shows it, or "" for one with none.
func value(
	v any,
) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		if n := strings.Count(strings.TrimRight(t, "\n"), "\n") + 1; n > 1 {
			return fmt.Sprintf("<%d lines>", n)
		}
		if strings.ContainsAny(t, " ") {
			return strconv.Quote(t)
		}

		return t
	case []any:
		if len(t) == 0 {
			return ""
		}
		items := make([]string, 0, len(t))
		for _, item := range t {
			items = append(items, value(item))
		}

		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		if len(t) == 0 {
			return ""
		}
		data, _ := json.Marshal(t)

		return string(data)
	default:
		return fmt.Sprint(t)
	}
}

// cut is s cut to width, with … where it was cut, or s when width is 0.
func cut(
	s string,
	width int,
) string {
	if r := []rune(s); width > 0 && len(r) > width {
		return string(r[:width-1]) + "…"
	}

	return s
}
