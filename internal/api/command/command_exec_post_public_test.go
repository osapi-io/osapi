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
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apicommand "github.com/retr0h/osapi/internal/api/command"
	"github.com/retr0h/osapi/internal/api/command/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/command"
)

type CommandExecPostPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apicommand.Command
	ctx           context.Context
}

func (s *CommandExecPostPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apicommand.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *CommandExecPostPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *CommandExecPostPublicTestSuite) TestPostCommandExec() {
	args := []string{"-la"}
	timeout := 30

	tests := []struct {
		name         string
		request      gen.PostCommandExecRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostCommandExecResponseObject)
	}{
		{
			name: "success",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Args:    &args,
					Timeout: &timeout,
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandExec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", &command.Result{
						Stdout:     "file1\nfile2",
						Stderr:     "",
						ExitCode:   0,
						DurationMs: 42,
						Changed:    true,
					}, "worker1", nil)
			},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				r, ok := resp.(gen.PostCommandExec202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("worker1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Stdout)
				s.Equal("file1\nfile2", *r.Results[0].Stdout)
				s.Require().NotNil(r.Results[0].ExitCode)
				s.Equal(0, *r.Results[0].ExitCode)
				s.Require().NotNil(r.Results[0].Changed)
				s.True(*r.Results[0].Changed)
			},
		},
		{
			name: "validation error missing command",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				r, ok := resp.(gen.PostCommandExec400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Command")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "validation error invalid timeout",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Timeout: intPtr(999),
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				r, ok := resp.(gen.PostCommandExec400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Timeout")
				s.Contains(*r.Error, "max")
			},
		},
		{
			name: "validation error empty target_hostname",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Args:    &args,
				},
				Params: gen.PostCommandExecParams{TargetHostname: strPtr("")},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				r, ok := resp.(gen.PostCommandExec400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "min")
			},
		},
		{
			name: "job client error",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Args:    &args,
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandExec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				_, ok := resp.(gen.PostCommandExec500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Args:    &args,
				},
				Params: gen.PostCommandExecParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandExecBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*command.Result{
						"server1": {
							Stdout:     "output1",
							ExitCode:   0,
							DurationMs: 10,
							Changed:    true,
						},
						"server2": {
							Stdout:     "output2",
							ExitCode:   0,
							DurationMs: 20,
							Changed:    true,
						},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				r, ok := resp.(gen.PostCommandExec202JSONResponse)
				s.True(ok)
				s.NotNil(r)
				s.Len(r.Results, 2)
				for _, result := range r.Results {
					s.Require().NotNil(result.Changed)
					s.True(*result.Changed)
				}
			},
		},
		{
			name: "broadcast error",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Args:    &args,
				},
				Params: gen.PostCommandExecParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandExecBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				_, ok := resp.(gen.PostCommandExec500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast with host errors",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Args:    &args,
				},
				Params: gen.PostCommandExecParams{TargetHostname: strPtr("_all")},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandExecBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*command.Result{
						"server1": {
							Stdout:     "output1",
							ExitCode:   0,
							DurationMs: 10,
						},
					}, map[string]string{
						"server2": "command not found",
					}, nil)
			},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				r, ok := resp.(gen.PostCommandExec202JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
			},
		},
		{
			name: "success with cwd",
			request: gen.PostCommandExecRequestObject{
				Body: &gen.PostCommandExecJSONRequestBody{
					Command: "ls",
					Cwd:     strPtr("/tmp"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandExec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), "/tmp", gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", &command.Result{
						Stdout:   "output",
						ExitCode: 0,
					}, "worker1", nil)
			},
			validateFunc: func(resp gen.PostCommandExecResponseObject) {
				_, ok := resp.(gen.PostCommandExec202JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostCommandExec(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestCommandExecPostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CommandExecPostPublicTestSuite))
}

func strPtr(
	s string,
) *string {
	return &s
}

func intPtr(
	i int,
) *int {
	return &i
}
