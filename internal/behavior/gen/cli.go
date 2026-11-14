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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
)

const (
	cobraPath       = "github.com/spf13/cobra"
	behaviorcmdPath = "github.com/osapi-io/osapi/internal/behaviorcmd"
	cmdtestPath     = "github.com/osapi-io/osapi/internal/behavior/behaviorcmdtest"
	suitePath       = "github.com/stretchr/testify/suite"
)

// flagFuncs maps a field's kind to the pflag function that declares its
// flag and the getter that reads it back. The getter returns the field's own
// type, so the value needs no conversion.
var flagFuncs = map[reflect.Kind][2]string{
	reflect.String:  {"String", "GetString"},
	reflect.Bool:    {"Bool", "GetBool"},
	reflect.Int:     {"Int", "GetInt"},
	reflect.Int32:   {"Int32", "GetInt32"},
	reflect.Int64:   {"Int64", "GetInt64"},
	reflect.Uint:    {"Uint", "GetUint"},
	reflect.Uint32:  {"Uint32", "GetUint32"},
	reflect.Uint64:  {"Uint64", "GetUint64"},
	reflect.Float64: {"Float64", "GetFloat64"},
}

// cliFlag is one input field's flag.
type cliFlag struct {
	behaviorcmd.Flag
	field reflect.StructField
	local string
}

// flags are m's input fields as flags, in declaration order.
func flags(
	b *behavior.Behavior,
	m *behavior.Method,
) ([]cliFlag, error) {
	var out []cliFlag
	for _, fl := range behaviorcmd.Flags(m.In) {
		field := m.In.Field(fl.Index)
		if _, ok := flagFuncs[field.Type.Kind()]; !ok {
			return nil, fmt.Errorf(
				"%s: field %s is a %s, which no flag can hold",
				b.Operation(m), field.Name, field.Type,
			)
		}
		out = append(out, cliFlag{Flag: fl, field: field, local: camel(fl.JSON)})
	}

	return out, nil
}

// cliFile is one behavior's generated commands, written the way the
// hand-written commands in cmd/ are: a package-level cobra.Command per
// command, an init that adds each to its parent and declares its flags, and
// a Run that reads the flags back and builds the method's typed input.
func cliFile(
	b *behavior.Behavior,
) (*jen.File, error) {
	pkg := pkgPath(b)
	f := genFile(b)
	for _, line := range wrap("Package gen holds what osapi generates from the "+b.Name+
		" behavior: its CLI commands in cli.go, its API routes in api.go and api_<method>.go, "+
		"and its part of the OpenAPI spec in api.yaml.", 76) {
		f.PackageComment(line)
	}

	parent := camel(b.Group, b.Name) + "Cmd"
	f.Var().Id(parent).Op("=").Op("&").Qual(cobraPath, "Command").Values(jen.Dict{
		jen.Id("Use"):   jen.Lit(b.Name),
		jen.Id("Short"): jen.Lit(b.Doc),
	})

	initBody := []jen.Code{
		jen.Qual(behaviorcmdPath, "Register").Call(jen.Lit(b.Group), jen.Id(parent)),
	}

	for _, m := range b.Methods {
		if m.In.PkgPath() != pkg {
			return nil, fmt.Errorf(
				"%s: input %s is not in %s; a behavior's types live in its own package",
				b.Operation(m), m.In, pkg,
			)
		}
		fls, err := flags(b, m)
		if err != nil {
			return nil, err
		}

		cmd := camel(b.Group, b.Name, m.Name) + "Cmd"
		var run []jen.Code
		values := jen.Dict{}
		for _, fl := range fls {
			run = append(run, jen.List(jen.Id(fl.local), jen.Id("_")).Op(":=").
				Id("cmd").
				Dot("Flags").Call().Dot(flagFuncs[fl.field.Type.Kind()][1]).Call(jen.Lit(fl.Name)))
			values[jen.Id(fl.field.Name)] = jen.Id(fl.local)
		}
		if len(run) > 0 {
			run = append(run, jen.Line())
		}
		run = append(run, jen.Qual(behaviorcmdPath, "Run").Call(
			jen.Id("cmd"),
			jen.Qual(pkg, "Behavior"),
			jen.Lit(m.Name),
			jen.Qual(pkg, m.In.Name()).Values(values),
		))

		fields := jen.Dict{
			jen.Id("Use"):   jen.Lit(m.Name),
			jen.Id("Short"): jen.Lit(m.Doc),
			jen.Id("Run"): jen.Func().Params(
				jen.Id("cmd").Op("*").Qual(cobraPath, "Command"),
				jen.Id("_").Index().String(),
			).Block(run...),
		}
		if m.Long != "" {
			fields[jen.Id("Long")] = jen.Lit(m.Long)
		}
		f.Line()
		f.Var().Id(cmd).Op("=").Op("&").Qual(cobraPath, "Command").Values(fields)

		initBody = append(initBody, jen.Line(), jen.Id(parent).Dot("AddCommand").Call(jen.Id(cmd)))
		for _, fl := range fls {
			kind := fl.field.Type.Kind()
			initBody = append(initBody, jen.Id(cmd).Dot("PersistentFlags").Call().
				Dot(flagFuncs[kind][0]).
				Call(jen.Lit(fl.Name), zero(kind), jen.Lit(fl.Usage)))
		}
		for _, fl := range fls {
			if fl.Required {
				initBody = append(initBody, jen.Id("_").Op("=").Id(cmd).
					Dot("MarkPersistentFlagRequired").Call(jen.Lit(fl.Name)))
			}
		}
	}

	f.Line()
	f.Func().Id("init").Params().Block(initBody...)

	return f, nil
}

// cliTestFile is the generated cli.go's test: each example's flags, given to
// its command, send the example's request.
func cliTestFile(
	b *behavior.Behavior,
) (*jen.File, error) {
	f := genTestFile(b)

	var rows []jen.Code
	for _, m := range b.Methods {
		cmdPath := strings.Fields(strings.TrimPrefix(b.CLICommand(m), behavior.CommandPrefix+" "))[1:]
		for _, ex := range m.Examples {
			method, path, body, err := behaviorcmd.Build(b, m, "server1", ex.Input)
			if err != nil {
				return nil, fmt.Errorf("%s: example %q: %w", b.Operation(m), ex.Doc, err)
			}
			want := jen.Qual(cmdtestPath, "Request").Add(multi("{", "}",
				jen.Id("Method").Op(":").Lit(method),
				jen.Id("Path").Op(":").Lit(path),
				jen.Id("Body").Op(":").Add(raw(string(body))),
			))
			args := append(append([]string{}, cmdPath...), "--target", "server1")
			args = append(args, behaviorcmd.Args(ex.Input)...)
			argList := make([]jen.Code, 0, len(args))
			for _, a := range args {
				argList = append(argList, jen.Lit(a))
			}
			rows = append(rows, multi("{", "}",
				jen.Id("name").Op(":").Lit("when the example "+strconv.Quote(ex.Doc)+" is run"),
				jen.Id("args").Op(":").Index().String().Values(argList...),
				jen.Id("want").Op(":").Add(want),
			))
		}
	}

	f.Type().Id("CLIPublicTestSuite").Struct(jen.Qual(suitePath, "Suite"))
	f.Line()
	comment(f, "TestCommands runs each example's command line and checks the request it sends.")
	f.Func().Params(jen.Id("s").Op("*").Id("CLIPublicTestSuite")).Id("TestCommands").Params().Block(
		jen.Id("tests").Op(":=").Index().Struct(
			jen.Id("name").String(),
			jen.Id("args").Index().String(),
			jen.Id("want").Qual(cmdtestPath, "Request"),
		).Add(multi("{", "}", rows...)),
		jen.Line(),
		jen.For(jen.List(jen.Id("_"), jen.Id("tc")).Op(":=").Range().Id("tests")).Block(
			jen.Id("s").Dot("Run").Call(jen.Id("tc").Dot("name"), jen.Func().Params().Block(
				jen.Id("s").Dot("Equal").Call(
					jen.Id("tc").Dot("want"),
					jen.Qual(cmdtestPath, "Sent").
						Call(jen.Id("s").Dot("T").Call(), jen.Id("tc").Dot("args")),
				),
			)),
		),
	)
	suiteRunner(f, "CLIPublicTestSuite")

	return f, nil
}

// allFile is the one file that imports every behavior's generated code, so
// a binary needs a single import to get every behavior's commands and
// routes.
func allFile(
	bs []*behavior.Behavior,
) *jen.File {
	f := jen.NewFilePathName(module+"/internal/behaviors/all/gen", "gen")
	f.HeaderComment(generated)
	for _, line := range wrap("Package gen imports the generated code of every behavior with an "+
		"API or a CLI: its commands and its routes.", 76) {
		f.PackageComment(line)
	}
	for _, b := range bs {
		if b.API || b.CLI {
			f.Anon(pkgPath(b) + "/gen")
		}
	}

	return f
}

func zero(
	kind reflect.Kind,
) jen.Code {
	switch kind {
	case reflect.String:
		return jen.Lit("")
	case reflect.Bool:
		return jen.False()
	default:
		return jen.Lit(0)
	}
}

// camel joins words as a Go identifier: sysctl, get is sysctlGet.
func camel(
	words ...string,
) string {
	var sb strings.Builder
	for _, w := range words {
		for _, part := range strings.FieldsFunc(w, func(r rune) bool { return r == '_' || r == '-' }) {
			if sb.Len() == 0 {
				sb.WriteString(strings.ToLower(part[:1]) + part[1:])
			} else {
				sb.WriteString(strings.ToUpper(part[:1]) + part[1:])
			}
		}
	}

	return sb.String()
}
