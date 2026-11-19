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
package command_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// update rewrites the scaffold's golden files from the current scaffold:
//
//	go test ./internal/cli/command -run TestScaffoldPublicTestSuite/TestGolden -update
var update = flag.Bool("update", false, "rewrite the scaffold's golden files")

const allGo = `// Package all registers every local group, one blank import each.
package all

import (
	// Each import registers its group.
	_ "github.com/osapi-io/osapi/internal/cli/commands/token"
)
`

type ScaffoldPublicTestSuite struct {
	suite.Suite
}

func (s *ScaffoldPublicTestSuite) TearDownSubTest() {
	command.ResetScaffold()
}

// checkout makes a minimal osapi checkout: its go.mod and the registry.
func (s *ScaffoldPublicTestSuite) checkout(
	all string,
) string {
	root := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module github.com/osapi-io/osapi\n\ngo 1.25\n"), 0o644))
	dir := filepath.Join(root, "internal", "cli", "commands", "all")
	s.Require().NoError(os.MkdirAll(dir, 0o755))
	if all != "" {
		s.Require().NoError(os.WriteFile(filepath.Join(dir, "all.go"), []byte(all), 0o644))
	}

	return root
}

// TestGolden compares each file the scaffold writes with its copy under
// testdata/scaffold, so a change to the scaffold shows in the diff of the
// files a person gets.
func (s *ScaffoldPublicTestSuite) TestGolden() {
	root := s.checkout(allGo)

	files, err := command.Scaffold(root, "widget", "Widgets of this host")
	s.Require().NoError(err)

	for _, rel := range files {
		got, err := os.ReadFile(filepath.Join(root, rel))
		s.Require().NoError(err)
		golden := filepath.Join("testdata", "scaffold", rel+".golden")
		if *update {
			s.Require().NoError(os.MkdirAll(filepath.Dir(golden), 0o755))
			s.Require().NoError(os.WriteFile(golden, got, 0o644))
		}
		want, err := os.ReadFile(golden)
		s.Require().NoError(err, "no golden file; run with -update")
		s.Equal(string(want), string(got), rel)
	}
}

func (s *ScaffoldPublicTestSuite) TestScaffoldErrors() {
	tests := []struct {
		name    string
		setup   func() string
		gname   string
		doc     string
		wantErr string
	}{
		{
			name:    "when the name is not a Go package name",
			setup:   func() string { return s.checkout(allGo) },
			gname:   "Wid-get",
			doc:     "x",
			wantErr: "must be lowercase letters and digits",
		},
		{
			name:    "when the description is empty",
			setup:   func() string { return s.checkout(allGo) },
			gname:   "widget",
			wantErr: "description must be one line",
		},
		{
			name:    "when the description has two lines",
			setup:   func() string { return s.checkout(allGo) },
			gname:   "widget",
			doc:     "one\ntwo",
			wantErr: "description must be one line",
		},
		{
			name:    "when the root is not an osapi checkout",
			setup:   func() string { return s.T().TempDir() },
			gname:   "widget",
			doc:     "x",
			wantErr: "is not an osapi checkout",
		},
		{
			name: "when the group exists already",
			setup: func() string {
				root := s.checkout(allGo)
				s.Require().NoError(os.MkdirAll(
					filepath.Join(root, "internal", "cli", "commands", "widget"), 0o755))

				return root
			},
			gname:   "widget",
			doc:     "x",
			wantErr: "already exists",
		},
		{
			name:    "when the registry is missing",
			setup:   func() string { return s.checkout("") },
			gname:   "widget",
			doc:     "x",
			wantErr: "no such file",
		},
		{
			name:    "when the registry has no import block",
			setup:   func() string { return s.checkout("package all\n") },
			gname:   "widget",
			doc:     "x",
			wantErr: "has no import block",
		},
		{
			name: "when a generated file is not Go",
			setup: func() string {
				command.BreakScaffold()

				return s.checkout(allGo)
			},
			gname:   "widget",
			doc:     "x",
			wantErr: "expected",
		},
		{
			name: "when a file cannot be written",
			setup: func() string {
				root := s.checkout(allGo)
				s.Require().
					NoError(os.Chmod(filepath.Join(root, "internal", "cli", "commands"), 0o555))
				s.T().Cleanup(func() {
					_ = os.Chmod(filepath.Join(root, "internal", "cli", "commands"), 0o755)
				})

				return root
			},
			gname:   "widget",
			doc:     "x",
			wantErr: "permission denied",
		},
		{
			name: "when the registry cannot be written",
			setup: func() string {
				root := s.checkout(allGo)
				s.Require().NoError(os.Chmod(
					filepath.Join(
						root,
						"internal",
						"cli",
						"commands",
						"all",
						"all.go",
					),
					0o444,
				))

				return root
			},
			gname:   "widget",
			doc:     "x",
			wantErr: "permission denied",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			root := tc.setup()

			_, err := command.Scaffold(root, tc.gname, tc.doc)

			s.ErrorContains(err, tc.wantErr)
		})
	}
}

func TestScaffoldPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ScaffoldPublicTestSuite))
}
