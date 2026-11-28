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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/behavior"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/cli/command/gen/genfile"
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
	// A generated page shows each time in UTC, so it does not change
	// with the time zone of the machine that writes it.
	time.Local = time.UTC

	if err := genCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "behavior gen:", err)
		os.Exit(1)
	}
}

func run(
	root string,
) error {
	w := &genfile.Writer{Root: root}

	for _, b := range behavior.All() {
		if !b.API && !b.CLI {
			continue
		}

		gen := filepath.Join(strings.TrimPrefix(pkgPath(b), module+"/"), "gen")
		if err := w.Clean(gen); err != nil {
			return err
		}
		if err := writeBehavior(w, b, gen); err != nil {
			return fmt.Errorf("%s: %w", b.Name, err)
		}
	}

	if err := w.Registry("internal/behaviors/all/gen", generated, wrap("Package gen imports the "+
		"generated code of every behavior with an API or a CLI: its commands and its routes.", 76),
		served(behavior.All())); err != nil {
		return err
	}

	return w.Flush()
}

// behavior writes b's gen directory.
func writeBehavior(
	w *genfile.Writer,
	b *behavior.Behavior,
	gen string,
) error {
	if b.API {
		// The behavior's endpoint group, which internal/behavior/gen/group
		// makes the behavior's spec from, once this file compiles.
		g, err := groupFile(b)
		if err != nil {
			return fmt.Errorf("group: %w", err)
		}
		if err := w.Go(filepath.Join(gen, "group.go"), g); err != nil {
			return err
		}
		// The routes and the decoders, gen/api.go and gen/decode.go, come
		// from the group too: internal/behavior/gen/group writes them with the
		// endpoint layer's route writer.
		// gen/api.go's test is one file, named for it, with one suite: the
		// routes, then the tests of each method.
		t := apiTestFile(b)
		for _, m := range b.Methods {
			if err := apiMethodTests(t, b, m); err != nil {
				return err
			}
		}
		if err := w.Go(filepath.Join(gen, "api_public_test.go"), t); err != nil {
			return err
		}
	}

	// The commands, gen/cli.go and its test, come from the group too:
	// internal/behavior/gen/group writes them with the endpoint layer's
	// command writer.

	return nil
}
