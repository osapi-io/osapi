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

package password_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/password"
)

type HashPublicTestSuite struct {
	suite.Suite
}

func (s *HashPublicTestSuite) TearDownSubTest() {
	password.ResetGenerate()
}

func (s *HashPublicTestSuite) TestHash() {
	tests := []struct {
		name         string
		plain        string
		setup        func()
		validateFunc func(hash string, err error)
	}{
		{
			name:  "produces a sha-512 crypt hash",
			plain: "correct horse battery staple",
			setup: func() {},
			validateFunc: func(hash string, err error) {
				s.NoError(err)
				s.True(strings.HasPrefix(hash, "$6$"))
				s.NotContains(hash, "correct horse battery staple")
			},
		},
		{
			name:  "empty password still hashes",
			plain: "",
			setup: func() {},
			validateFunc: func(hash string, err error) {
				s.NoError(err)
				s.True(strings.HasPrefix(hash, "$6$"))
			},
		},
		{
			name:  "two calls produce different hashes",
			plain: "same-password",
			setup: func() {},
			validateFunc: func(hash string, err error) {
				s.NoError(err)
				other, err := password.Hash("same-password")
				s.NoError(err)
				s.NotEqual(hash, other, "salts must be random per call")
			},
		},
		{
			name:  "when the underlying generator fails",
			plain: "correct horse battery staple",
			setup: func() {
				password.SetGenerate(func(_ []byte, _ []byte) (string, error) {
					return "", errors.New("salt decode failed")
				})
			},
			validateFunc: func(hash string, err error) {
				s.Error(err)
				s.Empty(hash)
				s.Contains(err.Error(), "hash password: salt decode failed")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			hash, err := password.Hash(tc.plain)
			tc.validateFunc(hash, err)
		})
	}
}

func TestHashPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HashPublicTestSuite))
}
