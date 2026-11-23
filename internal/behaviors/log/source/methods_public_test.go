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

package source_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/log/source"
	"github.com/osapi-io/osapi/internal/behaviors/log/source/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name  string
		names []string
		err   error
		want  []source.Source
	}{
		{
			name: "when the provider reads sources", names: []string{"cron", "sshd"},
			want: []source.Source{{Name: "cron"}, {Name: "sshd"}},
		},
		{name: "when the provider fails", err: errors.New("journal gone")},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			p := mocks.NewMockProvider(gomock.NewController(s.T()))
			p.EXPECT().Sources(gomock.Any()).Return(tc.names, tc.err)

			got, err := source.List(context.Background(), p, source.ListInput{})
			s.Equal(tc.want, got)
			s.Equal(tc.err, err)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
