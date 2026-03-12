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

package orchestrator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

// Compile-time interface check.
var _ orchestrator.RuntimeTarget = (*orchestrator.DockerTarget)(nil)

type DockerTargetPublicTestSuite struct {
	suite.Suite
}

func TestDockerTargetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTargetPublicTestSuite))
}

func (s *DockerTargetPublicTestSuite) TestNewDockerTarget() {
	tests := []struct {
		name         string
		targetName   string
		image        string
		validateFunc func(target *orchestrator.DockerTarget)
	}{
		{
			name:       "returns correct name and runtime",
			targetName: "web",
			image:      "ubuntu:24.04",
			validateFunc: func(target *orchestrator.DockerTarget) {
				s.Equal("web", target.Name())
				s.Equal("docker", target.Runtime())
				s.Equal("ubuntu:24.04", target.Image())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			target := orchestrator.NewDockerTarget(
				tt.targetName,
				tt.image,
				func(
					_ context.Context,
					_ string,
					_ []string,
				) (string, string, int, error) {
					return "", "", 0, nil
				},
			)
			tt.validateFunc(target)
		})
	}
}

func (s *DockerTargetPublicTestSuite) TestExecProvider() {
	tests := []struct {
		name         string
		provider     string
		operation    string
		data         []byte
		execFn       orchestrator.ExecFn
		validateFunc func(result []byte, err error, capturedCmd []string)
	}{
		{
			name:      "constructs correct command without data",
			provider:  "node.host",
			operation: "get",
			data:      nil,
			execFn: func(
				_ context.Context,
				_ string,
				_ []string,
			) (string, string, int, error) {
				return `{"hostname":"web-01"}`, "", 0, nil
			},
			validateFunc: func(result []byte, err error, _ []string) {
				s.NoError(err)
				s.Equal(`{"hostname":"web-01"}`, string(result))
			},
		},
		{
			name:      "constructs correct command with data",
			provider:  "network.dns",
			operation: "set",
			data:      []byte(`{"servers":["8.8.8.8"]}`),
			execFn:    nil, // set below to capture command
			validateFunc: func(_ []byte, _ error, capturedCmd []string) {
				s.Equal([]string{
					"/osapi", "provider", "run", "network.dns", "set",
					"--data", `{"servers":["8.8.8.8"]}`,
				}, capturedCmd)
			},
		},
		{
			name:      "returns error on non-zero exit",
			provider:  "node.host",
			operation: "get",
			data:      nil,
			execFn: func(
				_ context.Context,
				_ string,
				_ []string,
			) (string, string, int, error) {
				return "", "command not found", 1, nil
			},
			validateFunc: func(result []byte, err error, _ []string) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "failed (exit 1)")
				s.Contains(err.Error(), "command not found")
			},
		},
		{
			name:      "returns error on exec failure",
			provider:  "node.host",
			operation: "get",
			data:      nil,
			execFn: func(
				_ context.Context,
				_ string,
				_ []string,
			) (string, string, int, error) {
				return "", "", 0, fmt.Errorf("connection refused")
			},
			validateFunc: func(result []byte, err error, _ []string) {
				s.Error(err)
				s.Nil(result)
				s.Contains(err.Error(), "exec provider in container")
				s.Contains(err.Error(), "connection refused")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var capturedCmd []string

			execFn := tt.execFn
			if execFn == nil {
				execFn = func(
					_ context.Context,
					_ string,
					cmd []string,
				) (string, string, int, error) {
					capturedCmd = cmd

					return `{}`, "", 0, nil
				}
			}

			target := orchestrator.NewDockerTarget("web", "ubuntu:24.04", execFn)
			result, err := target.ExecProvider(
				context.Background(),
				tt.provider,
				tt.operation,
				tt.data,
			)
			tt.validateFunc(result, err, capturedCmd)
		})
	}
}
