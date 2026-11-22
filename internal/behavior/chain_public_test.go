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

package behavior_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type ChainPublicTestSuite struct {
	suite.Suite
}

func (s *ChainPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *ChainPublicTestSuite) TestRecord() {
	b := top()

	s.Run("a call under Record adds its step, with its input and output", func() {
		root := &behavior.Step{Behavior: "other", Method: "create"}
		out, err := behavior.Call(behavior.Record(context.Background(), root), b, get,
			fakeProvider{prefix: "v-"}, getIn{Key: "k"})
		s.Require().NoError(err)
		s.Require().Len(root.Steps, 1)
		s.Equal("sysctl", root.Steps[0].Behavior)
		s.Equal("list", root.Steps[0].Method, "the first method declared with get")
		s.Equal(getIn{Key: "k"}, root.Steps[0].Input)
		s.Equal(out, root.Steps[0].Output)
		s.Empty(root.Steps[0].Error)
	})

	s.Run("a failed call records its error, and no output", func() {
		root := &behavior.Step{}
		_, err := behavior.Call(behavior.Record(context.Background(), root), b, get,
			fakeProvider{}, getIn{Key: "boom"})
		s.Require().Error(err)
		s.Equal("boom", root.Steps[0].Error)
		s.Nil(root.Steps[0].Output)
	})

	s.Run("Wire gives the calls as a job's response carries them", func() {
		root := &behavior.Step{}
		ctx := behavior.Record(context.Background(), root)
		_, err := behavior.Call(ctx, b, get, fakeProvider{prefix: "v-"}, getIn{Key: "k"})
		s.Require().NoError(err)
		_, err = behavior.Call(ctx, b, get, fakeProvider{}, getIn{Key: "boom"})
		s.Require().Error(err)
		root.Steps[0].Steps = []*behavior.Step{
			{Behavior: "inner", Method: "update", Output: result{Changed: true}},
		}
		root.Steps[1].Output = "not an object"
		root.Steps = append(
			root.Steps,
			&behavior.Step{Behavior: "x", Method: "y", Output: make(chan int)},
		)

		wire := root.Wire()
		s.Require().Len(wire, 3)
		s.Equal("sysctl", wire[0].Behavior)
		s.Equal("list", wire[0].Method)
		s.Nil(wire[0].Changed, "an output with no changed field")
		s.GreaterOrEqual(wire[0].DurationMS, int64(0))
		s.Require().Len(wire[0].Steps, 1)
		s.Require().NotNil(wire[0].Steps[0].Changed)
		s.True(*wire[0].Steps[0].Changed)
		s.Equal("boom", wire[1].Error)
		s.Nil(wire[1].Changed, "an output that is not an object")
		s.Nil(wire[2].Changed, "an output that is not JSON")
		s.Nil((&behavior.Step{}).Wire(), "a method that called none")
	})

	s.Run("a call without Record records nothing", func() {
		_, err := behavior.Call(context.Background(), b, get, fakeProvider{}, getIn{Key: "k"})
		s.NoError(err)
	})

	s.Run("calls at the same time each add their step", func() {
		root := &behavior.Step{}
		ctx := behavior.Record(context.Background(), root)
		var wg sync.WaitGroup
		for range 20 {
			wg.Go(func() { _, _ = behavior.Call(ctx, b, get, fakeProvider{}, getIn{Key: "k"}) })
		}
		wg.Wait()
		s.Len(root.Steps, 20)
	})
}

type loginIn struct {
	User     string `json:"user"`
	Password string `json:"password,omitempty" secret:"true"`
}

func (s *ChainPublicTestSuite) TestRedact() {
	s.Equal(map[string]any{"user": "ops", "password": "[redacted]"},
		behavior.Redact(loginIn{User: "ops", Password: "hunter2"}))
	s.Equal(map[string]any{"user": "ops"}, behavior.Redact(loginIn{User: "ops"}),
		"an empty secret that the JSON leaves out stays out")
	s.Equal(
		getIn{Key: "k"},
		behavior.Redact(getIn{Key: "k"}),
		"a struct with no secret is unchanged",
	)
	s.Equal("text", behavior.Redact("text"))
	s.Nil(behavior.Redact(nil))
}

func (s *ChainPublicTestSuite) TestChainProblems() {
	missing := behavior.Define("missing", "x", behavior.Node)
	behavior.Register(behavior.Define("orphan", "x", behavior.Node, behavior.Internal))
	behavior.Register(behavior.Define("user", "x", behavior.Node, behavior.Requires(missing),
		behavior.Writes("relative")))

	rules := map[string]bool{}
	for _, p := range behavior.ChainProblems() {
		rules[p.Rule] = true
	}
	s.Equal(map[string]bool{behavior.RuleRequires: true, behavior.RuleInternal: true}, rules,
		"only the chain rules: the writes rule is behavior validate's")
	for _, p := range behavior.ChainProblems() {
		s.Equal(behavior.ChainChecks.Fix(p.Rule), p.Fix)
	}
}

func (s *ChainPublicTestSuite) TestTree() {
	files := behavior.Define("files", "x")
	netcfg := behavior.Define("netcfg", "x", behavior.Requires(files))
	disk := behavior.Define("disk", "x")
	s.Equal("files", behavior.Tree(files))
	s.Equal(
		"dns → netcfg → files",
		behavior.Tree(behavior.Define("dns", "x", behavior.Requires(netcfg))),
	)
	s.Equal("status → (disk, netcfg → files)",
		behavior.Tree(behavior.Define("status", "x", behavior.Requires(disk, netcfg))))
}

func (s *ChainPublicTestSuite) TestRequires() {
	files := behavior.Define("files", "x")
	state := behavior.Define("state", "x")

	b := behavior.Define("sysctl", "x", behavior.Requires(files), behavior.Requires(state))

	s.Equal([]*behavior.Behavior{files, state}, b.Requires)
}

func (s *ChainPublicTestSuite) TestChildSpans() {
	tests := []struct {
		name string
		b    func() *behavior.Behavior
		want []string
	}{
		{
			name: "when it declares nothing, there are none",
			b:    func() *behavior.Behavior { return behavior.Define("motd", "x") },
		},
		{
			name: "when it requires a behavior, each declaration of the two counts once",
			b: func() *behavior.Behavior {
				files := behavior.Define("files", "x", behavior.Internal,
					behavior.Sudo("chown"), behavior.Runs("stat"),
					behavior.Writes("*"), behavior.KeepsState)

				return behavior.Define("sysctl", "x",
					behavior.Requires(files, files),
					behavior.Sudo("sysctl"), behavior.Runs("stat"),
					behavior.Writes("/etc/sysctl.d/osapi-*.conf"))
			},
			want: []string{
				"behavior.files.<method>",
				"exec sysctl", "exec stat", "exec chown",
				"file.write", "file.remove", "file.chmod",
				"kv.get", "kv.put",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, behavior.ChildSpans(tc.b()))
		})
	}
}

func (s *ChainPublicTestSuite) TestWrites() {
	b := behavior.Define("conf", "x", behavior.Writes("/etc/a"), behavior.Writes("/etc/b"))

	s.Equal([]string{"/etc/a", "/etc/b"}, b.Writes)
	s.True(behavior.Define("files", "x", behavior.KeepsState).KeepsState)
}

func (s *ChainPublicTestSuite) TestFunc() {
	b := behavior.Define("sysctl", "x", behavior.Read("get", get))

	s.Equal("github.com/osapi-io/osapi/internal/behavior_test.get", b.Methods[0].Func())
}

func (s *ChainPublicTestSuite) TestInternal() {
	s.True(behavior.Define("files", "x", behavior.Internal).Internal)
	s.False(behavior.Define("sysctl", "x").Internal)
}

func (s *ChainPublicTestSuite) TestRequiredBy() {
	tests := []struct {
		name  string
		setup func() *behavior.Behavior
		want  []string
	}{
		{
			name: "when behaviors require it, each is listed once",
			setup: func() *behavior.Behavior {
				files := behavior.Register(behavior.Define("files", "x"))
				behavior.Register(behavior.Define("sysctl", "x", behavior.Requires(files, files)))
				behavior.Register(behavior.Define("cron", "x", behavior.Requires(files)))
				behavior.Register(behavior.Define("dns", "x"))

				return files
			},
			want: []string{"cron", "sysctl"},
		},
		{
			name: "when nothing requires it",
			setup: func() *behavior.Behavior {
				return behavior.Register(behavior.Define("files", "x"))
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			behavior.Reset()
			b := tc.setup()

			var got []string
			for _, r := range behavior.RequiredBy(b) {
				got = append(got, r.Name)
			}
			s.Equal(tc.want, got)
		})
	}
}

func TestChainPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ChainPublicTestSuite))
}
