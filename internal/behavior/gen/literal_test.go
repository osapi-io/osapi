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

package main

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/suite"
)

type LiteralTestSuite struct {
	suite.Suite
}

type literalInner struct {
	Name string
}

type literalIn struct {
	Inner *literalInner
	Any   any
	Vars  map[string]any
	Count uint64
	Ratio float32
}

type namedUint uint32

// source is c as the Go source it renders.
func source(
	c jen.Code,
) string {
	return fmt.Sprintf("%#v", c)
}

func (s *LiteralTestSuite) TestLiteral() {
	tests := []struct {
		name         string
		in           any
		validateFunc func(got jen.Code, err error)
	}{
		{
			name: "when a field is a pointer to a struct, it is the struct's address",
			in:   literalIn{Inner: &literalInner{Name: "a"}},
			validateFunc: func(got jen.Code, err error) {
				s.Require().NoError(err)
				s.Equal(`gen.literalIn{Inner: &gen.literalInner{Name: "a"}}`, source(got))
			},
		},
		{
			name: "when a field is an interface, it is its value",
			in:   literalIn{Any: "x"},
			validateFunc: func(got jen.Code, err error) {
				s.Require().NoError(err)
				s.Equal(`gen.literalIn{Any: "x"}`, source(got))
			},
		},
		{
			name: "when a map holds interface values, the map is map[string]any",
			in:   literalIn{Vars: map[string]any{"k": "v"}},
			validateFunc: func(got jen.Code, err error) {
				s.Require().NoError(err)
				s.Equal(`gen.literalIn{Vars: map[string]any{"k": "v"}}`, source(got))
			},
		},
		{
			name: "when an unsigned value is above the largest int, it is kept",
			in:   literalIn{Count: math.MaxUint64},
			validateFunc: func(got jen.Code, err error) {
				s.Require().NoError(err)
				s.Equal("gen.literalIn{Count: uint64(0xffffffffffffffff)}", source(got))
			},
		},
		{
			name: "when a float32 has a short form, it is written short",
			in:   literalIn{Ratio: 0.1},
			validateFunc: func(got jen.Code, err error) {
				s.Require().NoError(err)
				s.Equal("gen.literalIn{Ratio: float32(0.1)}", source(got))
			},
		},
		{
			name: "when a named unsigned type has a value, it is converted to it",
			in:   namedUint(7),
			validateFunc: func(got jen.Code, err error) {
				s.Require().NoError(err)
				s.Equal("gen.namedUint(uint64(0x7))", source(got))
			},
		},
		{
			name: "when a pointer is to a scalar, it has no literal",
			in:   new(int),
			validateFunc: func(_ jen.Code, err error) {
				s.ErrorContains(err, "has no Go literal")
			},
		},
		{
			name: "when a value is a func, it has no literal",
			in:   func() {},
			validateFunc: func(_ jen.Code, err error) {
				s.ErrorContains(err, "has no Go literal")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, err := literal(reflect.ValueOf(tc.in))

			tc.validateFunc(got, err)
		})
	}
}

func TestLiteralTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LiteralTestSuite))
}
