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

package sysctl_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	prov "github.com/osapi-io/osapi/internal/provider/node/sysctl"
	"github.com/osapi-io/osapi/internal/provider/node/sysctl/mocks"
	"github.com/osapi-io/osapi/pkg/behavior"
	"github.com/osapi-io/osapi/pkg/behavior/sysctl"
)

type BehaviorPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	mock *mocks.MockProvider
}

func (s *BehaviorPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mock = mocks.NewMockProvider(s.ctrl)
}

func (s *BehaviorPublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *BehaviorPublicTestSuite) TestBehavior() {
	assert.Same(s.T(), sysctl.Behavior, behavior.Lookup("sysctl"))
	assert.Empty(s.T(), behavior.ValidateBehavior(sysctl.Behavior))
}

func (s *BehaviorPublicTestSuite) TestList() {
	tests := []struct {
		name    string
		setup   func()
		want    []sysctl.Entry
		wantErr bool
	}{
		{
			name: "when the provider lists entries",
			setup: func() {
				s.mock.EXPECT().List(gomock.Any()).Return([]prov.Entry{
					{Key: "vm.swappiness", Value: "10"},
				}, nil)
			},
			want: []sysctl.Entry{{Key: "vm.swappiness", Value: "10"}},
		},
		{
			name: "when the provider fails",
			setup: func() {
				s.mock.EXPECT().List(gomock.Any()).Return(nil, errors.New("boom"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := sysctl.List(context.Background(), s.mock, sysctl.ListInput{})

			if tc.wantErr {
				s.Require().Error(err)

				return
			}
			s.Require().NoError(err)
			assert.Equal(s.T(), tc.want, got)
		})
	}
}

func (s *BehaviorPublicTestSuite) TestGet() {
	tests := []struct {
		name    string
		setup   func()
		want    sysctl.Entry
		wantErr bool
	}{
		{
			name: "when the provider returns the entry",
			setup: func() {
				s.mock.EXPECT().Get(gomock.Any(), "vm.swappiness").
					Return(&prov.Entry{Key: "vm.swappiness", Value: "10"}, nil)
			},
			want: sysctl.Entry{Key: "vm.swappiness", Value: "10"},
		},
		{
			name: "when the provider fails",
			setup: func() {
				s.mock.EXPECT().Get(gomock.Any(), "vm.swappiness").
					Return(nil, errors.New("boom"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := sysctl.Get(
				context.Background(), s.mock, sysctl.GetInput{Key: "vm.swappiness"},
			)

			if tc.wantErr {
				s.Require().Error(err)

				return
			}
			s.Require().NoError(err)
			assert.Equal(s.T(), tc.want, got)
		})
	}
}

func (s *BehaviorPublicTestSuite) TestCreate() {
	tests := []struct {
		name    string
		setup   func()
		want    sysctl.Result
		wantErr bool
	}{
		{
			name: "when the provider creates the entry",
			setup: func() {
				s.mock.EXPECT().Create(gomock.Any(), prov.Entry{Key: "vm.swappiness", Value: "10"}).
					Return(&prov.CreateResult{Key: "vm.swappiness", Changed: true}, nil)
			},
			want: sysctl.Result{Key: "vm.swappiness", Changed: true},
		},
		{
			name: "when the provider fails",
			setup: func() {
				s.mock.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("boom"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := sysctl.Create(context.Background(), s.mock,
				sysctl.CreateInput{Key: "vm.swappiness", Value: "10"})

			if tc.wantErr {
				s.Require().Error(err)

				return
			}
			s.Require().NoError(err)
			assert.Equal(s.T(), tc.want, got)
		})
	}
}

func (s *BehaviorPublicTestSuite) TestUpdate() {
	tests := []struct {
		name    string
		setup   func()
		want    sysctl.Result
		wantErr bool
	}{
		{
			name: "when the provider updates the entry",
			setup: func() {
				s.mock.EXPECT().Update(gomock.Any(), prov.Entry{Key: "vm.swappiness", Value: "10"}).
					Return(&prov.UpdateResult{Key: "vm.swappiness", Changed: false}, nil)
			},
			want: sysctl.Result{Key: "vm.swappiness", Changed: false},
		},
		{
			name: "when the provider fails",
			setup: func() {
				s.mock.EXPECT().Update(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("boom"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := sysctl.Update(context.Background(), s.mock,
				sysctl.UpdateInput{Key: "vm.swappiness", Value: "10"})

			if tc.wantErr {
				s.Require().Error(err)

				return
			}
			s.Require().NoError(err)
			assert.Equal(s.T(), tc.want, got)
		})
	}
}

func (s *BehaviorPublicTestSuite) TestDelete() {
	tests := []struct {
		name    string
		setup   func()
		want    sysctl.Result
		wantErr bool
	}{
		{
			name: "when the provider deletes the entry",
			setup: func() {
				s.mock.EXPECT().Delete(gomock.Any(), "vm.swappiness").
					Return(&prov.DeleteResult{Key: "vm.swappiness", Changed: true}, nil)
			},
			want: sysctl.Result{Key: "vm.swappiness", Changed: true},
		},
		{
			name: "when the provider fails",
			setup: func() {
				s.mock.EXPECT().Delete(gomock.Any(), "vm.swappiness").
					Return(nil, errors.New("boom"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := sysctl.Delete(
				context.Background(), s.mock, sysctl.DeleteInput{Key: "vm.swappiness"},
			)

			if tc.wantErr {
				s.Require().Error(err)

				return
			}
			s.Require().NoError(err)
			assert.Equal(s.T(), tc.want, got)
		})
	}
}

func TestBehaviorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(BehaviorPublicTestSuite))
}
