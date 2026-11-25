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

package sdk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/sdk"
	"github.com/osapi-io/osapi/internal/sdk/sdkspec"
	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

// deps is a small SDK: one type and one service with one method.
func (s *MethodsPublicTestSuite) deps(
	ts []types.TypeSpec,
	ss []client.ServiceSpec,
) sdk.Deps {
	return sdk.Deps{
		Types:    func() []types.TypeSpec { return ts },
		Services: func() []client.ServiceSpec { return ss },
	}
}

var (
	shade = types.TypeSpec{
		Name:   "Shade",
		Doc:    "Shade is a tone.",
		Values: []string{"light", "dark"},
	}
	network = client.ServiceSpec{
		Name: "Hostname",
		Doc:  "The hostname.",
		Methods: []client.MethodSpec{
			{Name: "Get", Doc: "Get it", HTTP: "GET /api/node/{hostname}/hostname", Examples: 1},
		},
	}
)

func (s *MethodsPublicTestSuite) TestList() {
	got, err := sdk.List(
		context.Background(),
		s.deps([]types.TypeSpec{shade}, []client.ServiceSpec{network}),
		sdk.ListInput{},
	)
	s.Require().NoError(err)
	s.Equal([]sdk.Item{
		{Kind: sdk.KindType, Name: "Shade", Count: 2, Doc: "Shade is a tone."},
		{Kind: sdk.KindService, Name: "Hostname", Count: 1, Doc: "The hostname."},
	}, got.Items)
}

func (s *MethodsPublicTestSuite) TestGet() {
	d := s.deps([]types.TypeSpec{shade}, []client.ServiceSpec{network})
	tests := []struct {
		name         string
		in           string
		validateFunc func(sdk.GetOutput, error)
	}{
		{
			name: "when the name is a type, it is the type",
			in:   "Shade",
			validateFunc: func(got sdk.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("Shade", got.Type.Name)
				s.Nil(got.Service)
			},
		},
		{
			name: "when the name has the package, it is the type",
			in:   "types.Shade",
			validateFunc: func(got sdk.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("Shade", got.Type.Name)
			},
		},
		{
			name: "when the name is a service, it is the service",
			in:   "Hostname",
			validateFunc: func(got sdk.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("Hostname", got.Service.Name)
			},
		},
		{
			name: "when nothing has the name, it says where to look",
			in:   "nope",
			validateFunc: func(_ sdk.GetOutput, err error) {
				s.ErrorContains(err, "osapi sdk list shows them")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(sdk.Get(context.Background(), d, sdk.GetInput{Name: tc.in}))
		})
	}
}

func (s *MethodsPublicTestSuite) TestSpec() {
	ts, ss := []types.TypeSpec{shade}, []client.ServiceSpec{network}
	got, err := sdk.Spec(context.Background(), s.deps(ts, ss), sdk.SpecInput{})
	s.Require().NoError(err)
	s.Equal(sdkspec.Spec(ts, ss).Sections(), got.Kinds)
}

func (s *MethodsPublicTestSuite) TestValidate() {
	tests := []struct {
		name         string
		ts           []types.TypeSpec
		validateFunc func(sdk.ValidateOutput, error)
	}{
		{
			name: "when nothing breaks a rule, it is valid",
			ts:   []types.TypeSpec{shade},
			validateFunc: func(got sdk.ValidateOutput, err error) {
				s.Require().NoError(err)
				s.True(got.Valid)
				s.Equal(1, got.Types)
			},
		},
		{
			name: "when a type breaks a rule, it fails with the fix",
			ts:   []types.TypeSpec{{Name: "Bare"}},
			validateFunc: func(got sdk.ValidateOutput, err error) {
				s.ErrorIs(err, command.FailedError{})
				s.False(got.Valid)
				s.Contains(got.Fixes, sdkspec.RuleValues)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(
				sdk.Validate(
					context.Background(),
					s.deps(tc.ts, []client.ServiceSpec{network}),
					sdk.ValidateInput{},
				),
			)
		})
	}
}

// TestKind lists the kinds of a row.
func (s *MethodsPublicTestSuite) TestKind() {
	s.True(sdk.KindService.Valid())
	s.False(sdk.Kind("package").Valid())
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
