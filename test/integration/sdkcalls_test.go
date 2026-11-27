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
package integration_test

// This file has no build tag: TestEachOperationHasAnSDKCall runs with the
// unit tests, as the gate of osapi-io/osapi#1599. With
// OSAPI_INTEGRATION_CLIENT=sdk, the integration suite sends each request of
// its CLI cases, refusals, checks and scenarios through the SDK's typed
// method for its operation (sdkreplay_test.go): the call that this file
// finds and builds.

import (
	"encoding"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/sdk/gen/sdkmodel"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// sdkCall is the typed method of the SDK for one operation of the spec:
// its service, such as Hostname, and its method, such as Update.
type sdkCall struct {
	service *sdkmodel.Service
	method  *sdkmodel.Method
}

// argKind is where one argument of a typed method comes from in a
// request.
type argKind int

const (
	// argPath is a parameter of the path: the target, or a key when the
	// method has no input.
	argPath argKind = iota
	// argQuery is a required parameter of the query.
	argQuery
	// argParams is the struct of the optional parameters of the query.
	argParams
	// argInput is the struct of the body, with the keys of the path.
	argInput
	// argForm is a field of a multipart form.
	argForm
	// argFile is the file of a multipart form, as an io.Reader.
	argFile
)

// arg is one argument of a typed method after its context.
type arg struct {
	kind argKind
	name string
}

// sdkRequest is the parts of one request that the arguments come from.
type sdkRequest struct {
	path  map[string]string
	query url.Values
	body  []byte
	form  *multipart.Form
}

var (
	sdkOnce sync.Once
	sdkByOp map[string]sdkCall
	errSDK  error
)

// sdkCalls is each operation of the full spec, by its method and path
// such as GET /api/node/{hostname}/ntp, with its typed method of the SDK.
func sdkCalls() (map[string]sdkCall, error) {
	sdkOnce.Do(func() {
		doc, err := openapi3.NewLoader().LoadFromFile(
			filepath.Join(repoRootDir(), "internal", "controller", "api", "gen", "api.yaml"),
		)
		if err != nil {
			errSDK = fmt.Errorf("load the full spec: %w", err)

			return
		}
		m, err := sdkmodel.New(doc)
		if err != nil {
			errSDK = fmt.Errorf("the model of the SDK: %w", err)

			return
		}
		sdkByOp = map[string]sdkCall{}
		for _, s := range m.Services {
			for _, meth := range s.Methods {
				sdkByOp[meth.HTTPMethod+" "+meth.Path] = sdkCall{service: s, method: meth}
			}
		}
	})

	return sdkByOp, errSDK
}

// repoRootDir is the root of the repository, from this package's
// directory, which go test runs in.
func repoRootDir() string {
	dir, _ := filepath.Abs(filepath.Join("..", ".."))

	return dir
}

// String is the call as a program writes it, such as Hostname.Update.
func (k sdkCall) String() string {
	return k.service.Name + "." + k.method.Name
}

// fn is the typed method of k on c.
func (k sdkCall) fn(
	c *client.Client,
) (reflect.Value, error) {
	svc := reflect.ValueOf(c).Elem().FieldByName(k.service.Name)
	if !svc.IsValid() {
		return reflect.Value{}, fmt.Errorf("client.Client has no service %s", k.service.Name)
	}
	f := svc.MethodByName(k.method.Name)
	if !f.IsValid() {
		return reflect.Value{}, fmt.Errorf("%s has no method %s", k.service.Name, k.method.Name)
	}

	return f, nil
}

// plan is the arguments of k's typed method after its context, in the
// order that internal/sdk/gen writes them.
func (k sdkCall) plan() ([]arg, error) {
	m := k.method
	var out []arg
	if m.Multipart {
		fields, err := m.FormArgs()
		if err != nil {
			return nil, err
		}
		for _, f := range fields {
			switch {
			case f.Path:
				out = append(out, arg{argPath, f.Name})
			case f.File:
				out = append(out, arg{argFile, f.Name})
			default:
				out = append(out, arg{argForm, f.Name})
			}
		}

		return out, nil
	}
	if t := m.Target(); t != nil {
		out = append(out, arg{argPath, t.Name})
	}
	if m.Input != nil {
		out = append(out, arg{argInput, m.Input.Name})
	} else {
		for _, p := range m.Keys() {
			out = append(out, arg{argPath, p.Name})
		}
		for _, p := range m.RequiredQuery() {
			out = append(out, arg{argQuery, p.Name})
		}
	}
	if len(m.OptionalQuery()) > 0 {
		out = append(out, arg{argParams, ""})
	}

	return out, nil
}

// readerType is io.Reader, the type of a form's file.
var readerType = reflect.TypeFor[io.Reader]()

// check is nil when fn, k's typed method, takes the arguments of k's plan
// and answers a Response or only an error.
func (k sdkCall) check(
	fn reflect.Type,
) error {
	plan, err := k.plan()
	if err != nil {
		return err
	}
	if fn.NumIn() != len(plan)+1 {
		return fmt.Errorf("takes %d arguments, and the spec gives %d", fn.NumIn(), len(plan)+1)
	}
	for i, a := range plan {
		t := fn.In(i + 1)
		switch a.kind {
		case argFile:
			if t != readerType {
				return fmt.Errorf("argument %d is %s, not io.Reader", i+1, t)
			}
		case argInput, argParams:
			if t.Kind() != reflect.Struct {
				return fmt.Errorf("argument %d is %s, not a struct", i+1, t)
			}
		default:
			if !scalar(t) {
				return fmt.Errorf("argument %d, %s, is %s, not a value of the path, "+
					"query or form", i+1, a.name, t)
			}
		}
		if a.kind == argInput {
			for _, p := range k.method.Keys() {
				if _, ok := t.FieldByName(p.GoName); !ok {
					return fmt.Errorf("%s has no field %s for the path's %s", t, p.GoName, p.Name)
				}
			}
		}
		if a.kind == argParams {
			for _, p := range k.method.OptionalQuery() {
				if _, ok := t.FieldByName(p.GoName); !ok {
					return fmt.Errorf("%s has no field %s for the query's %s", t, p.GoName, p.Name)
				}
			}
		}
	}
	switch fn.NumOut() {
	case 1:
	case 2:
		if _, ok := fn.Out(0).MethodByName("RawJSON"); !ok {
			return fmt.Errorf("answers %s, not a Response", fn.Out(0))
		}
	default:
		return fmt.Errorf("answers %d values", fn.NumOut())
	}

	return nil
}

// scalar is a type that one value of a path, a query or a form fills.
func scalar(
	t reflect.Type,
) bool {
	switch t.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Pointer:
		return scalar(t.Elem())
	default:
		return reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextUnmarshaler]())
	}
}

// sdkSkips are the operations that the SDK suite does not send through a
// typed method, by method and path, each with why and the open issue that
// makes it run, as osapi-io/<repo>#<n>. The gate requires both.
var sdkSkips = map[string]struct {
	why   string
	issue string
}{}

// SDKCallsTestSuite is the gate of osapi-io/osapi#1599: each operation of
// the spec, and so each command of the DSL that calls the API, has a typed
// method of the SDK that the SDK suite can build and call.
type SDKCallsTestSuite struct {
	suite.Suite
}

func (s *SDKCallsTestSuite) TestEachOperationHasAnSDKCall() {
	calls, err := sdkCalls()
	s.Require().NoError(err)
	s.Require().NotEmpty(calls)
	c := client.New("https://127.0.0.1:1", "gate")
	for op, k := range calls {
		if _, ok := sdkSkips[op]; ok {
			continue
		}
		fn, err := k.fn(c)
		if !s.NoError(err, "%s: %s", op, k) {
			continue
		}
		s.NoError(k.check(fn.Type()), "%s: %s", op, k)
	}
	for op, skip := range sdkSkips {
		s.Contains(calls, op, "sdkSkips names %s, which the spec does not have", op)
		s.GreaterOrEqual(len(skip.why), 40, "%s: say why the SDK suite cannot call it", op)
		s.Regexp(issueRef, skip.issue, "%s: a skip names the open issue that makes it run", op)
	}
}

// TestEachCommandHasAnSDKCall is the other half: each command of the DSL
// that calls the API names an operation that has a typed method, so its
// examples, refusals and checks reach the SDK.
func (s *SDKCallsTestSuite) TestEachCommandHasAnSDKCall() {
	calls, err := sdkCalls()
	s.Require().NoError(err)
	count := 0
	for _, g := range endpoint.All() {
		for _, c := range g.Commands() {
			m, ok := c.(*endpoint.Method)
			if !ok {
				continue
			}
			count++
			op := m.HTTPMethod() + " " + g.Path(m)
			_, skipped := sdkSkips[op]
			s.True(skipped || calls[op].method != nil,
				"%s %s: no typed method of the SDK sends %s", g.Name, m.CommandName(), op)
		}
	}
	s.NotZero(count)
}

func TestSDKCallsTestSuite(t *testing.T) {
	suite.Run(t, new(SDKCallsTestSuite))
}
