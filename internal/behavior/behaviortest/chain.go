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

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/secret"
)

// ChainRun is one example of a method, run on a fake host, with each Call
// that it made: the root step is the method, and its steps are the calls.
type ChainRun struct {
	Method   string                `json:"method"`
	Example  string                `json:"example"`
	Platform behavior.PlatformName `json:"platform"`
	Root     *behavior.Step        `json:"root"`
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
			var platform behavior.PlatformName
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
			root := behavior.NewStep(b.Name, m.Name, ex.Input)
			data, _ := json.Marshal(secret.Plain(ex.Input))
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
// examples of b: a call to a behavior that the caller does not require, a
// call that gives a secret to a field that is not a secret.Value, and a
// requirement of b that no example calls. osapi dsl chain validate
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
			for _, leak := range behavior.SecretLeaks(s, c) {
				add(
					run.Method,
					behavior.RuleSecret,
					fmt.Sprintf("example %q: %s", run.Example, leak),
				)
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
