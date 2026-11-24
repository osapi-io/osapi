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
package command_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// recorder is a Printer that writes what it is asked to print as text.
type recorder struct {
	b strings.Builder
}

func (r *recorder) Heading(title string) { fmt.Fprintf(&r.b, "# %s\n", title) }

func (r *recorder) KV(pairs ...string) {
	for i := 0; i+1 < len(pairs); i += 2 {
		fmt.Fprintf(&r.b, "%s: %s\n", pairs[i], pairs[i+1])
	}
}

func (r *recorder) Table(t command.Table) {
	fmt.Fprintf(&r.b, "# %s\n%s\n", t.Title, strings.Join(t.Headers, " | "))
	for _, row := range t.Rows {
		fmt.Fprintf(&r.b, "%s\n", strings.Join(row, " | "))
	}
	if len(t.Rows) == 0 {
		fmt.Fprintf(&r.b, "None. %s\n", t.Empty)
	}
}

func (r *recorder) Text(text string) { fmt.Fprintf(&r.b, "%s", text) }

func (r *recorder) Block(title string, pairs ...string) {
	fmt.Fprintf(&r.b, "* %s\n", title)
	for i := 0; i+1 < len(pairs); i += 2 {
		fmt.Fprintf(&r.b, "  %s| %s\n", pairs[i], pairs[i+1])
	}
}
func (r *recorder) Verdict(text string)             { fmt.Fprintf(&r.b, "[ok] %s\n", text) }
func (r *recorder) Status(word string) string       { return "<" + word + ">" }
func (r *recorder) Bytes(n int64) string            { return fmt.Sprintf("%dB", n) }
func (r *recorder) Time(t time.Time) string         { return t.UTC().Format(time.DateTime) }
func (r *recorder) Duration(d time.Duration) string { return d.String() }

func (r *recorder) Age(d time.Duration) string {
	if d > time.Hour {
		return "long ago"
	}

	return d.Truncate(time.Second).String() + " ago"
}

// keyed is a method that names the input field it acts on, as an endpoint
// method does.
type keyed struct {
	*command.LocalMethod
	key string
}

func (k keyed) Key() string { return k.key }

// outGroup is a registered group with the methods under test.
type outGroup struct {
	group
	cmds []command.Command
}

func (g outGroup) Commands() []command.Command { return g.cmds }

type (
	row struct {
		Name    string `json:"name"`
		Size    int    `json:"size"    render:"bytes"`
		Note    string `json:"note"                   table:"-"`
		Nothing string `json:"nothing"`
	}
	rows struct {
		_          struct{} `title:"Things"`
		TotalItems int      `               json:"total_items"`
		Items      []row    `               json:"items"       empty:"Add one."`
	}
	plain struct {
		Items []row `json:"items"`
	}
	term struct {
		Name string `json:"name"`
		Does string `json:"does" label:"-"`
		When string `json:"when" label:"Use"`
	}
	kind struct {
		Kind  string `json:"kind"  render:"-"`
		Title string `json:"title"`
		Terms []term `json:"terms"            table:"blocks"`
	}
	spec struct {
		Kinds []kind `json:"kinds" table:"sections"`
	}
	tagged struct {
		Tags []string `json:"tags" none:"none"`
	}
	terms struct {
		Terms []term `json:"terms" table:"blocks"`
	}
	flagRow struct {
		Flag string `json:"flag"`
	}
	cmdSection struct {
		Command string    `json:"command"`
		Method  string    `json:"method"`
		Flags   []flagRow `json:"flags"`
	}
	described struct {
		Name     string       `json:"name"`
		Commands []cmdSection `json:"commands" table:"sections"`
	}
	part struct {
		_     struct{} `title:"Part"`
		Value string   `             json:"value"`
	}
	envelope struct {
		Entry part `json:"entry"`
	}
	listEnvelope struct {
		Rows []row `json:"rows"`
	}
	detail struct {
		ID       string    `json:"id"                render:"short=4"`
		LastSeen *string   `json:"last_seen"                          none:"never"`
		Skipped  string    `json:"skipped,omitempty"`
		Off      bool      `json:"off,omitempty"`
		Hidden   string    `json:"hidden"            render:"-"`
		Part     *part     `json:"part,omitempty"`
		Rows     []row     `json:"rows,omitempty"`
		More     []row     `json:"more"`
		At       time.Time `json:"at"`
	}
	counted struct {
		TotalItems int            `json:"total_items"`
		Counts     map[string]int `json:"counts,omitempty"`
		Items      []row          `json:"items"`
	}
	values struct {
		On    bool              `json:"on,omitempty"`
		Count int               `json:"count,omitempty"`
		Tags  []string          `json:"tags,omitempty"`
		Meta  map[string]any    `json:"meta,omitempty"`
		Env   map[string]string `json:"env,omitempty"`
		Name  string            `json:"name,omitempty"`
	}
	osInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Size    int    `json:"size,omitempty" render:"bytes"`
	}
	cond struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	}
	host struct {
		Name  string  `json:"name"`
		OS    *osInfo `json:"os"`
		Conds []cond  `json:"conds"`
	}
	kept struct {
		Name  string `json:"name"`
		Conds []cond `json:"conds,omitempty" none:"-"`
	}
	node struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		Children []node `json:"children,omitempty"`
	}
	doc struct {
		Name string `json:"name"`
		Body string `json:"body" render:"body"`
	}
	nothing  struct{}
	deleteIn struct {
		MachineID string `json:"machine_id"`
	}
)

func handle[Out any](context.Context, struct{}, struct{}) (Out, error) {
	var out Out

	return out, nil
}

type OutputPublicTestSuite struct {
	suite.Suite
}

func (s *OutputPublicTestSuite) TestOutput() {
	g := outGroup{group: group{name: "zzderive"}}
	method := func(name string, out reflect.Type) *command.LocalMethod {
		m := command.Method[struct{}, struct{}](name).Handle(handle[struct{}])
		m.Out = out

		return m
	}
	g.cmds = []command.Command{
		method("list", reflect.TypeFor[rows]()),
		method("plain", reflect.TypeFor[plain]()),
		method("each", reflect.TypeFor[[]row]()),
		method("names", reflect.TypeFor[[]string]()),
		method("get", reflect.TypeFor[envelope]()),
		method("rows", reflect.TypeFor[listEnvelope]()),
		method("show", reflect.TypeFor[*detail]()),
		method("count", reflect.TypeFor[counted]()),
		method("values", reflect.TypeFor[values]()),
		method("map", reflect.TypeFor[map[string]string]()),
		method("hosts", reflect.TypeFor[[]host]()),
		method("doc", reflect.TypeFor[doc]()),
		method("spec", reflect.TypeFor[spec]()),
		method("terms", reflect.TypeFor[terms]()),
		method("tagged", reflect.TypeFor[tagged]()),
		method("described", reflect.TypeFor[described]()),
		method("nodes", reflect.TypeFor[[]node]()),
		method("kept", reflect.TypeFor[[]kept]()),
		keyed{method("delete", reflect.TypeFor[nothing]()), "machine_id"},
		method("drain", reflect.TypeFor[nothing]()).Says("draining"),
	}
	command.AddCommands(g)

	tests := []struct {
		name   string
		method string
		in     any
		answer string
		json   bool
		want   string
	}{
		{
			name:   "when the output is a list with its count, it is a table under the title",
			method: "list",
			answer: `{"total_items":5,"items":[{"name":"a","size":10,"note":"n"},{"name":"b","size":0}]}`,
			want:   "# Things (2 of 5)\nNAME | SIZE\na | 10B\nb | 0B\n",
		},
		{
			name:   "when the list is empty, it says None. and the hint",
			method: "list",
			answer: `{"total_items":0,"items":[]}`,
			want:   "# Things (0)\n\nNone. Add one.\n",
		},
		{
			name:   "when a list has no title, its heading is the group in the plural",
			method: "plain",
			answer: `{"items":[{"name":"a"}]}`,
			want:   "# Zzderives (1)\nNAME\na\n",
		},
		{
			name:   "when the output is a list, each item is a row",
			method: "each",
			answer: `[{"name":"a","size":1}]`,
			want:   "# Zzderives (1)\nNAME | SIZE\na | 1B\n",
		},
		{
			name:   "when the output is a list of values, each is a row",
			method: "names",
			answer: `["a","b"]`,
			want:   "# Zzderives (2)\nVALUE\na\nb\n",
		},
		{
			name:   "when the output is an envelope of an object, the object prints",
			method: "get",
			answer: `{"entry":{"value":"v"}}`,
			want:   "# Part\nValue: v\n",
		},
		{
			name:   "when the output is an envelope of a list, the list prints",
			method: "rows",
			answer: `{"rows":[{"name":"a"}]}`,
			want:   "# Zzderives (1)\nNAME\na\n",
		},
		{
			name:   "when the output is an object, it is a block, then its objects and lists",
			method: "show",
			answer: `{"id":"abcdefgh","last_seen":null,"skipped":"","off":false,"hidden":"x",` +
				`"part":{"value":"v"},"rows":[],"more":[{"name":"a","note":"n"}],"at":"2026-11-12T18:15:13Z"}`,
			want: "# Zzderive\nID: abcdefgh\nLast seen: never\nAt: 2026-11-12 18:15:13\n" +
				"# Part\nValue: v\n" +
				"# More (1)\nNAME\na\n",
		},
		{
			name:   "when a list with its count has no title, its heading is the group in the plural",
			method: "count",
			answer: `{"total_items":1,"items":[{"name":"a"}]}`,
			want:   "# Zzderives (1)\nNAME\na\n",
		},
		{
			name:   "when a list has more fields than its items and its count, they follow the table",
			method: "count",
			answer: `{"total_items":1,"counts":{"failed":1},"items":[{"name":"a"}]}`,
			want:   "# Zzderives (1)\nNAME\na\nCounts: failed:1\n",
		},
		{
			name:   "when a field with omitempty has its zero value, it is left out",
			method: "values",
			answer: `{"on":false,"count":0,"tags":[],"meta":{},"env":{},"name":"x"}`,
			want:   "# Zzderive\nName: x\n",
		},
		{
			name:   "when a field with omitempty has a value, it is shown",
			method: "values",
			answer: `{"on":true,"count":2,"tags":["a"],"meta":{"k":"v"},"env":{"a":"b"},"name":"x"}`,
			want:   "# Zzderive\nOn: yes\nCount: 2\nTags: a\nEnv: a:b\nName: x\n# Meta\nk: v\n",
		},
		{
			name:   "when the output is not an object or a list, it has only its heading",
			method: "map",
			answer: `{"a":"b"}`,
			want:   "# Zzderive\n",
		},
		{
			name:   "when a table cell is an object, it is its values; when a list of objects, the first field of each",
			method: "hosts",
			answer: `[{"name":"a","os":{"name":"Ubuntu","version":"24.04","size":2048},` +
				`"conds":[{"type":"HighLoad","reason":"r"},{"type":""},{"type":"DiskPressure"}]},` +
				`{"name":"b","os":"text","conds":"text"}]`,
			want: "# Zzderives (2)\nNAME | OS | CONDS\n" +
				"a | Ubuntu 24.04 2048B | HighLoad, DiskPressure\n" +
				"b | text | text\n",
		},
		{
			name:   "when a column has a none tag, it stays with no value in any row",
			method: "kept",
			answer: `[{"name":"a"},{"name":"b"}]`,
			want:   "# Zzderives (2)\nNAME | CONDS\na | -\nb | -\n",
		},
		{
			name:   "when a row has children of its own type, they are a tree under it",
			method: "nodes",
			answer: `[{"name":"controller","status":"ok","children":[` +
				`{"name":"api","status":"ok","children":[{"name":"v1","status":"ok"}]},` +
				`{"name":"metrics","status":"ok","children":[{"name":"m","status":"ok"}]}]},` +
				`{"name":"nats","status":"ok"}]`,
			want: "# Zzderives (2)\nNAME | STATUS\n" +
				"controller | ok\n" +
				"├─ api | ok\n" +
				"│  └─ v1 | ok\n" +
				"└─ metrics | ok\n" +
				"   └─ m | ok\n" +
				"nats | ok\n",
		},
		{
			name:   "when a field is tagged render:\"body\", it is the whole output, as it is",
			method: "doc",
			answer: `{"name":"a","body":"# Title\n\ntext\n"}`,
			want:   "# Title\n\ntext\n",
		},
		{
			name:   "when a list is tagged sections and blocks, each item is a section of blocks",
			method: "spec",
			answer: `{"kinds":[{"kind":"o","title":"Options","terms":[` +
				`{"name":"Flat()","does":"Adds no word.","when":"Alone."},{"name":"Long()","does":"Help."}]}]}`,
			want: "# Options\n* Flat()\n  | Adds no word.\n  Use| Alone.\n* Long()\n  | Help.\n",
		},
		{
			name:   "when a list of values is empty, it has no value, and none says so",
			method: "tagged",
			answer: `{"tags":[]}`,
			want:   "# Zzderive\nTags: none\n",
		},
		{
			name:   "when a list of blocks is the output, it has its heading with the count",
			method: "terms",
			answer: `{"terms":[{"name":"Flat()","does":"Adds no word."}]}`,
			want:   "# Zzderives (1)\n* Flat()\n  | Adds no word.\n",
		},
		{
			name:   "when a section has more than its list, its values print, then its list",
			method: "described",
			answer: `{"name":"token","commands":[{"command":"osapi controller token list","method":"list","flags":[{"flag":"--x"}]}]}`,
			want:   "# Zzderive\nName: token\n# osapi controller token list\nMethod: list\n# Flags (1)\nFLAG\n--x\n",
		},
		{
			name:   "when the method answers nothing, one line names the key and what it did",
			method: "delete",
			in:     deleteIn{MachineID: "m-1"},
			want:   "[ok] m-1 deleted\n",
		},
		{
			name:   "when the method has no key, the line names the group, and Says what it is",
			method: "drain",
			in:     struct{}{},
			want:   "[ok] zzderive draining\n",
		},
		{
			name:   "when --json is given to a method that answers nothing, it prints the key and the result",
			method: "delete",
			in:     deleteIn{MachineID: "m-1"},
			json:   true,
			want:   `{"machine_id":"m-1","result":"deleted"}`,
		},
		{
			name:   "when --json is given to a method with no key, the key is name",
			method: "drain",
			in:     struct{}{},
			json:   true,
			want:   `{"name":"zzderive","result":"draining"}`,
		},
		{
			name:   "when the method is not registered, the answer prints as JSON",
			method: "missing",
			answer: `{"a":1}`,
			want:   `{"a":1}`,
		},
		{
			name:   "when the answer does not decode, the error is reported",
			method: "list",
			answer: `{`,
			want:   "error: decode zzderive list: unexpected end of JSON input",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			r := &recorder{}
			s.Require().NoError(command.Attach(&cobra.Command{Use: "client"}, command.Deps{
				JSON:      func() bool { return tc.json },
				Error:     func(err error) { fmt.Fprintf(&r.b, "error: %s", err) },
				PrintJSON: func(body []byte) { r.b.Write(body) },
				Print:     r,
			}))
			var answer []byte
			if tc.answer != "" {
				answer = []byte(tc.answer)
			}

			command.Output(g, tc.method, tc.in, answer)

			s.Equal(tc.want, r.b.String())
		})
	}
}

func (s *OutputPublicTestSuite) TestFormat() {
	type fields struct {
		Bytes  int        `json:"bytes"  render:"bytes"`
		Age    int        `json:"age"    render:"age"`
		Dur    int64      `json:"dur"    render:"duration"`
		Ms     int        `json:"ms"     render:"duration=ms"`
		Status string     `json:"status" render:"status"`
		Short  string     `json:"short"  render:"short=3"`
		At     *time.Time `json:"at"`
		When   string     `json:"when"   render:"time"`
		Ref    string     `json:"ref"                         prefix:"@fact."`
		Plain  any        `json:"plain"`
		Load   float32    `json:"load"`
		Ratio  *float64   `json:"ratio"`
		hidden string     //nolint:unused // an unexported field is not a column
		NoJSON string
		Skip   string `json:"-"`
	}
	cols := map[string]command.Column{}
	for _, c := range command.Columns(reflect.TypeFor[*[]fields]()) {
		cols[c.JSON] = c
	}
	s.Len(cols, 12)
	s.Nil(command.Columns(reflect.TypeFor[string]()))
	r := &recorder{}
	recent := time.Now().Add(-30 * time.Second).Format(time.RFC3339)

	tests := []struct {
		name  string
		col   string
		value any
		want  string
	}{
		{"a size", "bytes", float64(3480), "3480B"},
		{"an age in seconds", "age", float64(30), "30s ago"},
		{"an age from a time", "age", recent, "30s ago"},
		{"an age that is not a time", "age", "soon", "soon"},
		{"a duration in nanoseconds", "dur", float64(time.Second), "1s"},
		{"a duration as a string", "dur", "1m0s", "1m0s"},
		{"a duration that does not parse", "dur", "long", "long"},
		{"milliseconds", "ms", float64(42), "42ms"},
		{"a status", "status", "ok", "<ok>"},
		{"a long value cut", "short", "abcdef", "abc..."},
		{"a short value kept", "short", "ab", "ab"},
		{"a time by its type", "at", "2026-11-12T18:15:13Z", "2026-11-12 18:15:13"},
		{"a zero time is no value", "at", "0001-01-01T00:00:00Z", ""},
		{"a time by its tag", "when", "2026-11-12T18:15:13Z", "2026-11-12 18:15:13"},
		{"a time that does not parse", "when", "later", "later"},
		{"a prefix", "ref", "x", "@fact.x"},
		{"no value", "plain", nil, ""},
		{"an empty string", "plain", "", ""},
		{"true", "plain", true, "yes"},
		{"false", "plain", false, "no"},
		{"a whole number", "plain", float64(4), "4"},
		{"a fraction", "plain", 0.25, "0.25"},
		{"a float field with a whole value", "load", float64(0), "0.00"},
		{"a float field", "load", 0.5, "0.50"},
		{"a pointer to a float field", "ratio", 1.25, "1.25"},
		{"a list of values", "plain", []any{"a", float64(1)}, "a, 1"},
		{"a list of objects", "plain", []any{map[string]any{}}, "1 item"},
		{"a list of lists", "plain", []any{[]any{}, []any{}}, "2 items"},
		{"an object of values", "plain", map[string]any{"b": "2", "a": "1"}, "a:1, b:2"},
		{"an object of objects", "plain", map[string]any{"a": map[string]any{}}, "1 field"},
		{"another value", "plain", "text", "text"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, cols[tc.col].Format(r, tc.value))
		})
	}

	s.Equal("abcdef", cols["short"].FormatKV(r, "abcdef"), "a block does not cut a value")
}

func (s *OutputPublicTestSuite) TestPlural() {
	g := outGroup{group: group{name: "zzlogs"}}
	m := command.Method[struct{}, struct{}]("list").Handle(handle[struct{}])
	m.Out = reflect.TypeFor[[]string]()
	g.cmds = []command.Command{m}
	command.AddCommands(g)
	r := &recorder{}
	s.Require().NoError(command.Attach(&cobra.Command{Use: "client"}, command.Deps{
		JSON:  func() bool { return false },
		Print: r,
	}))

	command.Output(g, "list", struct{}{}, []byte(`["a"]`))

	s.Equal("# Zzlogs (1)\nVALUE\na\n", r.b.String(), "a name in s is its own plural")
}

func (s *OutputPublicTestSuite) TestPrintsGroups() {
	g := outGroup{group: group{name: "zzprinted"}}
	m := command.Method[struct{}, struct{}]("list").Handle(handle[struct{}])
	m.Out = reflect.TypeFor[rows]()
	g.cmds = []command.Command{m}
	command.AddCommands(g)
	printed := ""
	command.PrintsGroups(func(group command.Group, method string, answer []byte) bool {
		if group.GroupName() != "zzprinted" {
			return false
		}
		printed = method + ":" + string(answer)

		return true
	})
	s.Require().NoError(command.Attach(&cobra.Command{Use: "client"}, command.Deps{
		JSON: func() bool { return false },
	}))

	command.Output(g, "list", struct{}{}, []byte(`{}`))

	s.Equal("list:{}", printed)
}

// TestDerive prints an answer of a method from its output type, with an age
// from the time that SetNow fixes, so a docs page is the same on each run.
func (s *OutputPublicTestSuite) TestDerive() {
	type aged struct {
		Name string `json:"name"`
		Seen string `json:"seen" render:"age"`
	}
	m := command.Method[struct{}, aged]("get")
	g := command.Define("zzderive", "Derive", m)
	t0 := time.Date(2026, time.January, 15, 10, 0, 0, 0, time.UTC)

	command.SetNow(t0)
	s.Equal(t0, command.Now())
	r := &recorder{}
	s.Require().NoError(command.Derive(r, g, m, struct{}{},
		[]byte(`{"name":"web-01","seen":"2026-01-15T09:59:30Z"}`)))
	s.Contains(r.b.String(), "30s ago")
	s.Error(
		command.Derive(&recorder{}, g, m, struct{}{}, []byte(`{"name":`)),
		"an answer that is not JSON",
	)

	command.SetNow(time.Time{})
	s.WithinDuration(time.Now(), command.Now(), time.Minute, "the clock, with no fixed time")
}

func (s *OutputPublicTestSuite) TestDone() {
	type keyed struct {
		Name  string `json:"name"`
		Force bool   `json:"force"`
		Count int    `json:"count"`
	}
	type named struct {
		Name string `json:"name"`
	}
	drain := command.Method[keyed, struct{}]("drain").Says("draining")
	g := command.Define("zzdone", "Done",
		drain,
		command.Method[keyed, named]("get"),
		command.Method[keyed, *struct{}]("frob"),
		command.Method[string, struct{}]("delete"),
	)

	tests := []struct {
		name   string
		method command.Command
		flags  string
		wantIn any
		want   string
		ok     bool
	}{
		{
			name:   "when the flag and its value are two words, after a bool flag",
			method: drain,
			flags:  "--force --name \"web-01\" --count 3",
			wantIn: keyed{Name: "web-01"},
			want:   `{"name":"zzdone","result":"draining"}`,
			ok:     true,
		},
		{
			name:   "when the flag has its value after =",
			method: drain,
			flags:  "-j --name=web-02",
			wantIn: keyed{Name: "web-02"},
			want:   `{"name":"zzdone","result":"draining"}`,
			ok:     true,
		},
		{
			name:   "when the last flag has no value",
			method: drain,
			flags:  "--name",
			wantIn: keyed{},
			want:   `{"name":"zzdone","result":"draining"}`,
			ok:     true,
		},
		{
			name:   "when the input is not a struct",
			method: g.Method("delete"),
			flags:  "--name web-01",
			wantIn: "",
			want:   `{"name":"zzdone","result":"deleted"}`,
			ok:     true,
		},
		{name: "when the method answers with a body", method: g.Method("get"), flags: "--name a"},
		{name: "when the output is a pointer", method: g.Method("frob"), flags: "--name a"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			in, body, ok := command.Done(g, tc.method, tc.flags)
			s.Equal(tc.ok, ok)
			if !tc.ok {
				return
			}
			s.Equal(tc.wantIn, in)
			s.JSONEq(tc.want, string(body))
		})
	}
}

func (s *OutputPublicTestSuite) TestRender() {
	r := &recorder{}
	s.Require().NoError(command.Render(r, "zzrender", &plain{Items: []row{{Name: "web-01"}}}))
	s.Contains(r.b.String(), "web-01", "a pointer prints as its value, as the CLI prints it")

	s.Error(command.Render(&recorder{}, "zzrender", func() {}), "a value JSON cannot hold")
}

func TestOutputPublicTestSuite(t *testing.T) {
	suite.Run(t, new(OutputPublicTestSuite))
}
