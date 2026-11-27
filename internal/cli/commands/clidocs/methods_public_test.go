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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/clidocs"
	"github.com/osapi-io/osapi/internal/guide"
)

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestGenerate() {
	root := &cobra.Command{Use: "osapi", Short: "The osapi CLI"}
	root.AddCommand(&cobra.Command{
		Use: "version", Short: "Show the version", Run: func(*cobra.Command, []string) {},
	})
	spec := guide.Spec{
		Kinds: []guide.Kind{{Kind: "a", Title: "Parts"}},
		Terms: []guide.Term{{Kind: "a", Name: "Key", Does: "x", When: "y"}},
	}
	page := clidocs.Page{Path: "docs/docs/sidebar/reference/cli/x.md", Title: "X", Intro: "x."}
	d := clidocs.Deps{
		Root: root,
		DSLs: func() []clidocs.DSL {
			return []clidocs.DSL{{Slug: "behavior", Title: "Behavior DSL", Spec: spec}}
		},
		Pages: func() ([]clidocs.Page, error) { return []clidocs.Page{page}, nil },
	}
	boom := errors.New("boom")

	tests := []struct {
		name         string
		root         func() string
		deps         clidocs.Deps
		validateFunc func(dir string, got clidocs.GenerateOutput, err error)
	}{
		{
			name: "when each part writes, it counts the pages of each under the root",
			root: s.T().TempDir,
			deps: d,
			validateFunc: func(dir string, got clidocs.GenerateOutput, err error) {
				s.Require().NoError(err)
				s.Positive(got.CLI)
				s.Equal(1, got.DSL)
				s.Equal(1, got.Reference)
				s.FileExists(filepath.Join(dir, "docs/docs/sidebar/development/dsl/behavior.md"))
				s.FileExists(filepath.Join(dir, "docs/docs/sidebar/reference/cli/x.md"))
			},
		},
		{
			name: "when the CLI pages cannot be written, it is the error",
			root: func() string {
				f := filepath.Join(s.T().TempDir(), "file")
				s.Require().NoError(os.WriteFile(f, nil, 0o644))

				return f
			},
			deps: d,
			validateFunc: func(_ string, _ clidocs.GenerateOutput, err error) {
				s.ErrorContains(err, "write the CLI reference into")
			},
		},
		{
			name: "when the DSL pages cannot be written, it is the error",
			root: func() string {
				dir := s.T().TempDir()
				s.Require().
					NoError(os.MkdirAll(filepath.Join(dir, "docs/docs/sidebar/development"), 0o755))
				s.Require().NoError(os.WriteFile(
					filepath.Join(dir, "docs/docs/sidebar/development/dsl"), nil, 0o644))

				return dir
			},
			deps: d,
			validateFunc: func(_ string, _ clidocs.GenerateOutput, err error) {
				s.ErrorContains(err, "write the DSL pages into")
			},
		},
		{
			name: "when the reference cannot be read, it is the error",
			root: s.T().TempDir,
			deps: clidocs.Deps{
				Root: d.Root, DSLs: d.DSLs,
				Pages: func() ([]clidocs.Page, error) { return nil, boom },
			},
			validateFunc: func(_ string, _ clidocs.GenerateOutput, err error) {
				s.ErrorIs(err, boom)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			dir := tc.root()
			got, err := clidocs.Generate(
				context.Background(),
				tc.deps,
				clidocs.GenerateInput{Root: dir},
			)
			tc.validateFunc(dir, got, err)
		})
	}

	wd, _ := os.Getwd()
	work := s.T().TempDir()
	s.Require().NoError(os.Chdir(work))
	defer func() { _ = os.Chdir(wd) }()
	_, err := clidocs.Generate(context.Background(), d, clidocs.GenerateInput{})
	s.Require().NoError(err, "with no root, it writes under the working directory")
	s.FileExists(filepath.Join(work, "docs/docs/sidebar/reference/cli/x.md"))
}

func (s *MethodsPublicTestSuite) TestWriteCLI() {
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
	got, err := clidocs.WriteCLI(context.Background(), d, clidocs.CLIInput{Dir: dir})
	s.Require().NoError(err)
	s.Equal(dir, got.Dir)
	s.Positive(got.Pages)

	wd, _ := os.Getwd()
	work := s.T().TempDir()
	s.Require().NoError(os.Chdir(work))
	defer func() { _ = os.Chdir(wd) }()
	got, err = clidocs.WriteCLI(context.Background(), d, clidocs.CLIInput{})
	s.Require().NoError(err, "with no dir, it writes to the docs site's")
	s.Equal("docs/docs/sidebar/usage/cli", got.Dir)

	file := filepath.Join(work, "file")
	s.Require().NoError(os.WriteFile(file, nil, 0o644))
	_, err = clidocs.WriteCLI(
		context.Background(),
		d,
		clidocs.CLIInput{Dir: filepath.Join(file, "pages")},
	)
	s.ErrorContains(err, "write the CLI reference into")
}

func (s *MethodsPublicTestSuite) TestWriteDSL() {
	spec := guide.Spec{
		Kinds: []guide.Kind{{Kind: "a", Title: "Parts"}},
		Terms: []guide.Term{{Kind: "a", Name: "Key", Does: "x", When: "y"}},
	}
	d := clidocs.Deps{DSLs: func() []clidocs.DSL {
		return []clidocs.DSL{
			{Slug: "behavior", Title: "Behavior DSL", Intro: "i", Position: 2, Spec: spec},
		}
	}}
	dir := s.T().TempDir()
	stale := filepath.Join(dir, "old.md")
	s.Require().NoError(os.WriteFile(stale, []byte(guide.Generated), 0o644))
	mine := filepath.Join(dir, "dsl.mdx")
	s.Require().NoError(os.WriteFile(mine, []byte("written by a person"), 0o644))
	kept := filepath.Join(dir, "notes.md")
	s.Require().NoError(os.WriteFile(kept, []byte("written by a person"), 0o644))

	got, err := clidocs.WriteDSL(context.Background(), d, clidocs.DSLInput{Dir: dir})

	s.Require().NoError(err)
	s.Equal(clidocs.DSLOutput{Dir: dir, Pages: 1}, got)
	page, err := os.ReadFile(filepath.Join(dir, "behavior.md"))
	s.Require().NoError(err)
	s.Contains(string(page), "# Behavior DSL")
	s.NoFileExists(stale, "a generated page that no DSL has is removed")
	s.FileExists(mine)
	s.FileExists(kept, "a page that a person wrote stays")

	wd, _ := os.Getwd()
	work := s.T().TempDir()
	s.Require().NoError(os.Chdir(work))
	defer func() { _ = os.Chdir(wd) }()
	got, err = clidocs.WriteDSL(context.Background(), d, clidocs.DSLInput{})
	s.Require().NoError(err)
	s.Equal(
		"docs/docs/sidebar/development/dsl",
		got.Dir,
		"with no dir, it writes to the docs site's",
	)

	file := filepath.Join(work, "file")
	s.Require().NoError(os.WriteFile(file, nil, 0o644))
	_, err = clidocs.WriteDSL(
		context.Background(),
		d,
		clidocs.DSLInput{Dir: filepath.Join(file, "dsl")},
	)
	s.ErrorContains(err, "write the DSL pages")
	s.Require().NoError(os.MkdirAll(filepath.Join(work, "ro"), 0o555))
	_, err = clidocs.WriteDSL(
		context.Background(),
		d,
		clidocs.DSLInput{Dir: filepath.Join(work, "ro")},
	)
	s.ErrorContains(err, "write the DSL pages", "a directory it cannot write into")
}

type refOut struct {
	_    struct{} `title:"Thing"`
	Name string   `              json:"name"`
}

func (s *MethodsPublicTestSuite) TestWriteReference() {
	pages := []clidocs.Page{
		{
			Path:  "docs/docs/sidebar/reference/behaviors/networking/dns/dns.md",
			Title: "DNS",
			Intro: "The resolvers.",
			Parts: []clidocs.Part{
				{Group: "behavior", Out: refOut{Name: "dns"}},
				{
					Group: "behavior", Out: refOut{Name: "get"}, Heading: "Method get",
					Link: func(string, string) string { return "" },
				},
			},
		},
		{
			Path: "docs/docs/sidebar/reference/behaviors/_catalog.mdx", Partial: true,
			Parts: []clidocs.Part{{Group: "behavior", Out: refOut{Name: "all"}}},
		},
		{
			Path: "docs/docs/sidebar/reference/behaviors/networking/_category_.json",
			Text: "{\"label\": \"Networking\"}\n",
		},
	}
	d := clidocs.Deps{Pages: func() ([]clidocs.Page, error) { return pages, nil }}
	root := s.T().TempDir()
	staleDir := filepath.Join(root, "docs/docs/sidebar/reference/chains")
	s.Require().NoError(os.MkdirAll(staleDir, 0o755))
	stale := filepath.Join(staleDir, "gone.md")
	s.Require().NoError(os.WriteFile(stale, []byte(guide.Generated), 0o644))
	landing := filepath.Join(staleDir, "chains.mdx")
	s.Require().NoError(os.WriteFile(landing, []byte("written by a person"), 0o644))
	topCategory := filepath.Join(staleDir, "_category_.json")
	s.Require().NoError(os.WriteFile(topCategory, []byte("{}"), 0o644))
	// A folder of an area that has no page now, with the folder of a page
	// that moved in it.
	goneArea := filepath.Join(root, "docs/docs/sidebar/reference/behaviors/kitchen")
	s.Require().NoError(os.MkdirAll(filepath.Join(goneArea, "oven"), 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(goneArea, "_category_.json"),
		[]byte("{}"), 0o644))
	s.Require().NoError(os.WriteFile(filepath.Join(goneArea, "oven", "oven.md"),
		[]byte(guide.Generated), 0o644))
	kept := filepath.Join(root, "docs/docs/sidebar/reference/behaviors/notes")
	s.Require().NoError(os.MkdirAll(kept, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(kept, "notes.md"),
		[]byte("written by a person"), 0o644))

	got, err := clidocs.WriteReference(
		context.Background(),
		d,
		clidocs.ReferenceInput{Root: root},
	)

	s.Require().NoError(err)
	s.Equal(3, got.Pages)
	page, err := os.ReadFile(filepath.Join(root, pages[0].Path))
	s.Require().NoError(err)
	s.Contains(
		string(page),
		"---\ntitle: DNS\n---\n\n"+guide.Generated+"\n\n# DNS\n\nThe resolvers.\n",
	)
	s.Contains(string(page), "\n## Thing\n")
	s.Contains(string(page), "\n## Method get\n", "a part's heading takes the place of its title")
	part, err := os.ReadFile(filepath.Join(root, pages[1].Path))
	s.Require().NoError(err)
	s.True(strings.HasPrefix(string(part), "{/* Code generated"), "a partial has no front matter")
	s.NoFileExists(stale, "a generated page that no command writes is removed")
	s.FileExists(landing, "a page that a person wrote stays")
	category, err := os.ReadFile(filepath.Join(root, pages[2].Path))
	s.Require().NoError(err)
	s.Equal(pages[2].Text, string(category), "a page with text is the text as it is")
	s.FileExists(topCategory, "the _category_.json of a reference directory stays")
	s.NoDirExists(goneArea, "a folder that no page is in now is removed, with its category")
	s.DirExists(kept, "a folder with a page that a person wrote stays")

	wd, _ := os.Getwd()
	s.Require().NoError(os.Chdir(root))
	defer func() { _ = os.Chdir(wd) }()
	_, err = clidocs.WriteReference(context.Background(), d, clidocs.ReferenceInput{})
	s.Require().NoError(err, "with no root, it writes under the working directory")

	_, err = clidocs.WriteReference(context.Background(), clidocs.Deps{
		Pages: func() ([]clidocs.Page, error) { return nil, errors.New("no registry") },
	}, clidocs.ReferenceInput{Root: root})
	s.ErrorContains(err, "no registry")

	bad := []clidocs.Page{{Path: "x.md", Parts: []clidocs.Part{{Group: "g", Out: func() {}}}}}
	_, err = clidocs.WriteReference(context.Background(), clidocs.Deps{
		Pages: func() ([]clidocs.Page, error) { return bad, nil },
	}, clidocs.ReferenceInput{Root: root})
	s.ErrorContains(err, "x.md", "an output that does not render names its page")

	file := filepath.Join(root, "file")
	s.Require().NoError(os.WriteFile(file, nil, 0o644))
	inFile := []clidocs.Page{{Path: "file/sub/x.md", Title: "X"}}
	_, err = clidocs.WriteReference(context.Background(), clidocs.Deps{
		Pages: func() ([]clidocs.Page, error) { return inFile, nil },
	}, clidocs.ReferenceInput{Root: root})
	s.Error(err, "a page under a file")

	s.Require().NoError(os.MkdirAll(filepath.Join(root, "ro"), 0o555))
	inRO := []clidocs.Page{{Path: "ro/x.md", Title: "X"}}
	_, err = clidocs.WriteReference(context.Background(), clidocs.Deps{
		Pages: func() ([]clidocs.Page, error) { return inRO, nil },
	}, clidocs.ReferenceInput{Root: root})
	s.Error(err, "a page in a directory it cannot write into")
}

// TestWriteReferenceTables writes the generated tables of a page that a
// person wrote, and keeps the rest of the page.
func (s *MethodsPublicTestSuite) TestWriteReferenceTables() {
	root := s.T().TempDir()
	features := filepath.Join(root, "docs/docs/sidebar/features")
	s.Require().NoError(os.MkdirAll(features, 0o755))
	metrics := filepath.Join(features, "metrics.md")
	s.Require().NoError(os.WriteFile(metrics, []byte("# Metrics\n\nTABLE\n"), 0o644))
	same := filepath.Join(features, "same.md")
	s.Require().NoError(os.WriteFile(same, []byte("# Same\n"), 0o644))
	s.Require().NoError(os.WriteFile(filepath.Join(features, "logo.svg"), []byte("TABLE"), 0o644))
	none := func() ([]clidocs.Page, error) { return nil, nil }
	var roots []string
	tables := func(r string, page string) (string, error) {
		roots = append(roots, r)

		return strings.ReplaceAll(page, "TABLE", "| a | b |"), nil
	}

	_, err := clidocs.WriteReference(context.Background(),
		clidocs.Deps{Pages: none, Tables: tables}, clidocs.ReferenceInput{Root: root})

	s.Require().NoError(err)
	got, _ := os.ReadFile(metrics)
	s.Equal("# Metrics\n\n| a | b |\n", string(got))
	got, _ = os.ReadFile(same)
	s.Equal("# Same\n", string(got), "a page with no table is as it was")
	got, _ = os.ReadFile(filepath.Join(features, "logo.svg"))
	s.Equal("TABLE", string(got), "only a markdown page is read")
	s.Equal([]string{root, root}, roots, "the tables read from the root")

	_, err = clidocs.WriteReference(context.Background(), clidocs.Deps{
		Pages:  none,
		Tables: func(string, string) (string, error) { return "", errors.New("no config") },
	}, clidocs.ReferenceInput{Root: root})
	s.ErrorContains(err, "metrics.md: no config", "the error names the page")

	s.Require().NoError(os.Chmod(metrics, 0o000))
	defer func() { _ = os.Chmod(metrics, 0o644) }()
	_, err = clidocs.WriteReference(context.Background(),
		clidocs.Deps{Pages: none, Tables: tables}, clidocs.ReferenceInput{Root: root})
	s.Error(err, "a page it cannot read")

	s.Require().NoError(os.Chmod(metrics, 0o644))
	s.Require().NoError(os.WriteFile(metrics, []byte("TABLE"), 0o644))
	restore := clidocs.FailHostRename("metrics.md")
	_, err = clidocs.WriteReference(context.Background(),
		clidocs.Deps{Pages: none, Tables: tables}, clidocs.ReferenceInput{Root: root})
	restore()
	s.Error(err, "a page it cannot write")
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}
