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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type APITestSuite struct {
	suite.Suite
}

type sizeIn struct {
	Size *int `json:"size"`
}

type sizeOut struct {
	Changed bool `json:"changed"`
}

// setSize is a write whose example input has no Go literal: a pointer to an
// int.
func setSize(
	_ context.Context,
	_ struct{},
	_ sizeIn,
) (sizeOut, error) {
	return sizeOut{}, nil
}

func (s *APITestSuite) TestResponseCases() {
	size := 3

	tests := []struct {
		name         string
		b            *behavior.Behavior
		validateFunc func(cases []httpCase, err error)
	}{
		{
			name: "when an example's input has no Go literal, it is an error, not a test of the wrong input",
			b: behavior.Define("size", "No literal", behavior.Node, behavior.API,
				behavior.Write("set", setSize).Describe("Set it").
					Example("A size", sizeIn{Size: &size}, sizeOut{}),
			),
			validateFunc: func(cases []httpCase, err error) {
				s.ErrorContains(err, `example "A size": Size: a *int has no Go literal here`)
				s.Nil(cases)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cases, err := responseCases(tc.b, tc.b.Method("set"))

			tc.validateFunc(cases, err)
		})
	}
}

func TestAPITestSuite(
	t *testing.T,
) {
	suite.Run(t, new(APITestSuite))
}
