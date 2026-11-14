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
	"encoding/json"
	"io"
	"reflect"

	"github.com/labstack/echo/v5"

	"github.com/osapi-io/osapi/internal/validation"
)

// Body decodes a write's JSON body into in. An empty body leaves in as it
// is, for Valid to report what is missing.
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

// Query sets *field from the query parameter name, when the request has it.
func Query[T any](
	c *echo.Context,
	name string,
	field *T,
) (string, bool) {
	return text(c.QueryParam(name), "query parameter "+name, field)
}

// Param sets *field from the path parameter name, the method's key.
func Param[T any](
	c *echo.Context,
	name string,
	field *T,
) (string, bool) {
	return text(c.Param(name), "path parameter "+name, field)
}

// text sets *field from value. A field of a string kind takes the text as it
// is; any other field takes it as the JSON it spells, 3 or true. An empty value
// leaves the field as it is.
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

// Valid checks in's validate tags.
func Valid(
	in any,
) (string, bool) {
	return validation.Struct(in)
}
