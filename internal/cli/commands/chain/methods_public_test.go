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

package chain_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/chain"
	"github.com/osapi-io/osapi/internal/guide"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	files  *behavior.Behavior
	disk   *behavior.Behavior
	netcfg *behavior.Behavior
	dns    *behavior.Behavior
	status *behavior.Behavior
	alone  *behavior.Behavior
}

func (s *MethodsPublicTestSuite) SetupTest() {
	behavior.Reset()
	s.files = behavior.Register(behavior.Define("files", "x", behavior.Node, behavior.Internal))
	s.netcfg = behavior.Register(behavior.Define("netcfg", "x", behavior.Node, behavior.Internal,
		behavior.Requires(s.files)))
	s.dns = behavior.Register(
		behavior.Define("dns", "x", behavior.Node, behavior.Requires(s.netcfg)),
	)
	s.disk = behavior.Register(behavior.Define("disk", "x", behavior.Node))
	s.status = behavior.Register(behavior.Define("status", "x", behavior.Node,
		behavior.Requires(s.disk, s.netcfg)))
	s.alone = behavior.Register(behavior.Define("alone", "x", behavior.Node))
}

func (s *MethodsPublicTestSuite) TearDownTest() {
	behavior.Reset()
}

func (s *MethodsPublicTestSuite) deps() chain.Deps {
	return chain.Deps{
		Behaviors: behavior.All,
		Runs: func(b *behavior.Behavior) []behaviortest.ChainRun {
			switch b {
			case s.dns:
				return []behaviortest.ChainRun{
					{
						Method:  "get",
						Example: "No call",
						Root:    &behavior.Step{Behavior: "dns", Method: "get"},
					},
					{Method: "update", Example: "A server", Root: &behavior.Step{
						Behavior: "dns", Method: "update",
						Input: map[string]any{
							"server": "9.9.9.9",
							"empty":  "",
							"none":   nil,
							"list":   []string{},
							"tags": []string{
								"a",
								"b",
							},
							"text": "two words",
							"obj":  map[string]any{},
						},
						Output: map[string]bool{"changed": true},
						Steps: []*behavior.Step{
							{
								Behavior: "netcfg",
								Method:   "update",
								Input: map[string]any{
									"content": "a\nb\nc\n",
									"meta":    map[string]string{"k": "v"},
								},
								Steps: []*behavior.Step{{
									Behavior: "files", Method: "update", Error: "disk full",
									Input: map[string]string{"path": strings.Repeat("x", 100)},
								}},
							},
						},
					}},
				}
			case s.status:
				return []behaviortest.ChainRun{
					{Method: "get", Example: "A list", Root: &behavior.Step{
						Behavior: "status", Method: "get", Output: []string{"a"},
						Steps: []*behavior.Step{{Behavior: "disk", Method: "get"}},
					}},
				}
			default:
				return nil
			}
		},
		Problems:    func() []behavior.Problem { return nil },
		RunProblems: func(*behavior.Behavior) []behavior.Problem { return nil },
		Spec: func() guide.Spec {
			return guide.Spec{
				Kinds: []guide.Kind{{Kind: "chain", Title: "What a chain is"}},
				Terms: []guide.Term{{Kind: "chain", Name: "A chain", Does: "x", When: "y"}},
			}
		},
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	got, err := chain.List(context.Background(), s.deps(), chain.ListInput{})

	s.Require().NoError(err)
	s.ElementsMatch([]chain.Summary{
		{
			Behavior: "files",
			Internal: true,
			Chain:    "files",
			Calls:    []string{},
			UsedBy:   []string{"netcfg"},
		},
		{
			Behavior: "netcfg", Internal: true, Chain: "netcfg → files",
			Calls: []string{"files"}, UsedBy: []string{"dns", "status"},
		},
		{
			Behavior: "dns",
			Chain:    "dns → netcfg → files",
			Calls:    []string{"netcfg"},
			UsedBy:   []string{},
		},
		{Behavior: "disk", Chain: "disk", Calls: []string{}, UsedBy: []string{"status"}},
		{
			Behavior: "status", Chain: "status → (disk, netcfg → files)",
			Calls: []string{"disk", "netcfg"}, UsedBy: []string{},
		},
	}, got.Items, "a behavior in no chain is left out")
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		in           chain.GetInput
		validateFunc func(chain.GetOutput, error)
	}{
		{
			name: "when the behavior calls others, each call is a step under its caller",
			in:   chain.GetInput{Name: "dns"},
			validateFunc: func(got chain.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("dns → netcfg → files", got.Chain)
				s.Len(got.Runs, 2)
				s.Require().Len(got.Examples, 1, "an example with no call has no section")
				s.Equal("A server (update)", got.Examples[0].Example)
				calls := got.Examples[0].Calls
				s.Require().Len(calls, 3)
				s.Equal(chain.Call{
					Call: "dns update", Sends: "server=9.9.9.9 tags=[a, b] text=\"two words\"",
					Gets: "changed=true",
				}, calls[0], "a field with no value is left out")
				s.Equal("└─ netcfg update", calls[1].Call)
				s.Equal("content=<3 lines> meta={\"k\":\"v\"}", calls[1].Sends)
				s.Equal("· └─ files update", calls[2].Call)
				s.Equal("error: disk full", calls[2].Gets)
				s.Len([]rune(calls[2].Sends), 70, "a long value is cut for the table")
				s.True(strings.HasSuffix(calls[2].Sends, "…"))
			},
		},
		{
			name: "when an output is not an object, it is shown as it is",
			in:   chain.GetInput{Name: "status"},
			validateFunc: func(got chain.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal(`["a"]`, got.Examples[0].Calls[0].Gets)
			},
		},
		{
			name: "when the behavior calls none, there is no step",
			in:   chain.GetInput{Name: "alone"},
			validateFunc: func(got chain.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("alone", got.Chain)
				s.Empty(got.Examples)
			},
		},
		{
			name: "when there is no such behavior, it says where to look",
			in:   chain.GetInput{Name: "nosuch"},
			validateFunc: func(_ chain.GetOutput, err error) {
				s.ErrorContains(err, "osapi behavior chain list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(chain.Get(context.Background(), s.deps(), tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestValidate() {
	s.Run("when no rule breaks, it is valid", func() {
		got, err := chain.Validate(context.Background(), s.deps(), chain.ValidateInput{})
		s.Require().NoError(err)
		s.True(got.Valid)
		s.Equal(6, got.Behaviors)
	})

	s.Run("when a rule breaks, each problem and its fix, and it fails", func() {
		d := s.deps()
		d.Problems = func() []behavior.Problem {
			return []behavior.Problem{
				{Behavior: "files", Rule: behavior.RuleInternal, Fix: "require it"},
			}
		}
		d.RunProblems = func(b *behavior.Behavior) []behavior.Problem {
			if b != s.dns {
				return nil
			}
			return []behavior.Problem{{Behavior: "dns", Rule: behavior.RuleUsed, Fix: "call it"}}
		}
		got, err := chain.Validate(context.Background(), d, chain.ValidateInput{})
		s.Equal(command.Fail(), err)
		s.False(got.Valid)
		s.Len(got.Problems, 2)
		s.Equal(
			map[string]any{behavior.RuleInternal: "require it", behavior.RuleUsed: "call it"},
			got.Fixes,
		)
	})
}

func (s *MethodsPublicTestSuite) TestSpec() {
	got, err := chain.Spec(context.Background(), s.deps(), chain.SpecInput{})
	s.Require().NoError(err)
	s.Require().Len(got.Kinds, 1)
	s.Equal("What a chain is", got.Kinds[0].Title)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
