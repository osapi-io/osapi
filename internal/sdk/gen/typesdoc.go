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

package main

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
)

// typesDocFile is pkg/sdk/types/types.gen.go: Doc of each closed set of the
// SDK, a type with Values, from its doc comment, and All, which lists each
// one. A caller reads what the SDK has from All, as osapi reads gohai's
// collectors from gohai.Collectors, and the spec describes the type with its
// Doc, so the comment, the code and the spec say the same.
func typesDocFile(
	dir string,
) (*jen.File, error) {
	fset := token.NewFileSet()
	entries, err := hostFS.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") ||
			name == typesDocName {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	// NewFromFiles fails only on a file outside fset, and each is in it.
	pkg, _ := doc.NewFromFiles(fset, files, "github.com/osapi-io/osapi/"+typesDir)

	f := jen.NewFile("types")
	f.HeaderComment(strings.TrimPrefix(header, "// "))
	var all []jen.Code
	for _, t := range pkg.Types {
		methods := map[string]bool{}
		for _, m := range t.Methods {
			methods[m.Name] = true
		}
		if !methods["Values"] || methods["Doc"] {
			continue
		}
		text := strings.Join(strings.Fields(t.Doc), " ")
		f.Comment("Doc is what " + t.Name + " is, from its doc comment.")
		f.Func().Params(jen.Id(t.Name)).Id("Doc").Params().String().Block(
			jen.Return(jen.Lit(text)),
		)
		f.Line()
		zero := jen.Id(t.Name).Call(jen.Lit(""))
		all = append(all, jen.Values(jen.Dict{
			jen.Id("Name"):   jen.Lit(t.Name),
			jen.Id("Doc"):    zero.Clone().Dot("Doc").Call(),
			jen.Id("Values"): jen.Id("names").Call(zero.Clone().Dot("Values").Call()),
		}))
	}
	if len(all) == 0 {
		return nil, nil
	}
	f.Comment("All is each closed set of the SDK, by name: what it is and its values.")
	f.Func().Id("All").Params().Index().Id("TypeSpec").Block(
		jen.Return(jen.Index().Id("TypeSpec").Values(all...)),
	)

	return f, nil
}

// typesDocName is the file of pkg/sdk/types that typesDocFile writes.
const typesDocName = "types.gen.go"
