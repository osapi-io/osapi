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

// Package routegen writes the route table and the decoders of an endpoint
// group, for the endpoint generator and the behavior generator alike, so a
// behavior's routes are the same code as an endpoint's.
package routegen

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/gen/genutil"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
)

const (
	serverPath = "github.com/osapi-io/osapi/internal/controller/endpoint/server"
	echoPath   = "github.com/labstack/echo/v5"
)

// Routes writes gen/api.go: the group's routes, each with its verb, its echo
// path, its permission and the handler that Serve makes, and the init that
// registers them.
func Routes(
	f *jen.File,
	g *endpoint.Group,
	pkg string,
) {
	f.Comment("Routes are the " + g.Name + " group's endpoints, as its definition declares them.")

	items := make([]jen.Code, 0, len(g.Methods))
	for _, m := range g.Methods {
		items = append(items, genutil.Multi("{", "}",
			jen.Id("Method").Op(":").Qual("net/http", verbConst(m.Verb)),
			jen.Id("Path").Op(":").Lit(server.Path(g, m)),
			jen.Id("Permission").Op(":").Lit(g.Permission(m)),
			jen.Id("Handler").Op(":").Qual(serverPath, "Serve").Call(
				jen.Id("s"),
				jen.Qual(pkg, "Group"),
				jen.Lit(m.Name),
				handler(m.Handler()),
				jen.Id(decoderName(m)),
			),
		))
	}

	f.Func().Id("Routes").Add(genutil.Params(jen.Id("s").Op("*").Qual(serverPath, "Server"))).
		Index().Qual(serverPath, "Route").Block(
		jen.Return(jen.Index().Qual(serverPath, "Route").Add(genutil.Multi("{", "}", items...))),
	)
	f.Line()
	f.Func().Id("init").Params().Block(
		jen.Qual(serverPath, "Register").Call(jen.Id("Routes")),
	)
}

// Decoders writes gen/decode.go: one function for each method that reads its
// input from the path, the query and the body, and validates it.
func Decoders(
	f *jen.File,
	g *endpoint.Group,
) {
	for i, m := range g.Methods {
		if i > 0 {
			f.Line()
		}
		in := jen.Qual(m.In.PkgPath(), m.In.Name())
		fail := func() jen.Code {
			return jen.Return(jen.Id("in"), jen.Id("msg"), jen.False())
		}
		check := func(call jen.Code) jen.Code {
			return jen.If(
				jen.List(jen.Id("msg"), jen.Id("ok")).Op(":=").Add(call),
				jen.Op("!").Id("ok"),
			).Block(fail())
		}

		body := []jen.Code{jen.Var().Id("in").Add(in.Clone())}
		var path, query []reflect.StructField
		var bodyField *reflect.StructField
		hasBody, hasForm := false, false
		for _, fld := range genutil.StructFields(m.In) {
			switch fld.Tag.Get("in") {
			case "path":
				path = append(path, fld)
			case "query":
				query = append(query, fld)
			case "body":
				bodyField = &fld
			case "form":
				hasForm = true
			default:
				hasBody = true
			}
		}
		switch {
		case hasForm:
			body = append(body, check(jen.Qual(serverPath, "Form").Call(
				jen.Id("c"), jen.Op("&").Id("in"))))
		case bodyField != nil:
			body = append(body, check(jen.Qual(serverPath, "Body").Call(
				jen.Id("c"), jen.Op("&").Id("in").Dot(bodyField.Name))))
		case hasBody:
			body = append(body, check(jen.Qual(serverPath, "Body").Call(
				jen.Id("c"), jen.Op("&").Id("in"))))
		}
		for _, fld := range path {
			body = append(body, check(jen.Qual(serverPath, "Param").Call(
				jen.Id("c"), jen.Lit(genutil.JSONName(fld)), jen.Op("&").Id("in").Dot(fld.Name))))
		}
		for _, fld := range query {
			body = append(body, check(jen.Qual(serverPath, "Query").Call(
				jen.Id("c"), jen.Lit(genutil.JSONName(fld)), jen.Op("&").Id("in").Dot(fld.Name))))
		}
		if len(genutil.StructFields(m.In)) > 0 {
			body = append(body, check(jen.Qual(serverPath, "Valid").Call(jen.Id("c"), jen.Id("in"))))
		}
		body = append(body, jen.Line(), jen.Return(jen.Id("in"), jen.Lit(""), jen.True()))

		f.Comment(decoderName(m) + " reads the " + m.Name + " method's input from the request.")
		f.Func().
			Id(decoderName(m)).
			Add(genutil.Params(jen.Id("c").Op("*").Qual(echoPath, "Context"))).
			Params(in.Clone(), jen.String(), jen.Bool()).
			Block(body...)
	}
}

func decoderName(
	m *endpoint.Method,
) string {
	return "decode" + strings.ToUpper(m.Name[:1]) + m.Name[1:]
}

// handler is the handler function, qualified by its package: a group's own
// package for an endpoint, a behavior's gen package for a behavior.
func handler(
	fn any,
) *jen.Statement {
	full := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	i := strings.LastIndex(full, ".")

	return jen.Qual(full[:i], full[i+1:])
}

func verbConst(
	verb string,
) string {
	switch verb {
	case http.MethodGet:
		return "MethodGet"
	case http.MethodPost:
		return "MethodPost"
	case http.MethodPut:
		return "MethodPut"
	default:
		return "MethodDelete"
	}
}
