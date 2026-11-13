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

package skills_test

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/failfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/skills"
)

type InstallPublicTestSuite struct {
	suite.Suite
}

// source is a skills directory with one skill of two files.
func source() fstest.MapFS {
	return fstest.MapFS{
		"osapi/SKILL.md":        {Data: []byte("skill\n")},
		"osapi/references/a.md": {Data: []byte("ref a\n")},
	}
}

// brokenFS fails to read one file, standing in for a damaged embed.
type brokenFS struct {
	fstest.MapFS
}

func (b brokenFS) Open(
	name string,
) (fs.File, error) {
	if name == "osapi/SKILL.md" {
		return nil, errors.New("unreadable")
	}

	return b.MapFS.Open(name)
}

func (b brokenFS) ReadFile(
	name string,
) ([]byte, error) {
	if name == "osapi/SKILL.md" {
		return nil, errors.New("unreadable")
	}

	return b.MapFS.ReadFile(name)
}

// failing returns a filesystem where fn fails for paths containing match.
func failing(
	base avfs.VFS,
	fn avfs.FnVFS,
	match string,
) avfs.VFS {
	vfs := failfs.New(base)
	_ = vfs.SetFailFunc(func(
		_ avfs.VFSBase,
		got avfs.FnVFS,
		fp *failfs.FailParam,
	) error {
		if got == fn && strings.Contains(fp.Path, match) {
			return fmt.Errorf("injected %v error", got)
		}

		return nil
	})

	return vfs
}

func statuses(
	results []skills.Result,
) map[string]skills.Status {
	out := map[string]skills.Status{}
	for _, r := range results {
		out[r.Path] = r.Status
	}

	return out
}

func (s *InstallPublicTestSuite) TestToolNames() {
	assert.Equal(s.T(), []string{"claude", "codex"}, skills.ToolNames())
}

func (s *InstallPublicTestSuite) TestInstall() {
	claude := skills.Tools["claude"]
	codex := skills.Tools["codex"]

	tests := []struct {
		name         string
		setup        func(avfs.VFS)
		vfs          func(avfs.VFS) avfs.VFS
		tools        []skills.Tool
		src          fs.FS
		wantErr      string
		validateFunc func(avfs.VFS, map[string]skills.Status)
	}{
		{
			name:  "when the repository is empty every file is created",
			tools: []skills.Tool{claude, codex},
			validateFunc: func(vfs avfs.VFS, got map[string]skills.Status) {
				assert.Equal(s.T(), map[string]skills.Status{
					".claude/skills/osapi/SKILL.md":        skills.Created,
					".claude/skills/osapi/references/a.md": skills.Created,
					"CLAUDE.md":                            skills.Created,
					".agents/skills/osapi/SKILL.md":        skills.Created,
					".agents/skills/osapi/references/a.md": skills.Created,
					"AGENTS.md":                            skills.Created,
				}, got)
				data, _ := vfs.ReadFile("/repo/CLAUDE.md")
				assert.True(s.T(), strings.HasPrefix(string(data), "<!-- osapi:begin -->"))
				assert.Contains(s.T(), string(data), "version 1.2.3")
			},
		},
		{
			name: "when run again nothing changes",
			setup: func(vfs avfs.VFS) {
				_, _ = skills.Install(vfs, "/repo", []skills.Tool{claude}, source(), "1.2.3")
			},
			tools: []skills.Tool{claude},
			validateFunc: func(_ avfs.VFS, got map[string]skills.Status) {
				for path, status := range got {
					assert.Equal(s.T(), skills.Unchanged, status, path)
				}
			},
		},
		{
			name: "when a skill file differs it is updated",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo/.claude/skills/osapi", 0o755)
				_ = vfs.WriteFile("/repo/.claude/skills/osapi/SKILL.md", []byte("old\n"), 0o644)
			},
			tools: []skills.Tool{claude},
			validateFunc: func(_ avfs.VFS, got map[string]skills.Status) {
				assert.Equal(s.T(), skills.Updated, got[".claude/skills/osapi/SKILL.md"])
			},
		},
		{
			name: "when an older version left a file it is removed, and other skills are kept",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo/.claude/skills/osapi/references", 0o755)
				_ = vfs.WriteFile(
					"/repo/.claude/skills/osapi/references/gone.md",
					[]byte("x"),
					0o644,
				)
				_ = vfs.MkdirAll("/repo/.claude/skills/mine", 0o755)
				_ = vfs.WriteFile("/repo/.claude/skills/mine/SKILL.md", []byte("x"), 0o644)
			},
			tools: []skills.Tool{claude},
			validateFunc: func(vfs avfs.VFS, got map[string]skills.Status) {
				assert.Equal(s.T(), skills.Removed, got[".claude/skills/osapi/references/gone.md"])
				_, err := vfs.Stat("/repo/.claude/skills/osapi/references/gone.md")
				assert.Error(s.T(), err)
				_, err = vfs.Stat("/repo/.claude/skills/mine/SKILL.md")
				assert.NoError(s.T(), err)
			},
		},
		{
			name: "when the instructions file has no section one is appended",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo", 0o755)
				_ = vfs.WriteFile("/repo/CLAUDE.md", []byte("# Mine"), 0o644)
			},
			tools: []skills.Tool{claude},
			validateFunc: func(vfs avfs.VFS, got map[string]skills.Status) {
				assert.Equal(s.T(), skills.Updated, got["CLAUDE.md"])
				data, _ := vfs.ReadFile("/repo/CLAUDE.md")
				assert.True(
					s.T(),
					strings.HasPrefix(string(data), "# Mine\n\n<!-- osapi:begin -->"),
				)
			},
		},
		{
			name: "when the instructions file ends in a newline the section follows a blank line",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo", 0o755)
				_ = vfs.WriteFile("/repo/CLAUDE.md", []byte("# Mine\n"), 0o644)
			},
			tools: []skills.Tool{claude},
			validateFunc: func(vfs avfs.VFS, _ map[string]skills.Status) {
				data, _ := vfs.ReadFile("/repo/CLAUDE.md")
				assert.True(
					s.T(),
					strings.HasPrefix(string(data), "# Mine\n\n<!-- osapi:begin -->"),
				)
			},
		},
		{
			name: "when the instructions file is empty it gets the section alone",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo", 0o755)
				_ = vfs.WriteFile("/repo/CLAUDE.md", []byte(""), 0o644)
			},
			tools: []skills.Tool{claude},
			validateFunc: func(vfs avfs.VFS, _ map[string]skills.Status) {
				data, _ := vfs.ReadFile("/repo/CLAUDE.md")
				assert.True(s.T(), strings.HasPrefix(string(data), "<!-- osapi:begin -->"))
			},
		},
		{
			name: "when the section exists only it is replaced",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo", 0o755)
				_ = vfs.WriteFile("/repo/CLAUDE.md", []byte(
					"before\n<!-- osapi:begin -->\nold\n<!-- osapi:end -->\nafter\n",
				), 0o644)
			},
			tools: []skills.Tool{claude},
			validateFunc: func(vfs avfs.VFS, _ map[string]skills.Status) {
				data, _ := vfs.ReadFile("/repo/CLAUDE.md")
				text := string(data)
				assert.True(s.T(), strings.HasPrefix(text, "before\n<!-- osapi:begin -->"))
				assert.True(s.T(), strings.HasSuffix(text, "<!-- osapi:end -->\nafter\n"))
				assert.NotContains(s.T(), text, "old")
			},
		},
		{
			name:    "when the source cannot be read",
			tools:   []skills.Tool{claude},
			src:     brokenFS{source()},
			wantErr: "install skills for claude",
		},
		{
			name: "when a directory cannot be created",
			vfs: func(base avfs.VFS) avfs.VFS {
				return failing(base, avfs.FnMkdirAll, "references")
			},
			tools:   []skills.Tool{claude},
			wantErr: "injected",
		},
		{
			name: "when a skill file cannot be written",
			vfs: func(base avfs.VFS) avfs.VFS {
				return failing(base, avfs.FnOpenFile, "SKILL.md")
			},
			tools:   []skills.Tool{claude},
			wantErr: "injected",
		},
		{
			name: "when a stale file cannot be removed",
			setup: func(vfs avfs.VFS) {
				_ = vfs.MkdirAll("/repo/.claude/skills/osapi", 0o755)
				_ = vfs.WriteFile("/repo/.claude/skills/osapi/gone.md", []byte("x"), 0o644)
			},
			vfs: func(base avfs.VFS) avfs.VFS {
				return failing(base, avfs.FnRemove, "gone.md")
			},
			tools:   []skills.Tool{claude},
			wantErr: "remove stale skill files",
		},
		{
			name: "when the instructions file cannot be written",
			vfs: func(base avfs.VFS) avfs.VFS {
				return failing(base, avfs.FnOpenFile, "CLAUDE.md")
			},
			tools:   []skills.Tool{claude},
			wantErr: "injected",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			base := memfs.New()
			if tc.setup != nil {
				tc.setup(base)
			}
			vfs := avfs.VFS(base)
			if tc.vfs != nil {
				vfs = tc.vfs(base)
			}
			src := fs.FS(source())
			if tc.src != nil {
				src = tc.src
			}

			results, err := skills.Install(vfs, "/repo", tc.tools, src, "1.2.3")

			if tc.wantErr != "" {
				s.Require().Error(err)
				assert.Contains(s.T(), err.Error(), tc.wantErr)

				return
			}
			s.Require().NoError(err)
			tc.validateFunc(base, statuses(results))
		})
	}
}

func TestInstallPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(InstallPublicTestSuite))
}
