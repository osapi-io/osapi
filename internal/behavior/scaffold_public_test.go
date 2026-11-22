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

package behavior_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

const allGo = `// Package all registers every behavior.
package all

import (
	// Each import registers its behavior.
	_ "github.com/osapi-io/osapi/internal/behaviors/sysctl"
)
`

// update rewrites the scaffold's golden files from the current scaffold:
//
//	go test ./internal/behavior -run TestScaffoldPublicTestSuite/TestGolden -update
var update = flag.Bool("update", false, "rewrite the scaffold's golden files")

type ScaffoldPublicTestSuite struct {
	suite.Suite
}

func (s *ScaffoldPublicTestSuite) TearDownSubTest() {
	behavior.ResetScaffold()
}

// checkout makes a minimal osapi checkout: its go.mod and internal/behaviors/all.
func (s *ScaffoldPublicTestSuite) checkout(
	all string,
) string {
	root := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module github.com/osapi-io/osapi\n\ngo 1.25\n"), 0o644))
	s.Require().NoError(os.MkdirAll(filepath.Join(root, "internal", "behaviors", "all"), 0o755))
	s.Require().NoError(os.WriteFile(
		filepath.Join(root, "internal", "behaviors", "all", "all.go"), []byte(all), 0o644))

	return root
}

// template is a scaffold of one template, for the cases a template is wrong.
func (s *ScaffoldPublicTestSuite) TestScaffold() {
	tests := []struct {
		name         string
		setup        func() string
		group        string
		bname        string
		goPackage    string
		doc          string
		wantErr      string
		validateFunc func(root string, files []string)
	}{
		{
			name:    "when the name is a Go keyword and no package is named",
			setup:   func() string { return s.checkout(allGo) },
			bname:   "package",
			doc:     "The packages on the host",
			wantErr: `"package" is a Go keyword, so it cannot be the Go package: pass --package`,
		},
		{
			name:      "when the package is not a Go package name",
			setup:     func() string { return s.checkout(allGo) },
			bname:     "package",
			goPackage: "pkg-s",
			doc:       "The packages on the host",
			wantErr:   `package "pkg-s" must be lowercase letters and digits`,
		},
		{
			name:      "when a package is named, the code lives there and the name stays",
			setup:     func() string { return s.checkout(allGo) },
			bname:     "package",
			goPackage: "packages",
			doc:       "The packages on the host",
			validateFunc: func(root string, files []string) {
				s.Contains(files, "internal/behaviors/packages/behavior.go")
				def, err := os.ReadFile(
					filepath.Join(root, "internal/behaviors/packages/behavior.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(def), "package packages")
				s.Contains(string(def), `"package", "The packages on the host",`)
				test, err := os.ReadFile(
					filepath.Join(root, "internal/behaviors/packages/behavior_public_test.go"))
				s.Require().NoError(err)
				s.Contains(string(test), "package packages_test")
				all, err := os.ReadFile(filepath.Join(root, "internal/behaviors/all/all.go"))
				s.Require().NoError(err)
				s.Contains(string(all), `"github.com/osapi-io/osapi/internal/behaviors/packages"`)
			},
		},
		{
			name:  "when the name is free, it writes the layout and registers it",
			setup: func() string { return s.checkout(allGo) },
			bname: "motd",
			doc:   `The "message" shown at login`,
			validateFunc: func(root string, files []string) {
				s.Equal([]string{
					"internal/behaviors/motd/behavior.go",
					"internal/behaviors/motd/behavior_public_test.go",
					"internal/behaviors/motd/debian.go",
					"internal/behaviors/motd/debian_public_test.go",
					"internal/behaviors/motd/methods.go",
					"internal/behaviors/motd/methods_public_test.go",
					"internal/behaviors/motd/mocks/generate.go",
					"internal/behaviors/motd/provider.go",
					"internal/behaviors/motd/settings.go",
					"internal/behaviors/motd/types.go",
					"internal/behaviors/all/all.go",
				}, files)

				def, err := os.ReadFile(filepath.Join(root, "internal/behaviors/motd/behavior.go"))
				s.Require().NoError(err)
				s.Contains(string(def), `"motd", "The \"message\" shown at login",`)
				s.Contains(string(def), "package motd")

				settings, err := os.ReadFile(
					filepath.Join(root, "internal/behaviors/motd/settings.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(settings), "OSAPI_BEHAVIORS_MOTD_<FIELD>")

				all, err := os.ReadFile(filepath.Join(root, "internal/behaviors/all/all.go"))
				s.Require().NoError(err)
				s.Contains(string(all), "import (\n"+
					"\t// Each import registers its behavior.\n"+
					"\t_ \"github.com/osapi-io/osapi/internal/behaviors/motd\"\n"+
					"\t_ \"github.com/osapi-io/osapi/internal/behaviors/sysctl\"\n)")
			},
		},
		{
			name:  "when a group is given, it nests the behavior under it",
			setup: func() string { return s.checkout(allGo) },
			group: "network",
			bname: "neighbors",
			doc:   "The host's ARP neighbors",
			validateFunc: func(root string, files []string) {
				s.Contains(files, "internal/behaviors/network/neighbors/behavior.go")

				def, err := os.ReadFile(
					filepath.Join(root, "internal/behaviors/network/neighbors/behavior.go"))
				s.Require().NoError(err)
				s.Contains(string(def), `behavior.Under(behavior.GroupNetwork),`)

				test, err := os.ReadFile(
					filepath.Join(
						root,
						"internal/behaviors/network/neighbors/behavior_public_test.go",
					),
				)
				s.Require().NoError(err)
				s.Contains(
					string(test),
					`"github.com/osapi-io/osapi/internal/behaviors/network/neighbors"`,
				)

				all, err := os.ReadFile(filepath.Join(root, "internal/behaviors/all/all.go"))
				s.Require().NoError(err)
				s.Contains(
					string(all),
					`_ "github.com/osapi-io/osapi/internal/behaviors/network/neighbors"`,
				)
			},
		},
		{
			name:    "when the group is not a Go package name",
			setup:   func() string { return s.checkout(allGo) },
			group:   "Net-work",
			bname:   "neighbors",
			doc:     "x",
			wantErr: "group \"Net-work\" must be lowercase",
		},
		{
			name:    "when the group is not one of KnownGroups",
			setup:   func() string { return s.checkout(allGo) },
			group:   "storage",
			bname:   "neighbors",
			doc:     "x",
			wantErr: "group \"storage\" is not one of KnownGroups",
		},
		{
			name:    "when the name is not a Go package name",
			setup:   func() string { return s.checkout(allGo) },
			bname:   "Mo-td",
			doc:     "x",
			wantErr: "lowercase letters and digits",
		},
		{
			name:    "when the description is empty",
			setup:   func() string { return s.checkout(allGo) },
			bname:   "motd",
			wantErr: "description must be one line",
		},
		{
			name:    "when the description is two lines",
			setup:   func() string { return s.checkout(allGo) },
			bname:   "motd",
			doc:     "one\ntwo",
			wantErr: "description must be one line",
		},
		{
			name:    "when root has no go.mod",
			setup:   func() string { return s.T().TempDir() },
			bname:   "motd",
			doc:     "x",
			wantErr: "is not an osapi checkout",
		},
		{
			name: "when root is another module",
			setup: func() string {
				root := s.checkout(allGo)
				s.Require().NoError(os.WriteFile(filepath.Join(root, "go.mod"),
					[]byte("module example.com/other\n"), 0o644))

				return root
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "is not an osapi checkout",
		},
		{
			name:    "when the behavior exists",
			setup:   func() string { return s.checkout(allGo) },
			bname:   "all",
			doc:     "x",
			wantErr: "internal/behaviors/all already exists",
		},
		{
			name: "when internal/behaviors/all is missing",
			setup: func() string {
				root := s.checkout(allGo)
				s.Require().
					NoError(os.RemoveAll(filepath.Join(root, "internal", "behaviors", "all")))

				return root
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "no such file",
		},
		{
			name:    "when all.go has no import block",
			setup:   func() string { return s.checkout("package all\n") },
			bname:   "motd",
			doc:     "x",
			wantErr: "has no import block",
		},
		{
			name:    "when all.go does not parse",
			setup:   func() string { return s.checkout("package all\n\nimport (\n\tbroken\n)\n") },
			bname:   "motd",
			doc:     "x",
			wantErr: "missing import path",
		},
		{
			name: "when the behavior's directory cannot be made",
			setup: func() string {
				root := s.checkout(allGo)
				dir := filepath.Join(root, "internal", "behaviors")
				s.Require().NoError(os.Chmod(dir, 0o555))
				s.T().Cleanup(func() { _ = os.Chmod(dir, 0o755) })

				return root
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "permission denied",
		},
		{
			name: "when the registration cannot be written",
			setup: func() string {
				root := s.checkout(allGo)
				s.Require().NoError(
					os.Chmod(filepath.Join(root, "internal", "behaviors", "all", "all.go"), 0o444))

				return root
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "permission denied",
		},
		{
			name: "when a generated file is not Go",
			setup: func() string {
				behavior.BreakScaffold()

				return s.checkout(allGo)
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "expected",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			root := tc.setup()

			files, err := behavior.Scaffold(root, tc.group, tc.bname, tc.goPackage, tc.doc)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(root, files)
		})
	}
}

// TestGolden compares each file that the scaffold writes with its copy
// under testdata/scaffold. The scaffold builds its files with jen, so the
// copies are where a reader sees them, and a change to the scaffold shows in
// the diff of its golden files.
func (s *ScaffoldPublicTestSuite) TestGolden() {
	tests := []struct {
		name  string
		group string
		bname string
		doc   string
	}{
		{
			name:  "motd",
			bname: "motd",
			doc:   `The "message" shown at login`,
		},
		{
			name:  "neighbors",
			group: "network",
			bname: "neighbors",
			doc:   "The neighbor table",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			root := s.checkout(allGo)

			files, err := behavior.Scaffold(root, tc.group, tc.bname, "", tc.doc)
			s.Require().NoError(err)

			for _, rel := range files {
				got, err := os.ReadFile(filepath.Join(root, rel))
				s.Require().NoError(err)
				golden := filepath.Join("testdata", "scaffold", tc.name, rel+".golden")
				if *update {
					s.Require().NoError(os.MkdirAll(filepath.Dir(golden), 0o755))
					s.Require().NoError(os.WriteFile(golden, got, 0o644))
				}
				want, err := os.ReadFile(golden)
				s.Require().NoError(err, "no golden file; run with -update")
				s.Equal(string(want), string(got), rel)
			}
		})
	}
}

func TestScaffoldPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ScaffoldPublicTestSuite))
}
