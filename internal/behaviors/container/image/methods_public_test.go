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

package image_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/container/image"
	"github.com/osapi-io/osapi/internal/behaviors/container/image/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

// TestMethods: each method hands its input to the provider and gives back
// what it answers.
func (s *MethodsPublicTestSuite) TestMethods() {
	ctx := context.Background()
	m := mocks.NewMockProvider(gomock.NewController(s.T()))
	m.EXPECT().List(gomock.Any()).Return([]image.Entry{{Repository: "alpine"}}, nil)
	m.EXPECT().
		Pull(gomock.Any(), "alpine").
		Return(image.PullResult{Image: "alpine", Changed: true}, nil)
	m.EXPECT().Delete(gomock.Any(), "alpine", true).Return(true, nil)

	list, err := image.List(ctx, m, image.ListInput{})
	s.Require().NoError(err)
	s.Equal([]image.Entry{{Repository: "alpine"}}, list)
	pulled, err := image.Pull(ctx, m, image.PullInput{Image: "alpine"})
	s.Require().NoError(err)
	s.True(pulled.Changed)
	got, err := image.Delete(ctx, m, image.DeleteInput{Image: "alpine", Force: true})
	s.Require().NoError(err)
	s.Equal(image.Result{Image: "alpine", Changed: true}, got)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
