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

package iface_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/network/iface"
	"github.com/osapi-io/osapi/internal/behaviors/network/iface/mocks"
)

// MethodsPublicTestSuite checks that each method passes its input to the
// provider and returns what it gives.
type MethodsPublicTestSuite struct {
	suite.Suite

	mock *mocks.MockProvider
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.mock = mocks.NewMockProvider(gomock.NewController(s.T()))
}

func (s *MethodsPublicTestSuite) TestList() {
	s.mock.EXPECT().Interfaces(gomock.Any()).Return([]iface.Entry{{Name: "eth0"}}, nil)
	got, err := iface.List(context.Background(), s.mock, iface.ListInput{})
	s.Require().NoError(err)
	s.Equal([]iface.Entry{{Name: "eth0"}}, got)
}

func (s *MethodsPublicTestSuite) TestGet() {
	s.mock.EXPECT().Interface(gomock.Any(), "eth0").Return(iface.Entry{Name: "eth0"}, nil)
	got, err := iface.Get(context.Background(), s.mock, iface.GetInput{Name: "eth0"})
	s.Require().NoError(err)
	s.Equal(iface.Entry{Name: "eth0"}, got)
}

func (s *MethodsPublicTestSuite) TestCreate() {
	s.mock.EXPECT().Create(gomock.Any(), iface.Config{Name: "eth1", DHCP4: true}).Return(true, nil)
	got, err := iface.Create(
		context.Background(),
		s.mock,
		iface.CreateInput{Name: "eth1", DHCP4: true},
	)
	s.Require().NoError(err)
	s.Equal(iface.Result{Name: "eth1", Changed: true}, got)
}

func (s *MethodsPublicTestSuite) TestUpdate() {
	boom := errors.New("boom")
	s.mock.EXPECT().Update(gomock.Any(), iface.Config{Name: "eth1"}).Return(false, boom)
	_, err := iface.Update(context.Background(), s.mock, iface.UpdateInput{Name: "eth1"})
	s.ErrorIs(err, boom)
}

func (s *MethodsPublicTestSuite) TestDelete() {
	s.mock.EXPECT().Remove(gomock.Any(), "eth1").Return(true, nil)
	got, err := iface.Delete(context.Background(), s.mock, iface.DeleteInput{Name: "eth1"})
	s.Require().NoError(err)
	s.Equal(iface.Result{Name: "eth1", Changed: true}, got)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
