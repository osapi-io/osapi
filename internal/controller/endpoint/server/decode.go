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

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/validation"
)

// Param sets *field from the path parameter name, the method's key.
func Param[T any](
	c *echo.Context,
	name string,
	field *T,
) (string, bool) {
	return text(c.Param(name), "path parameter "+name, field)
}

// Query sets *field from the query parameter name, when the request has it.
func Query[T any](
	c *echo.Context,
	name string,
	field *T,
) (string, bool) {
	return text(c.QueryParam(name), "query parameter "+name, field)
}

// Body decodes the JSON body into in. An empty body leaves in as it is, for
// Valid to report what is missing.
func Body(
	c *echo.Context,
	in any,
) (string, bool) {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return "read body: " + err.Error(), false
	}
	if len(body) == 0 {
		return "", true
	}
	if err := json.Unmarshal(body, in); err != nil {
		return "invalid JSON body: " + err.Error(), false
	}

	return "", true
}

// text sets *field from value: a string as it is, any other type as JSON,
// so a number or a boolean in the path or the query is read as its type.
func text[T any](
	value string,
	what string,
	field *T,
) (string, bool) {
	if value == "" {
		return "", true
	}
	if v := reflect.ValueOf(field).Elem(); v.Kind() == reflect.String {
		v.SetString(value)

		return "", true
	}
	if err := json.Unmarshal([]byte(value), field); err != nil {
		return "invalid " + what + ": " + err.Error(), false
	}

	return "", true
}

// Valid checks in's validate tags, and each string field tagged
// pattern:"..." against its regular expression.
func Valid(
	in any,
) (string, bool) {
	if msg, ok := validation.Struct(in); !ok {
		return msg, false
	}
	// A generated decoder calls Valid only for an input with fields, a
	// struct, as validation.Struct needs.
	v := reflect.ValueOf(in)
	for i := range v.NumField() {
		f := v.Type().Field(i)
		p := f.Tag.Get("pattern")
		if p == "" || v.Field(i).Kind() != reflect.String {
			continue
		}
		if !regexp.MustCompile(p).MatchString(v.Field(i).String()) {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "" {
				name = f.Name
			}

			return fmt.Sprintf("%s does not match %s", name, p), false
		}
	}

	return "", true
}

// maxForm is the most of a multipart form that the server holds in memory;
// the rest goes to temporary files.
const maxForm = 32 << 20

// Form reads the fields of in that are tagged in:"form" from a
// multipart/form-data body: a string from its value, and a []byte from the
// file of that name.
func Form(
	c *echo.Context,
	in any,
) (string, bool) {
	req := c.Request()
	if err := req.ParseMultipartForm(maxForm); err != nil {
		return fmt.Sprintf("the body is not a multipart form: %s", err), false
	}

	v := reflect.ValueOf(in).Elem()
	for i := range v.NumField() {
		f := v.Type().Field(i)
		if f.Tag.Get("in") != "form" {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if f.Type == reflect.TypeFor[[]byte]() {
			file, _, err := req.FormFile(name)
			if err != nil {
				continue
			}
			// ParseMultipartForm stored the part whole, in memory or in a
			// temporary file, so it reads to its end.
			data, _ := io.ReadAll(file)
			_ = file.Close()
			v.Field(i).SetBytes(data)

			continue
		}
		v.Field(i).SetString(req.FormValue(name))
	}

	return "", true
}
