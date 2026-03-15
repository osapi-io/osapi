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
	"errors"
	"testing"

	gopsutil "github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/suite"
)

type ProcessTestSuite struct {
	suite.Suite
}

func (suite *ProcessTestSuite) TestGetMetrics() {
	tests := []struct {
		name         string
		setup        func()
		teardown     func()
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
		{
			name: "returns error when CPUPercent fails",
			pid:  0,
			setup: func() {
				newProcessFn = func(_ int32) (*gopsutil.Process, error) {
					return &gopsutil.Process{}, nil
				}
				cpuPercentFn = func(_ *gopsutil.Process) (float64, error) {
					return 0, errors.New("cpu error")
				}
			},
			teardown: func() {
				newProcessFn = gopsutil.NewProcess
				cpuPercentFn = func(
					proc *gopsutil.Process,
				) (float64, error) {
					return proc.CPUPercent()
				}
			},
			validateFunc: func(got *Metrics, err error) {
				suite.Nil(got)
				suite.Error(err)
				suite.Contains(err.Error(), "get cpu percent")
			},
		},
		{
			name: "returns error when MemoryInfo fails",
			pid:  0,
			setup: func() {
				newProcessFn = func(_ int32) (*gopsutil.Process, error) {
					return &gopsutil.Process{}, nil
				}
				cpuPercentFn = func(_ *gopsutil.Process) (float64, error) {
					return 1.5, nil
				}
				memoryInfoFn = func(_ *gopsutil.Process) (uint64, error) {
					return 0, errors.New("memory error")
				}
			},
			teardown: func() {
				newProcessFn = gopsutil.NewProcess
				cpuPercentFn = func(
					proc *gopsutil.Process,
				) (float64, error) {
					return proc.CPUPercent()
				}
				memoryInfoFn = func(
					proc *gopsutil.Process,
				) (uint64, error) {
					info, err := proc.MemoryInfo()
					if err != nil {
						return 0, err
					}

					return info.RSS, nil
				}
			},
			validateFunc: func(got *Metrics, err error) {
				suite.Nil(got)
				suite.Error(err)
				suite.Contains(err.Error(), "get memory info")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}
			if tc.teardown != nil {
				defer tc.teardown()
			}

			p := &provider{pid: tc.pid}
			got, err := p.GetMetrics()
			tc.validateFunc(got, err)
		})
	}
}

func TestProcessTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessTestSuite))
}
