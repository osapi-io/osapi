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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/orchestrator"
)

type DeployPublicTestSuite struct {
	suite.Suite
}

func TestDeployPublicTestSuite(t *testing.T) {
	suite.Run(t, new(DeployPublicTestSuite))
}

func (suite *DeployPublicTestSuite) TestPrepare() {
	tests := []struct {
		name         string
		binaryURL    string
		skipPrepare  bool
		callTwice    bool
		execStdout   string
		execStderr   string
		execCode     int
		execErr      error
		validateFunc func(err error, capturedCmds [][]string)
	}{
		{
			name:      "downloads binary from custom URL",
			binaryURL: "https://example.com/osapi-linux",
			validateFunc: func(err error, capturedCmds [][]string) {
				assert.NoError(suite.T(), err)
				assert.Len(suite.T(), capturedCmds, 1)
				assert.Equal(suite.T(), "sh", capturedCmds[0][0])
				assert.Equal(suite.T(), "-c", capturedCmds[0][1])
				assert.Contains(suite.T(), capturedCmds[0][2], "https://example.com/osapi-linux")
				assert.Contains(suite.T(), capturedCmds[0][2], "curl")
				assert.Contains(suite.T(), capturedCmds[0][2], "chmod +x /osapi")
			},
		},
		{
			name:       "returns error on non-zero exit",
			binaryURL:  "https://example.com/osapi-linux",
			execStderr: "curl: (22) 404 Not Found",
			execCode:   22,
			validateFunc: func(err error, _ [][]string) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "deploy osapi binary (exit 22)")
				assert.Contains(suite.T(), err.Error(), "404 Not Found")
			},
		},
		{
			name:      "returns error on exec failure",
			binaryURL: "https://example.com/osapi-linux",
			execErr:   assert.AnError,
			validateFunc: func(err error, _ [][]string) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "deploy osapi binary")
			},
		},
		{
			name:        "skips preparation when configured",
			skipPrepare: true,
			validateFunc: func(err error, capturedCmds [][]string) {
				assert.NoError(suite.T(), err)
				assert.Empty(suite.T(), capturedCmds)
			},
		},
		{
			name:      "executes preparation only once across multiple calls",
			binaryURL: "https://example.com/osapi",
			callTwice: true,
			validateFunc: func(err error, capturedCmds [][]string) {
				assert.NoError(suite.T(), err)
				assert.Len(suite.T(), capturedCmds, 1)
			},
		},
		{
			name: "returns error when GitHub release resolution fails",
			validateFunc: func(err error, _ [][]string) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "resolve osapi binary")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var capturedCmds [][]string

			execFn := func(
				_ context.Context,
				_ string,
				cmd []string,
			) (string, string, int, error) {
				capturedCmds = append(capturedCmds, cmd)
				if tc.execErr != nil {
					return "", "", -1, tc.execErr
				}

				return tc.execStdout, tc.execStderr, tc.execCode, nil
			}

			target := orchestrator.NewDockerTarget("web", "ubuntu:24.04", execFn)
			if tc.binaryURL != "" {
				target.SetBinaryURL(tc.binaryURL)
			}
			if tc.skipPrepare {
				target.SetSkipPrepare(true)
			}

			if tc.callTwice {
				_ = target.Prepare(context.Background())
			}

			err := target.Prepare(context.Background())
			tc.validateFunc(err, capturedCmds)
		})
	}
}

func (suite *DeployPublicTestSuite) TestExecProvider() {
	tests := []struct {
		name         string
		binaryURL    string
		validateFunc func(result []byte, err error, cmds [][]string)
	}{
		{
			name: "returns prepare error",
			validateFunc: func(result []byte, err error, _ [][]string) {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
				assert.Contains(suite.T(), err.Error(), "resolve osapi binary")
			},
		},
		{
			name:      "triggers automatic preparation",
			binaryURL: "https://example.com/osapi",
			validateFunc: func(_ []byte, err error, cmds [][]string) {
				assert.NoError(suite.T(), err)
				// First call is the deploy script, second is the provider command.
				assert.Len(suite.T(), cmds, 2)
				assert.Equal(suite.T(), "sh", cmds[0][0])
				assert.Contains(suite.T(), cmds[0][2], "curl")
				assert.Equal(suite.T(), "/osapi", cmds[1][0])
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var cmds [][]string

			execFn := func(
				_ context.Context,
				_ string,
				cmd []string,
			) (string, string, int, error) {
				cmds = append(cmds, cmd)

				return `"ok"`, "", 0, nil
			}

			target := orchestrator.NewDockerTarget("web", "ubuntu:24.04", execFn)
			if tc.binaryURL != "" {
				target.SetBinaryURL(tc.binaryURL)
			}

			result, err := target.ExecProvider(
				context.Background(),
				"host",
				"get-hostname",
				nil,
			)
			tc.validateFunc(result, err, cmds)
		})
	}
}
