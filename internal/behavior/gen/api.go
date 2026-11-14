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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behaviorcmd"
	behaviorapi "github.com/osapi-io/osapi/internal/controller/api/behavior"
	"github.com/osapi-io/osapi/internal/validation"
)

// statusNames are the net/http names of the codes a behavior answers.
var statusNames = map[int]string{
	http.StatusOK:                  "StatusOK",
	http.StatusBadRequest:          "StatusBadRequest",
	http.StatusUnauthorized:        "StatusUnauthorized",
	http.StatusForbidden:           "StatusForbidden",
	http.StatusNotFound:            "StatusNotFound",
	http.StatusInternalServerError: "StatusInternalServerError",
}

// handlerName and decoderName are a method's generated function names:
// apiUpdate and decodeUpdate.
func handlerName(
	m *behavior.Method,
) string {
	return "api" + title(m.Name)
}

func decoderName(
	m *behavior.Method,
) string {
	return "decode" + title(m.Name)
}

// apiFile is b's gen/api.go: the table of its routes, registered with the
// runtime from init.
func apiFile(
	b *behavior.Behavior,
) *jen.File {
	f := genFile(b)

	routes := make([]jen.Code, 0, len(b.Methods))
	for _, m := range b.Methods {
		routes = append(routes, multi("{", "}",
			jen.Id("Method").Op(":").Qual(httpPath, methodConst(m)),
			jen.Id("Path").Op(":").Lit(behaviorapi.Path(b, m)),
			jen.Id("Permission").Op(":").Lit(b.Permission(m.Verb)),
			jen.Id("Handler").Op(":").Id(handlerName(m)).Call(jen.Id("s")),
		))
	}

	comment(f, fmt.Sprintf("Routes are %s's API routes: each method's HTTP method, its path, the "+
		"permission a token needs, and its handler.", b.Name))
	f.Func().Id("Routes").Add(params(jen.Id("s").Op("*").Qual(apiPath, "Server"))).
		Index().Qual(apiPath, "Route").Block(
		jen.Return(jen.Index().Qual(apiPath, "Route").Add(multi("{", "}", routes...))),
	)
	f.Line()
	f.Func().Id("init").Params().Block(jen.Qual(apiPath, "Register").Call(jen.Id("Routes")))

	return f
}

// apiMethodFile is b's gen/api_<method>.go: the method's handler, and the
// decoder that builds its typed input from the request.
func apiMethodFile(
	b *behavior.Behavior,
	m *behavior.Method,
) (*jen.File, error) {
	f := genFile(b)
	in := jen.Qual(pkgPath(b), m.In.Name())

	comment(f, fmt.Sprintf("%s serves %s %s, which needs %s: %s.",
		handlerName(m), m.HTTPMethod(), b.Route(m), b.Permission(m.Verb),
		strings.TrimSuffix(m.Doc, ".")))
	f.Func().Id(handlerName(m)).Add(params(jen.Id("s").Op("*").Qual(apiPath, "Server"))).
		Qual(echoPath, "HandlerFunc").Block(
		jen.Return(jen.Qual(apiPath, "Serve").Call(
			jen.Id("s"), jen.Qual(pkgPath(b), "Behavior"), jen.Lit(m.Name), jen.Id(decoderName(m)),
		)),
	)
	f.Line()

	fls := behaviorcmd.Flags(m.In)
	if len(fls) == 0 && m.In.NumField() == 0 {
		comment(f, fmt.Sprintf("%s builds %s, which takes nothing from the request.",
			decoderName(m), m.In.Name()))
		f.Func().Id(decoderName(m)).Add(params(jen.Id("_").Op("*").Qual(echoPath, "Context"))).
			Params(in.Clone(), jen.String(), jen.Bool()).Block(
			jen.Return(in.Clone().Values(), jen.Lit(""), jen.True()),
		)

		return f, nil
	}

	var steps []jen.Code
	var from []string
	fail := jen.Return(jen.Id("in"), jen.Id("msg"), jen.False())
	step := func(call jen.Code) {
		steps = append(steps, jen.If(
			jen.List(jen.Id("msg"), jen.Id("ok")).Op(":=").Add(call), jen.Op("!").Id("ok"),
		).Block(fail))
	}

	if m.Verb == behavior.VerbWrite {
		from = append(from, "the JSON body")
		step(jen.Qual(apiPath, "Body").Call(jen.Id("c"), jen.Op("&").Id("in")))
	} else {
		for _, fl := range fls {
			if fl.JSON == m.KeyField {
				continue
			}
			from = append(from, fl.JSON+" from the query")
			step(jen.Qual(apiPath, "Query").Call(
				jen.Id("c"), jen.Lit(fl.JSON), jen.Op("&").Id("in").Dot(m.In.Field(fl.Index).Name),
			))
		}
	}
	if m.KeyField != "" {
		field, ok := fieldByJSON(m.In, m.KeyField)
		if !ok {
			return nil, fmt.Errorf("%s: key %q is not an input field", b.Operation(m), m.KeyField)
		}
		from = append(from, m.KeyField+" from the path")
		step(jen.Qual(apiPath, "Param").Call(
			jen.Id("c"), jen.Lit(m.KeyField), jen.Op("&").Id("in").Dot(field.Name),
		))
	}
	step(jen.Qual(apiPath, "Valid").Call(jen.Id("in")))

	comment(f, fmt.Sprintf("%s builds %s from the request: %s, then checks its validate tags.",
		decoderName(m), m.In.Name(), strings.Join(from, ", then ")))
	body := append([]jen.Code{jen.Var().Id("in").Add(in.Clone()), jen.Line()}, steps...)
	body = append(body, jen.Line(), jen.Return(jen.Id("in"), jen.Lit(""), jen.True()))
	f.Func().Id(decoderName(m)).Add(params(jen.Id("c").Op("*").Qual(echoPath, "Context"))).
		Params(in.Clone(), jen.String(), jen.Bool()).Block(body...)

	return f, nil
}

// apiTestFile is gen/api.go's test: Routes lists every method with its
// HTTP method, path and permission.
func apiTestFile(
	b *behavior.Behavior,
) *jen.File {
	f := genTestFile(b)

	rows := make([]jen.Code, 0, len(b.Methods))
	for _, m := range b.Methods {
		rows = append(rows, multi("{", "}",
			jen.Id("method").Op(":").Qual(httpPath, methodConst(m)),
			jen.Id("path").Op(":").Lit(behaviorapi.Path(b, m)),
			jen.Id("permission").Op(":").Lit(b.Permission(m.Verb)),
		))
	}

	f.Type().Id("APIPublicTestSuite").Struct(jen.Qual(suitePath, "Suite"))
	f.Line()
	f.Func().Params(jen.Id("s").Op("*").Id("APIPublicTestSuite")).Id("TestRoutes").Params().Block(
		jen.Id("routes").Op(":=").Qual(pkgPath(b)+"/gen", "Routes").Call(
			jen.Qual(apiPath, "NewServer").Call(
				jen.Qual("log/slog", "Default").Call(), jen.Nil(), jen.Lit(""), jen.Nil(),
			),
		),
		jen.Id("want").Op(":=").Index().Struct(
			jen.Id("method").String(),
			jen.Id("path").String(),
			jen.Id("permission").String(),
		).Add(multi("{", "}", rows...)),
		jen.Line(),
		jen.Id("s").
			Dot("Require").
			Call().
			Dot("Len").
			Call(jen.Id("routes"), jen.Len(jen.Id("want"))),
		jen.For(jen.List(jen.Id("i"), jen.Id("w")).Op(":=").Range().Id("want")).Block(
			jen.Id("s").Dot("Run").Call(
				jen.Id("w").Dot("method").Op("+").Lit(" ").Op("+").Id("w").Dot("path"),
				jen.Func().Params().Block(
					jen.Id("s").
						Dot("Equal").
						Call(jen.Id("w").Dot("method"), jen.Id("routes").Index(jen.Id("i")).Dot("Method")),
					jen.Id("s").
						Dot("Equal").
						Call(jen.Id("w").Dot("path"), jen.Id("routes").Index(jen.Id("i")).Dot("Path")),
					jen.Id("s").
						Dot("Equal").
						Call(jen.Id("w").Dot("permission"), jen.Id("routes").Index(jen.Id("i")).Dot("Permission")),
					jen.Id("s").
						Dot("NotNil").
						Call(jen.Id("routes").Index(jen.Id("i")).Dot("Handler")),
				),
			),
		),
	)
	suiteRunner(f, "APIPublicTestSuite")

	return f
}

// httpCase is one generated request and what it should get.
type httpCase struct {
	name     string
	method   string
	path     string
	body     string
	token    string
	job      jen.Code
	code     int
	wantBody string
}

// apiMethodTestFile is gen/api_<method>.go's test: each example on one host
// and across hosts, each way a host or the job client can answer, each way
// the request can be wrong, and each token that must be refused, sent
// through the real handler and middleware and checked against the spec.
func apiMethodTestFile(
	b *behavior.Behavior,
	m *behavior.Method,
) (*jen.File, error) {
	f := genTestFile(b)
	name := "API" + title(m.Name) + "PublicTestSuite"

	responses, err := responseCases(b, m)
	if err != nil {
		return nil, err
	}
	validations, err := validationCases(b, m)
	if err != nil {
		return nil, err
	}
	rbac, err := rbacCases(b, m)
	if err != nil {
		return nil, err
	}

	f.Type().Id(name).Struct(
		jen.Qual(suitePath, "Suite"),
		jen.Line(),
		jen.Id("api").Op("*").Qual(apitestPath, "Harness"),
	)
	f.Line()
	f.Func().Params(jen.Id("s").Op("*").Id(name)).Id("SetupSuite").Params().Block(
		jen.Id("s").Dot("api").Op("=").Qual(apitestPath, "New").Call(
			jen.Id("s").Dot("T").Call(), jen.Qual(pkgPath(b)+"/gen", "Routes"),
		),
	)

	method := func(test string, doc string, cases []httpCase) {
		f.Line()
		comment(f, doc)
		f.Func().Params(jen.Id("s").Op("*").Id(name)).Id(test).Params().Block(
			jen.Id("s").Dot("api").Dot("Run").Call(
				jen.Id("s").Dot("T").Call(),
				jen.Index().Qual(apitestPath, "Case").Add(multi("{", "}", render(cases)...)),
			),
		)
	}

	method("Test"+title(m.Name),
		fmt.Sprintf("Test%s sends %s %s and checks each answer's response.",
			title(m.Name), m.HTTPMethod(), b.Route(m)), responses)
	method("Test"+title(m.Name)+"ValidationHTTP",
		fmt.Sprintf("Test%sValidationHTTP sends requests the API refuses before any job is queued.",
			title(m.Name)), validations)
	method("Test"+title(m.Name)+"RBACHTTP",
		fmt.Sprintf("Test%sRBACHTTP sends tokens that do not grant %s.",
			title(m.Name), b.Permission(m.Verb)), rbac)
	suiteRunner(f, name)

	return f, nil
}

// responseCases are the requests that reach a host: each example on one
// host and broadcast, then each other way a host or the job client answers.
func responseCases(
	b *behavior.Behavior,
	m *behavior.Method,
) ([]httpCase, error) {
	var cases []httpCase
	for _, ex := range m.Examples {
		out, err := json.Marshal(ex.Output)
		if err != nil {
			return nil, err
		}
		one, err := request(b, m, "server1", ex.Input)
		if err != nil {
			return nil, err
		}
		one.name = "when the example " + strconv.Quote(
			ex.Doc,
		) + " is sent, the host's result is returned"
		one.job = jobCall(b, m, "server1", ex.Input, completed("server1", string(out)))
		one.code = http.StatusOK
		one.wantBody = body(rows(m, "server1", out))
		cases = append(cases, one)

		all, err := request(b, m, "_all", ex.Input)
		if err != nil {
			return nil, err
		}
		all.name = "when the example " + strconv.Quote(ex.Doc) +
			" is broadcast, every host's result is returned"
		all.job = jobCall(b, m, "_all", ex.Input,
			completed("server1", string(out)),
			jen.Qual(apitestPath, "Skipped").Call(jen.Lit("server2")),
			jen.Qual(apitestPath, "Failed").Call(jen.Lit("server3"), jen.Lit("boom")),
		)
		all.code = http.StatusOK
		all.wantBody = body(append(rows(m, "server1", out),
			failedRow("server2", "skipped", "unsupported"),
			failedRow("server3", "failed", "boom")))
		cases = append(cases, all)
	}

	input := firstInput(m)
	base, err := request(b, m, "server1", input)
	if err != nil {
		return nil, err
	}
	all, err := request(b, m, "_all", input)
	if err != nil {
		return nil, err
	}

	add := func(c httpCase, name string, job jen.Code, code int, wantBody string) {
		c.name, c.job, c.code, c.wantBody = name, job, code, wantBody
		cases = append(cases, c)
	}
	add(base, "when the host skips it, its row says skipped",
		jobCall(b, m, "server1", input, jen.Qual(apitestPath, "Skipped").Call(jen.Lit("server1"))),
		http.StatusOK, body([]map[string]any{failedRow("server1", "skipped", "unsupported")}))
	add(base, "when the host fails it, its row says failed",
		jobCall(b, m, "server1", input,
			jen.Qual(apitestPath, "Failed").Call(jen.Lit("server1"), jen.Lit("boom"))),
		http.StatusOK, body([]map[string]any{failedRow("server1", "failed", "boom")}))
	add(base, "when the host never answers, its row says timeout",
		jobCall(b, m, "server1", input, jen.Qual(apitestPath, "TimedOut").Call(jen.Lit("server1"))),
		http.StatusOK, body([]map[string]any{failedRow("server1", "timeout", "")}))
	add(base, "when the job client fails, it is a 500",
		failing(jobCall(b, m, "server1", input), jen.Qual(apitestPath, "ErrUnreachable")),
		http.StatusInternalServerError, "")
	add(all, "when broadcast and the job client fails, it is a 500",
		failing(jobCall(b, m, "_all", input), jen.Qual(apitestPath, "ErrUnreachable")),
		http.StatusInternalServerError, "")
	if m.KeyField != "" {
		add(base, "when the host does not have it, it is a 404",
			failing(jobCall(b, m, "server1", input),
				jen.Qual(apitestPath, "NotFound").Call(jen.Lit("server1"))),
			http.StatusNotFound, "")
	}

	return cases, nil
}

// validationCases are the requests the API refuses with a 400.
func validationCases(
	b *behavior.Behavior,
	m *behavior.Method,
) ([]httpCase, error) {
	input := firstInput(m)
	base, err := request(b, m, "server1", input)
	if err != nil {
		return nil, err
	}

	var cases []httpCase
	add := func(c httpCase, name string) {
		c.name, c.code = name, http.StatusBadRequest
		cases = append(cases, c)
	}

	target, err := request(b, m, "nowhere", input)
	if err != nil {
		return nil, err
	}
	add(target, "when the target is not an agent, it is a 400")

	empty, err := request(b, m, "", input)
	if err != nil {
		return nil, err
	}
	add(empty, "when the target is empty, it is a 400")

	for _, fl := range behaviorcmd.Flags(m.In) {
		if !fl.Required || fl.JSON == m.KeyField {
			continue
		}
		c, err := request(b, m, "server1", without(input, fl.Index))
		if err != nil {
			return nil, err
		}
		add(c, "when "+fl.JSON+" is missing, it is a 400")
	}

	if m.Verb == behavior.VerbWrite {
		c := base
		c.body = "{"
		add(c, "when the body is not JSON, it is a 400")
		for _, fl := range behaviorcmd.Flags(m.In) {
			if fl.JSON == m.KeyField {
				continue
			}
			c := base
			c.body = `{"` + fl.JSON + `":[]}`
			add(c, "when "+fl.JSON+" has the wrong type, it is a 400")

			break
		}
	}

	for _, r := range m.Refusals {
		// A refusal the API cannot see, such as a parameter osapi does not
		// manage, is the provider's, which behaviortest checks; and a keyed
		// method with no key has no route.
		if _, ok := validation.Struct(r.Input); ok {
			continue
		}
		if m.KeyField != "" {
			if field, ok := fieldByJSON(m.In, m.KeyField); ok &&
				reflect.ValueOf(r.Input).FieldByIndex(field.Index).IsZero() {
				continue
			}
		}
		c, err := request(b, m, "server1", r.Input)
		if err != nil {
			return nil, err
		}
		add(c, "when sent "+lower(r.Doc)+", it is a 400")
	}

	return cases, nil
}

// rbacCases are the tokens that must be refused, and the admin role, which
// must not be.
func rbacCases(
	b *behavior.Behavior,
	m *behavior.Method,
) ([]httpCase, error) {
	base, err := request(b, m, "server1", firstInput(m))
	if err != nil {
		return nil, err
	}

	none, lacking := base, base
	none.name, none.token, none.code = "when there is no token, it is a 401", "NoToken", http.StatusUnauthorized
	lacking.name, lacking.token, lacking.code = "when the token lacks "+b.Permission(m.Verb)+
		", it is a 403", "Lacking", http.StatusForbidden
	cases := []httpCase{none, lacking}

	if m.Verb == behavior.VerbWrite {
		read := base
		read.name, read.token, read.code = "when the token has only the read role, it is a 403",
			"Read", http.StatusForbidden
		cases = append(cases, read)
	}

	admin := base
	admin.name, admin.token, admin.code = "when the token has the admin role, the request runs",
		"Admin", http.StatusOK
	admin.job = jobCall(b, m, "server1", firstInput(m),
		jen.Qual(apitestPath, "Skipped").Call(jen.Lit("server1")))
	admin.wantBody = body([]map[string]any{failedRow("server1", "skipped", "unsupported")})
	cases = append(cases, admin)

	return cases, nil
}

// request is the request for input on target: the key in the path, a
// read's other fields in the query, a write's input as its body.
func request(
	b *behavior.Behavior,
	m *behavior.Method,
	target string,
	input any,
) (httpCase, error) {
	c := httpCase{method: m.HTTPMethod(), token: "Read"}
	if m.Verb == behavior.VerbWrite {
		c.token = "Write"
	}

	path := strings.Replace(b.Route(m), "{hostname}", url.PathEscape(target), 1)
	v := reflect.ValueOf(input)
	if m.KeyField != "" {
		field, _ := fieldByJSON(m.In, m.KeyField)
		path = strings.Replace(path, "{"+m.KeyField+"}",
			url.PathEscape(fmt.Sprint(v.FieldByIndex(field.Index).Interface())), 1)
	}

	if m.Verb == behavior.VerbWrite {
		data, err := json.Marshal(input)
		if err != nil {
			return c, err
		}
		c.body = string(data)
	} else {
		q := url.Values{}
		for _, fl := range behaviorcmd.Flags(m.In) {
			if fv := v.Field(fl.Index); fl.JSON != m.KeyField && !fv.IsZero() {
				q.Set(fl.JSON, fmt.Sprint(fv.Interface()))
			}
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}
	c.path = path

	return c, nil
}

// render is each case as an apitest.Case literal.
func render(
	cases []httpCase,
) []jen.Code {
	out := make([]jen.Code, 0, len(cases))
	for _, c := range cases {
		fields := []jen.Code{
			jen.Id("Name").Op(":").Lit(c.name),
			jen.Id("Request").Op(":").Qual(apitestPath, "Request").Add(multi("{", "}",
				jen.Id("Method").Op(":").Qual(httpPath, "Method"+title(strings.ToLower(c.method))),
				jen.Id("Path").Op(":").Lit(c.path),
				bodyField(c.body),
			)),
			jen.Id("Token").Op(":").Qual(apitestPath, c.token),
		}
		if c.job != nil {
			fields = append(fields, jen.Id("Job").Op(":").Add(c.job))
		}
		fields = append(fields, jen.Id("WantCode").Op(":").Qual(httpPath, statusNames[c.code]))
		if c.wantBody != "" {
			fields = append(fields, jen.Id("WantBody").Op(":").Add(raw(c.wantBody)))
		}
		out = append(out, multi("{", "}", fields...))
	}

	return out
}

func bodyField(
	body string,
) jen.Code {
	if body == "" {
		return jen.Null()
	}

	return jen.Id("Body").Op(":").Add(raw(body))
}

// jobCall is the apitest call the job client expects for input on target.
func jobCall(
	b *behavior.Behavior,
	m *behavior.Method,
	target string,
	input any,
	replies ...jen.Code,
) *jen.Statement {
	fn := "Query"
	if m.Verb == behavior.VerbWrite {
		fn = "Modify"
	}
	in, err := literal(reflect.ValueOf(input))
	if err != nil {
		in = jen.Lit(fmt.Sprint(input))
	}
	args := append([]jen.Code{jen.Lit(target), jen.Lit(b.Operation(m)), in}, replies...)

	return jen.Qual(apitestPath, fn).Add(multi("(", ")", args...))
}

func failing(
	call *jen.Statement,
	err jen.Code,
) jen.Code {
	return call.Dot("Fails").Call(err)
}

func completed(
	host string,
	data string,
) jen.Code {
	return jen.Qual(apitestPath, "Completed").Call(jen.Lit(host), raw(data))
}

// rows are the result rows the handler makes of one host's output, as
// behaviorapi's rows does: one per element of a list, else one, each with
// the host's name and status.
func rows(
	m *behavior.Method,
	host string,
	out []byte,
) []map[string]any {
	var outputs []map[string]any
	if m.Out.Kind() == reflect.Slice {
		_ = json.Unmarshal(out, &outputs)
	} else {
		var one map[string]any
		if err := json.Unmarshal(out, &one); err == nil && one != nil {
			outputs = []map[string]any{one}
		}
	}
	if len(outputs) == 0 && m.Out.Kind() != reflect.Slice {
		outputs = []map[string]any{{}}
	}
	for _, row := range outputs {
		row["hostname"] = host
		row["status"] = "ok"
	}

	return outputs
}

func failedRow(
	host string,
	status string,
	msg string,
) map[string]any {
	return map[string]any{"hostname": host, "status": status, "error": msg}
}

// body is the response body for rows, with the rows in hostname order.
func body(
	results []map[string]any,
) string {
	sort.SliceStable(results, func(i, j int) bool {
		return fmt.Sprint(results[i]["hostname"]) < fmt.Sprint(results[j]["hostname"])
	})
	if results == nil {
		results = []map[string]any{}
	}
	data, _ := json.Marshal(map[string]any{"job_id": "job-1", "results": results})

	return string(data)
}

// firstInput is the first example's input, or the zero input when the
// method has none.
func firstInput(
	m *behavior.Method,
) any {
	if len(m.Examples) > 0 {
		return m.Examples[0].Input
	}

	return reflect.New(m.In).Elem().Interface()
}

// without is input with field i zeroed.
func without(
	input any,
	i int,
) any {
	v := reflect.New(reflect.TypeOf(input)).Elem()
	v.Set(reflect.ValueOf(input))
	v.Field(i).SetZero()

	return v.Interface()
}

func fieldByJSON(
	t reflect.Type,
	name string,
) (reflect.StructField, bool) {
	for _, fl := range behaviorcmd.Flags(t) {
		if fl.JSON == name {
			return t.Field(fl.Index), true
		}
	}

	return reflect.StructField{}, false
}

// methodConst is the net/http constant for m's HTTP method.
func methodConst(
	m *behavior.Method,
) string {
	return "Method" + title(strings.ToLower(m.HTTPMethod()))
}

func title(
	s string,
) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func lower(
	s string,
) string {
	if s == "" {
		return ""
	}

	return strings.ToLower(s[:1]) + s[1:]
}

// raw is s as a raw string literal, so JSON reads without escapes, or as a
// quoted one when s holds a backquote.
func raw(
	s string,
) *jen.Statement {
	if strings.Contains(s, "`") {
		return jen.Lit(s)
	}

	return jen.Id("`" + s + "`")
}
