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

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
)

type fakeProvider struct{ prefix string }

// secretIn has a field that carries a secret.
type secretIn struct {
	Name       string `json:"name"       doc:"test"`
	Passphrase string `json:"passphrase" doc:"test" secret:"true"`
}

func readSecret(
	_ context.Context,
	_ fakeProvider,
	_ secretIn,
) (entry, error) {
	return entry{}, nil
}

// leakyIn records a secret on the span, which validate refuses.
type leakyIn struct {
	Token string `json:"token" secret:"true" telemetry:"attr" validate:"omitempty" doc:"test"`
}

func readLeaky(
	_ context.Context,
	_ fakeProvider,
	_ leakyIn,
) (entry, error) {
	return entry{}, nil
}

// readBare takes a bare string, which has no fields to make flags of.
func readBare(
	_ context.Context,
	_ fakeProvider,
	_ string,
) (entry, error) {
	return entry{}, nil
}

// onDebian gives a fixture behavior the provider every rule needs.
var onDebian = behavior.Platform("debian", func(host.Host) (fakeProvider, error) {
	return fakeProvider{}, nil
})

type getIn struct {
	Key string `json:"key" validate:"required" doc:"The key"`
}

type entry struct {
	Key   string `json:"key"   doc:"The key"`
	Value string `json:"value" doc:"The value"`
}

type result struct {
	Key     string `json:"key"     doc:"test"`
	Changed bool   `json:"changed" doc:"test"`
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
		behavior.Action("restart", set).Key("key"),
		behavior.Action("reboot", set),
	)
}

func grouped() *behavior.Behavior {
	return behavior.Define("dns", "Resolver settings",
		behavior.Node,
		behavior.Under("network"),
		behavior.Read("get", get),
	)
}

// partIn is the input of a child: the parent's key, and the child's own.
type partIn struct {
	Key  string `json:"key"  validate:"required" doc:"The parent item"`
	Part string `json:"part" validate:"required" doc:"The part"`
}

func setPart(
	_ context.Context,
	_ fakeProvider,
	in partIn,
) (result, error) {
	return result{Key: in.Key, Changed: true}, nil
}

// child is a behavior whose items are under an item of top, keyed by key.
func child(
	parent *behavior.Behavior,
) *behavior.Behavior {
	return behavior.Define("part", "The parts of an item",
		behavior.Node,
		behavior.Of(parent),
		behavior.Read("list", get),
		behavior.Write("delete", setPart).Key("part"),
		behavior.Action("restart", setPart).Key("part"),
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
				assert.Len(s.T(), b.Methods, 8)
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

func (s *BehaviorPublicTestSuite) TestPermissionVerb() {
	b := top()
	tests := []struct {
		name   string
		method string
		want   behavior.Verb
	}{
		{
			name:   "when the method is an action, its permission is execute",
			method: "restart",
			want:   behavior.VerbExecute,
		},
		{
			name:   "when the method is a read, its permission is read",
			method: "get",
			want:   behavior.VerbRead,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			m := b.Method(tc.method)
			s.Require().NotNil(m)
			s.Equal(tc.want, m.PermissionVerb())
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
		{
			name: "when under a parent the parent's permission is the child's",
			b:    child(grouped()),
			verb: behavior.VerbWrite,
			want: "network:write",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, string(tc.b.Permission(tc.verb)))
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
		{
			name:   "when an action on one item",
			b:      top(),
			method: "restart",
			want:   "/api/node/{hostname}/sysctl/{key}/restart",
		},
		{
			name:   "when an action on the host",
			b:      top(),
			method: "reboot",
			want:   "/api/node/{hostname}/sysctl/reboot",
		},
		{
			name:   "when under a parent, the collection is under the parent's item",
			b:      child(top()),
			method: "list",
			want:   "/api/node/{hostname}/sysctl/{key}/part",
		},
		{
			name:   "when under a parent, an item has the two keys",
			b:      child(top()),
			method: "delete",
			want:   "/api/node/{hostname}/sysctl/{key}/part/{part}",
		},
		{
			name: "when under a parent that is under a group, the group comes first",
			b: child(
				behavior.Define(
					"iface",
					"x",
					behavior.Node,
					behavior.Under("network"),
					behavior.Read("get", get).Key("key"),
				),
			),
			method: "restart",
			want:   "/api/node/{hostname}/network/iface/{key}/part/{part}/restart",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, tc.b.Path(tc.b.Method(tc.method)))
		})
	}
}

func (s *BehaviorPublicTestSuite) TestServedAt() {
	b := top()
	m := b.Method("get")
	s.Equal("/api/node/{hostname}/sysctl/{key}", b.Path(m))
	s.Equal("osapi client node sysctl get", b.Command(m))
}

// TestChildTypeName: a child's names in the spec start with its parent's.
func (s *BehaviorPublicTestSuite) TestChildTypeName() {
	s.Equal("SysctlPart", child(top()).TypeName())
	s.Equal("NetworkDnsPart", child(grouped()).TypeName())
}

// TestPathKeys: a child's methods put the parent's key in the path before
// their own; a behavior with no parent has only its own.
func (s *BehaviorPublicTestSuite) TestPathKeys() {
	c := child(top())
	s.Equal("key", c.ParentKey())
	s.Equal([]string{"key"}, c.Method("list").PathKeys())
	s.Equal([]string{"key", "part"}, c.Method("delete").PathKeys())
	s.True(c.Method("delete").InPath("part"))
	s.False(c.Method("list").InPath("part"))
	s.Equal([]string{"key"}, top().Method("get").PathKeys())
	s.Empty(top().Method("list").PathKeys())
	s.Empty(top().ParentKey(), "a behavior of the host has no parent key")
	s.Empty(child(grouped()).ParentKey(), "a parent with no keyed method gives no key")
	s.Empty(
		behavior.Read("get", get).Key("key").PathKeys()[1:],
		"a method of no behavior has its own key only",
	)
}

func (s *BehaviorPublicTestSuite) TestExplain() {
	m := behavior.Read("get", get).Explain("Reads the value.")
	s.Equal("Reads the value.", m.Long)
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
	s.Equal("osapi client node sysctl part list", child(top()).Command(child(top()).Method("list")),
		"a child's command is under its parent's")

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
		{name: "when an action", method: "restart", want: "POST"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assert.Equal(s.T(), tc.want, top().Method(tc.method).HTTPMethod())
		})
	}
}

func (s *BehaviorPublicTestSuite) TestAction() {
	m := behavior.Action("restart", set)
	s.True(m.Action)
	s.Equal(behavior.VerbWrite, m.Verb, "an action is a write: <group>:write and a modify job")
	s.False(behavior.Write("update", set).Action)
	s.True(behavior.Action("restart", set).Always().AlwaysChanges)
}

func (s *BehaviorPublicTestSuite) TestNotAFact() {
	s.Equal("not a fact", behavior.Read("list", set).NotAFact("not a fact").NonFact)
	s.Empty(behavior.Read("list", set).NonFact, "a read goes through gohai unless it says why")
}

func (s *BehaviorPublicTestSuite) TestKey() {
	m := behavior.Read("get", get).Key("key")
	assert.Equal(s.T(), "key", m.KeyField)
}

func (s *BehaviorPublicTestSuite) TestDescribe() {
	m := behavior.Read("get", get).Describe("Get one")
	assert.Equal(s.T(), "Get one", m.Doc)
}

func (s *BehaviorPublicTestSuite) TestAdminOnly() {
	assert.False(s.T(), behavior.Write("update", set).Admin)
	assert.True(s.T(), behavior.Write("update", set).AdminOnly().Admin)
}

func (s *BehaviorPublicTestSuite) TestAdminRole() {
	tests := []struct {
		name      string
		m         *behavior.Method
		want      behavior.Verb
		wantAdmin bool
	}{
		{name: "when a read", m: behavior.Read("get", get), want: behavior.VerbRead},
		{name: "when a write", m: behavior.Write("update", set), want: behavior.VerbWrite},
		{name: "when an action", m: behavior.Action("exec", set), want: behavior.VerbExecute},
		{
			name: "when an admin-only action", m: behavior.Action("exec", set).AdminOnly(),
			want: behavior.VerbExecute, wantAdmin: true,
		},
		{
			name: "when a shell action, only the admin role gets its verb",
			m:    behavior.Action("shell", set).Shell(), want: behavior.VerbShell, wantAdmin: true,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.m.PermissionVerb())
			s.Equal(tc.wantAdmin, tc.m.AdminRole())
		})
	}
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
