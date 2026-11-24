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
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/sdk/gen/sdkmodel"
)

// exampleDir is the module of the SDK's examples, from the root of the
// repository.
const exampleDir = "examples/sdk/client"

// call is one example of a method, as the code that makes the call.
type call struct {
	// Name runs the example, such as update or update-2.
	Name string
	// Func is the name of the example's function, such as runUpdate2.
	Func string
	// Doc says what the example shows.
	Doc string
	// Method is the method the example calls.
	Method *sdkmodel.Method
	// Code is the call, such as c.NetworkDNS.Get(ctx, target, "eth0").
	Code *jen.Statement
	// Pre is the statements before the call, such as the read of the file
	// that an upload sends, or nil.
	Pre []jen.Code
}

// calls is each example of each method of s that the SDK has, in the order
// of the methods.
func calls(
	s *sdkmodel.Service,
) ([]*call, error) {
	var out []*call
	for _, m := range s.Methods {
		for i, ex := range m.Examples {
			code, pre, err := callCode(s, m, ex)
			if err != nil {
				return nil, fmt.Errorf("%s.%s: example %q: %w", s.Name, m.Name, ex.Doc, err)
			}
			c := &call{
				Name:   strings.ToLower(m.Name[:1]) + m.Name[1:],
				Func:   "run" + m.Name,
				Doc:    ex.Doc,
				Method: m,
				Code:   code,
				Pre:    pre,
			}
			if i > 0 {
				c.Name += "-" + strconv.Itoa(i+1)
				c.Func += strconv.Itoa(i + 1)
			}
			out = append(out, c)
		}
	}

	return out, nil
}

// callCode is the call of m with the input of ex, and the statements that
// come before it. The target of a node's method is the variable target,
// which the program reads from OSAPI_TARGET.
func callCode(
	s *sdkmodel.Service,
	m *sdkmodel.Method,
	ex *sdkmodel.Example,
) (*jen.Statement, []jen.Code, error) {
	if m.Multipart {
		return formCall(s, m, ex)
	}
	code, err := jsonCall(s, m, ex)

	return code, nil, err
}

// formCall is the call of m, a method that sends a multipart form, with
// the input of ex, and the read of the file that it sends. The method has
// the arguments that FormArgs gives. A field that ex does not set has its
// default.
func formCall(
	s *sdkmodel.Service,
	m *sdkmodel.Method,
	ex *sdkmodel.Example,
) (*jen.Statement, []jen.Code, error) {
	fields, err := m.FormArgs()
	if err != nil {
		return nil, nil, err
	}
	args := []jen.Code{jen.Id("ctx")}
	var pre []jen.Code
	for _, f := range fields {
		v, ok := ex.Input[f.Name]
		switch {
		case f.File:
			path, isPath := v.(string)
			if !isPath {
				return nil, nil, fmt.Errorf("it has no path of %s", f.Name)
			}
			pre = []jen.Code{
				jen.List(jen.Id("content"), jen.Err()).Op(":=").
					Qual("os", "ReadFile").Call(jen.Lit(path)),
				jen.If(jen.Err().Op("!=").Nil()).Block(jen.Return(jen.Err())),
				jen.Line(),
			}
			args = append(args, jen.Qual("bytes", "NewReader").Call(jen.Id("content")))

			continue
		case !ok && f.Default != nil:
			v = f.Default
		case !ok && f.Required:
			return nil, nil, fmt.Errorf("it has no %s", f.Name)
		case !ok:
			v = zeroJSON(f.Type)
		}
		code, err := value(f.Type, v)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", f.Name, err)
		}
		args = append(args, code)
	}

	return jen.Id("c").Dot(s.Name).Dot(m.Name).Call(args...), pre, nil
}

// jsonCall is the call of m, a method that the generator writes, with the
// input of ex.
func jsonCall(
	s *sdkmodel.Service,
	m *sdkmodel.Method,
	ex *sdkmodel.Example,
) (*jen.Statement, error) {
	args := []jen.Code{jen.Id("ctx")}
	if t := m.Target(); t != nil {
		if s.Node {
			args = append(args, jen.Id("target"))
		} else {
			v, err := required(t, ex)
			if err != nil {
				return nil, err
			}
			args = append(args, v)
		}
	}
	if m.Input != nil {
		v, err := structValue(m.Input.Name, m.Input.Fields, ex)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	} else {
		for _, p := range append(m.Keys(), m.RequiredQuery()...) {
			v, err := required(p, ex)
			if err != nil {
				return nil, err
			}
			args = append(args, v)
		}
	}
	if opt := m.OptionalQuery(); len(opt) > 0 {
		fields := make([]*sdkmodel.Field, 0, len(opt))
		for _, p := range opt {
			fields = append(fields, &sdkmodel.Field{Name: p.Name, GoName: p.GoName, Type: p.Type})
		}
		v, err := structValue(s.Name+m.Name+"Params", fields, ex)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	return jen.Id("c").Dot(s.Name).Dot(m.Name).Call(args...), nil
}

// required is the value of p in ex, which the call must have.
func required(
	p *sdkmodel.Param,
	ex *sdkmodel.Example,
) (jen.Code, error) {
	v, ok := ex.Input[p.Name]
	if !ok {
		return nil, fmt.Errorf("it has no %s", p.Name)
	}

	return value(p.Type, v)
}

// zeroJSON is the value of JSON that a field of type t has when nothing sets
// it: "", false or 0 for a scalar, "" for a named type such as an enum, and
// nil for any other type, which value refuses.
func zeroJSON(
	t *sdkmodel.TypeRef,
) any {
	if t.Kind == sdkmodel.Named {
		return ""
	}
	if t.Kind != sdkmodel.Scalar {
		return nil
	}
	switch t.Name {
	case "string":
		return ""
	case "bool":
		return false
	}

	return float64(0)
}

// structValue is a value of the SDK's struct name, with the fields of ex
// that it has, in the order of fields.
func structValue(
	name string,
	fields []*sdkmodel.Field,
	ex *sdkmodel.Example,
) (jen.Code, error) {
	dict := jen.Dict{}
	for _, f := range fields {
		v, ok := ex.Input[f.Name]
		if !ok {
			continue
		}
		code, err := value(f.Type, v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name, err)
		}
		dict[jen.Id(f.GoName)] = code
	}

	return jen.Qual(pkg, name).Values(dict), nil
}

// value is v, a value of JSON, as a Go literal of type t.
func value(
	t *sdkmodel.TypeRef,
	v any,
) (jen.Code, error) {
	switch t.Kind {
	case sdkmodel.Scalar:
		return scalar(t.Name, v)
	case sdkmodel.Slice:
		items, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("%v is not a list", v)
		}
		elem, err := literalType(t.Elem)
		if err != nil {
			return nil, err
		}
		codes := make([]jen.Code, 0, len(items))
		for _, item := range items {
			code, err := value(t.Elem, item)
			if err != nil {
				return nil, err
			}
			codes = append(codes, code)
		}

		return jen.Index().Add(elem).Values(codes...), nil
	case sdkmodel.Map:
		pairs, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%v is not a map", v)
		}
		elem, err := literalType(t.Elem)
		if err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(pairs))
		for k := range pairs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		dict := jen.Dict{}
		for _, k := range keys {
			code, err := value(t.Elem, pairs[k])
			if err != nil {
				return nil, err
			}
			dict[jen.Lit(k)] = code
		}

		return jen.Map(jen.String()).Add(elem).Values(dict), nil
	case sdkmodel.Named:
		// A named string, such as the enum ContentType, is its value
		// converted to the type.
		if text, ok := v.(string); ok {
			at := pkg
			if shared[t.Name] {
				at = typesPkg
			}

			return jen.Qual(at, t.Name).Call(jen.Lit(text)), nil
		}
	}

	return nil, fmt.Errorf("a value of %s has no literal", t)
}

// literalType is the Go type t, a scalar, as an example writes it.
func literalType(
	t *sdkmodel.TypeRef,
) (jen.Code, error) {
	if t.Kind != sdkmodel.Scalar || t.Name == "time.Time" {
		return nil, fmt.Errorf("a value of %s has no literal", t)
	}

	return jen.Id(t.Name), nil
}

// scalar is v as a literal of the Go type name.
func scalar(
	name string,
	v any,
) (jen.Code, error) {
	switch name {
	case "string":
		if s, ok := v.(string); ok {
			return jen.Lit(s), nil
		}
	case "bool":
		if b, ok := v.(bool); ok {
			return jen.Lit(b), nil
		}
	case "int", "int64":
		if n, ok := v.(float64); ok && n == float64(int64(n)) {
			return jen.Lit(int(n)), nil
		}
	case "float64":
		if n, ok := v.(float64); ok {
			return jen.Lit(n), nil
		}
	}

	return nil, fmt.Errorf("%v is not a %s", v, name)
}

// exampleFile is the program of the examples of s, or nil when s has none.
func exampleFile(
	s *sdkmodel.Service,
) (*jen.File, error) {
	cs, err := calls(s)
	if err != nil || len(cs) == 0 {
		return nil, err
	}
	file := snake(s.Name) + ".go"
	f := jen.NewFile("main")
	f.ImportName(pkg, "client")
	f.HeaderComment(strings.TrimPrefix(header, "// "))
	doc := "Package main runs the examples of the SDK's " + s.Name + " service, " +
		"which calls the API's " + s.Name + " group: " + lowerFirst(sentence(s.Doc)) +
		" Each example calls a method with the input of an example of the method's " +
		"definition, and prints the answer as JSON."
	for _, line := range strings.Split(wrap(doc), "\n") {
		f.PackageComment(line)
	}
	f.PackageComment("")
	f.PackageComment("Give the name of an example to run it. With no name, the program lists them:")
	f.PackageComment("")
	f.PackageComment("\tOSAPI_TOKEN=\"<jwt>\" go run " + file + " " + cs[0].Name)
	f.PackageComment("")
	env := "OSAPI_URL is the controller, http://localhost:8080 when it is not set."
	if s.Node {
		env += " OSAPI_TARGET is the hosts each call goes to, _any when it is not set."
	}
	for _, line := range strings.Split(wrap(env), "\n") {
		f.PackageComment(line)
	}

	params := []jen.Code{
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("c").Op("*").Qual(pkg, "Client"),
	}
	if s.Node {
		params = append(params, jen.Id("target").String())
	}

	rows := make([]jen.Code, 0, len(cs))
	for _, c := range cs {
		rows = append(rows, jen.Values(jen.Lit(c.Name), jen.Lit(c.Doc), jen.Id(c.Func)))
	}
	f.Comment("examples are the examples of the service, by the name that runs each one.")
	f.Var().Id("examples").Op("=").Index().Struct(
		jen.Id("name").String(),
		jen.Id("doc").String(),
		jen.Id("run").Func().Params(params...).Error(),
	).Values(multi(rows)...)

	for _, c := range cs {
		f.Line()
		f.Comment(c.Func + " runs the example " + c.Name + ": " + lowerFirst(sentence(c.Doc)))
		body := append([]jen.Code{}, c.Pre...)
		if c.Method.Result == nil {
			body = append(body,
				jen.If(jen.Err().Op(":=").Add(c.Code), jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				),
				jen.Qual("fmt", "Println").Call(jen.Lit("done")),
				jen.Line(),
				jen.Return(jen.Nil()),
			)
		} else {
			body = append(body,
				jen.List(jen.Id("resp"), jen.Err()).Op(":=").Add(c.Code),
				jen.If(jen.Err().Op("!=").Nil()).Block(jen.Return(jen.Err())),
				jen.Line(),
				jen.Return(jen.Id("show").Call(jen.Id("resp").Dot("Data"))),
			)
		}
		f.Func().Id(c.Func).Params(multi(params)...).Error().Block(body...)
	}

	f.Line()
	writeMain(f, s.Node)

	return f, nil
}

// writeMain writes main, which runs the example its argument names, and
// show, which prints an answer.
func writeMain(
	f *jen.File,
	node bool,
) {
	run := []jen.Code{jen.Id("ctx"), jen.Id("c")}
	var target []jen.Code
	if node {
		run = append(run, jen.Id("target"))
		target = []jen.Code{
			jen.Id("target").Op(":=").Qual("os", "Getenv").Call(jen.Lit("OSAPI_TARGET")),
			jen.If(jen.Id("target").Op("==").Lit("")).Block(
				jen.Id("target").Op("=").Lit("_any"),
			),
			jen.Line(),
		}
	}
	body := make([]jen.Code, 0, 16)
	body = append(body,
		jen.If(jen.Len(jen.Qual("os", "Args")).Op("!=").Lit(2)).Block(
			jen.Qual("fmt", "Println").Call(jen.Lit("Give the name of an example:")),
			jen.For(jen.List(jen.Id("_"), jen.Id("e")).Op(":=").Range().Id("examples")).Block(
				jen.Qual("fmt", "Printf").Call(
					jen.Lit("  %-12s %s\n"), jen.Id("e").Dot("name"), jen.Id("e").Dot("doc"),
				),
			),
			jen.Qual("os", "Exit").Call(jen.Lit(2)),
		),
		jen.Line(),
		jen.Id("url").Op(":=").Qual("os", "Getenv").Call(jen.Lit("OSAPI_URL")),
		jen.If(jen.Id("url").Op("==").Lit("")).Block(
			jen.Id("url").Op("=").Lit("http://localhost:8080"),
		),
		jen.Line(),
		jen.Id("token").Op(":=").Qual("os", "Getenv").Call(jen.Lit("OSAPI_TOKEN")),
		jen.If(jen.Id("token").Op("==").Lit("")).Block(
			jen.Qual("log", "Fatal").Call(jen.Lit("OSAPI_TOKEN is required")),
		),
		jen.Line(),
	)
	body = append(body, target...)
	body = append(
		body,
		jen.Id("c").Op(":=").Qual(pkg, "New").Call(jen.Id("url"), jen.Id("token")),
		jen.Id("ctx").Op(":=").Qual("context", "Background").Call(),
		jen.For(jen.List(jen.Id("_"), jen.Id("e")).Op(":=").Range().Id("examples")).Block(
			jen.If(jen.Id("e").Dot("name").Op("!=").Qual("os", "Args").Index(jen.Lit(1))).Block(
				jen.Continue(),
			),
			jen.If(
				jen.Err().Op(":=").Id("e").Dot("run").Call(run...),
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Qual("log", "Fatalf").
					Call(jen.Lit("%s: %v"), jen.Id("e").Dot("name"), jen.Err()),
			),
			jen.Line(),
			jen.Return(),
		),
		jen.Qual("log", "Fatalf").
			Call(jen.Lit("no example %s"), jen.Qual("os", "Args").Index(jen.Lit(1))),
	)
	f.Func().Id("main").Params().Block(body...)
	f.Line()
	f.Comment("show prints v as JSON.")
	f.Func().Id("show").Params(multi([]jen.Code{jen.Id("v").Any()})...).Error().Block(
		jen.List(jen.Id("out"), jen.Err()).Op(":=").Qual("encoding/json", "MarshalIndent").
			Call(jen.Id("v"), jen.Lit(""), jen.Lit("  ")),
		jen.If(jen.Err().Op("!=").Nil()).Block(jen.Return(jen.Err())),
		jen.Qual("fmt", "Println").Call(jen.String().Call(jen.Id("out"))),
		jen.Line(),
		jen.Return(jen.Nil()),
	)
}
