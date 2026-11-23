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

// Package command is the endpoint layer's runner of the generated CLI
// commands of the endpoint groups. Each group's gen/cli.go makes one cobra
// command for each method and registers it with the command layer,
// internal/cli/command. Run builds the method's request from its typed
// input, sends it through the SDK client that Use set, and gives the answer
// to the command layer, which prints it.
package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Caller sends one request to the API. The SDK client satisfies it.
type Caller interface {
	Send(
		ctx context.Context,
		method string,
		path string,
		contentType string,
		body []byte,
	) (*client.Response[json.RawMessage], error)
}

// caller returns the SDK client, which Use sets.
var caller func() Caller

// Use sets what Run sends each request through: a function that returns
// the SDK client, so the client is made only when a command runs.
func Use(
	c func() Caller,
) {
	caller = c
}

// Run sends method m of g with in, and gives the answer to the command
// layer to print. The answer decodes into the type of the method's
// renderer, so the generated command does not name its output type.
func Run[In any](
	cmd *cobra.Command,
	g *endpoint.Group,
	method string,
	in In,
) {
	m := g.Method(method)
	path, body, contentType, err := Build(g, m, in)
	if err != nil {
		clicommand.Report(err)

		return
	}

	resp, err := caller().Send(cmd.Context(), m.HTTPMethod(), path, contentType, body)
	data, ok := answered(m, err)
	if err != nil && !ok {
		clicommand.Report(err)

		return
	}
	if ok {
		resp = client.NewResponse(data, []byte(data))
	}
	clicommand.SetTraceID(resp.TraceID())
	clicommand.Output(g, method, in, resp.RawJSON())
}

// Build is the request that sends in to method m of g: the key in the path,
// the query fields in the query, and the other fields as the JSON body.
func Build(
	g *endpoint.Group,
	m *endpoint.Method,
	in any,
) (string, []byte, string, error) {
	v := reflect.ValueOf(in)
	path := g.Path(m)
	query := url.Values{}
	body := map[string]any{}
	form := &bytes.Buffer{}
	parts := multipart.NewWriter(form)
	hasForm := false
	var whole any

	if v.Kind() == reflect.Struct {
		for i := range v.NumField() {
			f := v.Type().Field(i)
			if !f.IsExported() {
				continue
			}
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "" {
				name = f.Name
			}
			if name == "-" {
				continue
			}
			value := v.Field(i)
			switch f.Tag.Get("in") {
			case "path":
				path = strings.Replace(path, "{"+name+"}",
					url.PathEscape(fmt.Sprint(value.Interface())), 1)
			case "query":
				if !value.IsZero() {
					query.Set(name, fmt.Sprint(value.Interface()))
				}
			case "body":
				whole = value.Interface()
			case "form":
				hasForm = true
				formPart(parts, name, value)
			default:
				body[name] = value.Interface()
			}
		}
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	switch {
	case hasForm:
		_ = parts.Close()

		return path, form.Bytes(), parts.FormDataContentType(), nil
	case whole != nil:
		data, err := json.Marshal(whole)

		return path, data, jsonType, err
	case len(body) > 0:
		data, err := json.Marshal(body)

		return path, data, jsonType, err
	}

	return path, nil, jsonType, nil
}

// jsonType is the content type of a JSON body.
const jsonType = "application/json"

// formPart writes a field of a form: a []byte as the file of that name,
// and each other value as text. A zero value is not sent, so the server
// applies its default.
func formPart(
	parts *multipart.Writer,
	name string,
	value reflect.Value,
) {
	if value.IsZero() {
		return
	}
	if data, ok := value.Interface().([]byte); ok {
		file, _ := parts.CreateFormFile(name, name)
		_, _ = file.Write(data)

		return
	}
	_ = parts.WriteField(name, fmt.Sprint(value.Interface()))
}

// answered is the body of an error answer whose status the method
// .Answers with its output, such as a probe's 503: it prints as an answer.
func answered(
	m *endpoint.Method,
	err error,
) (json.RawMessage, bool) {
	var a interface{ Answer() (int, []byte) }
	if !errors.As(err, &a) {
		return nil, false
	}
	code, body := a.Answer()
	for _, ans := range m.Answering() {
		if ans.Code == code {
			return body, true
		}
	}

	return nil, false
}
