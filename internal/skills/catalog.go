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

package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Skill is one skill that the binary has.
type Skill struct {
	// Name is the directory of the skill, and the name in its frontmatter.
	Name string `json:"name"`
	// Description is the description in its frontmatter.
	Description string `json:"description" table:"-"`
	// Init is true when `osapi init` installs the skill.
	Init bool `json:"init" label:"Installed by init"`
	// References are the other files of the skill, relative to its directory.
	References []string `json:"references" table:"-"`
}

// ErrNotFound is the error for a skill or a reference that the binary does not
// have.
var ErrNotFound = errors.New("not found")

// Catalog lists the skills in installed, which `osapi init` writes, and in
// dev, which it does not. Each is a directory of skills. The list is sorted by
// name.
func Catalog(
	installed fs.FS,
	dev fs.FS,
) ([]Skill, error) {
	var out []Skill
	for _, src := range []struct {
		fsys fs.FS
		init bool
	}{{installed, true}, {dev, false}} {
		dirs, err := fs.ReadDir(src.fsys, ".")
		if err != nil {
			return nil, fmt.Errorf("list skills: %w", err)
		}
		for _, d := range dirs {
			if !d.IsDir() {
				continue
			}
			s, err := describe(src.fsys, d.Name())
			if err != nil {
				return nil, err
			}
			s.Init = src.init
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out, nil
}

// Read returns a file of the skill name: its SKILL.md when reference is
// empty, or the reference file, relative to the directory of the skill.
func Read(
	installed fs.FS,
	dev fs.FS,
	name string,
	reference string,
) ([]byte, error) {
	file := "SKILL.md"
	if reference != "" {
		file = path.Clean(reference)
	}
	p := path.Join(name, file)
	if !fs.ValidPath(p) || strings.Contains(name, "/") || !strings.HasPrefix(p, name+"/") {
		return nil, fmt.Errorf("skill %q, file %q: %w", name, reference, ErrNotFound)
	}
	for _, fsys := range []fs.FS{installed, dev} {
		if _, err := fs.Stat(fsys, path.Join(name, "SKILL.md")); err != nil {
			continue
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, fmt.Errorf("skill %q, file %q: %w", name, reference, ErrNotFound)
		}

		return data, nil
	}

	return nil, fmt.Errorf("skill %q: %w", name, ErrNotFound)
}

// describe reads the frontmatter and the reference files of one skill.
func describe(
	fsys fs.FS,
	name string,
) (Skill, error) {
	data, err := fs.ReadFile(fsys, path.Join(name, "SKILL.md"))
	if err != nil {
		return Skill{}, fmt.Errorf("skill %q: %w", name, err)
	}
	s := Skill{Name: name, References: []string{}}
	for _, line := range strings.Split(string(data), "\n") {
		if d, ok := strings.CutPrefix(line, "description: "); ok {
			s.Description = d

			break
		}
	}
	_ = fs.WalkDir(fsys, name, func(p string, d fs.DirEntry, _ error) error {
		if d != nil && !d.IsDir() && p != path.Join(name, "SKILL.md") {
			s.References = append(s.References, strings.TrimPrefix(p, name+"/"))
		}

		return nil
	})

	return s, nil
}
