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
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

// Run checks b with Check and reports each problem as a test error.
func Run(
	t testing.TB,
	b *behavior.Behavior,
) {
	t.Helper()
	for _, p := range Check(b) {
		t.Error(p)
	}
}

// Check runs every example of every method of b on a fake host, once per
// platform b declares, with that platform's provider, and returns what went
// wrong. For each example:
//
//   - the output is the example's Output;
//   - each declared effect holds: a file written or removed, a command run;
//   - a write that reports changed declares at least one effect;
//   - a write run a second time reports changed: false;
//   - with each filesystem, command and store call failing in turn, and
//     with each store read returning a damaged record in turn, the method
//     returns rather than panics.
//
// For each refusal declared with Fails, the call fails with an error that
// contains its text. And each string input field is probed with hostile
// values, as checkProbes describes.
func Check(
	b *behavior.Behavior,
) []string {
	var problems []string
	for _, pp := range b.Platforms {
		for _, m := range b.Methods {
			for _, ex := range m.Examples {
				where := fmt.Sprintf("%s %s on %s, %q", b.Name, m.Name, pp.Name, ex.Doc)
				for _, p := range checkExample(b, m, pp.Name, ex) {
					problems = append(problems, where+": "+p)
				}
				for _, p := range checkProbes(b, m, pp.Name, ex) {
					problems = append(problems, where+": "+p)
				}
			}
			for _, r := range m.Refusals {
				where := fmt.Sprintf("%s %s on %s, refusal %q", b.Name, m.Name, pp.Name, r.Doc)
				if p := checkRefusal(b, m, pp.Name, r); p != "" {
					problems = append(problems, where+": "+p)
				}
			}
		}
	}

	return problems
}

func checkExample(
	b *behavior.Behavior,
	m *behavior.Method,
	platform string,
	ex behavior.Example,
) []string {
	h, p, problem := prepare(b, platform, ex)
	if problem != "" {
		return []string{problem}
	}

	var problems []string
	before, gets := h.calls, h.gets
	out, err := call(m, p, ex.Input)
	if errors.Is(err, errUnsupported) {
		return []string{fmt.Sprintf("%v: %s", err, strings.Join(h.unsupported, "; "))}
	}
	if err != nil {
		return []string{fmt.Sprintf("failed: %v", err)}
	}
	if got, want := jsonOf(out), jsonOf(ex.Output); got != want {
		problems = append(problems, fmt.Sprintf("output %s, want %s", got, want))
	}
	problems = append(problems, effects(h, ex.Effects)...)
	problems = append(problems, h.undeclared...)

	if m.Verb == behavior.VerbWrite {
		if changed, _ := behavior.Changed(ex.Output); changed && len(ex.Effects) == 0 {
			problems = append(problems,
				"a write declares no effects; add behavior.Wrote, Removed or Ran to the example")
		}
		again, err := call(m, p, ex.Input)
		if changed, _ := behavior.Changed(again); err == nil && changed {
			problems = append(problems, "running it again reports changed: true")
		}
	}

	// Each filesystem, command or store call the example makes fails in
	// turn, on a fresh host set up the same way.
	for i := 1; i <= h.calls-before; i++ {
		fh, fp, _ := prepare(b, platform, ex)
		fh.failAt = fh.calls + i
		if msg := survives(m, fp, ex.Input); msg != "" {
			problems = append(problems, fmt.Sprintf("with call %d failing: %s", i, msg))
		}
	}

	// Each store read returns a damaged record in turn: one that is not
	// JSON, one that is JSON with nothing in it, and a file state record
	// whose metadata is empty.
	for i := 1; i <= h.gets-gets; i++ {
		for _, damaged := range []string{"not json", "{}", `{"metadata":{}}`} {
			fh, fp, _ := prepare(b, platform, ex)
			fh.corruptAt, fh.corrupt = fh.gets+i, []byte(damaged)
			if msg := survives(m, fp, ex.Input); msg != "" {
				problems = append(problems,
					fmt.Sprintf("with store read %d returning %q: %s", i, damaged, msg))
			}
		}
	}

	return problems
}

// checkRefusal runs r on a fake host and reports what is wrong with how the
// method failed, or "".
func checkRefusal(
	b *behavior.Behavior,
	m *behavior.Method,
	platform string,
	r behavior.Refusal,
) string {
	_, p, problem := prepare(b, platform, r.Example)
	if problem != "" {
		return problem
	}

	out, err := call(m, p, r.Input)
	switch {
	case err == nil:
		return fmt.Sprintf("succeeded with %s, want an error containing %q", jsonOf(out), r.Error)
	case !strings.Contains(err.Error(), r.Error):
		return fmt.Sprintf("failed with %q, want it to contain %q", err, r.Error)
	default:
		return ""
	}
}

// prepare builds a fake host for platform, sets it up as ex's Given says,
// and builds b's provider on it.
func prepare(
	b *behavior.Behavior,
	platform string,
	ex behavior.Example,
) (*fakeHost, any, string) {
	h := newHost(b, platform)

	p, err := host.Build(b, h.Host)
	if err != nil || p == nil {
		return nil, nil, fmt.Sprintf("no provider on the fake host: %v", err)
	}

	for _, g := range ex.Given {
		switch g.Kind {
		case behavior.GivenFile:
			_ = h.FS.MkdirAll(filepath.Dir(g.Path), 0o755)
			_ = h.FS.WriteFile(g.Path, []byte(g.Content), 0o644)
		case behavior.GivenOutput:
			if g.After != "" {
				h.later = append(h.later, g)
			} else {
				h.outputs[g.Command] = g.Content
			}
		case behavior.GivenCall:
			m := methodFor(b, g.Input)
			if m == nil {
				return nil, nil, fmt.Sprintf("no method takes %T, the input After names", g.Input)
			}
			if _, err := call(m, p, g.Input); err != nil {
				return nil, nil, fmt.Sprintf("After %s failed: %v", m.Name, err)
			}
		}
	}
	h.ran = nil

	return h, p, ""
}

// methodFor is the method of b whose input type is input's.
func methodFor(
	b *behavior.Behavior,
	input any,
) *behavior.Method {
	for _, m := range b.Methods {
		if m.In == reflect.TypeOf(input) {
			return m
		}
	}

	return nil
}

// call runs m with input. A call the fake host has no answer for stops the
// method; it is returned as errUnsupported, and any other panic goes on.
func call(
	m *behavior.Method,
	p any,
	input any,
) (out any, err error) {
	defer func() {
		if r := recover(); r != nil {
			if r != errUnsupported { //nolint:errorlint // the value panicked with, not a wrapped error
				panic(r)
			}
			out, err = nil, errUnsupported
		}
	}()
	data, _ := json.Marshal(input)

	return m.Invoke(context.Background(), p, data)
}

// survives runs the method and reports a panic.
func survives(
	m *behavior.Method,
	p any,
	input any,
) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprintf("panicked: %v", r)
		}
	}()
	_, _ = call(m, p, input)

	return ""
}

// effects checks each effect against the host.
func effects(
	h *fakeHost,
	want []behavior.Effect,
) []string {
	var problems []string
	for _, e := range want {
		switch e.Kind {
		case behavior.EffectWrote:
			got, err := h.FS.ReadFile(e.Path)
			if err != nil {
				problems = append(problems, fmt.Sprintf("did not write %s", e.Path))
			} else if string(got) != e.Content {
				problems = append(problems, fmt.Sprintf("wrote %s as %q, want %q", e.Path, got, e.Content))
			}
		case behavior.EffectRemoved:
			if _, err := h.FS.Stat(e.Path); err == nil {
				problems = append(problems, fmt.Sprintf("did not remove %s", e.Path))
			}
		case behavior.EffectRan:
			if !slices.Contains(h.ran, e.Command) {
				problems = append(
					problems,
					fmt.Sprintf("did not run %q; ran %q", e.Command, h.ran),
				)
			}
		}
	}

	return problems
}

func jsonOf(
	v any,
) string {
	data, _ := json.Marshal(v)

	return string(data)
}
