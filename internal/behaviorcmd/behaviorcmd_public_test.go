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

package behaviorcmd_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
	"github.com/osapi-io/osapi/internal/behaviorcmd/mocks"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type provider struct{}

type input struct {
	Key     string   `json:"key"   validate:"required" doc:"The key"`
	Force   bool     `json:"force"                     doc:"Force it"`
	Count   int64    `json:"count"`
	Ratio   float32  `json:"ratio"`
	Tags    []string `json:"tags"`
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

func read(_ context.Context, _ provider, _ input) ([]entry, error)      { return nil, nil }
func readOne(_ context.Context, _ provider, _ input) (*entry, error)    { return nil, nil }
func readString(_ context.Context, _ provider, _ input) (string, error) { return "", nil }
func write(_ context.Context, _ provider, _ input) (result, error)      { return result{}, nil }
func writeEntry(_ context.Context, _ provider, _ input) (entry, error)  { return entry{}, nil }

type BehaviorcmdPublicTestSuite struct {
	suite.Suite
}

func (s *BehaviorcmdPublicTestSuite) TestFlags() {
	tests := []struct {
		name string
		t    reflect.Type
		want []behaviorcmd.Flag
	}{
		{
			name: "when the input has every kind of field",
			t:    reflect.TypeFor[input](),
			want: []behaviorcmd.Flag{
				{
					Name:     "key",
					JSON:     "key",
					Kind:     reflect.String,
					Usage:    "The key (required)",
					Required: true,
					Index:    0,
				},
				{Name: "force", JSON: "force", Kind: reflect.Bool, Usage: "Force it", Index: 1},
				{Name: "count", JSON: "count", Kind: reflect.Int, Index: 2},
				{Name: "ratio", JSON: "ratio", Kind: reflect.Float64, Index: 3},
			},
		},
		{
			name: "when a field is secret, its help says so",
			t: reflect.TypeFor[struct {
				Token string `json:"token" doc:"The token" secret:"true"`
			}](),
			want: []behaviorcmd.Flag{{
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
			s.Equal(tc.want, behaviorcmd.Flags(tc.t))
		})
	}
}

func (s *BehaviorcmdPublicTestSuite) TestArgs() {
	s.Nil(behaviorcmd.Args(nil))
	s.Equal(
		[]string{"--key", "k", "--count", "2"},
		behaviorcmd.Args(input{Key: "k", Count: 2, Tags: []string{"a"}}),
		"zero fields and fields no flag holds are left out",
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
			for _, c := range behaviorcmd.Columns(tc.t) {
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
		r            behaviorcmd.Result
		validateFunc func(sec cli.Section)
	}{
		{
			name: "when a read renders every column by its render tag",
			m:    list,
			r: behaviorcmd.Result{Results: []map[string]any{
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
			}},
			validateFunc: func(sec cli.Section) {
				s.Equal(
					[]string{"HOSTNAME", "STATUS", "KEY", "SIZE", "AGE", "RATIO", "EXTRA"},
					sec.Headers,
				)
				s.Equal([]string{"web-01", "ok", "k", "2.0 KB", "1m", "0.5", ""}, sec.Rows[0])
				s.Len(sec.Errors, 1)
			},
		},
		{
			name: "when a write folds changed into the status",
			m:    create,
			r: behaviorcmd.Result{Results: []map[string]any{
				{"hostname": "web-01", "status": "ok", "key": "k", "changed": true},
			}},
			validateFunc: func(sec cli.Section) {
				s.Equal("changed", sec.Rows[0][1])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(behaviorcmd.Table(tc.m, tc.r))
		})
	}
}

func (s *BehaviorcmdPublicTestSuite) TestANSIToHTML() {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "when plain text is escaped", in: "a <b>", want: "a &lt;b&gt;"},
		{
			name: "when bold and true color",
			in:   "\x1b[1;38;2;0;213;255mHi\x1b[0m",
			want: `<span style="color:#00d5ff;font-weight:bold">Hi</span>`,
		},
		{
			name: "when a 256 color",
			in:   "\x1b[38;5;222mskip\x1b[39m!",
			want: `<span style="color:#ffdf87">skip</span>!`,
		},
		{
			name: "when bold is turned off",
			in:   "\x1b[1mA\x1b[22mB",
			want: `<span style="font-weight:bold">A</span>B`,
		},
		{name: "when a sequence is not finished", in: "a\x1b[1", want: "a"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, behaviorcmd.ANSIToHTML(tc.in))
		})
	}
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
			Example("Every key", input{Key: "k", Force: true, Count: 2}, []entry{{Key: "k"}}),
		behavior.Platform(
			"debian",
			func(host.Host) (provider, error) { return provider{}, nil },
		),
		behavior.Sudo("kvctl"),
		behavior.Runs("kvstat"),
		behavior.Write("create", write).Describe("Create one").
			Example("One key", input{Key: "k"}, result{Key: "k"},
				behavior.Wrote("/var/kv/k", "v\n"),
				behavior.Removed("/var/kv/old"),
				behavior.Ran("kvctl", "sync")).
			Fails("A key that is taken", input{Key: "k"}, "already exists").
			Fails("No key", input{}, "key must not be empty").
			Fails("A key on two lines", input{Key: "a\nb"}, "line breaks"),
	)

	s.Equal("docs/docs/sidebar/behaviors/kv", behaviorcmd.DocsDir(b))
	index := behaviorcmd.IndexPage(b)
	s.Contains(index, "# Kv")
	s.Contains(index, "| API | `/api/next/node/{hostname}/store/kv`")
	s.Contains(index, "| CLI | `osapi client next node store kv` |")
	s.Contains(index, "| [`list`](list.mdx) | read | `store:read` | List &lt;them&gt; |")
	s.Contains(index, "None. A behavior declares settings with `behavior.Config`.")
	s.Contains(index, "On `debian`. On any other platform its methods report `skipped`.")
	s.Contains(index, "It runs `kvctl` with sudo")
	s.Contains(index, "It runs `kvstat`, which the host must have installed.")

	// A behavior with settings lists them; one without API or CLI says
	// only where its jobs run.
	behavior.Reset()
	defer behavior.Reset()
	files := behavior.Register(behavior.Define(
		"files", "Files it keeps", behavior.Node, behavior.Internal,
		behavior.Read("get", read).Describe("Read one"),
	))
	quiet := behavior.Register(behavior.Define(
		"quiet", "Runs only as jobs", behavior.Node,
		behavior.Requires(files),
		behavior.Config(kvSettings{Dir: "/var/kv"}),
		behavior.Read("list", read),
	))
	apiOnly := behavior.Define("served", "Served, no commands", behavior.Node, behavior.API,
		behavior.Read("get", read))

	internal := behaviorcmd.IndexPage(files)
	s.Contains(internal, "Nowhere: it is internal.")
	s.NotContains(internal, "| Jobs |")
	s.Contains(internal, "| Method | Verb | Description |")
	s.Contains(internal, "| `get` | read | Read one |", "no method page to link to")
	s.Contains(
		internal,
		"| Nothing | [quiet](../quiet/quiet.md) |",
		"what it is built on and used by",
	)
	s.NotContains(internal, "<DocCardList />")
	s.Contains(behaviorcmd.IndexPage(quiet), "| [files](../files/files.md) | Nothing |")

	catalog := behaviorcmd.CatalogPage([]*behavior.Behavior{files, b, quiet, apiOnly})
	s.Contains(catalog, "| [files](files/files.md) | Files it keeps | Internal | - | quiet |")
	s.Contains(catalog, "| [kv](kv/kv.md) | Key values | API and CLI | - | - |")
	s.Contains(catalog, "| [quiet](quiet/quiet.md) | Runs only as jobs | Jobs only | files | - |")
	s.Contains(catalog, "| [served](served/served.md) | Served, no commands | API | - | - |")

	index = behaviorcmd.IndexPage(quiet)
	s.Contains(index, "Nowhere yet")
	s.NotContains(index, "| API |")
	s.NotContains(index, "| CLI |")
	s.Contains(
		index,
		"| `behaviors.quiet.dir` | `--quiet-dir` | `OSAPI_BEHAVIORS_QUIET_DIR` | string | `/var/kv` | Where it writes. |",
	)
	s.Contains(
		index,
		"| `behaviors.quiet.ratio` | `--quiet-ratio` | `OSAPI_BEHAVIORS_QUIET_RATIO` | float | `0` |",
	)

	page, err := behaviorcmd.MethodPage(b, b.Method("list"), 2)
	s.Require().NoError(err)
	s.Contains(page, "List &lt;them&gt;.")
	s.Contains(page, "All of &#123;them&#125;.")
	s.Contains(page, "--key k --force true --count 2")
	s.Contains(page, "mac-01", "the fleet rendering shows the skipped host")
	s.Contains(page, "| `--key` | The key | required |")

	page, err = behaviorcmd.MethodPage(b, b.Method("create"), 3)
	s.Require().NoError(err)
	s.Contains(page, `"changed":false`)
	s.Contains(page, "- writes `/var/kv/k`: `v`")
	s.Contains(page, "- removes `/var/kv/old`")
	s.Contains(page, "- runs `kvctl sync`")
	s.Contains(page, "| A key that is taken | `--key k` | already exists |")
	s.Contains(page, "| No key |  | key must not be empty |")
	s.Contains(page, `| A key on two lines | `+"`"+`--key "a\nb"`+"`"+` | line breaks |`,
		"a newline is quoted, not printed into the table row")

	// A write whose output has no changed field reports one, and an
	// example with no input runs the command bare.
	bare := behavior.Write("apply", writeEntry).
		Describe("Apply").
		Example("Apply all", nil, entry{Key: "k"})
	page, err = behaviorcmd.MethodPage(b, bare, 4)
	s.Require().NoError(err)
	s.Contains(page, `"changed":true`)

	bad := []struct {
		name string
		m    *behavior.Method
	}{
		{
			name: "when an example output cannot be encoded",
			m:    behavior.Read("get", readOne).Example("x", input{}, struct{ C chan int }{}),
		},
		{
			name: "when a list example's elements are not objects",
			m:    behavior.Read("list", read).Example("x", input{}, []string{"a"}),
		},
		{
			name: "when an example output is not an object",
			m:    behavior.Read("get", readString).Example("x", input{}, "a"),
		},
	}
	for _, tc := range bad {
		s.Run(tc.name, func() {
			_, err := behaviorcmd.MethodPage(b, tc.m, 1)
			s.Error(err)
		})
	}
}

func (s *BehaviorcmdPublicTestSuite) TestRun() {
	b := behavior.Define("kv", "Key values", behavior.Node, behavior.Read("list", read))
	body := []byte(`{"job_id":"job-1","results":[{"hostname":"web-01","status":"ok","key":"k"}]}`)
	ok := client.NewResponse(json.RawMessage(body), body)

	tests := []struct {
		name string
		resp *client.Response[json.RawMessage]
		err  error
		json bool
		want string
	}{
		{name: "when the call succeeds the table prints", resp: ok, want: "web-01"},
		{
			name: "when --json the body prints",
			resp: ok,
			json: true,
			want: `"job_id":"job-1"`,
		},
		{
			name: "when the call fails the error is reported",
			err:  errors.New("boom"),
			want: "boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			behaviorcmd.ResetRegistered()
			caller := mocks.NewMockCaller(gomock.NewController(s.T()))
			caller.EXPECT().Request(gomock.Any(), "GET", gomock.Any(), gomock.Nil()).
				Return(tc.resp, tc.err)
			var reported string
			parent := &cobra.Command{Use: "node"}
			parent.PersistentFlags().String("target", "_all", "")
			cmd := &cobra.Command{Use: "list"}
			behaviorcmd.Register("store", &cobra.Command{Use: "kv"})
			behaviorcmd.Register("store", cmd)
			behaviorcmd.Attach(parent, behaviorcmd.Deps{
				Caller: func() behaviorcmd.Caller { return caller },
				JSON:   func() bool { return tc.json },
				Error:  func(err error) { reported = err.Error() },
			})

			out := cli.Capture(func() { behaviorcmd.Run(cmd, b, "list", input{}) })
			s.True(
				strings.Contains(out, tc.want) || reported == tc.want,
				"got %q / %q",
				out,
				reported,
			)
			s.Len(parent.Commands(), 1, "both commands sit under one store group")
		})
	}
}

func TestBehaviorcmdPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorcmdPublicTestSuite))
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
	behaviorcmd.SettingFlags(fs, v)

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
