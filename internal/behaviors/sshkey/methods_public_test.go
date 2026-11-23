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

package sshkey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behaviors/sshkey"
	"github.com/osapi-io/osapi/internal/behaviors/sshkey/mocks"
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
	k := sshkey.Key{Type: "ssh-ed25519", Fingerprint: "SHA256:x"}

	s.mock.EXPECT().Keys(gomock.Any(), "alice").Return([]sshkey.Key{k}, nil)
	keys, err := sshkey.List(ctx, s.mock, sshkey.ListInput{Name: "alice"})
	s.NoError(err)
	s.Equal([]sshkey.Key{k}, keys)

	s.mock.EXPECT().Add(gomock.Any(), "alice", "ssh-ed25519 AAAA").Return("SHA256:x", true, nil)
	res, err := sshkey.Create(
		ctx,
		s.mock,
		sshkey.CreateInput{Name: "alice", Key: "ssh-ed25519 AAAA"},
	)
	s.NoError(err)
	s.Equal(sshkey.Result{Fingerprint: "SHA256:x", Changed: true}, res)

	s.mock.EXPECT().Remove(gomock.Any(), "alice", "SHA256:x").Return(false, boom)
	res, err = sshkey.Delete(
		ctx,
		s.mock,
		sshkey.DeleteInput{Name: "alice", Fingerprint: "SHA256:x"},
	)
	s.ErrorIs(err, boom)
	s.Equal(sshkey.Result{Fingerprint: "SHA256:x"}, res)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
