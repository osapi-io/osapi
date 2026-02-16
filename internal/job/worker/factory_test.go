// Copyright (c) 2025 John Dewey

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

package worker

import (
	"log/slog"
	"testing"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/stretchr/testify/suite"
)

type FactoryTestSuite struct {
	suite.Suite
}

func (s *FactoryTestSuite) TestCreateProvidersUbuntuPlatform() {
	tests := []struct {
		name string
	}{
		{
			name: "creates ubuntu providers when platform is ubuntu",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			original := factoryHostInfoFn
			defer func() { factoryHostInfoFn = original }()

			factoryHostInfoFn = func() (*host.InfoStat, error) {
				return &host.InfoStat{
					Platform: "Ubuntu",
				}, nil
			}

			factory := NewProviderFactory(slog.Default())
			hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider := factory.CreateProviders()

			s.NotNil(hostProvider)
			s.NotNil(diskProvider)
			s.NotNil(memProvider)
			s.NotNil(loadProvider)
			s.NotNil(dnsProvider)
			s.NotNil(pingProvider)
		})
	}
}

func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}
