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

package registry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/provider/registry"
)

type RegistryPublicTestSuite struct {
	suite.Suite
}

func (suite *RegistryPublicTestSuite) SetupTest() {}

func (suite *RegistryPublicTestSuite) TearDownTest() {}

func (suite *RegistryPublicTestSuite) TestLookup() {
	tests := []struct {
		name         string
		register     bool
		provider     string
		operation    string
		validateFunc func(spec *registry.OperationSpec, found bool)
	}{
		{
			name:      "when registered provider found",
			register:  true,
			provider:  "container",
			operation: "create",
			validateFunc: func(spec *registry.OperationSpec, found bool) {
				suite.True(found)
				suite.NotNil(spec)
				suite.NotNil(spec.NewParams)
				suite.NotNil(spec.Run)
			},
		},
		{
			name:      "when unregistered provider not found",
			register:  false,
			provider:  "nonexistent",
			operation: "create",
			validateFunc: func(spec *registry.OperationSpec, found bool) {
				suite.False(found)
				suite.Nil(spec)
			},
		},
		{
			name:      "when registered provider wrong operation",
			register:  true,
			provider:  "container",
			operation: "nonexistent",
			validateFunc: func(spec *registry.OperationSpec, found bool) {
				suite.False(found)
				suite.Nil(spec)
			},
		},
		{
			name:      "when run returns expected result",
			register:  true,
			provider:  "container",
			operation: "create",
			validateFunc: func(spec *registry.OperationSpec, found bool) {
				suite.Require().True(found)
				suite.Require().NotNil(spec)

				params := spec.NewParams()
				suite.Equal("default-params", params)

				result, err := spec.Run(context.Background(), params)
				suite.NoError(err)
				suite.Equal("created", result)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			r := registry.New()

			if tc.register {
				r.Register(registry.Registration{
					Name: "container",
					Operations: map[string]registry.OperationSpec{
						"create": {
							NewParams: func() any {
								return "default-params"
							},
							Run: func(
								_ context.Context,
								_ any,
							) (any, error) {
								return "created", nil
							},
						},
					},
				})
			}

			spec, found := r.Lookup(tc.provider, tc.operation)
			tc.validateFunc(spec, found)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestRegistryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryPublicTestSuite))
}
