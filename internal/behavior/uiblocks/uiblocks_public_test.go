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

package uiblocks_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/uiblocks"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
)

type provider struct{}

type itemIn struct {
	Name   string            `json:"name"   validate:"required"            doc:"The item"`
	Mode   string            `json:"mode"   validate:"omitempty,oneof=a b" doc:"The mode"`
	Count  int               `json:"count"  validate:"omitempty"           doc:"How many"`
	On     bool              `json:"on"     validate:"omitempty"           doc:"Whether"`
	Tags   []string          `json:"tags"   validate:"omitempty"           doc:"Tags"`
	Labels map[string]string `json:"labels" validate:"omitempty"           doc:"Labels"`
	Token  string            `json:"token"  validate:"omitempty"           doc:"A secret" secret:"true"`
	Skip   string            `json:"-"`
	_      string
}

type row struct {
	Name    string `json:"name"    doc:"The item"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
	Skip    string `json:"-"`
	_       string
}

func read(context.Context, provider, itemIn) ([]row, error) { return nil, nil }

func write(context.Context, provider, itemIn) (row, error) { return row{}, nil }

func none(context.Context, provider, *int) (string, error) { return "", nil }

type BlocksPublicTestSuite struct {
	suite.Suite
}

func (s *BlocksPublicTestSuite) TestBlocks() {
	b := behavior.Define("item", "Items", behavior.Node, behavior.API,
		behavior.Read("list", read),
		behavior.Write("update", write).Key("name"),
		behavior.Read("get", none),
	)
	inner := behavior.Define(
		"inner",
		"x",
		behavior.Node,
		behavior.Internal,
		behavior.Read("list", read),
	)

	got := uiblocks.Blocks([]*behavior.Behavior{b, inner})
	s.Require().Len(got, 3, "an internal behavior has no block")

	list := got[0]
	s.Equal("item-list", list.Type)
	s.Equal("GET", list.HTTPMethod)
	s.Equal("/api/node/{hostname}/item", list.Path)
	s.Equal("item:read", list.Permission)
	s.False(list.Write)
	s.Equal([]uiblocks.Field{
		{Name: "name", In: "query", Kind: "string", Required: true, Doc: "The item"},
		{Name: "mode", In: "query", Kind: "string", Doc: "The mode", Options: []string{"a", "b"}},
		{Name: "count", In: "query", Kind: "number", Doc: "How many"},
		{Name: "on", In: "query", Kind: "bool", Doc: "Whether"},
		{Name: "tags", In: "query", Kind: "list", Doc: "Tags"},
		{Name: "labels", In: "query", Kind: "object", Doc: "Labels"},
		{Name: "token", In: "query", Kind: "string", Doc: "A secret", Secret: true},
	}, list.Fields)
	s.Equal([]uiblocks.Column{
		{Name: "name", Doc: "The item"},
		{Name: "changed", Doc: "Whether the host changed"},
	}, list.Columns)

	update := got[1]
	s.Equal("PUT", update.HTTPMethod)
	s.True(update.Write)
	s.Equal("path", update.Fields[0].In, "the key goes in the path")
	s.Equal("body", update.Fields[1].In, "a write's other fields go in the body")

	s.Empty(got[2].Fields, "an input that is not a struct has no fields")
	s.Empty(got[2].Columns, "an output that is not a struct has no columns")

	ts := uiblocks.TypeScript(got)
	s.Contains(ts, `"item-list": {"type":"item-list"`)
	s.Contains(ts, `export type BehaviorName = "item";`)
}

// TestFileIsCurrent: the UI's file is what the registry makes.
func (s *BlocksPublicTestSuite) TestFileIsCurrent() {
	ts := uiblocks.TypeScript(uiblocks.Blocks(behavior.All()))
	file, err := os.ReadFile("../../../" + uiblocks.File)
	s.Require().NoError(err)
	s.Equal(ts, string(file), "run go generate ./internal/behaviors/all")
}

func TestBlocksPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BlocksPublicTestSuite))
}
