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

package behavior

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/imports"
)

const (
	behaviorPath     = module + "/internal/behavior"
	areaPath         = module + "/internal/area"
	behaviortestPath = module + "/internal/behavior/behaviortest"
	hostPath         = module + "/internal/behavior/host"
	atomicfilePath   = module + "/internal/atomicfile"
	avfsPath         = "github.com/avfs/avfs"
	memfsPath        = "github.com/avfs/avfs/vfs/memfs"
	gomockPath       = "go.uber.org/mock/gomock"
	suitePath        = "github.com/stretchr/testify/suite"
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

// scaffoldSource is one file of a new behavior: where it goes under the
// behavior's directory, and its code.
type scaffoldSource struct {
	rel  string
	file *jen.File
}

// scaffold is what a new behavior is called and where it lives.
type scaffold struct {
	// name is the behavior's name, in its routes, commands and docs.
	name string
	// goPackage is the Go package and directory it lives in.
	goPackage string
	path      string
	doc       string
	// requires are the behaviors it is built on, in the order given.
	requires []requirement
}

// pkg is the import path of the new behavior's package.
func (sc scaffold) pkg() string {
	return module + "/internal/behaviors/" + sc.path
}

// scaffoldSources builds the files of a new behavior; a test swaps it.
var scaffoldSources = func(
	sc scaffold,
) []scaffoldSource {
	return []scaffoldSource{
		{"behavior.go", sc.behaviorFile()},
		{"behavior_public_test.go", sc.behaviorTestFile()},
		{"debian.go", sc.debianFile()},
		{"debian_public_test.go", sc.debianTestFile()},
		{"methods.go", sc.methodsFile()},
		{"methods_public_test.go", sc.methodsTestFile()},
		{"mocks/generate.go", sc.mocksFile()},
		{"provider.go", sc.providerFile()},
		{"settings.go", sc.settingsFile()},
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
		Comments:   true,
		TabIndent:  true,
		TabWidth:   8,
		FormatOnly: true,
	})
}

// newFile is a file of the behavior's package.
func (sc scaffold) newFile() *jen.File {
	f := jen.NewFilePathName(sc.pkg(), sc.goPackage)
	f.HeaderComment(license)
	f.ImportName(areaPath, "area")
	f.ImportName(behaviorPath, "behavior")
	for _, r := range sc.requires {
		f.ImportName(r.path, r.pkg)
	}

	return f
}

// newTestFile is a file of the behavior's external test package.
func (sc scaffold) newTestFile() *jen.File {
	f := jen.NewFilePathName(sc.pkg()+"_test", sc.goPackage+"_test")
	f.HeaderComment(license)
	f.ImportName(sc.pkg(), sc.goPackage)
	f.ImportName(sc.pkg()+"/mocks", "mocks")
	f.ImportName(suitePath, "suite")
	f.ImportName(gomockPath, "gomock")
	f.ImportName(behaviorPath, "behavior")
	f.ImportName(behaviortestPath, "behaviortest")
	for _, r := range sc.requires {
		f.ImportName(r.path, r.pkg)
		f.ImportAlias(r.path+"/mocks", r.pkg+"mocks")
	}

	return f
}

// comments adds each line as its own // comment; jen renders a comment that
// has a newline as a block comment.
func comments(
	f *jen.File,
	lines ...string,
) {
	for _, l := range lines {
		f.Comment(l)
	}
}

// wrap breaks text into lines of at most width characters, at spaces, for
// a comment.
func wrap(
	text string,
	width int,
) []string {
	var lines []string
	line := ""
	for _, w := range strings.Fields(text) {
		if line != "" && len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += w
	}

	return append(lines, line)
}

// multi lays out a group one item to a line, as the house style writes
// parameter lists and literals.
func multi(
	open string,
	closing string,
	items ...jen.Code,
) *jen.Statement {
	return jen.Custom(
		jen.Options{Open: open, Close: closing, Separator: ",", Multi: true},
		items...)
}

// params is a parameter list, one parameter to a line.
func params(
	items ...jen.Code,
) *jen.Statement {
	return multi("(", ")", items...)
}

// field is a key and its value in a composite literal.
func field(
	key string,
	value jen.Code,
) *jen.Statement {
	return jen.Id(key).Op(":").Add(value)
}

// tag is a struct tag with its keys in the order given. jen's Tag sorts the
// keys, and the house order is json, then validate, then doc.
func tag(
	pairs ...string,
) *jen.Statement {
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, fmt.Sprintf("%s:%q", pairs[i], pairs[i+1]))
	}

	return jen.Op("`" + strings.Join(parts, " ") + "`")
}

// octal is a file mode, written as Go code writes one.
func octal(
	mode int,
) *jen.Statement {
	return jen.Op(fmt.Sprintf("0o%o", mode))
}

// suiteRunner adds the test function that runs a suite.
func suiteRunner(
	f *jen.File,
	name string,
) {
	f.Line()
	f.Func().Id("Test" + name).Add(params(jen.Id("t").Op("*").Qual("testing", "T"))).Block(
		jen.Qual(suitePath, "Run").Call(jen.Id("t"), jen.New(jen.Id(name))),
	)
}

// suiteMethod adds a method of the test suite s.
func suiteMethod(
	f *jen.File,
	suite string,
	name string,
	body ...jen.Code,
) {
	f.Func().Params(jen.Id("s").Op("*").Id(suite)).Id(name).Params().Block(body...)
}

// s is a call on the suite's s, such as s.Equal(a, b).
func s(
	method string,
	args ...jen.Code,
) *jen.Statement {
	return jen.Id("s").Dot(method).Call(args...)
}

func (sc scaffold) behaviorFile() *jen.File {
	f := sc.newFile()
	f.PackageComment(
		fmt.Sprintf("Package %s is the %s behavior: %s.", sc.goPackage, sc.name, sc.doc),
	)
	b := func(name string) *jen.Statement { return jen.Qual(behaviorPath, name) }
	// next breaks a method chain after its dot, one method to a line.
	next := func(method string, args ...jen.Code) *jen.Statement {
		return jen.Op(".").Line().Id(method).Call(args...)
	}
	example := func(value string) jen.Code { return field("Value", jen.Lit(value)) }

	opts := []jen.Code{jen.Lit(sc.name).Op(",").Lit(sc.doc), b("Node")}
	opts = append(opts,
		b("API"),
		// The scaffold writes a config file, so its blocks go under Config in
		// the UI; the author moves it to the area that fits.
		b("Area").Call(jen.Qual(areaPath, "Config")),
		b("CLI"),
		b("Config").Call(jen.Id("defaults")),
	)
	if len(sc.requires) > 0 {
		required := make([]jen.Code, 0, len(sc.requires))
		for _, r := range sc.requires {
			required = append(required, jen.Qual(r.path, "Behavior"))
		}
		opts = append(opts, b("Requires").Call(required...))
	}
	opts = append(opts,
		b("Platform").Call(b("PlatformDebian"), jen.Id("newDebian")),
		b("Writes").Call(jen.Id("defaultPath")),
		b("Read").Call(jen.Lit("get"), jen.Id("Get")).
			Add(next("Describe", jen.Lit("Get the "+sc.name+" value"))).
			Add(next("Example", jen.Lit("Before anything is set"),
				jen.Id("GetInput").Values(), jen.Id("Entry").Values())).
			Add(next("Example", jen.Lit("The current value"), jen.Id("GetInput").Values(),
				jen.Id("Entry").Values(example("example")),
				jen.Line().Add(b("After").Call(jen.Id("UpdateInput").Values(example("example")))).
					Op(",").Line())),
		b("Write").Call(jen.Lit("update"), jen.Id("Update")).
			Add(next("Describe", jen.Lit("Set the "+sc.name+" value"))).
			Add(next("Explain",
				jen.Lit("Writes the value to behaviors."+sc.name+".path. Setting the value it ").
					Op("+").Line().Lit("already has reports changed: false."))).
			Add(next("Example", jen.Lit("A new value"), jen.Id("UpdateInput").Values(example("example")),
				jen.Id("Result").Values(field("Changed", jen.True())),
				jen.Line().Add(b("Wrote").Call(jen.Id("defaultPath"), jen.Lit("example"))).
					Op(",").Line().
					Add(b("Reads").Call(jen.Id("GetInput").Values(), jen.Id("Entry").Values(example("example")))).
					Op(",").Line())),
	)

	comments(f,
		"Behavior is the "+sc.name+" definition. The routes, the agent's dispatch, the",
		"CLI commands, the spec, the docs pages and the tests of every layer above",
		"the provider are built from it. osapi dsl behavior spec get lists every option.",
	)
	f.Var().
		Id("Behavior").
		Op("=").
		Add(b("Register").Call(b("Define").Add(multi("(", ")", opts...))))

	return f
}

func (sc scaffold) typesFile() *jen.File {
	f := sc.newFile()
	comments(f,
		"GetInput is get's input. Each field becomes a flag, a request field and a",
		"spec property; this one has none.",
	)
	f.Type().Id("GetInput").Struct()
	f.Line()
	comments(f,
		"UpdateInput is update's input. update has no key, so the field that says",
		`what it changes is tagged telemetry:"attr", and the trace shows it.`,
		"osapi dsl behavior spec get lists each tag.",
		"",
		"A field that takes one of a closed set of values, such as a state or a",
		"mode, is a named type with Values and Valid, and validate:\"enum\": use the",
		"type of pkg/sdk/types when it has one (osapi dsl behavior spec get lists them), or",
		"declare one here. validate refuses oneof.",
	)
	f.Type().Id("UpdateInput").Struct(
		jen.Id("Value").String().Add(tag("json", "value", "validate", "required",
			"doc", "The value to set", "telemetry", "attr")),
	)
	f.Line()
	comments(f, "Entry is get's output. Each field becomes a column and a JSON field.")
	f.Type().Id("Entry").Struct(
		jen.Id("Value").String().Add(tag("json", "value", "doc", "The current value")),
	)
	f.Line()
	comments(f, "Result is update's output. Every write reports whether it changed the host.")
	f.Type().Id("Result").Struct(
		jen.Id("Changed").Bool().Add(tag("json", "changed", "doc", "Whether the host changed")),
	)

	return f
}

func (sc scaffold) providerFile() *jen.File {
	f := sc.newFile()
	ctx := jen.Id("ctx").Qual("context", "Context")
	comments(f,
		"Provider is what each platform implements. A platform's implementation is",
		"<platform>.go, such as debian.go, declared in the definition with",
		"behavior.Platform; on a platform without one, the methods report skipped.",
		"mocks/ generates its double for the methods' tests.",
	)
	f.Type().Id("Provider").Interface(
		jen.Id("Get").Params(ctx).Params(jen.String(), jen.Error()),
		jen.Id("Set").Params(ctx, jen.Id("value").String()).
			Params(jen.Id("changed").Bool(), jen.Id("err").Error()),
	)

	return f
}

func (sc scaffold) settingsFile() *jen.File {
	f := sc.newFile()
	comments(f, "defaultPath is where the value is kept unless the config says otherwise.")
	f.Const().Id("defaultPath").Op("=").Lit("/etc/osapi/" + sc.name)
	f.Line()
	comments(f,
		"Settings is the agent's standing config for "+sc.name+", under",
		"behaviors."+sc.name+". Each field gets an OSAPI_BEHAVIORS_"+
			strings.ToUpper(sc.name)+"_<FIELD>",
		"variable and an osapi agent start flag. A value that changes per call",
		"belongs in an input type, not here.",
	)
	f.Type().Id("Settings").Struct(
		jen.Id("Path").String().Add(tag("mapstructure", "path",
			"doc", "File the "+sc.name+" value is kept in.",
			"validate", "required,startswith=/")),
	)
	f.Line()
	f.Var().Id("defaults").Op("=").Id("Settings").Values(field("Path", jen.Id("defaultPath")))

	return f
}

func (sc scaffold) methodsFile() *jen.File {
	f := sc.newFile()
	ctx := jen.Id("ctx").Qual("context", "Context")
	failed := func(zero jen.Code) jen.Code {
		return jen.If(jen.Err().Op("!=").Nil()).Block(jen.Return(zero, jen.Err()))
	}

	comments(f, "Get reads the value from the host.")
	f.Func().Id("Get").Add(params(ctx, jen.Id("p").Id("Provider"), jen.Id("_").Id("GetInput"))).
		Params(jen.Id("Entry"), jen.Error()).Block(
		jen.List(jen.Id("value"), jen.Err()).Op(":=").Id("p").Dot("Get").Call(jen.Id("ctx")),
		failed(jen.Id("Entry").Values()),
		jen.Line(),
		jen.Return(jen.Id("Entry").Values(field("Value", jen.Id("value"))), jen.Nil()),
	)
	f.Line()
	comments(f, "Update sets the value on the host.")
	f.Func().
		Id("Update").
		Add(params(ctx, jen.Id("p").Id("Provider"), jen.Id("in").Id("UpdateInput"))).
		Params(jen.Id("Result"), jen.Error()).
		Block(
			jen.List(jen.Id("changed"), jen.Err()).Op(":=").Id("p").Dot("Set").
				Call(jen.Id("ctx"), jen.Id("in").Dot("Value")),
			failed(jen.Id("Result").Values()),
			jen.Line(),
			jen.Return(jen.Id("Result").Values(field("Changed", jen.Id("changed"))), jen.Nil()),
		)

	return f
}

func (sc scaffold) debianFile() *jen.File {
	f := sc.newFile()
	f.ImportName(avfsPath, "avfs")
	f.ImportName(hostPath, "host")
	f.ImportName(atomicfilePath, "atomicfile")
	d := func() *jen.Statement { return jen.Id("d") }
	recv := jen.Params(jen.Id("d").Op("*").Id("debian"))
	ctx := jen.Id("_").Qual("context", "Context")
	setCtx := jen.Id("ctx").Qual("context", "Context")

	comments(f,
		"debian keeps the value in a file. debian_public_test.go tests it on an",
		"in-memory filesystem, and the examples in behavior.go run it on a fake",
		"host.",
	)
	fields := []jen.Code{
		jen.Id("fs").Qual(avfsPath, "VFS"),
		jen.Id("path").String(),
	}
	values := make([]jen.Code, 0, 2+len(sc.requires))
	values = append(values,
		field("fs", jen.Id("h").Dot("FS")),
		field("path", jen.Qual(hostPath, "Settings").Index(jen.Id("Settings")).
			Call(jen.Id("h")).Dot("Path")),
	)
	var body []jen.Code
	var missing jen.Code
	for _, r := range sc.requires {
		v := r.pkg + "Provider"
		doc := wrap(r.pkg+" is the provider of the "+r.name+" behavior, which this one "+
			"requires. Call its methods with behavior.Call, so the trace shows each call "+
			"as a step:", 75)
		for _, l := range doc {
			fields = append(fields, jen.Comment(l))
		}
		fields = append(fields,
			jen.Comment("//"),
			jen.Comment("//\tbehavior.Call(ctx, "+r.pkg+".Behavior, "+r.pkg+"."+r.fn+
				", d."+r.pkg+", in)"),
			jen.Id(r.pkg).Qual(r.path, "Provider"),
		)
		values = append(values, field(r.pkg, jen.Id(v)))
		body = append(body, jen.Id(v).Op(":=").Qual(hostPath, "Require").
			Types(jen.Qual(r.path, "Provider")).Call(jen.Id("h"), jen.Qual(r.path, "Behavior")))
		isNil := jen.Id(v).Op("==").Nil()
		if missing == nil {
			missing = isNil
		} else {
			missing = jen.Add(missing).Op("||").Line().Add(isNil)
		}
	}
	f.Type().Id("debian").Struct(fields...)
	f.Line()
	doc := "newDebian is the provider on Debian and its family, built by the agent at " +
		"startup with the host and the behavior's settings."
	literal := jen.Op("&").Id("debian").Values(values...)
	if missing != nil {
		doc += " It declines a host where a behavior it requires has no provider, and " +
			"the methods report skipped there."
		literal = jen.Op("&").Id("debian").Add(multi("{", "}", values...))
		body = append(body,
			jen.If(missing).Block(jen.Return(jen.Nil(), jen.Qual(behaviorPath, "ErrUnsupported"))),
			jen.Line(),
		)
	}
	comments(f, wrap(doc, 77)...)
	body = append(body, jen.Return(literal, jen.Nil()))
	f.Func().Id("newDebian").Add(params(jen.Id("h").Qual(hostPath, "Host"))).
		Params(jen.Id("Provider"), jen.Error()).Block(body...)
	f.Line()
	comments(f, "Get reads the value; a value never set reads as empty.")
	f.Func().Add(recv.Clone()).Id("Get").Add(params(ctx.Clone())).
		Params(jen.String(), jen.Error()).Block(
		jen.List(jen.Id("data"), jen.Err()).Op(":=").Add(d()).Dot("fs").Dot("ReadFile").
			Call(d().Dot("path")),
		jen.If(jen.Qual("errors", "Is").Call(jen.Err(), jen.Qual("io/fs", "ErrNotExist"))).Block(
			jen.Return(jen.Lit(""), jen.Nil()),
		),
		jen.Line(),
		jen.Return(jen.String().Call(jen.Id("data")), jen.Err()),
	)
	f.Line()
	comments(f, "Set writes the value, and reports whether it changed.")
	f.Func().Add(recv.Clone()).Id("Set").Add(params(setCtx, jen.Id("value").String())).
		Params(jen.Bool(), jen.Error()).Block(
		jen.If(
			jen.List(jen.Id("current"), jen.Err()).Op(":=").Add(d()).Dot("fs").Dot("ReadFile").
				Call(d().Dot("path")),
			jen.Err().Op("==").Nil().Op("&&").String().Call(jen.Id("current")).Op("==").Id("value"),
		).Block(jen.Return(jen.False(), jen.Nil())),
		jen.Comment("A directory that cannot be made fails the write below, which reports it."),
		jen.Id("_").Op("=").Add(d()).Dot("fs").Dot("MkdirAll").
			Call(jen.Qual("path/filepath", "Dir").Call(d().Dot("path")), octal(0o755)),
		jen.Line(),
		jen.Return(jen.True(), jen.Qual(atomicfilePath, "Write").Call(jen.Id("ctx"), d().Dot("fs"),
			d().Dot("path"), jen.Index().Byte().Call(jen.Id("value")), octal(0o644))),
	)

	return f
}

func (sc scaffold) mocksFile() *jen.File {
	f := jen.NewFilePathName(sc.pkg()+"/mocks", "mocks")
	f.HeaderComment(license)
	f.PackageComment("Package mocks holds the generated double of the " + sc.name +
		" provider, for the")
	f.PackageComment("methods' tests.")
	f.Line()
	// A comment that starts with // is written as it is; jen puts a space
	// after the // of any other, which a directive cannot have.
	f.Comment("//go:generate go tool go.uber.org/mock/mockgen -source=../provider.go " +
		"-destination=provider.gen.go -package=mocks")

	return f
}

func (sc scaffold) behaviorTestFile() *jen.File {
	f := sc.newTestFile()
	const suite = "BehaviorPublicTestSuite"
	def := jen.Qual(sc.pkg(), "Behavior")

	f.Type().Id(suite).Struct(jen.Qual(suitePath, "Suite"))
	f.Line()
	suiteMethod(f, suite, "TestBehavior",
		s("Same", def.Clone(), jen.Qual(behaviorPath, "Lookup").Call(jen.Lit(sc.name))),
		s("Empty", jen.Qual(behaviorPath, "ValidateBehavior").Call(def.Clone())),
	)
	f.Line()
	comments(f,
		"TestExamples runs every example and refusal on a fake host with each",
		"platform's provider, through the methods, and checks what each changes.",
	)
	suiteMethod(f, suite, "TestExamples",
		jen.Qual(behaviortestPath, "Run").Call(s("T"), def.Clone()),
	)
	suiteRunner(f, suite)

	return f
}

// table is a table test: tests := []struct{...}{...}, one case to a block.
func table(
	fields []jen.Code,
	cases ...[]jen.Code,
) jen.Code {
	rows := make([]jen.Code, 0, len(cases))
	for _, c := range cases {
		rows = append(rows, multi("{", "}", c...))
	}

	return jen.Id("tests").Op(":=").Index().Struct(fields...).Add(multi("{", "}", rows...))
}

// eachCase runs body as a subtest for each case of the table.
func eachCase(
	body ...jen.Code,
) jen.Code {
	return jen.For(jen.List(jen.Id("_"), jen.Id("tc")).Op(":=").Range().Id("tests")).Block(
		s("Run", jen.Id("tc").Dot("name"), jen.Func().Params().Block(body...)),
	)
}

func (sc scaffold) methodsTestFile() *jen.File {
	f := sc.newTestFile()
	const suite = "MethodsPublicTestSuite"
	q := func(name string) *jen.Statement { return jen.Qual(sc.pkg(), name) }
	bg := jen.Qual("context", "Background").Call()
	fails := []jen.Code{
		field("name", jen.Lit("when the provider fails")),
		field("err", jen.Qual("errors", "New").Call(jen.Lit("permission denied"))),
	}
	check := []jen.Code{
		jen.Line(),
		jen.If(jen.Id("tc").Dot("err").Op("!=").Nil()).Block(
			s("ErrorIs", jen.Err(), jen.Id("tc").Dot("err")),
			jen.Line(),
			jen.Return(),
		),
		s("Require").Dot("NoError").Call(jen.Err()),
		s("Equal", jen.Id("tc").Dot("want"), jen.Id("got")),
	}

	f.Type().Id(suite).Struct(
		jen.Qual(suitePath, "Suite"),
		jen.Line(),
		jen.Id("ctrl").Op("*").Qual(gomockPath, "Controller"),
		jen.Id("mock").Op("*").Qual(sc.pkg()+"/mocks", "MockProvider"),
	)
	f.Line()
	suiteMethod(f, suite, "SetupSubTest",
		jen.Id("s").Dot("ctrl").Op("=").Qual(gomockPath, "NewController").Call(s("T")),
		jen.Id("s").Dot("mock").Op("=").Qual(sc.pkg()+"/mocks", "NewMockProvider").
			Call(jen.Id("s").Dot("ctrl")),
	)
	f.Line()
	suiteMethod(f, suite, "TearDownSubTest", jen.Id("s").Dot("ctrl").Dot("Finish").Call())
	f.Line()
	suiteMethod(f, suite, "TestGet",
		table(
			[]jen.Code{
				jen.Id("name").String(),
				jen.Id("value").String(),
				jen.Id("err").Error(),
				jen.Id("want").Add(q("Entry")),
			},
			[]jen.Code{
				field("name", jen.Lit("when the provider reads the value")),
				field("value", jen.Lit("hello")),
				field("want", q("Entry").Values(field("Value", jen.Lit("hello")))),
			},
			fails,
		),
		jen.Line(),
		eachCase(append([]jen.Code{
			jen.Id("s").Dot("mock").Dot("EXPECT").Call().Dot("Get").
				Call(jen.Qual(gomockPath, "Any").Call()).
				Dot("Return").Call(jen.Id("tc").Dot("value"), jen.Id("tc").Dot("err")),
			jen.Line(),
			jen.List(jen.Id("got"), jen.Err()).Op(":=").Add(q("Get")).
				Call(bg.Clone(), jen.Id("s").Dot("mock"), q("GetInput").Values()),
		}, check...)...),
	)
	f.Line()
	suiteMethod(f, suite, "TestUpdate",
		table(
			[]jen.Code{
				jen.Id("name").String(),
				jen.Id("changed").Bool(),
				jen.Id("err").Error(),
				jen.Id("want").Add(q("Result")),
			},
			[]jen.Code{
				field("name", jen.Lit("when the provider writes the value")),
				field("changed", jen.True()),
				field("want", q("Result").Values(field("Changed", jen.True()))),
			},
			[]jen.Code{
				field("name", jen.Lit("when the value is already set")),
				field("want", q("Result").Values()),
			},
			fails,
		),
		jen.Line(),
		eachCase(append([]jen.Code{
			jen.Id("s").Dot("mock").Dot("EXPECT").Call().Dot("Set").
				Call(jen.Qual(gomockPath, "Any").Call(), jen.Lit("hello")).
				Dot("Return").Call(jen.Id("tc").Dot("changed"), jen.Id("tc").Dot("err")),
			jen.Line(),
			jen.List(jen.Id("got"), jen.Err()).Op(":=").Add(q("Update")).Add(multi("(", ")",
				bg.Clone(),
				jen.Id("s").Dot("mock"),
				q("UpdateInput").Values(field("Value", jen.Lit("hello"))),
			)),
		}, check...)...),
	)
	suiteRunner(f, suite)

	return f
}

func (sc scaffold) debianTestFile() *jen.File {
	f := sc.newTestFile()
	f.ImportName(memfsPath, "memfs")
	f.ImportName(hostPath, "host")
	const suite = "DebianPublicTestSuite"
	q := func(name string) *jen.Statement { return jen.Qual(sc.pkg(), name) }
	bg := jen.Qual("context", "Background").Call()
	provider := func() *jen.Statement { return jen.Id("s").Dot("provider") }
	setHello := jen.Func().Params().Block(
		jen.List(jen.Id("_"), jen.Err()).Op(":=").Add(provider()).Dot("Set").
			Call(bg.Clone(), jen.Lit("hello")),
		s("Require").Dot("NoError").Call(jen.Err()),
	)
	noSetup := jen.Func().Params().Block()

	f.Type().Id(suite).Struct(
		jen.Qual(suitePath, "Suite"),
		jen.Line(),
		jen.Id("provider").Add(q("Provider")),
	)
	f.Line()
	hostFields := []jen.Code{
		field("FS", jen.Qual(memfsPath, "New").Call()),
		field(
			"Settings",
			q("Settings").Values(field("Path", jen.Lit("/etc/osapi/"+sc.name))),
		),
	}
	if len(sc.requires) > 0 {
		// Each required behavior's provider is a mock, which the tests of
		// the calls that the author writes set up.
		required := make([]jen.Code, 0, len(sc.requires))
		for _, r := range sc.requires {
			required = append(required, jen.Qual(r.path, "Behavior").Dot("Name").Op(":").
				Qual(r.path+"/mocks", "NewMockProvider").
				Call(jen.Id("ctrl")))
		}
		hostFields = append(hostFields,
			field("Required", jen.Map(jen.String()).Any().Add(multi("{", "}", required...))))
	}
	var setup []jen.Code
	if len(sc.requires) > 0 {
		setup = append(setup,
			jen.Id("ctrl").Op(":=").Qual(gomockPath, "NewController").Call(s("T")))
	}
	suiteMethod(f, suite, "SetupSubTest", append(setup,
		jen.List(jen.Id("p"), jen.Err()).Op(":=").Add(q("Behavior")).Dot("Provider").Call(
			jen.Lit("debian"),
			jen.Qual(hostPath, "Host").Add(multi("{", "}", hostFields...)),
		),
		s("Require").Dot("NoError").Call(jen.Err()),
		provider().Op("=").Id("p").Assert(q("Provider")),
	)...)
	if len(sc.requires) > 0 {
		f.Line()
		names := make([]string, 0, len(sc.requires))
		for _, r := range sc.requires {
			names = append(names, r.name)
		}
		comments(f,
			"TestNewDebian: the provider requires "+strings.Join(names, " and ")+
				", so on a host where",
			"one has no provider it declines, and the methods report skipped.",
		)
		suiteMethod(f, suite, "TestNewDebian",
			jen.List(jen.Id("_"), jen.Err()).Op(":=").Add(q("Behavior")).Dot("Provider").
				Call(jen.Lit("debian"), jen.Qual(hostPath, "Host").Values()),
			jen.Line(),
			s("ErrorIs", jen.Err(), jen.Qual(behaviorPath, "ErrUnsupported")),
		)
	}
	f.Line()
	suiteMethod(f, suite, "TestGet",
		table(
			[]jen.Code{
				jen.Id("name").String(),
				jen.Id("setup").Func().Params(),
				jen.Id("want").String(),
			},
			[]jen.Code{
				field("name", jen.Lit("when the value was never set, it reads as empty")),
				field("setup", noSetup.Clone()),
			},
			[]jen.Code{
				field("name", jen.Lit("when the value was set, it reads back")),
				field("setup", setHello.Clone()),
				field("want", jen.Lit("hello")),
			},
		),
		jen.Line(),
		eachCase(
			jen.Id("tc").Dot("setup").Call(),
			jen.Line(),
			jen.List(jen.Id("got"), jen.Err()).Op(":=").Add(provider()).Dot("Get").Call(bg.Clone()),
			jen.Line(),
			s("Require").Dot("NoError").Call(jen.Err()),
			s("Equal", jen.Id("tc").Dot("want"), jen.Id("got")),
		),
	)
	f.Line()
	suiteMethod(f, suite, "TestSet",
		table(
			[]jen.Code{
				jen.Id("name").String(),
				jen.Id("setup").Func().Params(),
				jen.Id("wantChanged").Bool(),
			},
			[]jen.Code{
				field("name", jen.Lit("when the value is new, it is written")),
				field("setup", noSetup.Clone()),
				field("wantChanged", jen.True()),
			},
			[]jen.Code{
				field("name", jen.Lit("when the file already holds the value, nothing changes")),
				field("setup", setHello.Clone()),
			},
		),
		jen.Line(),
		eachCase(
			jen.Id("tc").Dot("setup").Call(),
			jen.Line(),
			jen.List(jen.Id("changed"), jen.Err()).Op(":=").Add(provider()).Dot("Set").
				Call(bg.Clone(), jen.Lit("hello")),
			jen.Line(),
			s("Require").Dot("NoError").Call(jen.Err()),
			s("Equal", jen.Id("tc").Dot("wantChanged"), jen.Id("changed")),
		),
	)
	suiteRunner(f, suite)

	return f
}
