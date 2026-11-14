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

// Command gen writes the files built from behaviors. For every behavior,
// its docs pages. For one with behavior.API, in its gen/ directory: api.go,
// the table of its routes; api_<method>.go, each method's handler; a test
// for each; and api.yaml, its part of the spec. For one with behavior.CLI,
// cli.go and its test. Then the OpenAPI spec of every behavior, and
// internal/behaviors/all/gen, which imports every behavior's generated code. The
// Go files are built with jen and formatted the way the repository's own
// are. It runs from the go:generate directive in internal/behaviors/all.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/openapi"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	_ "github.com/osapi-io/osapi/internal/validation"
)

// genCmd is the generator's command. It is a cobra command with its flags
// bound through viper, the same way osapi's own commands are, so --root can
// also come from OSAPI_GEN_ROOT.
var genCmd = &cobra.Command{
	Use:           "gen",
	Short:         "Write every file generated from the registered behaviors",
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return run(viper.GetString("root"))
	},
}

func init() {
	genCmd.Flags().String("root", ".", "Repository root to write into")
	_ = viper.BindPFlag("root", genCmd.Flags().Lookup("root"))

	viper.SetEnvPrefix("osapi_gen")
	viper.AutomaticEnv()
}

func main() {
	if err := genCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "behavior gen:", err)
		os.Exit(1)
	}
}

// writer writes the generated files under root and remembers the Go ones,
// relative to root, where the formatter runs, to format them once at the
// end.
type writer struct {
	root string
	gofs []string
}

func run(
	root string,
) error {
	w := &writer{root: root}

	spec, err := openapi.OpenAPI(behavior.All())
	if err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	if err := w.write(behavior.SpecPath, spec); err != nil {
		return err
	}

	for _, b := range behavior.All() {
		// Every behavior has a docs page, internal ones included, so the
		// docs list everything the binary has.
		dir := behaviorcmd.DocsDir(b)
		if err := w.write(filepath.Join(dir, b.Name+".md"), []byte(behaviorcmd.IndexPage(b))); err != nil {
			return err
		}
		if !b.API && !b.CLI {
			continue
		}

		gen := filepath.Join(strings.TrimPrefix(pkgPath(b), module+"/"), "gen")
		if err := w.clean(gen); err != nil {
			return err
		}
		if err := w.behavior(b, gen); err != nil {
			return fmt.Errorf("%s: %w", b.Name, err)
		}

		if !b.CLI {
			continue
		}
		for i, m := range b.Methods {
			page, err := behaviorcmd.MethodPage(b, m, i+2)
			if err != nil {
				return err
			}
			// .mdx, because the page holds the terminal block.
			if err := w.write(filepath.Join(dir, m.Name+".mdx"), []byte(page)); err != nil {
				return err
			}
		}
	}

	catalog := filepath.Join(filepath.Dir(behaviorcmd.DocsDir(behavior.All()[0])), "_catalog.mdx")
	if err := w.write(catalog, []byte(behaviorcmd.CatalogPage(behavior.All()))); err != nil {
		return err
	}
	if err := w.goFile("internal/behaviors/all/gen/gen.go", allFile(behavior.All())); err != nil {
		return err
	}

	return w.format()
}

// behavior writes b's gen directory.
func (w *writer) behavior(
	b *behavior.Behavior,
	gen string,
) error {
	if b.API {
		// This behavior's part of the spec, beside its code, so a review of
		// the behavior shows the API it produces.
		snippet, err := openapi.OpenAPI([]*behavior.Behavior{b})
		if err != nil {
			return fmt.Errorf("spec: %w", err)
		}
		if err := w.write(filepath.Join(gen, "api.yaml"), snippet); err != nil {
			return err
		}
		if err := w.goFile(filepath.Join(gen, "api.go"), apiFile(b)); err != nil {
			return err
		}
		if err := w.goFile(filepath.Join(gen, "api_public_test.go"), apiTestFile(b)); err != nil {
			return err
		}
		for _, m := range b.Methods {
			f, err := apiMethodFile(b, m)
			if err != nil {
				return err
			}
			if err := w.goFile(filepath.Join(gen, "api_"+m.Name+".go"), f); err != nil {
				return err
			}
			t, err := apiMethodTestFile(b, m)
			if err != nil {
				return err
			}
			if err := w.goFile(filepath.Join(gen, "api_"+m.Name+"_public_test.go"), t); err != nil {
				return err
			}
		}
	}

	if b.CLI {
		f, err := cliFile(b)
		if err != nil {
			return fmt.Errorf("commands: %w", err)
		}
		if err := w.goFile(filepath.Join(gen, "cli.go"), f); err != nil {
			return err
		}
		t, err := cliTestFile(b)
		if err != nil {
			return fmt.Errorf("commands: %w", err)
		}
		if err := w.goFile(filepath.Join(gen, "cli_public_test.go"), t); err != nil {
			return err
		}
	}

	return nil
}

// clean removes the generated files in dir, so a method that is gone
// leaves no handler behind.
func (w *writer) clean(
	dir string,
) error {
	entries, err := os.ReadDir(filepath.Join(w.root, dir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		p := filepath.Join(w.root, dir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(data), "// Code generated") ||
			strings.HasPrefix(string(data), "# Code generated") {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}

	return nil
}

// goFile renders f to rel.
func (w *writer) goFile(
	rel string,
	f *jen.File,
) error {
	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	// jen puts every import in one block; goimports groups them the way the
	// repository does: the standard library, then other modules, then this
	// one.
	imports.LocalPrefix = module
	src, err := imports.Process(rel, buf.Bytes(), &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: true,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	w.gofs = append(w.gofs, rel)

	return w.write(rel, src)
}

// format runs the repository's formatters over every Go file written, so
// the generated code reads, and passes the format check, like the rest.
func (w *writer) format() error {
	args := append([]string{
		"tool", "github.com/segmentio/golines",
		"--base-formatter=go tool mvdan.cc/gofumpt", "-w",
	}, w.gofs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = w.root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("format: %w: %s", err, out)
	}

	return nil
}

func (w *writer) write(
	rel string,
	data []byte,
) error {
	p := filepath.Join(w.root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	return os.WriteFile(p, data, 0o644)
}
