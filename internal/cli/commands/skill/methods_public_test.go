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

package skill_test

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/avfs/avfs/vfs/memfs"
	"github.com/avfs/avfs/vfs/rofs"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/skill"
)

// unreadable is a source of skills that cannot be read.
type unreadable struct{}

func (unreadable) Open(string) (fs.File, error) {
	return nil, fs.ErrPermission
}

type MethodsPublicTestSuite struct {
	suite.Suite

	deps skill.Deps
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.deps = skill.Deps{
		Installed: fstest.MapFS{
			"osapi/SKILL.md": {
				Data: []byte("---\nname: osapi\ndescription: Operate hosts.\n---\nbody\n"),
			},
			"osapi/references/targets.md": {Data: []byte("targets\n")},
		},
		Dev: fstest.MapFS{},
	}
}

func (s *MethodsPublicTestSuite) TestList() {
	got, err := skill.List(context.Background(), s.deps, skill.ListInput{})
	s.Require().NoError(err)
	s.Require().Len(got.Items, 1)
	s.Equal("osapi", got.Items[0].Name)
	s.True(got.Items[0].Init)

	_, err = skill.List(
		context.Background(),
		skill.Deps{Installed: unreadable{}, Dev: fstest.MapFS{}},
		skill.ListInput{},
	)
	s.Error(err, "a source that cannot be read is an error")
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		in           skill.GetInput
		validateFunc func(skill.GetOutput, error)
	}{
		{
			name: "when the skill is there, its SKILL.md is the content",
			in:   skill.GetInput{Name: "osapi"},
			validateFunc: func(got skill.GetOutput, err error) {
				s.Require().NoError(err)
				s.Contains(got.Content, "body")
			},
		},
		{
			name: "when a reference is given, its file is the content",
			in:   skill.GetInput{Name: "osapi", Reference: "references/targets.md"},
			validateFunc: func(got skill.GetOutput, err error) {
				s.Require().NoError(err)
				s.Equal("targets\n", got.Content)
				s.Equal("references/targets.md", got.Reference)
			},
		},
		{
			name: "when the skill is not there, it says to list them",
			in:   skill.GetInput{Name: "nosuch"},
			validateFunc: func(_ skill.GetOutput, err error) {
				s.ErrorContains(err, "run osapi skill list")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(skill.Get(context.Background(), s.deps, tc.in))
		})
	}
}

func (s *MethodsPublicTestSuite) TestInstall() {
	src := fstest.MapFS{
		"osapi/SKILL.md": {
			Data: []byte("---\nname: osapi\ndescription: Operate hosts.\n---\nbody\n"),
		},
	}
	version := func() string { return "v1.2.3" }

	tests := []struct {
		name         string
		deps         func() skill.Deps
		in           skill.InstallInput
		validateFunc func(skill.InstallOutput, error)
	}{
		{
			name: "when no tool is given, it installs for claude into the current directory",
			deps: func() skill.Deps {
				fs := memfs.New()
				s.Require().NoError(fs.MkdirAll("/repo", 0o755))
				s.Require().NoError(fs.Chdir("/repo"))

				return skill.Deps{FS: fs, Installed: src, Version: version}
			},
			validateFunc: func(got skill.InstallOutput, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"claude"}, got.Tools)
				s.Equal("v1.2.3", got.Version)
				s.Positive(got.Changed)
				s.Equal(len(got.Files), got.Changed, "each file is new")
			},
		},
		{
			name: "when a path and tools are given, it installs for each",
			deps: func() skill.Deps {
				fs := memfs.New()
				s.Require().NoError(fs.MkdirAll("/ops", 0o755))

				return skill.Deps{FS: fs, Installed: src, Version: version}
			},
			in: skill.InstallInput{Path: "/ops", Tools: []string{"claude", "codex"}},
			validateFunc: func(got skill.InstallOutput, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"claude", "codex"}, got.Tools)
			},
		},
		{
			name: "when the file system is read only, it is the error",
			deps: func() skill.Deps {
				return skill.Deps{FS: rofs.New(memfs.New()), Installed: src, Version: version}
			},
			in: skill.InstallInput{Path: "/repo"},
			validateFunc: func(_ skill.InstallOutput, err error) {
				s.Error(err)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.validateFunc(skill.Install(context.Background(), tc.deps(), tc.in))
		})
	}
}

// TestInstallAgain counts only the files that change: a second run on the same
// repository changes none.
func (s *MethodsPublicTestSuite) TestInstallAgain() {
	fs := memfs.New()
	s.Require().NoError(fs.MkdirAll("/repo", 0o755))
	d := skill.Deps{
		FS: fs,
		Installed: fstest.MapFS{
			"osapi/SKILL.md": {Data: []byte("---\nname: osapi\ndescription: x\n---\n")},
		},
		Version: func() string { return "v1" },
	}
	_, err := skill.Install(context.Background(), d, skill.InstallInput{Path: "/repo"})
	s.Require().NoError(err)

	got, err := skill.Install(context.Background(), d, skill.InstallInput{Path: "/repo"})
	s.Require().NoError(err)
	s.Zero(got.Changed)
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
