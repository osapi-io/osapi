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

package group_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/group"
	"github.com/osapi-io/osapi/internal/behaviors/group/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	mock *mocks.MockProvider
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mock = mocks.NewMockProvider(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) TestMethods() {
	ctx := context.Background()
	boom := errors.New("boom")
	dev := group.Account{Name: "dev", GID: 1}

	s.mock.EXPECT().List(gomock.Any()).Return([]group.Account{dev}, nil)
	list, err := group.List(ctx, s.mock, group.ListInput{})
	s.NoError(err)
	s.Equal([]group.Account{dev}, list)

	s.mock.EXPECT().Get(gomock.Any(), "dev").Return(dev, nil)
	got, err := group.Get(ctx, s.mock, group.GetInput{Name: "dev"})
	s.NoError(err)
	s.Equal(dev, got)

	in := group.CreateInput{Name: "dev", GID: 1}
	s.mock.EXPECT().Create(gomock.Any(), in).Return(true, nil)
	res, err := group.Create(ctx, s.mock, in)
	s.NoError(err)
	s.Equal(group.Result{Name: "dev", Changed: true}, res)

	s.mock.EXPECT().Update(gomock.Any(), "dev", []string{"a"}).Return(false, boom)
	res, err = group.Update(ctx, s.mock, group.UpdateInput{Name: "dev", Members: []string{"a"}})
	s.ErrorIs(err, boom)
	s.Equal(group.Result{Name: "dev"}, res)

	s.mock.EXPECT().Delete(gomock.Any(), "dev").Return(true, nil)
	res, err = group.Delete(ctx, s.mock, group.DeleteInput{Name: "dev"})
	s.NoError(err)
	s.Equal(group.Result{Name: "dev", Changed: true}, res)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
