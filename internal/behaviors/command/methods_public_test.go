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

package command_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/command"
	"github.com/osapi-io/osapi/internal/behaviors/command/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestMethods() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	p := mocks.NewMockProvider(ctrl)
	ctx := context.Background()

	in := command.ExecInput{Command: "true"}
	p.EXPECT().Exec(gomock.Any(), in).Return(command.Result{Changed: true}, nil)
	got, err := command.Exec(ctx, p, in)
	s.NoError(err)
	s.True(got.Changed)

	sh := command.ShellInput{Command: "true"}
	p.EXPECT().Shell(gomock.Any(), sh).Return(command.Result{ExitCode: 2}, nil)
	got, err = command.Shell(ctx, p, sh)
	s.NoError(err)
	s.Equal(2, got.ExitCode)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
