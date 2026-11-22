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

package behaviortest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

type echoIn struct {
	Text string `json:"text"`
}

type echoOut struct {
	Text string `json:"text"`
}

type echoP struct{}

func echoGet(
	_ context.Context,
	_ echoP,
	in echoIn,
) (echoOut, error) {
	if in.Text == "fail" {
		return echoOut{}, errors.New("refused")
	}

	return echoOut(in), nil
}

// echo is a building block that answers with its input.
var echo = behavior.Define("echo", "Answers with its input", behavior.Node, behavior.Internal,
	behavior.Platform("debian", func(host.Host) (echoP, error) { return echoP{}, nil }),
	behavior.Read("get", echoGet),
)

// callerP calls echo, or does not, as its behavior's test needs.
type callerP struct{ calls bool }

func callerGet(
	ctx context.Context,
	p callerP,
	in echoIn,
) (echoOut, error) {
	if !p.calls {
		return echoOut(in), nil
	}

	return behavior.Call(ctx, echo, echoGet, echoP{}, in)
}

func caller(
	name string,
	calls bool,
	opts ...behavior.Option,
) *behavior.Behavior {
	return behavior.Define(name, "Calls echo", append([]behavior.Option{
		behavior.Node,
		behavior.Platform(
			"debian",
			func(host.Host) (callerP, error) { return callerP{calls: calls}, nil },
		),
		behavior.Read("get", callerGet).
			Example("A text", echoIn{Text: "hi"}, echoOut{Text: "hi"}).
			Example("Not on this platform", echoIn{Text: "mac"}, echoOut{Text: "mac"},
				behavior.OnPlatform(behavior.PlatformDarwin)).
			Example("An example the fake host cannot set up", echoIn{Text: "x"}, echoOut{Text: "x"},
				behavior.After(42)),
	}, opts...)...)
}

type ChainPublicTestSuite struct {
	suite.Suite
}

func (s *ChainPublicTestSuite) SetupTest() {
	behavior.Reset()
	behavior.Register(echo)
}

func (s *ChainPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *ChainPublicTestSuite) TestChains() {
	b := behavior.Register(caller("user", true, behavior.Requires(echo)))

	runs := behaviortest.Chains(b)

	s.Require().Len(runs, 1, "only the example that runs on a platform it serves, and that sets up")
	s.Equal("get", runs[0].Method)
	s.Equal("A text", runs[0].Example)
	s.Equal(behavior.PlatformDebian, runs[0].Platform)
	root := runs[0].Root
	s.Equal("user", root.Behavior)
	s.Equal(echoIn{Text: "hi"}, root.Input)
	s.Equal(echoOut{Text: "hi"}, root.Output)
	s.Require().Len(root.Steps, 1)
	s.Equal("echo", root.Steps[0].Behavior)
	s.Equal("get", root.Steps[0].Method)
	s.Empty(behaviortest.ChainProblems(b))
}

func (s *ChainPublicTestSuite) TestChainFails() {
	b := behavior.Register(behavior.Define(
		"user",
		"Calls echo",
		behavior.Node,
		behavior.Requires(echo),
		behavior.Platform(
			"debian",
			func(host.Host) (callerP, error) { return callerP{calls: true}, nil },
		),
		behavior.Read("get", callerGet).Example("A refused text", echoIn{Text: "fail"}, echoOut{}),
	))

	runs := behaviortest.Chains(b)

	s.Require().Len(runs, 1)
	s.Equal("refused", runs[0].Root.Error)
	s.Nil(runs[0].Root.Output)
	s.Equal("refused", runs[0].Root.Steps[0].Error)
}

func (s *ChainPublicTestSuite) TestChainProblems() {
	tests := []struct {
		name string
		b    *behavior.Behavior
		want behavior.Problem
	}{
		{
			name: "when a requirement is never called, it is the used rule",
			b:    caller("unused", false, behavior.Requires(echo)),
			want: behavior.Problem{
				Behavior: "unused", Rule: behavior.RuleUsed,
				Message: "unused requires echo, and no example calls it",
				Fix:     behavior.ChainChecks.Fix(behavior.RuleUsed),
			},
		},
		{
			name: "when a call goes to a behavior that is not required, it is the called rule",
			b:    caller("rogue", true),
			want: behavior.Problem{
				Behavior: "rogue", Method: "get", Rule: behavior.RuleCalled,
				Message: `example "A text": rogue get calls echo get, which rogue does not require`,
				Fix:     behavior.ChainChecks.Fix(behavior.RuleCalled),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			behavior.Register(tc.b)
			s.Equal([]behavior.Problem{tc.want}, behaviortest.ChainProblems(tc.b))
		})
	}

	s.Run("Run fails the tests on each one, with its fix", func() {
		b := behavior.Register(caller("lazy", false, behavior.Requires(echo)))
		r := &recorder{}
		behaviortest.Run(r, b)
		s.Contains(r.errs, "lazy requires echo, and no example calls it. Fix: "+
			behavior.ChainChecks.Fix(behavior.RuleUsed))
	})
}

func TestChainPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ChainPublicTestSuite))
}
