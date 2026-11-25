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

package catalog_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	"github.com/osapi-io/osapi/internal/controller/endpoints/catalog"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// mode is a mode of a test, a closed set for the enum rule.
type mode string

func (mode) Values() []mode { return []mode{"a", "b"} }

func (v mode) Valid() bool { return v == "a" || v == "b" }

type provider struct{}

type itemIn struct {
	Name   string            `json:"name"   validate:"required"       doc:"The item"`
	Mode   mode              `json:"mode"   validate:"omitempty,enum" doc:"The mode"`
	Count  int               `json:"count"  validate:"omitempty"      doc:"How many"`
	On     bool              `json:"on"     validate:"omitempty"      doc:"Whether"`
	Tags   []string          `json:"tags"   validate:"omitempty"      doc:"Tags"`
	Labels map[string]string `json:"labels" validate:"omitempty"      doc:"Labels"`
	Token  secret.Value      `json:"token"  validate:"omitempty"      doc:"A secret"`
	Skip   string            `json:"-"`
	_      string
}

type row struct {
	Name    string   `json:"name"    doc:"The item"`
	Changed bool     `json:"changed" doc:"Whether the host changed"`
	Size    int      `json:"size"    doc:"Its size"                 render:"bytes" label:"Bytes"`
	Usage   usage    `json:"usage"   doc:"Its usage"`
	Disks   []*usage `json:"disks"   doc:"Its disks"`
	Tags    []string `json:"tags"    doc:"Its tags"`
	Ptr     *usage   `json:"ptr"     doc:"A pointer"`
	Hidden  string   `json:"hidden"  doc:"Not shown"                render:"-"`
	Skip    string   `json:"-"`
	_       string
}

type usage struct {
	Total int `json:"total" doc:"Total bytes" render:"bytes"`
}

func read(context.Context, provider, itemIn) ([]row, error) { return nil, nil }

func write(context.Context, provider, itemIn) (row, error) { return row{}, nil }

func none(context.Context, provider, *int) (string, error) { return "", nil }

func build(any) (provider, error) { return provider{}, nil }

// noRuns is a registry whose examples call nothing.
func noRuns(*behavior.Behavior) []behaviortest.ChainRun { return nil }

type BehaviorsPublicTestSuite struct {
	suite.Suite
}

// TestBehaviorsLeaveOutWhatTheStepSets: the form of a method has no field
// that the method's .Before step sets.
func (s *BehaviorsPublicTestSuite) TestBehaviorsLeaveOutWhatTheStepSets() {
	type pwIn struct {
		Password secret.Value `json:"password" doc:"The password"`
		Hash     string       `json:"hash"     doc:"The hash"     before:"set"`
	}
	b := behavior.Define("acct", "Accounts", behavior.Node, behavior.API, behavior.Area("config"),
		behavior.Write("update", func(context.Context, struct{}, pwIn) (struct{}, error) {
			return struct{}{}, nil
		}),
	)
	got := catalog.Behaviors([]*behavior.Behavior{b}, noRuns)
	s.Require().Len(got, 1)
	s.Require().Len(got[0].Methods[0].Fields, 1)
	s.Equal("password", got[0].Methods[0].Fields[0].Name)
}

func (s *BehaviorsPublicTestSuite) TestBehaviors() {
	inner := behavior.Define(
		"inner",
		"Inner items",
		behavior.Node,
		behavior.Internal,
		behavior.Read("list", read),
	)
	b := behavior.Define(
		"item",
		"Items",
		behavior.Node,
		behavior.API,
		behavior.Area("hardware"),
		behavior.Requires(inner),
		behavior.Platform[any, provider](behavior.PlatformDebian, build),
		behavior.Platform[any, provider](behavior.PlatformDarwin, build),
		behavior.Read("list", read).
			Example("Lists the items", itemIn{Name: "a", Token: "s3cret"}, []row{{Name: "a"}}),
		behavior.Write("update", write).Key("name").On(behavior.PlatformDebian),
		behavior.Read("get", none),
		behavior.Action("restart", write).Disrupts("It stops the item"),
		behavior.Write("signal", write).PermitAs(types.VerbExecute),
	)
	child := behavior.Define("part", "Parts", behavior.Node, behavior.API, behavior.Of(b),
		behavior.Read("list", read),
	)
	cliOnly := behavior.Define("tool", "A tool", behavior.Node, behavior.CLI,
		behavior.Read("list", read),
	)
	runs := func(bh *behavior.Behavior) []behaviortest.ChainRun {
		if bh != b {
			return nil
		}
		step := func(name, method string, steps ...*behavior.Step) *behavior.Step {
			return &behavior.Step{Behavior: name, Method: method, Steps: steps}
		}

		return []behaviortest.ChainRun{
			{Method: "update", Root: step("item", "update",
				step("inner", "list"), step("inner", "list", step("deep", "get")))},
			{Method: "update", Root: step("item", "update",
				step("inner", "list", step("deep", "set")), step("other", "get"))},
		}
	}

	got := catalog.Behaviors([]*behavior.Behavior{b, inner, child, cliOnly}, runs)
	s.Require().Len(got, 3, "a behavior that only the CLI has is not in the catalog")

	item := got[0]
	s.Equal("item", item.Name)
	s.Equal("Item", item.Title)
	s.Equal("hardware", item.Area)
	s.Equal("Items", item.Doc)
	s.False(item.Internal)
	s.Empty(item.Parent)
	s.Equal([]string{"debian", "darwin"}, item.Platforms)
	s.Equal([]string{"inner"}, item.Uses)
	s.Equal([]string{}, item.UsedBy)
	s.Equal("item → inner", item.Chain)
	s.Require().Len(item.Methods, 5)

	list := item.Methods[0]
	s.Equal("list", list.Name)
	s.Equal(types.KindRead, list.Kind)
	s.Equal("GET", list.HTTPMethod)
	s.Equal("/api/node/{hostname}/item", list.Path)
	s.Equal(types.Permission("item:read"), list.Permission)
	s.Equal([]string{"debian", "darwin"}, list.Platforms)
	s.Equal("itemIn { name*, mode, count, on, tags, labels, token }", list.Input)
	s.Equal("[]row { name, changed, size, usage, disks, tags, ptr, hidden }", list.Output)
	s.Equal([]catalog.CatalogExample{
		{Doc: "Lists the items", Input: map[string]any{
			"name": "a", "mode": "", "count": float64(0), "on": false, "tags": nil,
			"labels": nil, "token": "[redacted]",
		}, Output: []row{{Name: "a"}}},
	}, list.Examples, "an example's secret is redacted")
	s.Equal(
		[]catalog.CatalogCall{},
		list.Calls,
		"a method whose examples call nothing has no calls",
	)
	s.Equal([]catalog.CatalogField{
		{Name: "name", In: "query", Kind: "string", Required: true, Doc: "The item"},
		{Name: "mode", In: "query", Kind: "string", Doc: "The mode", Options: []string{"a", "b"}},
		{Name: "count", In: "query", Kind: "number", Doc: "How many"},
		{Name: "on", In: "query", Kind: "bool", Doc: "Whether"},
		{Name: "tags", In: "query", Kind: "list", Doc: "Tags"},
		{Name: "labels", In: "query", Kind: "object", Doc: "Labels"},
		{Name: "token", In: "query", Kind: "string", Doc: "A secret", Secret: true},
	}, list.Fields)
	total := []catalog.CatalogColumn{
		{Name: "total", Label: "Total", Doc: "Total bytes", Render: "bytes", Kind: "value"},
	}
	s.Equal([]catalog.CatalogColumn{
		{Name: "name", Label: "Name", Doc: "The item", Kind: "value"},
		{Name: "changed", Label: "Changed", Doc: "Whether the host changed", Kind: "value"},
		{Name: "size", Label: "Bytes", Doc: "Its size", Render: "bytes", Kind: "value"},
		{Name: "usage", Label: "Usage", Doc: "Its usage", Kind: "object", Columns: total},
		{Name: "disks", Label: "Disks", Doc: "Its disks", Kind: "list", Columns: total},
		{Name: "tags", Label: "Tags", Doc: "Its tags", Kind: "values"},
		{Name: "ptr", Label: "Ptr", Doc: "A pointer", Kind: "object", Columns: total},
	}, list.Columns, "the CLI's columns: labels, render tags, nested objects and lists; render:\"-\" is left out")

	update := item.Methods[1]
	s.Equal(types.KindWrite, update.Kind)
	s.Equal("PUT", update.HTTPMethod)
	s.Equal([]string{"debian"}, update.Platforms, "On limits the platforms")
	s.Equal(types.FieldInPath, update.Fields[0].In, "the key goes in the path")
	s.Equal(types.FieldInBody, update.Fields[1].In, "a write's other fields go in the body")
	s.Equal([]catalog.CatalogCall{
		{Behavior: "inner", Method: "list", Calls: []catalog.CatalogCall{
			{Behavior: "deep", Method: "get"},
			{Behavior: "deep", Method: "set"},
		}},
		{Behavior: "other", Method: "get"},
	}, update.Calls, "the calls of each example, merged, each once, in the order that they first ran")

	get := item.Methods[2]
	s.Empty(get.Fields, "an input that is not a struct has no fields")
	s.Empty(get.Columns, "an output that is not a struct has no columns")

	restart := item.Methods[3]
	s.Equal(types.KindAction, restart.Kind)
	s.True(restart.Disruptive)
	s.Equal("It stops the item", restart.Disruption)
	s.Equal(types.Permission("item:execute"), restart.Permission)

	s.Equal(types.KindAction, item.Methods[4].Kind, "the verb of the permission decides the kind")

	in := got[1]
	s.True(in.Internal)
	s.Equal([]string{}, in.UsedBy, "used by reads the registry, which these are not in")
	s.Empty(in.Methods[0].Path, "an internal behavior has no route")
	s.Empty(in.Methods[0].Permission, "an internal behavior has no permission")
	s.Equal([]string{}, in.Platforms)

	s.Equal("item", got[2].Parent)
}

func TestBehaviorsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorsPublicTestSuite))
}
