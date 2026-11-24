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
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// The chain rules, by the name a problem reports. RuleRequires and
// RuleInternal hold when the behaviors are built; RuleCalled and RuleUsed
// hold when the examples run, because only a run shows which behaviors a
// provider calls.
const (
	// RuleCalled: each call of a provider goes to a behavior that its
	// behavior requires.
	RuleCalled = "called"
	// RuleUsed: an example calls each behavior that a behavior requires.
	RuleUsed = "used"
)

// ChainChecks are the rules of a chain: the behaviors that a behavior
// requires, and the calls that it makes to them. osapi dsl chain
// validate checks each one; dsl behavior validate checks the ones that hold
// when the behaviors are built, and go test checks the ones that need a run.
var ChainChecks = guide.Checks{
	{
		Rule:     RuleRequires,
		Requires: "Each required behavior is registered, and no behavior requires itself through others.",
		Fix:      "Import the required behavior in internal/behaviors/all, or remove the requirement that makes the cycle.",
	},
	{
		Rule:     RuleInternal,
		Requires: "Another behavior requires each internal behavior. Nothing else can call it, because it has no API and no CLI.",
		Fix:      "Add behavior.Requires(it) to the behavior that uses it, or remove behavior.Internal if an operator calls it.",
	},
	{
		Rule:     RuleCalled,
		Requires: "A provider calls only the behaviors that its behavior requires. The agent builds the provider of each required behavior, and no other.",
		Fix:      "Add the behavior that it calls to behavior.Requires, and read its provider with host.Require.",
	},
	{
		Rule:     RuleUsed,
		Requires: "An example calls each behavior that a behavior requires. A requirement that no example calls is not tested, or is not needed.",
		Fix:      "Add an example whose provider calls the required behavior, or remove the requirement.",
	},
}

// ChainProblems are the problems of the chain rules that hold when the
// behaviors are built: RuleRequires and RuleInternal.
func ChainProblems() []Problem {
	var problems []Problem
	for _, p := range chainProblems(All()) {
		if p.Rule == RuleRequires || p.Rule == RuleInternal {
			p.Fix = ChainChecks.Fix(p.Rule)
			problems = append(problems, p)
		}
	}

	return problems
}

// A Step is one call of a method in a chain, with the calls it made. The
// examples record them: Record gives a context, and each Call under it adds
// its step, so osapi dsl chain get shows what each call sent and got.
type Step struct {
	Behavior string  `json:"behavior"`
	Method   string  `json:"method"`
	Input    any     `json:"input,omitempty"`
	Output   any     `json:"output,omitempty"`
	Error    string  `json:"error,omitempty"`
	Steps    []*Step `json:"steps,omitempty"`
	// Duration is how long the call ran. It is left out of the JSON of a
	// test, which compares the same chain run after run.
	Duration time.Duration `json:"-"`

	mu sync.Mutex
}

// Wire are the calls that s made, as a job's response carries them: each
// step's behavior, method, whether it changed the host, its error and its
// duration. The input and the output stay out: an input is redacted only by
// its top-level secret tags, an output not at all, and the content of a
// managed file can be large. The job's request and response have the input
// and the output of the method at the top.
func (s *Step) Wire() []client.Step {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Steps) == 0 {
		return nil
	}
	out := make([]client.Step, 0, len(s.Steps))
	for _, c := range s.Steps {
		w := client.Step{
			Behavior:   c.Behavior,
			Method:     c.Method,
			Changed:    changedOf(c.Output),
			Error:      c.Error,
			DurationMS: c.Duration.Milliseconds(),
			Steps:      c.Wire(),
		}
		out = append(out, w)
	}

	return out
}

// changedOf is the changed field of an output, or nil when it has none.
func changedOf(
	out any,
) *bool {
	data, err := json.Marshal(out)
	if err != nil {
		return nil
	}
	var fields struct {
		Changed *bool `json:"changed"`
	}
	if json.Unmarshal(data, &fields) != nil {
		return nil
	}

	return fields.Changed
}

type stepKey struct{}

// Record is ctx with s as the step that each Call under it adds its own
// step to. Without it, Call records nothing.
func Record(
	ctx context.Context,
	s *Step,
) context.Context {
	return context.WithValue(ctx, stepKey{}, s)
}

// Redact is in as a step records it: a struct with each field tagged
// secret:"true" as [redacted], as the audit log and the spans leave it out.
// Anything else is in unchanged.
func Redact(
	in any,
) any {
	t := reflect.TypeOf(in)
	if t == nil || t.Kind() != reflect.Struct {
		return in
	}
	var secrets []string
	for i := range t.NumField() {
		if f := t.Field(i); Secret(f) {
			secrets = append(secrets, strings.Split(f.Tag.Get("json"), ",")[0])
		}
	}
	if len(secrets) == 0 {
		return in
	}
	data, _ := json.Marshal(in)
	var fields map[string]any
	_ = json.Unmarshal(data, &fields)
	for _, name := range secrets {
		if _, ok := fields[name]; ok {
			fields[name] = "[redacted]"
		}
	}

	return fields
}

// Tree is b and the behaviors that it calls, down the chain:
// dns → netplan → managedfile, or status → (hostname, uptime) when a
// behavior calls more than one.
func Tree(
	b *Behavior,
) string {
	if len(b.Requires) == 0 {
		return b.Name
	}
	parts := make([]string, 0, len(b.Requires))
	for _, r := range b.Requires {
		parts = append(parts, Tree(r))
	}
	if len(parts) == 1 {
		return b.Name + " → " + parts[0]
	}

	return b.Name + " → (" + strings.Join(parts, ", ") + ")"
}

// add appends c to the steps of s. Calls can run at the same time, as
// status reads its behaviors.
func (s *Step) add(
	c *Step,
) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Steps = append(s.Steps, c)
}

// Requires declares the behaviors this one is built on. The agent builds
// each required behavior's provider before this one's and hands it to this
// one's builder, which reads it with host.Require; this behavior's provider
// then calls the required behavior's methods as Go calls, on the same host in
// the same job. Chaining behaviors across hosts or steps is a workflow, not a
// requirement.
func Requires(
	behaviors ...*Behavior,
) Option {
	return optionFunc(func(b *Behavior) { b.Requires = append(b.Requires, behaviors...) })
}

// Internal marks a behavior that only other behaviors use. It gets no route,
// no command, no spec and no permissions; dsl behavior list and dsl behavior get still
// show it, with what uses it. A primitive can start internal and be exposed
// later by removing the option.
var Internal Option = optionFunc(func(b *Behavior) { b.Internal = true })

// RequiredBy are the registered behaviors that require b, in registry order.
func RequiredBy(
	b *Behavior,
) []*Behavior {
	var out []*Behavior
	for _, other := range All() {
		for _, r := range other.Requires {
			if r == b {
				out = append(out, other)

				break
			}
		}
	}

	return out
}

// chainProblems are the requirements of the registered behaviors that cannot
// be built: one that is not registered, and a cycle.
func chainProblems(
	all []*Behavior,
) []Problem {
	var problems []Problem
	for _, b := range all {
		for _, r := range b.Requires {
			if r == nil || Lookup(r.Name) != r {
				name := "<nil>"
				if r != nil {
					name = r.Name
				}
				problems = append(problems, Problem{
					Behavior: b.Name,
					Rule:     RuleRequires,
					Message:  "requires " + name + ", which is not registered",
				})
			}
		}
		if cycle := cycleFrom(b, b, nil); cycle != nil {
			problems = append(problems, Problem{
				Behavior: b.Name,
				Rule:     RuleRequires,
				Message:  "requires itself: " + strings.Join(cycle, " -> "),
			})
		}
		if b.Internal && len(RequiredBy(b)) == 0 {
			problems = append(problems, Problem{
				Behavior: b.Name,
				Rule:     RuleInternal,
				Message: "is internal, but no behavior requires it, so nothing can call it: " +
					"require it from the behavior that needs it, or remove behavior.Internal " +
					"if an operator calls it",
			})
		}
		for _, p := range b.Writes {
			if _, err := path.Match(p, "/"); p != "*" &&
				(err != nil || !strings.HasPrefix(p, "/")) {
				problems = append(problems, Problem{
					Behavior: b.Name,
					Rule:     RuleWrites,
					Message: fmt.Sprintf(
						"writes %q, which is not \"*\" or an absolute path pattern",
						p,
					),
				})
			}
		}
	}

	return problems
}

// cycleFrom is the path from at back to start through requirements, or nil
// when there is none. path is the way taken so far.
func cycleFrom(
	start *Behavior,
	at *Behavior,
	path []string,
) []string {
	path = append(path, at.Name)
	for _, r := range at.Requires {
		if r == nil {
			continue
		}
		if r == start {
			return append(path, r.Name)
		}
		if slices.Contains(path, r.Name) {
			continue
		}
		if cycle := cycleFrom(start, r, path); cycle != nil {
			return cycle
		}
	}

	return nil
}

// Call runs fn, a method of b, with p, b's provider, inside the span and
// the metrics of that method, as the agent runs a method. A behavior calls
// the behaviors it requires through Call, so a trace shows each one as its
// own step: behavior.sysctl.create, then behavior.managedfile.update under
// it. A direct call of fn skips the span, and behaviortest.Run fails it.
//
// Call checks in against its validate tags before fn runs, as the API does
// for a served method, so an internal behavior's rules hold for each
// caller. An input that breaks one is an error, and fn does not run.
//
//	result, err := behavior.Call(ctx, managedfile.Behavior, managedfile.Update, d.files, in)
func Call[P any, In any, Out any](
	ctx context.Context,
	b *Behavior,
	fn func(ctx context.Context, p P, in In) (Out, error),
	p P,
	in In,
) (Out, error) {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	var m *Method
	for _, candidate := range b.Methods {
		if candidate.fn == name {
			m = candidate
			break
		}
	}
	if m == nil {
		var zero Out

		return zero, fmt.Errorf("%s is not a method of behavior %s", name, b.Name)
	}

	var step *Step
	if parent, ok := ctx.Value(stepKey{}).(*Step); ok {
		step = &Step{Behavior: b.Name, Method: m.Name, Input: Redact(in)}
		parent.add(step)
		ctx = Record(ctx, step)
	}

	var out Out
	start := time.Now()
	_, err := m.run(ctx, func(ctx context.Context) (any, any, error) {
		if err := check(ctx, m, in); err != nil {
			return nil, in, err
		}
		var err error
		out, err = fn(ctx, p, in)

		return out, in, err
	})
	if step != nil {
		step.Duration = time.Since(start)
		if err != nil {
			step.Error = err.Error()
		} else {
			step.Output = out
		}
	}

	return out, err
}

// Func is the name of the method's Go function, such as
// github.com/osapi-io/osapi/internal/behaviors/managedfile.Update.
func (m *Method) Func() string {
	return m.fn
}

// The kinds of the chain spec that the behavior spec does not have.
const (
	KindChain    = "chain"
	KindFailStep = "fail"
)

// ChainSpec is every term of a chain: what one is, how a behavior builds on
// others, how it calls them, and what happens when a step fails. osapi
// dsl chain spec get prints it, with the building blocks of the binary.
var ChainSpec = guide.Spec{
	Kinds: []guide.Kind{
		{Kind: KindChain, Title: "What a chain is"},
		{Kind: KindCompose, Title: "Building on other behaviors"},
		{Kind: KindCall, Title: "Calls between behaviors"},
		{Kind: KindFailStep, Title: "When a step fails"},
	},
	Terms: []guide.Term{
		{
			Kind: KindChain,
			Name: "A chain",
			Does: "The behaviors that a behavior calls, and the ones that they call: dns calls netplan, and netplan calls managedfile. The chain has the name of the behavior at its top. It is not declared on its own: Requires on each behavior makes it.",
			When: "osapi dsl chain list shows each chain; osapi dsl chain get --name <behavior> shows its tree.",
		},
		{
			Kind: KindChain,
			Name: "The calls of an example",
			Does: "osapi dsl chain get runs each example of the behavior on the fake host that go test uses, and records each Call: the behavior, the method, the input, and the output or the error. Thus the tree shows what the code does now.",
			When: "To see what a behavior sends to the behaviors that it calls, before a change or when one fails.",
		},
		{
			Kind: KindCompose,
			Name: "Use a building block",
			Does: "A building block is an internal behavior that does one generic job, such as writing a file that osapi manages. Its spans, checks and records come with it. \"Building blocks\" lists the ones in this binary.",
			When: "Before a provider does generic work itself: a managed file, a service reload, a call to an outside API. Require the block with Requires, and call it with Call. osapi dsl behavior create --requires <block> writes the Requires, the host.Require and a field of the provider; you write each Call.",
		},
		{
			Kind: KindCompose,
			Name: "Make a building block",
			Does: "Generic work that no block does yet becomes a new internal behavior. The behavior that needs it requires it.",
			When: "When the work is not about the one thing that the behavior manages, or a second behavior would need it.",
		},
		{
			Kind: KindCompose,
			Name: "One host, one step",
			Does: "A behavior acts on one host, in one job. Its building blocks run on the same host, in the same job.",
			When: "Work across hosts, or in ordered steps, is an osapi-orchestrator workflow, not a behavior.",
		},
		{
			Kind: KindCall,
			Name: "Call(ctx, other.Behavior, other.Method, p, in)",
			Does: "Runs a method of a required behavior in its own span, so the trace shows it as a step. It checks in against its validate tags first, as the API does for a served method: an input that breaks one is an error, and the method does not run.",
			When: "Each time a provider uses a required behavior. A direct call fails the tests.",
		},
		{
			Kind: KindCall,
			Name: "host.Require[P](h)",
			Does: "Returns the provider of a required behavior, in a builder.",
			When: "In the builder of a behavior that uses Requires.",
		},
		{
			Kind: KindFailStep,
			Name: "The error goes up",
			Does: "Call returns the error of a step to the provider that called it. The provider returns it, with what it was doing, or handles it. The job fails with the message, and the trace marks the span of the step that failed.",
			When: "In each provider that calls another behavior.",
		},
		{
			Kind: KindFailStep,
			Name: "A step undoes its own change",
			Does: "A building block that changes the host and then fails puts the host back before it returns the error: netplan gives its file back the content that it had, or removes a new file, when netplan generate refuses it. The caller does not undo the work of a block.",
			When: "In a building block whose change can fail after it started.",
		},
		{
			Kind: KindFailStep,
			Name: "Skipped on a platform",
			Does: "A behavior that a chain needs, and that cannot run on a platform, makes the behavior at the top report skipped there too.",
			When: "When a building block declares fewer platforms than the behavior that requires it.",
		},
	},
}

// CurrentChainSpec is ChainSpec, with "Building blocks": each internal
// behavior in the registry, with what it does and its methods. Thus the list
// agrees with the binary.
func CurrentChainSpec() guide.Spec {
	spec := guide.Spec{
		Kinds: append([]guide.Kind(nil), ChainSpec.Kinds...),
		Terms: append([]guide.Term(nil), ChainSpec.Terms...),
	}
	spec.Kinds = append(spec.Kinds, guide.Kind{Kind: KindBlock, Title: "Building blocks"})
	for _, b := range All() {
		if !b.Internal {
			continue
		}
		names := make([]string, 0, len(b.Methods))
		for _, m := range b.Methods {
			names = append(names, m.Name)
		}
		spec.Terms = append(spec.Terms, guide.Term{
			Kind: KindBlock,
			Name: b.Name,
			Does: b.Doc + ". Methods: " + strings.Join(names, ", ") + ".",
			When: "Require it, and call its methods with Call. osapi dsl behavior get --name " +
				b.Name + " shows its inputs.",
		})
	}

	return spec
}
