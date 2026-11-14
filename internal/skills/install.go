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

// Package skills writes the skills compiled into osapi into a repository, for
// the agent tools that read them.
package skills

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/avfs/avfs"
)

// Tool is an agent tool osapi can install skills for.
type Tool struct {
	// Name is what --tool takes.
	Name string
	// SkillsDir is where the tool reads skills, relative to the repository.
	SkillsDir string
	// Instructions is the file the tool reads on every session, where osapi
	// keeps a short managed section pointing at the skill.
	Instructions string
}

// Tools are the supported tools, by name.
var Tools = map[string]Tool{
	"claude": {Name: "claude", SkillsDir: ".claude/skills", Instructions: "CLAUDE.md"},
	"codex":  {Name: "codex", SkillsDir: ".agents/skills", Instructions: "AGENTS.md"},
}

// ToolNames lists the supported tools, sorted.
func ToolNames() []string {
	names := make([]string, 0, len(Tools))
	for n := range Tools {
		names = append(names, n)
	}
	sort.Strings(names)

	return names
}

// Status says what Install did to one file.
type Status string

const (
	// Created means the file did not exist.
	Created Status = "created"
	// Updated means the file existed with different content.
	Updated Status = "updated"
	// Unchanged means the file already had this content.
	Unchanged Status = "unchanged"
	// Removed means the file was left over from an older osapi.
	Removed Status = "removed"
)

// Result is one file Install looked at.
type Result struct {
	Path   string `json:"path"`
	Status Status `json:"status"`
}

// The markers around osapi's section of an instructions file. Everything
// between them is osapi's and is rewritten on every install; everything
// outside them is the repository's own and is never touched.
const (
	beginMarker = "<!-- osapi:begin -->"
	endMarker   = "<!-- osapi:end -->"
)

// section is osapi's part of a tool's instructions file.
func section(
	version string,
) string {
	return beginMarker + `
## osapi

This repository operates hosts through osapi (version ` + version + `). Use the
osapi skill to change hosts and the osapi-operations skill to run osapi itself. Run ` + "`osapi help <command> --json`" + ` before a command and
` + "`osapi behavior get --name <behavior> --method <method> --json`" + ` before a request,
and pass ` + "`--json`" + ` to every command. This section is written by
` + "`osapi init`" + `; edit outside it.
` + endMarker + "\n"
}

// Install writes every skill in src, a directory of skills, into root for each
// tool, removes files an older osapi wrote that this one no longer ships, and
// writes osapi's section of each tool's instructions file. Running it again
// is how the skills are upgraded to match the binary.
func Install(
	vfs avfs.VFS,
	root string,
	tools []Tool,
	src fs.FS,
	version string,
) ([]Result, error) {
	var results []Result

	for _, tool := range tools {
		r, err := installSkills(vfs, root, tool, src)
		if err != nil {
			return results, err
		}
		results = append(results, r...)

		res, err := writeSection(vfs, filepath.Join(root, tool.Instructions), version)
		if err != nil {
			return results, err
		}
		results = append(results, Result{Path: tool.Instructions, Status: res})
	}

	return results, nil
}

func installSkills(
	vfs avfs.VFS,
	root string,
	tool Tool,
	src fs.FS,
) ([]Result, error) {
	var results []Result
	shipped := map[string]bool{}
	skillDirs := map[string]bool{}

	err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		skillDirs[strings.SplitN(p, "/", 2)[0]] = true

		data, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}

		rel := path.Join(tool.SkillsDir, p)
		shipped[rel] = true
		status, err := writeFile(vfs, filepath.Join(root, filepath.FromSlash(rel)), data)
		if err != nil {
			return err
		}
		results = append(results, Result{Path: rel, Status: status})

		return nil
	})
	if err != nil {
		return results, fmt.Errorf("install skills for %s: %w", tool.Name, err)
	}

	removed, err := removeStale(vfs, root, tool, skillDirs, shipped)

	return append(results, removed...), err
}

// removeStale deletes files inside osapi's own skill directories that this
// version does not ship. Other skills in the same tool directory are not
// osapi's and are left alone.
func removeStale(
	vfs avfs.VFS,
	root string,
	tool Tool,
	skillDirs map[string]bool,
	shipped map[string]bool,
) ([]Result, error) {
	var results []Result

	names := make([]string, 0, len(skillDirs))
	for n := range skillDirs {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		dir := filepath.Join(root, filepath.FromSlash(path.Join(tool.SkillsDir, name)))
		err := avfs.WalkDir(vfs, dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			if shipped[rel] {
				return nil
			}
			if err := vfs.Remove(p); err != nil {
				return err
			}
			results = append(results, Result{Path: rel, Status: Removed})

			return nil
		})
		if err != nil {
			return results, fmt.Errorf("remove stale skill files: %w", err)
		}
	}

	return results, nil
}

func writeFile(
	vfs avfs.VFS,
	p string,
	data []byte,
) (Status, error) {
	old, err := vfs.ReadFile(p)
	if err == nil && bytes.Equal(old, data) {
		return Unchanged, nil
	}
	status := Updated
	if err != nil {
		status = Created
	}

	if err := vfs.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", filepath.Dir(p), err)
	}
	if err := vfs.WriteFile(p, data, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", p, err)
	}

	return status, nil
}

// writeSection puts osapi's section in an instructions file: replacing the
// one between the markers, or appending one when there is none.
func writeSection(
	vfs avfs.VFS,
	p string,
	version string,
) (Status, error) {
	text := section(version)

	old, err := vfs.ReadFile(p)
	if err != nil {
		return writeFile(vfs, p, []byte(text))
	}

	content := string(old)
	begin := strings.Index(content, beginMarker)
	end := strings.Index(content, endMarker)

	var next string
	switch {
	case begin >= 0 && end > begin:
		after := content[end+len(endMarker):]
		after = strings.TrimPrefix(after, "\n")
		next = content[:begin] + text + after
	default:
		sep := "\n"
		if strings.HasSuffix(content, "\n") || content == "" {
			sep = ""
		}
		if content != "" {
			sep += "\n"
		}
		next = content + sep + text
	}

	return writeFile(vfs, p, []byte(next))
}
