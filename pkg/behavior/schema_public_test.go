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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/behavior"
)

type SchemaPublicTestSuite struct {
	suite.Suite
}

type nested struct {
	Name string `json:"name"`
}

type everything struct {
	Plain     string            `doc:"no json tag"`
	Named     string            `                  json:"named,omitempty"`
	Commaonly string            `                  json:",omitempty"`
	Skipped   string            `                  json:"-"`
	unexport  string            //nolint:unused
	Flag      bool              `                  json:"flag"`
	Count     int               `                  json:"count"           validate:"min=1,max=10"`
	Ratio     float64           `                  json:"ratio"           validate:"min=0.5"`
	Size      uint8             `                  json:"size"`
	Tags      []string          `                  json:"tags"            validate:"min=1,max=3"`
	Exact     string            `                  json:"exact"           validate:"len=4"`
	Mode      string            `                  json:"mode"            validate:"required,oneof=on off"`
	Addr      string            `                  json:"addr"            validate:"omitempty,ip"`
	Custom    string            `                  json:"custom"          validate:"test_custom"`
	Formatted string            `                  json:"formatted"       validate:"test_format"`
	Labels    map[string]string `                  json:"labels"`
	Child     *nested           `                  json:"child"`
	Items     [2]nested         `                  json:"items"           validate:"dive"`
}

type badKind struct {
	C chan int `json:"c"`
}

type badSlice struct {
	C []chan int `json:"c"`
}

type badMap struct {
	C map[string]chan int `json:"c"`
}

type badTag struct {
	X string `json:"x" validate:"not_registered"`
}

type badBound struct {
	X string `json:"x" validate:"min=lots"`
}

func (s *SchemaPublicTestSuite) SetupSuite() {
	behavior.ValidatorSchema("test_custom", behavior.Pattern(`^x+$`))
	behavior.ValidatorSchema("test_format", behavior.Format("hostname"))
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
