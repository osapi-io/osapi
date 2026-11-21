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

package endpoint_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// AuditPublicTestSuite covers what the DSL audit added to the endpoint
// layer (osapi #974, #977, #982): the optional rule, secret in the spec,
// and dive on the items of a list.
type AuditPublicTestSuite struct {
	suite.Suite
}

func (s *AuditPublicTestSuite) TestSentEmpty() {
	type body struct {
		Note   string `json:"note"           validate:"max=9,omitempty"`
		Kept   string `json:"kept,omitempty" validate:"omitempty"`
		Need   string `json:"need"           validate:"required"`
		Hidden string `json:"-"              validate:"omitempty"`
		NoName string `                      validate:"omitempty"`
	}
	fs := make([]reflect.StructField, 0, 5)
	for f := range reflect.TypeFor[body]().Fields() {
		fs = append(fs, f)
	}
	s.Equal([]string{"note"}, endpoint.SentEmpty(fs),
		"omitempty anywhere in validate, and no json name is not a body field")
}

type pointerBodyIn struct {
	ID   string         `json:"id" in:"path" validate:"required"`
	Body *sentEmptyBody `          in:"body" validate:"required"`
}

func (s *AuditPublicTestSuite) TestSentEmptyThroughAPointerBody() {
	g := endpoint.Define("node", "The nodes",
		endpoint.SubDoc("one", "One node"),
		endpoint.Post[pointerBodyIn, Node]("reject").Sub("one"))
	problems := endpoint.Validate(g)
	msgs := make([]string, 0, len(problems))
	for _, p := range problems {
		msgs = append(msgs, p.Message)
	}
	s.Contains(msgs, "optional body field reason is sent as its zero value when it is empty.")
}

type secretIn struct {
	Password string   `json:"password" validate:"required"                   secret:"true"`
	Tags     []string `json:"tags"     validate:"omitempty,max=3,dive,max=9"`
	Bare     string   `json:"bare"     validate:"omitempty,dive,max=9"`
}

func (s *AuditPublicTestSuite) TestSecretAndDiveInTheSpec() {
	doc, err := endpoint.OpenAPI(
		endpoint.Define("node", "The nodes", endpoint.Post[secretIn, Node]("create")), "x")
	s.Require().NoError(err)
	var body map[string]any
	for _, path := range doc.Paths.Map() {
		schema := path.Post.RequestBody.Value.Content["application/json"].Schema.Value
		pw := schema.Properties["password"].Value
		s.True(pw.WriteOnly, "a secret is write-only")
		s.Equal("password", pw.Format)
		tags := schema.Properties["tags"].Value
		s.Equal(uint64(3), *tags.MaxItems, "max before dive is the list's")
		s.Equal(uint64(9), *tags.Items.Value.MaxLength, "max after dive is each item's")
		s.Nil(schema.Properties["bare"].Value.MaxLength, "dive on a string constrains no item")
		body = map[string]any{"ok": true}
	}
	s.NotNil(body)
}

func TestAuditPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AuditPublicTestSuite))
}
