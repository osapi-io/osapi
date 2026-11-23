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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/user"
	"github.com/osapi-io/osapi/internal/behaviors/user/mocks"
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
	alice := user.Account{Name: "alice", UID: 1000}

	s.mock.EXPECT().List(gomock.Any()).Return([]user.Account{alice}, nil)
	list, err := user.List(ctx, s.mock, user.ListInput{})
	s.NoError(err)
	s.Equal([]user.Account{alice}, list)

	s.mock.EXPECT().Get(gomock.Any(), "alice").Return(alice, nil)
	got, err := user.Get(ctx, s.mock, user.GetInput{Name: "alice"})
	s.NoError(err)
	s.Equal(alice, got)

	in := user.CreateInput{Name: "alice"}
	s.mock.EXPECT().Create(gomock.Any(), in).Return(true, nil)
	res, err := user.Create(ctx, s.mock, in)
	s.NoError(err)
	s.Equal(user.Result{Name: "alice", Changed: true}, res)

	up := user.UpdateInput{Name: "alice", Shell: "/bin/sh"}
	s.mock.EXPECT().Update(gomock.Any(), up).Return(false, boom)
	res, err = user.Update(ctx, s.mock, up)
	s.ErrorIs(err, boom)
	s.Equal(user.Result{Name: "alice"}, res)

	s.mock.EXPECT().Delete(gomock.Any(), "alice").Return(true, nil)
	res, err = user.Delete(ctx, s.mock, user.DeleteInput{Name: "alice"})
	s.NoError(err)
	s.Equal(user.Result{Name: "alice", Changed: true}, res)

	s.mock.EXPECT().SetPassword(gomock.Any(), "alice", "$6$x$y").Return(nil)
	res, err = user.SetPassword(
		ctx,
		s.mock,
		user.PasswordInput{Name: "alice", PasswordHash: "$6$x$y"},
	)
	s.NoError(err)
	s.Equal(user.Result{Name: "alice", Changed: true}, res)

	s.mock.EXPECT().SetPassword(gomock.Any(), "alice", "").Return(boom)
	res, err = user.SetPassword(ctx, s.mock, user.PasswordInput{Name: "alice"})
	s.ErrorIs(err, boom)
	s.Equal(user.Result{Name: "alice"}, res)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
