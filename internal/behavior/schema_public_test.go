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
	"reflect"
	"testing"

	"github.com/osapi-io/osapi/internal/validation/rule"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type SchemaPublicTestSuite struct {
	suite.Suite
}

type nested struct {
	Name string `json:"name" doc:"test"`
}

type everything struct {
	Plain     string            `doc:"no json tag"`
	Named     string            `doc:"test"        json:"named,omitempty"`
	Commaonly string            `doc:"test"        json:",omitempty"`
	Skipped   string            `                  json:"-"`
	unexport  string            //nolint:unused
	Flag      bool              `doc:"test"        json:"flag"`
	Count     int               `doc:"test"        json:"count"           validate:"min=1,max=10"`
	Ratio     float64           `doc:"test"        json:"ratio"           validate:"min=0.5"`
	Size      uint8             `doc:"test"        json:"size"`
	Tags      []string          `doc:"test"        json:"tags"            validate:"min=1,max=3"`
	Exact     string            `doc:"test"        json:"exact"           validate:"len=4"`
	Mode      string            `doc:"test"        json:"mode"            validate:"required,oneof=on off"`
	Addr      string            `doc:"test"        json:"addr"            validate:"omitempty,ip"`
	Custom    string            `doc:"test"        json:"custom"          validate:"test_custom"`
	Formatted string            `doc:"test"        json:"formatted"       validate:"test_format"`
	Labels    map[string]string `doc:"test"        json:"labels"`
	Child     *nested           `doc:"test"        json:"child"`
	Items     [2]nested         `doc:"test"        json:"items"           validate:"dive"`
	Pass      string            `doc:"test"        json:"pass"                                             secret:"true"`
	Dir       string            `doc:"test"        json:"dir"             validate:"startswith=/etc"`
}

type badKind struct {
	C chan int `json:"c" doc:"test"`
}

type badSlice struct {
	C []chan int `json:"c" doc:"test"`
}

type badMap struct {
	C map[string]chan int `json:"c" doc:"test"`
}

type badTag struct {
	X string `json:"x" validate:"not_registered" doc:"test"`
}

type badBound struct {
	X string `json:"x" validate:"min=lots" doc:"test"`
}

func (s *SchemaPublicTestSuite) SetupSuite() {
	rule.Register("test_custom", rule.Pattern(`^x+$`))
	rule.Register("test_format", rule.Format("hostname"))
}

func (s *SchemaPublicTestSuite) TestSchemaFor() {
	tests := []struct {
		name         string
		t            reflect.Type
		wantErr      string
		validateFunc func(*behavior.Schema)
	}{
		{
			name: "when every supported kind and tag is used",
			t:    reflect.TypeFor[*everything](),
			validateFunc: func(sch *behavior.Schema) {
				assert.Equal(s.T(), "object", sch.Type)
				assert.Equal(s.T(), []string{
					"Plain", "named", "Commaonly", "flag", "count", "ratio", "size", "tags",
					"exact", "mode", "addr", "custom", "formatted", "labels", "child", "items",
					"pass", "dir",
				}, sch.Order)
				assert.Equal(s.T(), []string{"mode"}, sch.Required)

				p := sch.Properties
				assert.Equal(s.T(), "no json tag", p["Plain"].Description)
				assert.Equal(s.T(), "boolean", p["flag"].Type)
				assert.Equal(s.T(), "integer", p["count"].Type)
				assert.Equal(s.T(), 1.0, *p["count"].Minimum)
				assert.Equal(s.T(), 10.0, *p["count"].Maximum)
				assert.Equal(s.T(), "number", p["ratio"].Type)
				assert.Equal(s.T(), 0.5, *p["ratio"].Minimum)
				assert.Nil(s.T(), p["ratio"].Maximum)
				assert.Equal(s.T(), "integer", p["size"].Type)
				assert.Equal(s.T(), "array", p["tags"].Type)
				assert.Equal(s.T(), 1, *p["tags"].MinItems)
				assert.Equal(s.T(), 3, *p["tags"].MaxItems)
				assert.Equal(s.T(), 4, *p["exact"].MinLength)
				assert.Equal(s.T(), 4, *p["exact"].MaxLength)
				assert.Equal(s.T(), []string{"on", "off"}, p["mode"].Enum)
				assert.Equal(s.T(), "ip", p["addr"].Format)
				assert.Equal(s.T(), `^x+$`, p["custom"].Pattern)
				assert.Equal(s.T(), "hostname", p["formatted"].Format)
				assert.Equal(s.T(), "password", p["pass"].Format, "a secret is a password field")
				assert.True(s.T(), p["pass"].WriteOnly, "a secret is never shown back")
				assert.False(s.T(), p["formatted"].WriteOnly)
				assert.Equal(s.T(), `^/etc`, p["dir"].Pattern, "startswith is a prefix pattern")
				assert.Equal(s.T(), "string", p["labels"].AdditionalProperties.Type)
				assert.Equal(s.T(), "object", p["child"].Type)
				assert.Equal(s.T(), "string", p["items"].Items.Properties["name"].Type)
				assert.NotContains(s.T(), p, "Skipped")
				assert.NotContains(s.T(), p, "unexport")
			},
		},
		{
			name:    "when a field has no schema",
			t:       reflect.TypeFor[badKind](),
			wantErr: "kind chan",
		},
		{
			name:    "when a slice element has no schema",
			t:       reflect.TypeFor[badSlice](),
			wantErr: "kind chan",
		},
		{
			name:    "when a map value has no schema",
			t:       reflect.TypeFor[badMap](),
			wantErr: "kind chan",
		},
		{
			name:    "when a tag is not registered",
			t:       reflect.TypeFor[badTag](),
			wantErr: "not_registered",
		},
		{
			name:    "when a bound is not a number",
			t:       reflect.TypeFor[badBound](),
			wantErr: "not a number",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			sch, err := behavior.SchemaFor(tc.t)
			if tc.wantErr != "" {
				s.Require().Error(err)
				assert.Contains(s.T(), err.Error(), tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(sch)
		})
	}
}

func TestSchemaPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SchemaPublicTestSuite))
}

type shapeIn struct {
	Key   string `json:"key"   validate:"required"`
	Value string `json:"value" validate:"omitempty"`
	Skip  string `json:"-"`
}

type emptyIn struct{}

type spanTree struct {
	Name     string     `json:"name"`
	Children []spanTree `json:"children,omitempty"`
}

func (s *SchemaPublicTestSuite) TestShape() {
	in, err := behavior.SchemaFor(reflect.TypeFor[shapeIn]())
	s.Require().NoError(err)
	s.Equal("shapeIn { key*, value }", behavior.Shape(reflect.TypeFor[shapeIn](), in, true))
	s.Equal("shapeIn { key, value }", behavior.Shape(reflect.TypeFor[shapeIn](), in, false),
		"an output marks no field required")
	list, err := behavior.SchemaFor(reflect.TypeFor[[]shapeIn]())
	s.Require().NoError(err)
	s.Equal("[]shapeIn { key, value }", behavior.Shape(reflect.TypeFor[[]shapeIn](), list, false))
	s.Equal("emptyIn", behavior.Shape(reflect.TypeFor[emptyIn](), &behavior.Schema{}, true),
		"a type with no field is its name")
}

func (s *SchemaPublicTestSuite) TestFieldShape() {
	s.Equal("shapeIn { key*, value }", behavior.FieldShape(reflect.TypeFor[shapeIn](), true),
		"a field with no json name, or json \"-\", or unexported, is left out")
	s.Equal("spanTree { name, children }", behavior.FieldShape(reflect.TypeFor[spanTree](), false),
		"a type that holds itself has a shape")
	s.Equal(
		"[]spanTree { name, children }",
		behavior.FieldShape(reflect.TypeFor[[]spanTree](), false),
	)
	s.Equal("string", behavior.FieldShape(reflect.TypeFor[string](), true))
	s.Equal("", behavior.FieldShape(reflect.TypeFor[struct{ quiet int }](), true))
}
