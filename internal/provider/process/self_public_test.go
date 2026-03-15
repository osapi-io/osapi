// Copyright (c) 2024 John Dewey

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

package process_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/process"
)

type ProcessPublicTestSuite struct {
	suite.Suite
}

func (suite *ProcessPublicTestSuite) SetupTest() {}

func (suite *ProcessPublicTestSuite) TearDownTest() {}

func (suite *ProcessPublicTestSuite) TestGetMetrics() {
	tests := []struct {
		name         string
		validateFunc func(got *process.Metrics, err error)
	}{
		{
			name: "returns non-nil metrics",
			validateFunc: func(got *process.Metrics, err error) {
				suite.NoError(err)
				suite.NotNil(got)
			},
		},
		{
			name: "goroutines is positive",
			validateFunc: func(got *process.Metrics, err error) {
				suite.NoError(err)
				suite.Require().NotNil(got)
				suite.Greater(got.Goroutines, 0)
			},
		},
		{
			name: "rss is positive",
			validateFunc: func(got *process.Metrics, err error) {
				suite.NoError(err)
				suite.Require().NotNil(got)
				suite.Greater(got.RSSBytes, int64(0))
			},
		},
		{
			name: "cpu is non-negative",
			validateFunc: func(got *process.Metrics, err error) {
				suite.NoError(err)
				suite.Require().NotNil(got)
				suite.GreaterOrEqual(got.CPUPercent, float64(0))
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			p := process.New()

			got, err := p.GetMetrics()

			tc.validateFunc(got, err)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestProcessPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessPublicTestSuite))
}
