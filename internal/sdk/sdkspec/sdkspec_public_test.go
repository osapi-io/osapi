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

package sdkspec_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/sdk/sdkspec"
	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type SDKSpecPublicTestSuite struct {
	suite.Suite
}

// TestSpec has the packages, the calls, a section of the types and one of
// the services, each service with its methods.
func (s *SDKSpecPublicTestSuite) TestSpec() {
	spec := sdkspec.Spec(
		[]types.TypeSpec{{Name: "Shade", Doc: "Shade is a tone.", Values: []string{"light"}}},
		[]client.ServiceSpec{
			{Name: "Hostname", Doc: "The hostname.", Methods: []client.MethodSpec{{Name: "Get"}}},
		},
	)

	s.Len(spec.Of(sdkspec.KindPackage), 2)
	s.Len(spec.Of(sdkspec.KindCall), 2)
	s.Len(spec.Of("type"), 1)
	services := spec.Of(sdkspec.KindService)
	s.Require().Len(services, 1)
	s.Equal("c.Hostname", services[0].Name)
	s.Equal("The hostname. Methods: Get.", services[0].Does)
}

// TestValidate reports each rule that the SDK breaks, and nothing for the
// SDK of this build.
func (s *SDKSpecPublicTestSuite) TestValidate() {
	tests := []struct {
		name  string
		ts    []types.TypeSpec
		ss    []client.ServiceSpec
		rules []string
	}{
		{
			name: "when it is the SDK of this build, nothing breaks a rule",
			ts:   types.All(),
			ss:   client.Services(),
		},
		{
			name:  "when a type has no doc and no values, both are reported",
			ts:    []types.TypeSpec{{Name: "Bare"}},
			rules: []string{sdkspec.RuleDoc, sdkspec.RuleValues},
		},
		{
			name: "when a service has no doc and is not on the client, both are reported",
			ss: []client.ServiceSpec{{Name: "Nope", Methods: []client.MethodSpec{
				{Name: "Get", Doc: "Get it", Examples: 1},
			}}},
			rules: []string{sdkspec.RuleDoc, sdkspec.RuleClient},
		},
		{
			name: "when a method has no doc, no example and is not on the service, each is reported",
			ss: []client.ServiceSpec{
				{Name: "Hostname", Doc: "The hostname.", Methods: []client.MethodSpec{
					{Name: "Rename"},
				}},
			},
			rules: []string{sdkspec.RuleDoc, sdkspec.RuleExample, sdkspec.RuleClient},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			var got []string
			for _, p := range sdkspec.Validate(tc.ts, tc.ss) {
				got = append(got, p.Rule)
				s.NotEmpty(p.Fix, "each problem has its fix")
			}
			s.Equal(tc.rules, got)
		})
	}
}

func TestSDKSpecPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SDKSpecPublicTestSuite))
}
