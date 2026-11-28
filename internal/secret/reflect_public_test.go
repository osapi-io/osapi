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

package secret_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/secret"
)

type ReflectPublicTestSuite struct {
	suite.Suite
}

func TestReflectPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReflectPublicTestSuite))
}

// selfRef is a type that refers to itself, with a secret field.
type selfRef struct {
	Next *selfRef     `json:"next"`
	Pass secret.Value `json:"pass"`
}

func (s *ReflectPublicTestSuite) TestIs() {
	tests := []struct {
		name string
		t    reflect.Type
		want bool
	}{
		{name: "when the type is a secret", t: reflectTypeOf[secret.Value](), want: true},
		{
			name: "when the type is a map of secrets",
			t:    reflectTypeOf[map[string]secret.Value](),
			want: true,
		},
		{
			name: "when the type is a pointer to a slice of secrets",
			t:    reflectTypeOf[*[]secret.Value](),
			want: true,
		},
		{name: "when the type is a string", t: reflectTypeOf[string]()},
		{name: "when the type is a struct of secrets", t: reflectTypeOf[inner]()},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, secret.Is(tc.t))
		})
	}
}

func (s *ReflectPublicTestSuite) TestFields() {
	tests := []struct {
		name string
		t    reflect.Type
		want []string
	}{
		{
			name: "when the input has secrets in each shape",
			t:    reflectTypeOf[*input](),
			want: []string{"env", "key", "list", "pair", "password", "ptr", "token"},
		},
		{
			name: "when the type refers to itself",
			t:    reflectTypeOf[selfRef](),
			want: []string{"pass"},
		},
		{name: "when the type is not a struct", t: reflectTypeOf[string]()},
		{
			name: "when the field has no json tag, it is named for the field",
			t:    reflectTypeOf[struct{ Untagged secret.Value }](),
			want: []string{"Untagged"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := secret.Fields(tc.t)
			if tc.want == nil {
				s.Empty(got)

				return
			}
			s.Equal(tc.want, got)
		})
	}
}
