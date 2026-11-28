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

type ValidatePublicTestSuite struct {
	suite.Suite
}

func TestValidatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ValidatePublicTestSuite))
}

// sentEmptyFields is a struct whose fields cover each case of SentEmpty.
type sentEmptyFields struct {
	Note   string `json:"note"           validate:"max=9,omitempty"`
	Kept   string `json:"kept,omitempty" validate:"omitempty"`
	Need   string `json:"need"           validate:"required"`
	Hidden string `json:"-"              validate:"omitempty"`
	NoName string `                      validate:"omitempty"`
}

func (s *ValidatePublicTestSuite) TestSentEmpty() {
	tests := []struct {
		name         string
		fields       []reflect.StructField
		validateFunc func(got []string)
	}{
		{
			name:   "when omitempty is anywhere in validate and the field has a json name",
			fields: fieldsOf[sentEmptyFields](),
			validateFunc: func(got []string) {
				s.Equal([]string{"note"}, got)
			},
		},
		{
			name: "when there are no fields",
			validateFunc: func(got []string) {
				s.Empty(got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(endpoint.SentEmpty(tt.fields))
		})
	}
}

// fieldsOf returns the fields of T, in order.
func fieldsOf[T any]() []reflect.StructField {
	fs := make([]reflect.StructField, 0)
	for f := range reflect.TypeFor[T]().Fields() {
		fs = append(fs, f)
	}

	return fs
}
