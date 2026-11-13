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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/behavior"
)

type fakeProvider struct{ prefix string }

type getIn struct {
	Key string `json:"key" validate:"required" doc:"The key"`
}

type entry struct {
	Key   string `json:"key"   doc:"The key"`
	Value string `json:"value" doc:"The value"`
}

type result struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
}

func get(
	_ context.Context,
	p fakeProvider,
	in getIn,
) (entry, error) {
	if in.Key == "boom" {
		return entry{}, errors.New("boom")
	}

	return entry{Key: in.Key, Value: p.prefix + in.Key}, nil
}

func set(
	_ context.Context,
	_ fakeProvider,
	in getIn,
) (result, error) {
	return result{Key: in.Key, Changed: true}, nil
}

// top is a behavior at the top level; grouped is one under a group.
func top() *behavior.Behavior {
	return behavior.Define("sysctl", "Kernel parameters",
		behavior.Node,
		behavior.Read("list", get),
		behavior.Read("get", get).Key("key"),
		behavior.Write("create", set),
		behavior.Write("update", set).Key("key"),
		behavior.Write("delete", set).Key("key"),
		behavior.Write("apply", set),
	)
}

func grouped() *behavior.Behavior {
	return behavior.Define("dns", "Resolver settings",
		behavior.Node,
		behavior.Under("network"),
		behavior.Read("get", get),
	)
}

type BehaviorPublicTestSuite struct {
	suite.Suite
}

func (s *BehaviorPublicTestSuite) TestDefine() {
	tests := []struct {
		name         string
		b            *behavior.Behavior
		validateFunc func(*behavior.Behavior)
	}{
		{
			name: "when placed on a node with methods",
			b:    top(),
			validateFunc: func(b *behavior.Behavior) {
				assert.Equal(s.T(), "sysctl", b.Name)
				assert.Equal(s.T(), "Kernel parameters", b.Doc)
				assert.Equal(s.T(), behavior.NodePlacement, b.Placement)
				assert.Empty(s.T(), b.Group)
				assert.Len(s.T(), b.Methods, 6)
			},
		},
		{
			name: "when nested under a group",
			b:    grouped(),
			validateFunc: func(b *behavior.Behavior) {
				assert.Equal(s.T(), "network", b.Group)
			},
		},
		{
			name: "when given no options",
			b:    behavior.Define("bare", ""),
			validateFunc: func(b *behavior.Behavior) {
				assert.Empty(s.T(), b.Placement)
				assert.Empty(s.T(), b.Methods)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(tc.b)
		})
	}
}

func (s *BehaviorPublicTestSuite) TestPermission() {
	tests := []struct {
		name string
		b    *behavior.Behavior
		verb behavior.Verb
		want string
	}{
		{
			name: "when a read at the top level",
			b:    top(),
			verb: behavior.VerbRead,
			want: "sysctl:read",
		},
		{
			name: "when a write at the top level",
			b:    top(),
			verb: behavior.VerbWrite,
			want: "sysctl:write",
		},
		{
			name: "when under a group the group is the prefix",
			b:    grouped(),
			verb: behavior.VerbRead,
			want: "network:read",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, tc.b.Permission(tc.verb))
		})
	}
}

func (s *BehaviorPublicTestSuite) TestMethod() {
	tests := []struct {
		name   string
		method string
		found  bool
	}{
		{name: "when the method exists", method: "get", found: true},
		{name: "when the method does not exist", method: "nope", found: false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := top().Method(tc.method)
			assert.Equal(s.T(), tc.found, m != nil)
		})
	}
}

func (s *BehaviorPublicTestSuite) TestOperation() {
	b := top()
	assert.Equal(s.T(), "sysctl.get", b.Operation(b.Method("get")))
}

func (s *BehaviorPublicTestSuite) TestPath() {
	tests := []struct {
		name   string
		b      *behavior.Behavior
		method string
		want   string
	}{
		{
			name:   "when addressing the collection",
			b:      top(),
			method: "list",
			want:   "/api/node/{hostname}/sysctl",
		},
		{
			name:   "when addressing one item by key",
			b:      top(),
			method: "get",
			want:   "/api/node/{hostname}/sysctl/{key}",
		},
		{
			name:   "when under a group",
			b:      grouped(),
			method: "get",
			want:   "/api/node/{hostname}/network/dns",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, tc.b.Path(tc.b.Method(tc.method)))
		})
	}
}

func (s *BehaviorPublicTestSuite) TestCommand() {
	tests := []struct {
		name string
		b    *behavior.Behavior
		want string
	}{
		{name: "when at the top level", b: top(), want: "osapi client node sysctl get"},
		{name: "when under a group", b: grouped(), want: "osapi client node network dns get"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, tc.b.Command(tc.b.Method("get")))
		})
	}
}

func (s *BehaviorPublicTestSuite) TestHTTPMethod() {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{name: "when a read", method: "get", want: "GET"},
		{name: "when a create", method: "create", want: "POST"},
		{name: "when an update", method: "update", want: "PUT"},
		{name: "when a delete", method: "delete", want: "DELETE"},
		{name: "when any other write", method: "apply", want: "POST"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, top().Method(tc.method).HTTPMethod())
		})
	}
}

func (s *BehaviorPublicTestSuite) TestKey() {
	m := behavior.Read("get", get).Key("key")
	assert.Equal(s.T(), "key", m.KeyField)
}

func (s *BehaviorPublicTestSuite) TestDescribe() {
	m := behavior.Read("get", get).Describe("Get one")
	assert.Equal(s.T(), "Get one", m.Doc)
}

func (s *BehaviorPublicTestSuite) TestExample() {
	m := behavior.Read("get", get).
		Example("One", getIn{Key: "a"}, entry{Key: "a"}).
		Example("Two", getIn{Key: "b"}, entry{Key: "b"})

	s.Require().Len(m.Examples, 2)
	assert.Equal(s.T(), behavior.Example{
		Doc: "One", Input: getIn{Key: "a"}, Output: entry{Key: "a"},
	}, m.Examples[0])
}

func (s *BehaviorPublicTestSuite) TestProviderType() {
	assert.Equal(s.T(), "fakeProvider", behavior.Read("get", get).ProviderType().Name())
}

func (s *BehaviorPublicTestSuite) TestRead() {
	m := behavior.Read("get", get)

	assert.Equal(s.T(), behavior.VerbRead, m.Verb)
	assert.Equal(s.T(), "getIn", m.In.Name())
	assert.Equal(s.T(), "entry", m.Out.Name())
}

func (s *BehaviorPublicTestSuite) TestWrite() {
	m := behavior.Write("create", set)

	assert.Equal(s.T(), behavior.VerbWrite, m.Verb)
	assert.Equal(s.T(), "result", m.Out.Name())
}

func (s *BehaviorPublicTestSuite) TestInvoke() {
	m := behavior.Read("get", get)

	tests := []struct {
		name     string
		provider any
		input    []byte
		want     any
		wantErr  string
	}{
		{
			name:     "when the input decodes it calls the method",
			provider: fakeProvider{prefix: "v-"},
			input:    []byte(`{"key":"a"}`),
			want:     entry{Key: "a", Value: "v-a"},
		},
		{
			name:     "when the input is empty the zero input is used",
			provider: fakeProvider{prefix: "v-"},
			want:     entry{Value: "v-"},
		},
		{
			name:     "when the provider is the wrong type",
			provider: "not a provider",
			input:    []byte(`{"key":"a"}`),
			wantErr:  "provider is string",
		},
		{
			name:     "when the input does not decode",
			provider: fakeProvider{},
			input:    []byte(`{`),
			wantErr:  "decode input",
		},
		{
			name:     "when the method fails",
			provider: fakeProvider{},
			input:    []byte(`{"key":"boom"}`),
			wantErr:  "boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := m.Invoke(context.Background(), tc.provider, tc.input)
			if tc.wantErr != "" {
				s.Require().Error(err)
				assert.Contains(s.T(), err.Error(), tc.wantErr)

				return
			}
			s.Require().NoError(err)
			assert.Equal(s.T(), tc.want, got)
		})
	}
}

func TestBehaviorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorPublicTestSuite))
}
