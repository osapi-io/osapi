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

package route_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/network/route"
	"github.com/osapi-io/osapi/internal/behaviors/network/route/mocks"
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
	s.mock.EXPECT().Routes(gomock.Any()).Return([]route.Entry{{Interface: "eth0"}}, nil)
	got, err := route.List(context.Background(), s.mock, route.ListInput{})
	s.Require().NoError(err)
	s.Equal([]route.Entry{{Interface: "eth0"}}, got)
}

func (s *MethodsPublicTestSuite) TestGet() {
	s.mock.EXPECT().InterfaceRoutes(gomock.Any(), "eth0").Return([]route.Entry{}, nil)
	got, err := route.Get(context.Background(), s.mock, route.GetInput{InterfaceName: "eth0"})
	s.Require().NoError(err)
	s.Equal([]route.Entry{}, got)
}

func (s *MethodsPublicTestSuite) TestCreate() {
	in := route.Config{InterfaceName: "eth1", Routes: []string{"10.0.0.0/8 via 10.0.1.1"}}
	s.mock.EXPECT().Create(gomock.Any(), in).Return(true, nil)
	got, err := route.Create(context.Background(), s.mock, route.CreateInput(in))
	s.Require().NoError(err)
	s.Equal(route.Result{InterfaceName: "eth1", Changed: true}, got)
}

func (s *MethodsPublicTestSuite) TestUpdate() {
	boom := errors.New("boom")
	s.mock.EXPECT().Update(gomock.Any(), route.Config{InterfaceName: "eth1"}).Return(false, boom)
	_, err := route.Update(context.Background(), s.mock, route.UpdateInput{InterfaceName: "eth1"})
	s.ErrorIs(err, boom)
}

func (s *MethodsPublicTestSuite) TestDelete() {
	s.mock.EXPECT().Remove(gomock.Any(), "eth1").Return(true, nil)
	got, err := route.Delete(context.Background(), s.mock, route.DeleteInput{InterfaceName: "eth1"})
	s.Require().NoError(err)
	s.Equal(route.Result{InterfaceName: "eth1", Changed: true}, got)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
