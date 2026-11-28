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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviors/managedfile"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
)

// testFS is the host's filesystem, through avfs. The generator under test writes real files, and the test reads them back.
var testFS avfs.VFS = osfs.NewWithNoIdm()

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

// SetupSubTest registers the behaviors a scaffold may require: managedfile,
// as a real behavior's package; chained, whose method lives in this test
// package; and empty, which has no method.
func (s *ScaffoldPublicTestSuite) SetupSubTest() {
	behavior.Register(managedfile.Behavior)
	behavior.Register(behavior.Define("chained", "x", behavior.Read("get", get)))
	behavior.Register(behavior.Define("empty", "x"))
}

func (s *ScaffoldPublicTestSuite) TearDownSubTest() {
	behavior.ResetScaffold()
}

// checkout makes a minimal osapi checkout: its go.mod and internal/behaviors/all.
func (s *ScaffoldPublicTestSuite) checkout(
	all string,
) string {
	root := s.T().TempDir()
	s.Require().NoError(avfs.WriteFile(testFS, filepath.Join(root, "go.mod"),
		[]byte("module github.com/osapi-io/osapi\n\ngo 1.25\n"), 0o644))
	s.Require().NoError(testFS.MkdirAll(filepath.Join(root, "internal", "behaviors", "all"), 0o755))
	s.Require().NoError(avfs.WriteFile(testFS,
		filepath.Join(root, "internal", "behaviors", "all", "all.go"), []byte(all), 0o644))

	return root
}

// template is a scaffold of one template, for the cases a template is wrong.
func (s *ScaffoldPublicTestSuite) TestScaffold() {
	tests := []struct {
		name         string
		setup        func() string
		bname        string
		goPackage    string
		doc          string
		requires     []string
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
				def, err := avfs.ReadFile(testFS,
					filepath.Join(root, "internal/behaviors/packages/behavior.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(def), "package packages")
				s.Contains(string(def), `"package", "The packages on the host",`)
				test, err := avfs.ReadFile(testFS,
					filepath.Join(root, "internal/behaviors/packages/behavior_public_test.go"))
				s.Require().NoError(err)
				s.Contains(string(test), "package packages_test")
				all, err := avfs.ReadFile(
					testFS,
					filepath.Join(root, "internal/behaviors/all/all.go"),
				)
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

				def, err := avfs.ReadFile(
					testFS,
					filepath.Join(root, "internal/behaviors/motd/behavior.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(def), `"motd", "The \"message\" shown at login",`)
				s.Contains(string(def), "package motd")

				settings, err := avfs.ReadFile(testFS,
					filepath.Join(root, "internal/behaviors/motd/settings.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(settings), "OSAPI_BEHAVIORS_MOTD_<FIELD>")

				all, err := avfs.ReadFile(
					testFS,
					filepath.Join(root, "internal/behaviors/all/all.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(all), "import (\n"+
					"\t// Each import registers its behavior.\n"+
					"\t_ \"github.com/osapi-io/osapi/internal/behaviors/motd\"\n"+
					"\t_ \"github.com/osapi-io/osapi/internal/behaviors/sysctl\"\n)")
			},
		},
		{
			name:     "when a behavior is required, it writes the wiring of each",
			setup:    func() string { return s.checkout(allGo) },
			bname:    "motd",
			doc:      "x",
			requires: []string{"managedfile", "chained"},
			validateFunc: func(root string, _ []string) {
				def, err := avfs.ReadFile(
					testFS,
					filepath.Join(root, "internal/behaviors/motd/behavior.go"),
				)
				s.Require().NoError(err)
				s.Contains(string(def),
					"behavior.Requires(managedfile.Behavior, behavior_test.Behavior),")
				debian, err := avfs.ReadFile(testFS,
					filepath.Join(root, "internal/behaviors/motd/debian.go"))
				s.Require().NoError(err)
				s.Contains(string(debian),
					"if managedfileProvider == nil ||\n\t\tbehavior_testProvider == nil {")
			},
		},
		{
			name:     "when a required behavior is not registered",
			setup:    func() string { return s.checkout(allGo) },
			bname:    "motd",
			doc:      "x",
			requires: []string{"nosuch"},
			wantErr: `requires "nosuch", which is not a registered behavior: ` +
				"osapi dsl behavior list names each one",
		},
		{
			name:     "when a required behavior is named twice",
			setup:    func() string { return s.checkout(allGo) },
			bname:    "motd",
			doc:      "x",
			requires: []string{"managedfile", "managedfile"},
			wantErr:  `requires "managedfile" twice`,
		},
		{
			name:     "when a required behavior has no method",
			setup:    func() string { return s.checkout(allGo) },
			bname:    "motd",
			doc:      "x",
			requires: []string{"empty"},
			wantErr:  `requires "empty", which has no method to call`,
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
				s.Require().NoError(avfs.WriteFile(testFS, filepath.Join(root, "go.mod"),
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
					NoError(testFS.RemoveAll(filepath.Join(root, "internal", "behaviors", "all")))

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
				s.Require().NoError(testFS.Chmod(dir, 0o555))
				s.T().Cleanup(func() { _ = testFS.Chmod(dir, 0o755) })

				return root
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "permission denied",
		},
		{
			name: "when the registration cannot be written, nothing is left",
			setup: func() string {
				root := s.checkout(allGo)
				// all.go is written through a temp file in its directory,
				// so the directory, not the file, is what refuses it.
				dir := filepath.Join(root, "internal", "behaviors", "all")
				s.Require().NoError(testFS.Chmod(dir, 0o555))
				s.T().Cleanup(func() { _ = testFS.Chmod(dir, 0o755) })

				return root
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "permission denied",
			validateFunc: func(root string, files []string) {
				s.Nil(files)
				s.NoDirExists(filepath.Join(root, "internal", "behaviors", "motd"),
					"the behavior's files are removed")
				all, err := avfs.ReadFile(testFS,
					filepath.Join(root, "internal", "behaviors", "all", "all.go"))
				s.Require().NoError(err)
				s.Equal(allGo, string(all), "all.go is as it was")
			},
		},
		{
			name:      "when the name is registered under another package",
			setup:     func() string { return s.checkout(allGo) },
			bname:     "chained",
			goPackage: "other",
			doc:       "x",
			wantErr:   `behavior "chained" is already registered`,
			validateFunc: func(root string, _ []string) {
				s.NoDirExists(filepath.Join(root, "internal", "behaviors", "other"))
			},
		},
		{
			name: "when all.go already imports the package",
			setup: func() string {
				return s.checkout(strings.Replace(allGo, "behaviors/sysctl", "behaviors/motd", 1))
			},
			bname:   "motd",
			doc:     "x",
			wantErr: "already imports",
			validateFunc: func(root string, _ []string) {
				s.NoDirExists(filepath.Join(root, "internal", "behaviors", "motd"))
			},
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

			files, err := behavior.Scaffold(
				root, tc.bname, tc.goPackage, tc.doc, tc.requires)

			if tc.wantErr != "" {
				s.ErrorContains(err, tc.wantErr)
				if tc.validateFunc != nil {
					tc.validateFunc(root, files)
				}

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
		name     string
		bname    string
		doc      string
		requires []string
	}{
		{
			name:  "motd",
			bname: "motd",
			doc:   `The "message" shown at login`,
		},
		{
			name:     "ntpconf",
			bname:    "ntpconf",
			doc:      "The NTP config",
			requires: []string{"managedfile"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			root := s.checkout(allGo)

			files, err := behavior.Scaffold(root, tc.bname, "", tc.doc, tc.requires)
			s.Require().NoError(err)

			for _, rel := range files {
				got, err := avfs.ReadFile(testFS, filepath.Join(root, rel))
				s.Require().NoError(err)
				golden := filepath.Join("testdata", "scaffold", tc.name, rel+".golden")
				if *update {
					s.Require().NoError(testFS.MkdirAll(filepath.Dir(golden), 0o755))
					s.Require().NoError(avfs.WriteFile(testFS, golden, got, 0o644))
				}
				want, err := avfs.ReadFile(testFS, golden)
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
