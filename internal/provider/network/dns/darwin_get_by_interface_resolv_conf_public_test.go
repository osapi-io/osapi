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

package dns_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/network/dns"
)

type DarwinGetResolvConfByInterfacePublicTestSuite struct {
	suite.Suite
}

func (suite *DarwinGetResolvConfByInterfacePublicTestSuite) SetupTest() {}

func (suite *DarwinGetResolvConfByInterfacePublicTestSuite) TearDownTest() {}

func (suite *DarwinGetResolvConfByInterfacePublicTestSuite) TestGetResolvConfByInterface() {
	tests := []struct {
		name string
		want *dns.Config
	}{
		{
			name: "when GetResolvConfByInterface returns mock data",
			want: &dns.Config{
				DNSServers:    []string{"8.8.8.8", "1.1.1.1"},
				SearchDomains: []string{"local", "example.com"},
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			darwin := dns.NewDarwinProvider()

			got, err := darwin.GetResolvConfByInterface("en0")

			suite.NoError(err)
			suite.NotNil(got)
			suite.Equal(tc.want, got)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestDarwinGetResolvConfByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(DarwinGetResolvConfByInterfacePublicTestSuite))
}
