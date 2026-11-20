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
package command

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/imports"
)

const (
	commandPath = module + "/internal/cli/command"
	suitePath   = "github.com/stretchr/testify/suite"
)

// license heads every file the scaffold writes, as it heads every Go file in
// the repository.
const license = `// Copyright (c) 2026 John Dewey

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
// DEALINGS IN THE SOFTWARE.`

// scaffoldSource is one file of a new group: where it goes under the
// group's directory, and its code.
type scaffoldSource struct {
	rel  string
	file *jen.File
}

// groupScaffold is what a new group is called.
type groupScaffold struct {
	name string
	doc  string
}

func (sc groupScaffold) pkg() string {
	return module + "/internal/cli/commands/" + sc.name
}

// typeName is a type of the group, with the group's name in front, as the
// types of an endpoint group have.
func (sc groupScaffold) typeName(
	suffix string,
) string {
	return strings.ToUpper(sc.name[:1]) + sc.name[1:] + suffix
}

// scaffoldSources builds the files of a new group; a test swaps it.
var scaffoldSources = func(
	sc groupScaffold,
) []scaffoldSource {
	return []scaffoldSource{
		{sc.name + ".go", sc.groupFile()},
		{"deps.go", sc.depsFile()},
		{"methods.go", sc.methodsFile()},
		{"methods_public_test.go", sc.methodsTestFile()},
		{"types.go", sc.typesFile()},
	}
}

// renderGo renders f as the repository writes Go: goimports groups the
// imports as the standard library, then other modules, then this one.
func renderGo(
	f *jen.File,
) ([]byte, error) {
	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return nil, err
	}
	imports.LocalPrefix = module

	return imports.Process("scaffold.go", buf.Bytes(), &imports.Options{
		Comments: true, TabIndent: true, TabWidth: 8, FormatOnly: true,
	})
}

func (sc groupScaffold) newFile() *jen.File {
	f := jen.NewFilePathName(sc.pkg(), sc.name)
	f.HeaderComment(license)
	f.ImportName(commandPath, "command")

	return f
}

func comments(
	f *jen.File,
	lines ...string,
) {
	for _, l := range lines {
		f.Comment(l)
	}
}

func multiline(
	open string,
	closing string,
	items ...jen.Code,
) *jen.Statement {
	return jen.Custom(
		jen.Options{Open: open, Close: closing, Separator: ",", Multi: true},
		items...)
}

func paramList(
	items ...jen.Code,
) *jen.Statement {
	return multiline("(", ")", items...)
}

func tagged(
	pairs ...string,
) *jen.Statement {
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, fmt.Sprintf("%s:%q", pairs[i], pairs[i+1]))
	}

	return jen.Op("`" + strings.Join(parts, " ") + "`")
}

func (sc groupScaffold) groupFile() *jen.File {
	f := sc.newFile()
	f.PackageComment(
		fmt.Sprintf("Package %s is the %s group of local commands: %s.", sc.name, sc.name, sc.doc),
	)
	c := func(name string) *jen.Statement { return jen.Qual(commandPath, name) }
	next := func(method string, args ...jen.Code) *jen.Statement {
		return jen.Op(".").Line().Id(method).Call(args...)
	}

	comments(f,
		"Group is the "+sc.name+" group's definition. The commands, their flags,",
		"their help and their pages come from it. osapi cli command spec get lists",
		"every option.",
	)
	f.Var().Id("Group").Op("=").Add(c("Register").Call(c("Define").Add(multiline("(", ")",
		jen.Lit(sc.name).Op(",").Lit(sc.doc),
		c("Uses").Index(jen.Id("Deps")).Call(),
		c("Method").Index(jen.List(jen.Id("ListInput"), jen.Id(sc.typeName("List")))).
			Call(jen.Lit("list")).
			Add(next("Short", jen.Lit("List the "+sc.name+" items"))).
			Add(next("Example", jen.Lit(""))).
			Add(next("Handle", jen.Id("List"))),
		c("Method").Index(jen.List(jen.Id("GetInput"), jen.Id(sc.typeName("Item")))).
			Call(jen.Lit("get")).
			Add(next("Short", jen.Lit("Get one "+sc.name+" item"))).
			Add(next("Example", jen.Lit("--name example"))).
			Add(next("Fails", jen.Lit("no name"), jen.Id("GetInput").Values())).
			Add(next("Fails", jen.Lit("a name over 255 characters"),
				jen.Id("GetInput").Values(jen.Dict{
					jen.Id("Name"): jen.Qual("strings", "Repeat").Call(jen.Lit("a"), jen.Lit(256)),
				}))).
			Add(next("Handle", jen.Id("Get"))),
	))))

	return f
}

func (sc groupScaffold) depsFile() *jen.File {
	f := sc.newFile()
	comments(f,
		"Deps is what the group's methods use, such as the config or the file",
		"system. Give each one as an interface, so a test can fake it. While it",
		"is empty, the CLI needs no wiring; once it has a field, give it with",
		"command.Provide in cmd/commands.go, as the token group's is.",
	)
	f.Type().Id("Deps").Struct()

	return f
}

func (sc groupScaffold) typesFile() *jen.File {
	f := sc.newFile()
	item, list := sc.typeName("Item"), sc.typeName("List")

	comments(f, "ListInput is list's input, which has none.")
	f.Type().Id("ListInput").Struct()
	f.Line()
	comments(f,
		"GetInput names the item to read. Each field is a flag of the command,",
		"named for its json name, with its doc as its help.",
	)
	f.Type().Id("GetInput").Struct(
		jen.Id("Name").String().Add(tagged("json", "name",
			"validate", "required,min=1,max=255", "doc", "The name of the item")),
	)
	f.Line()
	comments(f, item+" is one item. With --json, it prints as JSON.")
	f.Type().Id(item).Struct(
		jen.Id("Name").String().Add(tagged("json", "name")),
	)
	f.Line()
	comments(f, list+" is the items, with the total.")
	f.Type().Id(list).Struct(
		jen.Id("Items").Index().Id(item).Add(tagged("json", "items")),
		jen.Id("Total").Int().Add(tagged("json", "total")),
	)

	return f
}

func (sc groupScaffold) methodsFile() *jen.File {
	f := sc.newFile()
	item, list := sc.typeName("Item"), sc.typeName("List")
	ctx := jen.Id("_").Qual("context", "Context")

	comments(f, "ErrNotFound is an item that is not there.")
	f.Var().Id("ErrNotFound").Op("=").Qual("errors", "New").Call(jen.Lit("not found"))
	f.Line()
	comments(f,
		"items stand in for what the group reads, such as a file: a method gets",
		"it from Deps.",
	)
	f.Var().Id("items").Op("=").Index().Id(item).Values(
		jen.Values(jen.Id("Name").Op(":").Lit("example")))
	f.Line()
	comments(f, "List is every item.")
	f.Func().Id("List").Add(paramList(ctx, jen.Id("_").Id("Deps"), jen.Id("_").Id("ListInput"))).
		Params(jen.Id(list), jen.Error()).Block(
		jen.Return(jen.Id(list).Values(
			jen.Id("Items").Op(":").Id("items"),
			jen.Id("Total").Op(":").Len(jen.Id("items")),
		), jen.Nil()),
	)
	f.Line()
	comments(f,
		"Get is the item with the name. An item that is not there is ErrNotFound,",
		"with the name in the message.",
	)
	f.Func().
		Id("Get").
		Add(paramList(ctx.Clone(), jen.Id("_").Id("Deps"), jen.Id("in").Id("GetInput"))).
		Params(jen.Id(item), jen.Error()).
		Block(
			jen.For(jen.List(jen.Id("_"), jen.Id("it")).Op(":=").Range().Id("items")).Block(
				jen.If(jen.Id("it").Dot("Name").Op("==").Id("in").Dot("Name")).Block(
					jen.Return(jen.Id("it"), jen.Nil()),
				),
			),
			jen.Line(),
			jen.Return(jen.Id(item).Values(), jen.Qual("fmt", "Errorf").Call(
				jen.Lit("no item named %q: %w"), jen.Id("in").Dot("Name"), jen.Id("ErrNotFound"))),
		)

	return f
}

func (sc groupScaffold) methodsTestFile() *jen.File {
	f := jen.NewFilePathName(sc.pkg()+"_test", sc.name+"_test")
	f.HeaderComment(license)
	f.ImportName(sc.pkg(), sc.name)
	f.ImportName(suitePath, "suite")
	q := func(name string) *jen.Statement { return jen.Qual(sc.pkg(), name) }
	bg := jen.Qual("context", "Background").Call()
	const suite = "MethodsPublicTestSuite"

	f.Type().Id(suite).Struct(jen.Qual(suitePath, "Suite"))
	f.Line()
	f.Func().Params(jen.Id("s").Op("*").Id(suite)).Id("TestList").Params().Block(
		jen.List(jen.Id("got"), jen.Err()).Op(":=").Add(q("List")).
			Call(bg.Clone(), q("Deps").Values(), q("ListInput").Values()),
		jen.Line(),
		jen.Id("s").Dot("Require").Call().Dot("NoError").Call(jen.Err()),
		jen.Id("s").Dot("Equal").Call(jen.Lit(1), jen.Id("got").Dot("Total")),
	)
	f.Line()
	f.Func().Params(jen.Id("s").Op("*").Id(suite)).Id("TestGet").Params().Block(
		jen.Id("tests").Op(":=").Index().Struct(
			jen.Id("name").String(),
			jen.Id("item").String(),
			jen.Id("wantErr").Error(),
		).Add(multiline("{", "}",
			multiline("{", "}",
				jen.Id("name").Op(":").Lit("when the item is there, it is the answer"),
				jen.Id("item").Op(":").Lit("example"),
			),
			multiline("{", "}",
				jen.Id("name").Op(":").Lit("when the item is not there, it is ErrNotFound"),
				jen.Id("item").Op(":").Lit("nosuch"),
				jen.Id("wantErr").Op(":").Add(q("ErrNotFound")),
			),
		)),
		jen.Line(),
		jen.For(jen.List(jen.Id("_"), jen.Id("tc")).Op(":=").Range().Id("tests")).Block(
			jen.Id("s").Dot("Run").Call(jen.Id("tc").Dot("name"), jen.Func().Params().Block(
				jen.List(jen.Id("got"), jen.Err()).Op(":=").Add(q("Get")).Call(
					bg.Clone(), q("Deps").Values(), q("GetInput").Values(
						jen.Id("Name").Op(":").Id("tc").Dot("item"))),
				jen.Line(),
				jen.If(jen.Id("tc").Dot("wantErr").Op("!=").Nil()).Block(
					jen.Id("s").Dot("ErrorIs").Call(jen.Err(), jen.Id("tc").Dot("wantErr")),
					jen.Line(),
					jen.Return(),
				),
				jen.Id("s").Dot("Require").Call().Dot("NoError").Call(jen.Err()),
				jen.Id("s").Dot("Equal").Call(jen.Id("tc").Dot("item"), jen.Id("got").Dot("Name")),
			)),
		),
	)
	f.Line()
	f.Func().Id("Test" + suite).Add(paramList(jen.Id("t").Op("*").Qual("testing", "T"))).Block(
		jen.Qual(suitePath, "Run").Call(jen.Id("t"), jen.New(jen.Id(suite))),
	)

	return f
}
