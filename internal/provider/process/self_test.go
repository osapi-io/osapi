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

package process

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProcessTestSuite struct {
	suite.Suite
}

func (suite *ProcessTestSuite) TestGetMetricsInvalidPID() {
	tests := []struct {
		name         string
		pid          int32
		validateFunc func(got *Metrics, err error)
	}{
		{
			name: "returns error for invalid pid",
			pid:  -99999,
			validateFunc: func(got *Metrics, err error) {
				suite.Nil(got)
				suite.Error(err)
				suite.Contains(err.Error(), "get process")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			p := &provider{pid: tc.pid}

			got, err := p.GetMetrics()

			tc.validateFunc(got, err)
		})
	}
}

func TestProcessTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessTestSuite))
}
