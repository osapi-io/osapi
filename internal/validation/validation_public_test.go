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

package validation_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/validation"
)

type ValidationPublicTestSuite struct {
	suite.Suite
}

func (s *ValidationPublicTestSuite) TestStruct() {
	type testStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	tests := []struct {
		name     string
		input    any
		wantOK   bool
		contains []string
	}{
		{
			name: "when valid struct",
			input: testStruct{
				Name:  "test",
				Email: "test@example.com",
			},
			wantOK: true,
		},
		{
			name: "when missing required field",
			input: testStruct{
				Email: "test@example.com",
			},
			wantOK:   false,
			contains: []string{"Name", "required"},
		},
		{
			name: "when invalid email",
			input: testStruct{
				Name:  "test",
				Email: "not-an-email",
			},
			wantOK:   false,
			contains: []string{"Email", "email"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			errMsg, ok := validation.Struct(tt.input)
			s.Equal(tt.wantOK, ok)

			if !ok {
				for _, c := range tt.contains {
					s.Contains(errMsg, c)
				}
			}
		})
	}
}

func (s *ValidationPublicTestSuite) TestVar() {
	tests := []struct {
		name     string
		field    any
		tag      string
		wantOK   bool
		contains []string
	}{
		{
			name:   "when valid field",
			field:  "hello",
			tag:    "required",
			wantOK: true,
		},
		{
			name:     "when empty required field",
			field:    "",
			tag:      "required",
			wantOK:   false,
			contains: []string{"required"},
		},
		{
			name:     "when invalid email",
			field:    "not-an-email",
			tag:      "email",
			wantOK:   false,
			contains: []string{"email"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			errMsg, ok := validation.Var(tt.field, tt.tag)
			s.Equal(tt.wantOK, ok)

			if !ok {
				for _, c := range tt.contains {
					s.Contains(errMsg, c)
				}
			}
		})
	}
}

func (s *ValidationPublicTestSuite) TestInstance() {
	tests := []struct {
		name string
	}{
		{
			name: "when returns shared validator instance",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			v := validation.Instance()
			s.NotNil(v)
		})
	}
}

func TestValidationPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationPublicTestSuite))
}
