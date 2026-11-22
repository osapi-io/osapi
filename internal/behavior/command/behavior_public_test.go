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
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/spf13/cobra"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/command"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
	"github.com/osapi-io/osapi/internal/config"
)

type provider struct{}

type input struct {
	Key     string   `json:"key"   validate:"required" doc:"The key"`
	Force   bool     `json:"force"                     doc:"Force it"`
	Count   int      `json:"count"`
	Ratio   float32  `json:"ratio"`
	Tags    []string `json:"tags"`
	Big     int64    `json:"big"`
	NoJSON  string
	Skipped string `json:"-"`
}

type entry struct {
	Key   string  `json:"key"`
	Size  int     `json:"size"  render:"bytes"`
	Age   int     `json:"age"   render:"age"`
	Ratio float64 `json:"ratio"`
	Note  string  `json:"note"  render:"-"`
	Extra string  `json:"extra"`
}

type result struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
}

func read(_ context.Context, _ provider, _ input) ([]entry, error)     { return nil, nil }
func readOne(_ context.Context, _ provider, _ input) (*entry, error)   { return nil, nil }
func write(_ context.Context, _ provider, _ input) (result, error)     { return result{}, nil }
func writeEntry(_ context.Context, _ provider, _ input) (entry, error) { return entry{}, nil }
func readText(_ context.Context, _ provider, _ input) (string, error)  { return "", nil }

type BehaviorcmdPublicTestSuite struct {
	suite.Suite
}

func (s *BehaviorcmdPublicTestSuite) TestFlags() {
	tests := []struct {
		name string
		t    reflect.Type
		want []command.Flag
	}{
		{
			name: "when the input has every kind of field",
			t:    reflect.TypeFor[input](),
			want: []command.Flag{
				{
					Name:     "key",
					JSON:     "key",
					Kind:     reflect.String,
					Usage:    "The key",
					Required: true,
					Index:    0,
				},
				{Name: "force", JSON: "force", Kind: reflect.Bool, Usage: "Force it", Index: 1},
				{Name: "count", JSON: "count", Kind: reflect.Int, Index: 2},
				{Name: "tags", JSON: "tags", Kind: reflect.Slice, Index: 4},
			},
		},
		{
			name: "when a field is a map of strings, it is a key=value flag",
			t: reflect.TypeFor[struct {
				Vars map[string]string `json:"vars" doc:"Template vars"`
				Nums map[string]int    `json:"nums"`
			}](),
			want: []command.Flag{
				{Name: "vars", JSON: "vars", Kind: reflect.Map, Usage: "Template vars", Index: 0},
			},
		},
		{
			name: "when a field is secret, its help says so",
			t: reflect.TypeFor[struct {
				Token string `json:"token" doc:"The token" secret:"true"`
			}](),
			want: []command.Flag{{
				Name:  "token",
				JSON:  "token",
				Kind:  reflect.String,
				Usage: "The token (secret: kept out of the audit log and traces)",
			}},
		},
		{name: "when the input is not a struct", t: reflect.TypeFor[string]()},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, command.Flags(tc.t))
		})
	}
}

func (s *BehaviorcmdPublicTestSuite) TestArgs() {
	s.Nil(command.Args(nil))
	s.Equal(
		[]string{"--key", "k", "--count", "2", "--tags", "a", "--tags", "b"},
		command.Args(input{Key: "k", Count: 2, Tags: []string{"a", "b"}, Ratio: 0.5, Big: 9}),
		"zero fields and fields no flag holds are left out; a list is the flag once each",
	)
	s.Equal(
		[]string{"--vars", "a=1", "--vars", "b=2"},
		command.Args(struct {
			Vars map[string]string `json:"vars"`
		}{Vars: map[string]string{"b": "2", "a": "1"}}),
		"a map is one flag for each key, in order",
	)
}

func (s *BehaviorcmdPublicTestSuite) TestColumns() {
	tests := []struct {
		name string
		t    reflect.Type
		want []string
	}{
		{
			name: "when a list hides render:- and changed",
			t:    reflect.TypeFor[[]entry](),
			want: []string{"key", "size", "age", "ratio", "extra"},
		},
		{name: "when a pointer", t: reflect.TypeFor[*result](), want: []string{"key"}},
		{name: "when not a struct", t: reflect.TypeFor[string]()},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var got []string
			for _, c := range command.Columns(tc.t) {
				got = append(got, c.JSON)
			}
			s.Equal(tc.want, got)
		})
	}
}

func (s *BehaviorcmdPublicTestSuite) TestTable() {
	list := behavior.Read("list", read)
	create := behavior.Write("create", write)
	tests := []struct {
		name         string
		m            *behavior.Method
		r            command.Result
		validateFunc func(sec cli.Section)
	}{
		{
			name: "when a read renders every column by its render tag",
			m:    list,
			r: command.Result{Results: []map[string]any{
				{
					"hostname": "web-01",
					"status":   "ok",
					"key":      "k",
					"size":     float64(2048),
					"age":      float64(90),
					"ratio":    0.5,
					"extra":    nil,
				},
				{"hostname": "web-02", "status": "skipped", "error": "unsupported"},
				{"hostname": "web-03", "status": "ok", "ratio": float64(2)},
			}},
			validateFunc: func(sec cli.Section) {
				s.Equal(
					[]string{"HOSTNAME", "STATUS", "KEY", "SIZE", "AGE", "RATIO", "EXTRA"},
					sec.Headers,
				)
				s.Equal([]string{"web-01", "ok", "k", "2.0 KB", "1m", "0.50", ""}, sec.Rows[0])
				s.Equal("2.00", sec.Rows[1][5], "a float prints with two decimals")
				s.Len(sec.Errors, 1)
			},
		},
		{
			name: "when a write folds changed into the status",
			m:    create,
			r: command.Result{Results: []map[string]any{
				{"hostname": "web-01", "status": "ok", "key": "k", "changed": true},
			}},
			validateFunc: func(sec cli.Section) {
				s.Equal("changed", sec.Rows[0][1])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(command.Table(tc.m, tc.r))
		})
	}
}

func (s *BehaviorcmdPublicTestSuite) TestTitle() {
	s.Equal("DNS", command.Title(behavior.Define("dns", "x")), "an initialism is in capitals")
	s.Equal("Sysctl", command.Title(behavior.Define("sysctl", "x")))
}

func (s *BehaviorcmdPublicTestSuite) TestDocs() {
	b := behavior.Define(
		"kv",
		"Key values",
		behavior.Node,
		behavior.API,
		behavior.CLI,
		behavior.Under("store"),
		behavior.Read("list", read).Describe("List <them>").Explain("All of {them}.").
			Example("Every key", input{Key: "k", Force: true, Count: 2}, []entry{{Key: "k"}}).
			Example("No keys", input{Key: "none"}, []entry{}),
		behavior.Platform(
			"debian",
			func(host.Host) (provider, error) { return provider{}, nil },
		),
		behavior.Sudo("kvctl"),
		behavior.Runs("kvstat"),
		behavior.RunsOnPlatform(behavior.PlatformDarwin, "vm_stat"),
		behavior.RunsIfPresent("zfs"),
		behavior.Write("create", write).Describe("Create one").On("debian").
			Sudo("kvsync").Runs("kvcheck").
			Example("One key", input{Key: "k"}, result{Key: "k"},
				behavior.Wrote("/var/kv/k", "v\n"),
				behavior.Removed("/var/kv/old"),
				behavior.Ran("kvctl", "sync")).
			Example("A | B", input{Key: "p"}, result{Key: "p"},
				behavior.Wrote("/var/kv/p", "a\nb\n")).
			Fails("A key that is taken", input{Key: "k"}, "already exists").
			Fails("No key", input{}, "key must not be empty").
			Fails("A key on two lines", input{Key: "a\nb"}, "line breaks"),
	)

	s.Equal(
		"  # Every key\n  osapi client node store kv list --target web-01 --key k --force --count 2\n\n"+
			"  # No keys\n  osapi client node store kv list --target web-01 --key none",
		command.Example(b, b.Method("list")),
	)
	list := command.Annotations(b, b.Method("list"))
	s.Equal("store:read", list[reference.Permission])
	out := list[reference.Result]
	s.True(strings.HasPrefix(
		out,
		"Every key, on one host:\n\n```ansi\n$ osapi client node store kv list --target web-01 --key k --force --count 2\n",
	),
		"each example's command and output, in a block that the site shows in color")
	s.Contains(out, "On each host, where one cannot run it:\n\n```ansi\n"+
		"$ osapi client node store kv list --target _all --key k --force --count 2\n",
		"the fleet's output")
	s.Contains(out, "web-02", "the fleet's second host answers")
	s.Contains(out, "mac-01", "the fleet's Mac")
	s.Contains(out, "skipped")
	s.Contains(out, "operation not supported on this OS family")
	s.Contains(out, "With `--json`, the API's response:\n\n```json\n{\n  \"job_id\"",
		"the response as JSON")
	s.Contains(out, "\x1b[", "the CLI's colors")
	s.Contains(out, "web-01", "one host answers")
	s.Contains(out, "job 550e8400-e29b-41d4-a716-446655440000")
	s.Contains(out, "trace 26e448c9dbe1c4a8d216a98c675079c9")
	s.Contains(out, "No keys, on one host:\n\n```ansi\n", "an empty list still shows its output")

	create := command.Annotations(b, b.Method("create"))
	s.Equal("store:write", create[reference.Permission])
	result := create[reference.Result]
	s.Contains(
		result,
		"| One key | writes `/var/kv/k`: `v`; removes `/var/kv/old`; runs `kvctl sync` |",
	)
	s.Contains(result, "| A key that is taken | `--key k` | already exists |")
	s.Contains(result, "| No key |  | key must not be empty |")
	s.Contains(result, `| A key on two lines | `+"`"+`--key "a\nb"`+"`"+` | line breaks |`,
		"a newline is quoted, not printed into the table row")
	s.Contains(result, `| A \| B | writes `+"`"+`/var/kv/p`+"`"+`: `+"`"+`a\nb`+"`"+` |`,
		"a pipe and a newline stay in their cell")

	// A behavior that is not served by the API needs no permission, and an
	// example with no input runs the command bare.
	b.API = false
	bare := behavior.Write("apply", writeEntry).
		Describe("Apply").
		Example("Apply all", nil, entry{Key: "k"})
	s.NotContains(command.Annotations(b, bare), reference.Permission)
	s.Contains(command.Annotations(b, bare)[reference.Result], "changed",
		"a write's row has changed")

	// An output that is not an object, such as text, has no table to show.
	text := behavior.Read("text", readText).Example("Some text", nil, "hello")
	s.Empty(command.Output(b, text))
	b.API = true
	s.Equal(
		"  # Apply all\n  osapi client node store kv apply --target web-01",
		command.Example(b, bare),
	)
	s.Empty(command.Example(b, behavior.Read("get", readOne)), "no examples")

	s.Equal(
		map[string]string{
			reference.Notes: "The [kv behavior](../../../../../../reference/behaviors/kv/kv.md) page tells " +
				"where the behavior runs, what it is built on and its settings.",
		},
		command.GroupAnnotations(b),
	)
}

// TestScreen renders the osapi:output marker of a page that a person wrote:
// one host, the fleet with hosts that skip and fail, the --json body, and a
// value that is the example's host becomes each host's own.
func (s *BehaviorcmdPublicTestSuite) TestScreen() {
	behavior.Register(behavior.Define("zzscreen", "Screens", behavior.Node, behavior.API,
		behavior.Read("list", read).Example("Every key", input{Key: "k"}, []entry{{Key: "web-01"}}),
		behavior.Read("bare", read),
	))

	tests := []struct {
		name     string
		args     []string
		contains []string
		missing  []string
	}{
		{
			name:     "when the marker names too little",
			args:     []string{"zzscreen", "list"},
			contains: []string{"osapi:output needs a behavior, a method and a host"},
		},
		{
			name:     "when no behavior has the name",
			args:     []string{"nope", "list", "web-01"},
			contains: []string{"no behavior nope with a method list and an example"},
		},
		{
			name:     "when the method has no example",
			args:     []string{"zzscreen", "bare", "web-01"},
			contains: []string{"no behavior zzscreen with a method bare"},
		},
		{
			name: "when one host answers",
			args: []string{"zzscreen", "list", "web-01"},
			contains: []string{
				"```ansi\n$ osapi client node zzscreen list --target web-01 --key k\n",
				"\x1b[",
			},
		},
		{
			name: "when the fleet answers, with a host that skips and one that fails",
			args: []string{"zzscreen", "list", "db-01", "mac-01:skipped", "web-02:failed"},
			contains: []string{
				"--target _all --key k",
				"db-01", "operation not supported on this OS family", "permission denied",
			},
			missing: []string{"web-01"},
		},
		{
			name: "when the marker asks for the JSON body",
			args: []string{"zzscreen", "list", "web-01", "--json"},
			contains: []string{
				"$ osapi client node zzscreen list --target web-01 --key k --json\n{\"job_id\":",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := command.Screen(tc.args)
			for _, c := range tc.contains {
				s.Contains(got, c)
			}
			for _, m := range tc.missing {
				s.NotContains(got, m, "a value that is the example's host is each host's own")
			}
		})
	}
}

// TestPrint writes a method's answer as the job, the trace and the table,
// which the endpoint command runtime calls as each behavior method's
// renderer.
func (s *BehaviorcmdPublicTestSuite) TestPrint() {
	b := behavior.Define("kv", "Key values", behavior.Node, behavior.Read("list", read))
	out := map[string]any{
		"job_id":  "job-1",
		"results": []map[string]any{{"hostname": "web-01", "status": "ok", "key": "k"}},
	}

	got := cli.Capture(func() { command.Print(b, "list", out, "4bf92f3577b34da6a3ce929d0e0e4736") })

	s.Contains(got, "job-1")
	s.Contains(got, "4bf92f3577b34da6a3ce929d0e0e4736")
	s.Contains(got, "web-01")
}

// TestPrintResults prints a behavior's answer as its result table through
// the command layer, which registers no renderer for it.
func (s *BehaviorcmdPublicTestSuite) TestPrintResults() {
	behavior.Reset()
	defer behavior.Reset()
	behavior.Register(
		behavior.Define("kv", "Key values", behavior.Node, behavior.Read("list", read)),
	)
	printed := ""
	s.Require().NoError(clicommand.Attach(&cobra.Command{Use: "client"}, clicommand.Deps{
		JSON:      func() bool { return false },
		PrintJSON: func(body []byte) { printed = string(body) },
	}))
	answer := []byte(`{"job_id":"job-1","results":[{"hostname":"web-01","status":"ok","key":"k"}]}`)

	tests := []struct {
		name   string
		group  string
		method string
		want   string
	}{
		{"when the group is a behavior, its result table prints", "kv", "list", "web-01"},
		{"when the behavior has no such method, the answer prints as JSON", "kv", "get", ""},
		{"when the group is not a behavior, the answer prints as JSON", "zzother", "list", ""},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			printed = ""
			got := cli.Capture(func() {
				clicommand.Output(tc.group, tc.method, struct{}{}, answer)
			})

			if tc.want != "" {
				s.Contains(got, tc.want)
				s.Empty(printed)

				return
			}
			s.Equal(string(answer), printed)
		})
	}
}

// kvSettings is a behavior's settings, one of each kind of flag.
type kvSettings struct {
	Dir     string   `mapstructure:"dir"     doc:"Where it writes."`
	Level   int      `mapstructure:"level"   doc:"How much it says."`
	Verbose bool     `mapstructure:"verbose" doc:"Say everything."`
	Ratio   float64  `mapstructure:"ratio"   doc:"How much."`
	Tags    []string `mapstructure:"tags"    doc:"Labels."`
}

func (s *BehaviorcmdPublicTestSuite) TestSettingFlags() {
	behavior.Reset()
	defer behavior.Reset()
	b := behavior.Register(behavior.Define("kv", "Key values", behavior.Node,
		behavior.Config(kvSettings{Dir: "/var/kv", Level: 1, Ratio: 0.5, Tags: []string{"a"}})))

	fs := pflag.NewFlagSet("agent", pflag.ContinueOnError)
	v := viper.New()
	command.SettingFlags(fs, v)

	for _, name := range []string{"kv-dir", "kv-level", "kv-verbose", "kv-ratio", "kv-tags"} {
		s.NotNil(fs.Lookup(name), name)
	}
	s.Equal("/var/kv", fs.Lookup("kv-dir").DefValue)

	s.Require().
		NoError(fs.Parse([]string{"--kv-dir", "/srv/kv", "--kv-level", "4", "--kv-verbose"}))
	var c config.Config
	_, err := config.Read(v, filepath.Join(s.T().TempDir(), "missing.yaml"), &c)
	s.Require().NoError(err)

	got := behavior.Settings[kvSettings](b)
	s.Equal("/srv/kv", got.Dir, "a flag wins over the default")
	s.Equal(4, got.Level)
	s.True(got.Verbose)
	s.Equal(0.5, got.Ratio)
	s.Equal([]string{"a"}, got.Tags)
}

// TestExampleFlags is an example's flags as the help writes them: the
// target, then each field the input sets.
func (s *BehaviorcmdPublicTestSuite) TestExampleFlags() {
	s.Equal("--target web-01", command.ExampleFlags(input{}))
	s.Equal("--target web-01 --key k", command.ExampleFlags(input{Key: "k"}))
}

func TestBehaviorcmdPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorcmdPublicTestSuite))
}
