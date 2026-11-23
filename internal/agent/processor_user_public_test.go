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

package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/provider/node/user"
	userMocks "github.com/osapi-io/osapi/internal/provider/node/user/mocks"
)

type ProcessorUserPublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
}

func (s *ProcessorUserPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *ProcessorUserPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ProcessorUserPublicTestSuite) newProcessor(
	userProvider user.Provider,
) agent.ProcessorFunc {
	return agent.NewNodeProcessor(
		userProvider,
		slog.Default(),
	)
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserOperation() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "nil provider returns error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: nil,
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "user provider not available")
				s.Nil(result)
			},
		},
		{
			name: "invalid user operation missing sub-operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid user operation: user")
				s.Nil(result)
			},
		},
		{
			name: "unsupported user sub-operation",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.unknown",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unsupported user operation: user.unknown")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var userProvider user.Provider
			if tt.setupMock != nil {
				userProvider = tt.setupMock()
			}

			processor := s.newProcessor(userProvider)
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserList() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful user list",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().ListUsers(gomock.Any()).Return([]user.User{
					{Name: "root", UID: 0, GID: 0, Home: "/root", Shell: "/bin/bash"},
					{Name: "john", UID: 1000, GID: 1000, Home: "/home/john", Shell: "/bin/bash"},
				}, nil)
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)
				var users []user.User
				decodeErr := json.Unmarshal(result, &users)
				s.NoError(decodeErr)
				s.Len(users, 2)
				s.Equal("root", users[0].Name)
				s.Equal("john", users[1].Name)
			},
		},
		{
			name: "user list provider error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.list",
				Data:      json.RawMessage(`{}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().ListUsers(gomock.Any()).Return(nil, errors.New("permission denied"))
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "permission denied")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.newProcessor(tt.setupMock())
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserGet() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful user get",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.get",
				Data:      json.RawMessage(`{"name":"john"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().GetUser(gomock.Any(), "john").Return(&user.User{
					Name: "john", UID: 1000, GID: 1000,
					Home: "/home/john", Shell: "/bin/bash",
				}, nil)
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)
				var u user.User
				decodeErr := json.Unmarshal(result, &u)
				s.NoError(decodeErr)
				s.Equal("john", u.Name)
				s.Equal(1000, u.UID)
			},
		},
		{
			name: "user get with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.get",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unmarshal user get data")
				s.Nil(result)
			},
		},
		{
			name: "user get provider error",
			jobRequest: job.Request{
				Type:      job.TypeQuery,
				Category:  "node",
				Operation: "user.get",
				Data:      json.RawMessage(`{"name":"missing"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().GetUser(gomock.Any(), "missing").Return(nil, errors.New("not found"))
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "not found")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.newProcessor(tt.setupMock())
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserCreate() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful user create",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.create",
				Data:      json.RawMessage(`{"name":"newuser","shell":"/bin/bash"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().CreateUser(gomock.Any(), user.CreateUserOpts{
					Name:  "newuser",
					Shell: "/bin/bash",
				}).Return(&user.Result{
					Name:    "newuser",
					Changed: true,
				}, nil)
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)
				var r user.Result
				decodeErr := json.Unmarshal(result, &r)
				s.NoError(decodeErr)
				s.Equal("newuser", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "user create with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.create",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unmarshal user create data")
				s.Nil(result)
			},
		},
		{
			name: "user create provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.create",
				Data:      json.RawMessage(`{"name":"existing"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("user already exists"))
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "user already exists")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.newProcessor(tt.setupMock())
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserUpdate() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful user update",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.update",
				Data:      json.RawMessage(`{"name":"john","opts":{"shell":"/bin/zsh"}}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().UpdateUser(gomock.Any(), "john", user.UpdateUserOpts{
					Shell: "/bin/zsh",
				}).Return(&user.Result{
					Name:    "john",
					Changed: true,
				}, nil)
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)
				var r user.Result
				decodeErr := json.Unmarshal(result, &r)
				s.NoError(decodeErr)
				s.Equal("john", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "user update with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.update",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unmarshal user update data")
				s.Nil(result)
			},
		},
		{
			name: "user update provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.update",
				Data:      json.RawMessage(`{"name":"john","opts":{"shell":"/bin/zsh"}}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "not found")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.newProcessor(tt.setupMock())
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserDelete() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful user delete",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.delete",
				Data:      json.RawMessage(`{"name":"olduser"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().DeleteUser(gomock.Any(), "olduser").Return(&user.Result{
					Name:    "olduser",
					Changed: true,
				}, nil)
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)
				var r user.Result
				decodeErr := json.Unmarshal(result, &r)
				s.NoError(decodeErr)
				s.Equal("olduser", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "user delete with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.delete",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unmarshal user delete data")
				s.Nil(result)
			},
		},
		{
			name: "user delete provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.delete",
				Data:      json.RawMessage(`{"name":"missing"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().DeleteUser(gomock.Any(), "missing").Return(nil, errors.New("not found"))
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "not found")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.newProcessor(tt.setupMock())
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func (s *ProcessorUserPublicTestSuite) TestProcessUserPassword() {
	tests := []struct {
		name         string
		jobRequest   job.Request
		setupMock    func() user.Provider
		validateFunc func(json.RawMessage, error)
	}{
		{
			name: "successful password change",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.password",
				Data:      json.RawMessage(`{"name":"john","password_hash":"$6$abcd$deadbeef"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().
					ChangePassword(gomock.Any(), "john", "$6$abcd$deadbeef").
					Return(&user.Result{
						Name:    "john",
						Changed: true,
					}, nil)
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.NoError(err)
				s.NotNil(result)
				var r user.Result
				decodeErr := json.Unmarshal(result, &r)
				s.NoError(decodeErr)
				s.Equal("john", r.Name)
				s.True(r.Changed)
			},
		},
		{
			name: "password change with invalid JSON data",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.password",
				Data:      json.RawMessage(`invalid json`),
			},
			setupMock: func() user.Provider {
				return userMocks.NewMockProvider(s.mockCtrl)
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "unmarshal user password data")
				s.Nil(result)
			},
		},
		{
			name: "password change provider error",
			jobRequest: job.Request{
				Type:      job.TypeModify,
				Category:  "node",
				Operation: "user.password",
				Data:      json.RawMessage(`{"name":"john","password_hash":"$6$abcd$deadbeef"}`),
			},
			setupMock: func() user.Provider {
				m := userMocks.NewMockProvider(s.mockCtrl)
				m.EXPECT().
					ChangePassword(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("permission denied"))
				return m
			},
			validateFunc: func(result json.RawMessage, err error) {
				s.Error(err)
				s.Contains(err.Error(), "permission denied")
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			processor := s.newProcessor(tt.setupMock())
			tt.validateFunc(processor(context.Background(), tt.jobRequest))
		})
	}
}

func TestProcessorUserPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ProcessorUserPublicTestSuite))
}
