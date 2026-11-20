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
package all_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
)

// repoRoot is the repository root, from this package's directory.
const repoRoot = "../../.."

type SpecPublicTestSuite struct {
	suite.Suite
}

// TestSpec loads every spec the behaviors' endpoint groups produced, each
// behavior's gen/api.yaml, with kin-openapi, the library
// oapi-codegen reads specs with, and validates it, references to the shared
// spec included.
func (s *SpecPublicTestSuite) TestSpec() {
	var specs []string
	for _, pattern := range []string{
		"internal/behaviors/*/gen/api*.yaml",
		"internal/behaviors/*/*/gen/api*.yaml",
	} {
		found, err := filepath.Glob(filepath.Join(repoRoot, pattern))
		s.Require().NoError(err)
		specs = append(specs, found...)
	}
	s.Require().NotEmpty(specs)

	for _, spec := range specs {
		s.Run(spec, func() {
			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			doc, err := loader.LoadFromFile(spec)
			s.Require().NoError(err)
			s.NoError(doc.Validate(context.Background()))
		})
	}
}

func TestSpecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecPublicTestSuite))
}
