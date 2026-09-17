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

package user_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/node/user"
)

// ValidatePasswordInputPublicTestSuite exercises validatePasswordInput
// directly through ValidatePasswordInputForTest (see export_test.go). Its
// callers, CreateUser and ChangePassword, both validate the name via
// validateAccountName first, which already rejects a colon or line break —
// so this suite is the only way to reach the name-side branch.
type ValidatePasswordInputPublicTestSuite struct {
	suite.Suite
}

func (s *ValidatePasswordInputPublicTestSuite) TestValidatePasswordInput() {
	tests := []struct {
		name         string
		userName     string
		passwordHash string
		validateFunc func(err error)
	}{
		{
			name:         "when name contains a colon",
			userName:     "john:root",
			passwordHash: "$6$abcd$deadbeef",
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid user name: must not contain a colon or line break")
			},
		},
		{
			name:         "when name contains a line break",
			userName:     "john\nroot:pwned",
			passwordHash: "$6$abcd$deadbeef",
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid user name: must not contain a colon or line break")
			},
		},
		{
			name:         "when name and hash are valid",
			userName:     "john",
			passwordHash: "$6$abcd$deadbeef",
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := user.ValidatePasswordInputForTest(tc.userName, tc.passwordHash)
			tc.validateFunc(err)
		})
	}
}

func TestValidatePasswordInputPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ValidatePasswordInputPublicTestSuite))
}
