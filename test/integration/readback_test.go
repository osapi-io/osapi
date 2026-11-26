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
package integration_test

// This file has no build tag: its tests run with the unit tests, and
// CommandSmokeSuite, in cli_test.go, uses its functions against a live
// controller.

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// outRef is a reference to a field of the example's own --json output, in
// the flags or a want of a ReadBack, such as {out.id}.
var outRef = regexp.MustCompile(`\{out\.([^}]+)\}`)

// usesOut reports whether a ReadBack of e names a field of the example's
// output, so the suite runs the example with --json.
func usesOut(
	e command.Example,
) bool {
	for _, rb := range e.Checks {
		if outRef.MatchString(rb.Flags) {
			return true
		}
		for _, f := range rb.Fields {
			if s, ok := f.Want.(string); ok && outRef.MatchString(s) {
				return true
			}
		}
	}

	return false
}

// fillOut puts the fields of out, the example's decoded --json output, in
// place of each {out.<path>} of s.
func fillOut(
	s string,
	out any,
) (string, error) {
	var failed error
	filled := outRef.ReplaceAllStringFunc(s, func(ref string) string {
		path := outRef.FindStringSubmatch(ref)[1]
		v, ok, err := lookup(out, path)
		switch {
		case err != nil:
			failed = err
		case !ok:
			failed = fmt.Errorf("the example's output has no %s", path)
		default:
			return fmt.Sprint(v)
		}

		return ref
	})

	return filled, failed
}

// selector is a path segment that selects an element of a list by a
// field, such as [id=1a2b3c4d].
var selector = regexp.MustCompile(`^\[([^=\]]+)=([^\]]*)\]$`)

// lookup is the value at path of doc, a decoded JSON document, and whether
// it is there. A path is JSON names joined with a dot, and a segment
// [name=value] selects the first element of a list whose name field is
// value. It fails on a path that does not fit the document's shape, such as
// a name on a list.
func lookup(
	doc any,
	path string,
) (any, bool, error) {
	cur := doc
	for _, seg := range splitPath(path) {
		if m := selector.FindStringSubmatch(seg); m != nil {
			list, ok := cur.([]any)
			if !ok {
				return nil, false, fmt.Errorf(
					"%s: %s selects from a %T, not a list",
					path,
					seg,
					cur,
				)
			}
			found := false
			for _, el := range list {
				obj, ok := el.(map[string]any)
				if ok && obj[m[1]] != nil && fmt.Sprint(obj[m[1]]) == m[2] {
					cur, found = el, true

					break
				}
			}
			if !found {
				return nil, false, nil
			}

			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf(
				"%s: %s is a field of a %T, not an object",
				path,
				seg,
				cur,
			)
		}
		v, ok := obj[seg]
		if !ok {
			return nil, false, nil
		}
		cur = v
	}

	return cur, true, nil
}

// splitPath splits a path at each dot outside a selector, so a selector's
// value can hold a dot, as a file name does.
func splitPath(
	path string,
) []string {
	var segs []string
	depth, start := 0, 0
	for i, r := range path {
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		case '.':
			if depth == 0 {
				segs = append(segs, path[start:i])
				start = i + 1
			}
		}
	}

	return append(segs, path[start:])
}

// match is why the field f of doc does not have its want, or "". fill puts
// the suite's value in place of a name that the want uses, as for the
// example's flags.
func match(
	doc any,
	f command.Field,
	fill func(string) string,
) string {
	got, ok, err := lookup(doc, fill(f.Path))
	if err != nil {
		return err.Error()
	}
	if f.Want == command.Absent {
		if ok {
			return fmt.Sprintf("%s is %v, and must not be there", f.Path, got)
		}

		return ""
	}
	if !ok {
		return f.Path + " is not in the output"
	}
	wants := []any{f.Want}
	if c, isChoice := f.Want.(command.Choice); isChoice {
		wants = c
	}
	var shown []any
	for _, want := range wants {
		if s, isString := want.(string); isString {
			want = fill(s)
		}
		b, err := json.Marshal(want)
		if err != nil {
			return err.Error()
		}
		var w any
		if err := json.Unmarshal(b, &w); err != nil {
			return err.Error()
		}
		if reflect.DeepEqual(w, got) {
			return ""
		}
		shown = append(shown, w)
	}
	if len(shown) == 1 {
		return fmt.Sprintf("%s is %v, want %v", f.Path, got, shown[0])
	}

	return fmt.Sprintf("%s is %v, want one of %v", f.Path, got, shown)
}

// ReadBackTestSuite tests the functions that run a ReadBack.
type ReadBackTestSuite struct {
	suite.Suite
}

func (s *ReadBackTestSuite) TestLookup() {
	doc := map[string]any{
		"max_jobs": float64(4),
		"labels":   map[string]any{"group": "web"},
		"items": []any{
			map[string]any{"id": "a", "revoked": false},
			map[string]any{"id": "app.conf", "revoked": true},
		},
	}
	tests := []struct {
		name    string
		path    string
		want    any
		found   bool
		wantErr string
	}{
		{name: "a field", path: "max_jobs", want: float64(4), found: true},
		{name: "a nested field", path: "labels.group", want: "web", found: true},
		{name: "a selected element", path: "items.[id=app.conf].revoked", want: true, found: true},
		{name: "a missing field", path: "labels.dc"},
		{name: "no element selected", path: "items.[id=b].revoked"},
		{name: "a selector on an object", path: "labels.[id=a]", wantErr: "not a list"},
		{name: "a name on a list", path: "items.id", wantErr: "not an object"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, found, err := lookup(doc, tt.path)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.NoError(err)
			s.Equal(tt.found, found)
			s.Equal(tt.want, got)
		})
	}
}

func (s *ReadBackTestSuite) TestMatch() {
	doc := map[string]any{"max_jobs": float64(4), "name": "web-01", "labels": map[string]any{}}
	same := func(v string) string { return v }
	tests := []struct {
		name  string
		field command.Field
		fill  func(string) string
		want  string
	}{
		{name: "a number", field: command.Field{Path: "max_jobs", Want: 4}, fill: same},
		{
			name: "a string is not a number", field: command.Field{Path: "max_jobs", Want: "4"},
			fill: same, want: "max_jobs is 4, want 4",
		},
		{
			name: "a value the suite fills", field: command.Field{Path: "name", Want: "host"},
			fill: func(v string) string { return strings.ReplaceAll(v, "host", "web-01") },
		},
		{
			name:  "absent",
			field: command.Field{Path: "labels.group", Want: command.Absent},
			fill:  same,
		},
		{
			name: "absent but there", field: command.Field{Path: "name", Want: command.Absent},
			fill: same, want: "must not be there",
		},
		{
			name: "one of", field: command.Field{Path: "name", Want: command.OneOf("a", "web-01")},
			fill: same,
		},
		{
			name: "none of", field: command.Field{Path: "name", Want: command.OneOf("a", "b")},
			fill: same, want: "want one of [a b]",
		},
		{
			name: "missing", field: command.Field{Path: "id", Want: "x"}, fill: same,
			want: "id is not in the output",
		},
		{
			name: "a bad path", field: command.Field{Path: "name.x", Want: "x"}, fill: same,
			want: "not an object",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := match(doc, tt.field, tt.fill)
			if tt.want == "" {
				s.Empty(got)

				return
			}
			s.Contains(got, tt.want)
		})
	}
}

func (s *ReadBackTestSuite) TestFillOut() {
	out := map[string]any{"id": "1a2b", "job": map[string]any{"id": "j1"}}
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "a field", in: "--id {out.id}", want: "--id 1a2b"},
		{name: "a nested field", in: "{out.job.id}", want: "j1"},
		{name: "nothing to fill", in: "--id x", want: "--id x"},
		{name: "a missing field", in: "{out.name}", wantErr: "has no name"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := fillOut(tt.in, out)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)

				return
			}
			s.NoError(err)
			s.Equal(tt.want, got)
		})
	}
}

func (s *ReadBackTestSuite) TestUsesOut() {
	s.True(usesOut(command.Example{Checks: []command.ReadBack{{Flags: "--id {out.id}"}}}))
	s.True(usesOut(command.Example{Checks: []command.ReadBack{
		{Fields: []command.Field{{Path: "id", Want: "{out.id}"}}},
	}}))
	s.False(usesOut(command.Example{Checks: []command.ReadBack{{Flags: "--id x"}}}))
}

func TestReadBackTestSuite(t *testing.T) {
	suite.Run(t, new(ReadBackTestSuite))
}
