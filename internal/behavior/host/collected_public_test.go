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

package host_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

type CollectedPublicTestSuite struct {
	suite.Suite
}

type info struct{ Name string }

func (s *CollectedPublicTestSuite) TestCollected() {
	failed := errors.New("collector failed")
	tests := []struct {
		name         string
		got          any
		err          error
		validateFunc func(*info, error)
	}{
		{
			name: "a result of the type",
			got:  &info{Name: "a"},
			validateFunc: func(got *info, err error) {
				s.Require().NoError(err)
				s.Equal("a", got.Name)
			},
		},
		{
			name: "a collector's error",
			err:  failed,
			validateFunc: func(got *info, err error) {
				s.ErrorIs(err, failed)
				s.ErrorContains(err, "read it: ")
				s.Nil(got)
			},
		},
		{
			name: "a result of another type",
			got:  "a",
			validateFunc: func(got *info, err error) {
				s.ErrorContains(err, "read it: unexpected result string")
				s.Nil(got)
			},
		},
		{
			name: "no result",
			validateFunc: func(got *info, err error) {
				s.ErrorContains(err, "read it: unexpected result <nil>")
				s.Nil(got)
			},
		},
		{
			name: "a nil result of the type",
			got:  (*info)(nil),
			validateFunc: func(got *info, err error) {
				s.ErrorContains(err, "unexpected result")
				s.Nil(got)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(host.Collected[info](tc.got, tc.err, "read it"))
		})
	}
}

func TestCollectedPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CollectedPublicTestSuite))
}
