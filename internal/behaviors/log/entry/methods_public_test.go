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

package entry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/log/entry"
	"github.com/osapi-io/osapi/internal/behaviors/log/entry/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name string
		out  []entry.Entry
		err  error
	}{
		{name: "when the provider reads entries", out: []entry.Entry{{Message: "x"}}},
		{name: "when the provider fails", err: errors.New("journal gone")},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			p := mocks.NewMockProvider(gomock.NewController(s.T()))
			in := entry.ListInput{Unit: "nginx.service"}
			p.EXPECT().Entries(gomock.Any(), in).Return(tc.out, tc.err)

			got, err := entry.List(context.Background(), p, in)
			s.Equal(tc.out, got)
			s.Equal(tc.err, err)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
