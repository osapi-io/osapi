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

package file_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/controller/api/file"
)

type ValidateFilePublicTestSuite struct {
	suite.Suite
}

func (suite *ValidateFilePublicTestSuite) TestValidateFileName() {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "when valid name",
			input: "nginx.conf",
			valid: true,
		},
		{
			name:  "when valid name with path chars",
			input: "app.conf.tmpl",
			valid: true,
		},
		{
			name:  "when empty name",
			input: "",
			valid: false,
		},
		{
			name:  "when name exceeds 255 chars",
			input: strings.Repeat("a", 256),
			valid: false,
		},
		{
			name:  "when name at max 255 chars",
			input: strings.Repeat("a", 255),
			valid: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			errMsg, ok := file.ExportValidateFileName(tc.input)

			if tc.valid {
				suite.True(ok)
				suite.Empty(errMsg)
			} else {
				suite.False(ok)
				suite.NotEmpty(errMsg)
			}
		})
	}
}

func TestValidateFilePublicTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateFilePublicTestSuite))
}
