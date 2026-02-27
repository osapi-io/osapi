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
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
)

type FactoryPublicTestSuite struct {
	suite.Suite
}

func (s *FactoryPublicTestSuite) TestNewProviderFactory() {
	tests := []struct {
		name string
	}{
		{
			name: "creates factory with logger",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			factory := agent.NewProviderFactory(slog.Default())

			s.NotNil(factory)
		})
	}
}

func (s *FactoryPublicTestSuite) TestCreateProviders() {
	tests := []struct {
		name string
	}{
		{
			name: "creates all providers",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			factory := agent.NewProviderFactory(slog.Default())

			hostProvider, diskProvider, memProvider, loadProvider, dnsProvider, pingProvider, commandProvider := factory.CreateProviders()

			s.NotNil(hostProvider)
			s.NotNil(diskProvider)
			s.NotNil(memProvider)
			s.NotNil(loadProvider)
			s.NotNil(dnsProvider)
			s.NotNil(pingProvider)
			s.NotNil(commandProvider)
		})
	}
}

func TestFactoryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryPublicTestSuite))
}
