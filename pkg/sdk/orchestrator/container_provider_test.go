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

package orchestrator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ContainerProviderTestSuite struct {
	suite.Suite
}

func TestContainerProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerProviderTestSuite))
}

func (suite *ContainerProviderTestSuite) TestRunMarshalError() {
	tests := []struct {
		name         string
		params       any
		validateFunc func(err error)
	}{
		{
			name:   "returns error when params cannot be marshaled",
			params: make(chan int),
			validateFunc: func(err error) {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "marshal")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			target := &DockerTarget{
				name: "test",
				execFn: func(
					_ context.Context,
					_ string,
					_ []string,
				) (string, string, int, error) {
					return `{}`, "", 0, nil
				},
			}

			cp := NewContainerProvider(target)
			_, err := run[CommandResult](
				context.Background(),
				cp,
				"test",
				"op",
				tc.params,
			)
			tc.validateFunc(err)
		})
	}
}
