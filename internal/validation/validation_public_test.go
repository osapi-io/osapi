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

func (s *ValidationPublicTestSuite) TestAlphanumOrFact() {
	tests := []struct {
		name   string
		field  string
		wantOK bool
	}{
		{
			name:   "when alphanumeric value",
			field:  "eth0",
			wantOK: true,
		},
		{
			name:   "when fact reference",
			field:  "@fact.interface.primary",
			wantOK: true,
		},
		{
			name:   "when fact custom reference",
			field:  "@fact.custom.mykey",
			wantOK: true,
		},
		{
			name:   "when non-alphanum non-fact value",
			field:  "eth-0!",
			wantOK: false,
		},
		{
			name:   "when empty value",
			field:  "",
			wantOK: false,
		},
		{
			name:   "when partial fact prefix",
			field:  "@fact",
			wantOK: false,
		},
		{
			name:   "when at-sign without fact",
			field:  "@notfact.x",
			wantOK: false,
		},
		{
			name:   "when unknown fact key",
			field:  "@fact.primary_interface",
			wantOK: false,
		},
		{
			name:   "when fact with bare custom prefix",
			field:  "@fact.custom.",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(tt.field, "required,alphanum_or_fact")
			s.Equal(tt.wantOK, ok)
		})
	}
}

func (s *ValidationPublicTestSuite) TestIpOrFact() {
	tests := []struct {
		name   string
		field  string
		wantOK bool
	}{
		{
			name:   "when valid IPv4",
			field:  "1.1.1.1",
			wantOK: true,
		},
		{
			name:   "when valid IPv6",
			field:  "::1",
			wantOK: true,
		},
		{
			name:   "when fact reference",
			field:  "@fact.custom.gateway",
			wantOK: true,
		},
		{
			name:   "when fact interface primary",
			field:  "@fact.interface.primary",
			wantOK: true,
		},
		{
			name:   "when invalid address",
			field:  "not-an-ip",
			wantOK: false,
		},
		{
			name:   "when empty value",
			field:  "",
			wantOK: false,
		},
		{
			name:   "when partial fact prefix",
			field:  "@fact",
			wantOK: false,
		},
		{
			name:   "when at-sign without fact",
			field:  "@notfact.x",
			wantOK: false,
		},
		{
			name:   "when unknown fact key",
			field:  "@fact.primary_interface",
			wantOK: false,
		},
		{
			name:   "when fact with bare custom prefix",
			field:  "@fact.custom.",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, ok := validation.Var(tt.field, "required,ip_or_fact")
			s.Equal(tt.wantOK, ok)
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
