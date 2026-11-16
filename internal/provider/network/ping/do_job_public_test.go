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

package ping_test

import (
	"context"
	"testing"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/provider/network/ping"
)

// blockedPinger is a ping that does not end until the test releases it, so a
// job can end while it runs. It is not a gomock double, because the ping goes
// on after Do returns, past the end of the test.
type blockedPinger struct {
	release chan struct{}
}

func (p *blockedPinger) SetPrivileged(bool) {}

func (p *blockedPinger) SetCount(int) {}

func (p *blockedPinger) SetTimeout(time.Duration) {}

func (p *blockedPinger) Run() error {
	<-p.release

	return nil
}

func (p *blockedPinger) Statistics() *probing.Statistics {
	return &probing.Statistics{}
}

type DoJobPublicTestSuite struct {
	suite.Suite
}

func (s *DoJobPublicTestSuite) TestDoStopsWhenTheJobEnds() {
	tests := []struct {
		name string
		do   func(pinger ping.Pinger) func(context.Context, string) (*ping.Result, error)
	}{
		{
			name: "when the job ends on debian it stops, and says so",
			do: func(pinger ping.Pinger) func(context.Context, string) (*ping.Result, error) {
				p := ping.NewDebianProvider()
				p.NewPingerFn = func(string) (ping.Pinger, error) { return pinger, nil }

				return p.Do
			},
		},
		{
			name: "when the job ends on darwin it stops, and says so",
			do: func(pinger ping.Pinger) func(context.Context, string) (*ping.Result, error) {
				p := ping.NewDarwinProvider()
				p.NewPingerFn = func(string) (ping.Pinger, error) { return pinger, nil }

				return p.Do
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			pinger := &blockedPinger{release: make(chan struct{})}
			defer close(pinger.release)

			job, cancel := context.WithCancel(context.Background())
			cancel()

			result, err := tc.do(pinger)(job, "1.1.1.1")

			s.Nil(result)
			s.ErrorIs(err, context.Canceled)
			s.Contains(err.Error(), "ping stopped, the job ended")
		})
	}
}

func TestDoJobPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DoJobPublicTestSuite))
}
