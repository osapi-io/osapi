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

package clidocs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/clidocs"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestGenerate() {
	root := &cobra.Command{Use: "osapi", Short: "The osapi CLI"}
	root.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Show the version",
			Run:   func(*cobra.Command, []string) {},
		},
	)
	d := clidocs.Deps{Root: root}

	dir := s.T().TempDir()
	got, err := clidocs.Generate(context.Background(), d, clidocs.GenerateInput{Dir: dir})
	s.Require().NoError(err)
	s.Equal(dir, got.Dir)
	s.Positive(got.Pages)

	wd, _ := os.Getwd()
	work := s.T().TempDir()
	s.Require().NoError(os.Chdir(work))
	defer func() { _ = os.Chdir(wd) }()
	got, err = clidocs.Generate(context.Background(), d, clidocs.GenerateInput{})
	s.Require().NoError(err, "with no dir, it writes to the docs site's")
	s.Equal("docs/docs/sidebar/usage/cli", got.Dir)

	file := filepath.Join(work, "file")
	s.Require().NoError(os.WriteFile(file, nil, 0o644))
	_, err = clidocs.Generate(
		context.Background(),
		d,
		clidocs.GenerateInput{Dir: filepath.Join(file, "pages")},
	)
	s.ErrorContains(err, "write the CLI reference into")
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
