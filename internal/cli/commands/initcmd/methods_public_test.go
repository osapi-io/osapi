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

// Package token is the token group of local commands: osapi token generate,
// roles list and validate, which sign and read the bearer tokens of the API
// with the signing key of the config. They run in the CLI, and call no API.
package initcmd_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/avfs/avfs/vfs/rofs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/initcmd"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestInit() {
	src := fstest.MapFS{
		"osapi/SKILL.md": {
			Data: []byte("---\nname: osapi\ndescription: Operate hosts.\n---\nbody\n"),
		},
	}
	version := func() string { return "v1.2.3" }

	tests := []struct {
		name         string
		deps         func() initcmd.Deps
		in           initcmd.Input
		validateFunc func(initcmd.Output, error)
	}{
		{
			name: "when no tool is given, it installs for claude into the current directory",
			deps: func() initcmd.Deps {
				fs := memfs.New()
				s.Require().NoError(fs.MkdirAll("/repo", 0o755))
				s.Require().NoError(fs.Chdir("/repo"))

				return initcmd.Deps{FS: fs, Source: src, Version: version}
			},
			validateFunc: func(got initcmd.Output, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"claude"}, got.Tools)
				s.Equal("v1.2.3", got.Version)
				s.Positive(got.Changed)
				s.Equal(len(got.Files), got.Changed, "each file is new")
			},
		},
		{
			name: "when a path and tools are given, it installs for each",
			deps: func() initcmd.Deps {
				fs := memfs.New()
				s.Require().NoError(fs.MkdirAll("/ops", 0o755))

				return initcmd.Deps{FS: fs, Source: src, Version: version}
			},
			in: initcmd.Input{Path: "/ops", Tools: []string{"claude", "codex"}},
			validateFunc: func(got initcmd.Output, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"claude", "codex"}, got.Tools)
			},
		},
		{
			name: "when the file system is read only, it is the error",
			deps: func() initcmd.Deps {
				return initcmd.Deps{FS: rofs.New(memfs.New()), Source: src, Version: version}
			},
			in: initcmd.Input{Path: "/repo"},
			validateFunc: func(_ initcmd.Output, err error) {
				s.Error(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(initcmd.Init(context.Background(), tc.deps(), tc.in))
		})
	}
}

// TestInitAgain counts only the files that change: a second run on the same
// repository changes none.
func (s *MethodsPublicTestSuite) TestInitAgain() {
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/repo", 0o755))
	d := initcmd.Deps{
		FS: fs,
		Source: fstest.MapFS{
			"osapi/SKILL.md": {Data: []byte("---\nname: osapi\ndescription: x\n---\n")},
		},
		Version: func() string { return "v1" },
	}
	_, err := initcmd.Init(context.Background(), d, initcmd.Input{Path: "/repo"})
	s.Require().NoError(err)

	got, err := initcmd.Init(context.Background(), d, initcmd.Input{Path: "/repo"})
	s.Require().NoError(err)
	s.Zero(got.Changed)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
