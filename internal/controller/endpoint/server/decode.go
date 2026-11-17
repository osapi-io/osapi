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
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/labstack/echo/v5"

	behaviorserver "github.com/osapi-io/osapi/internal/behavior/server"
	"github.com/osapi-io/osapi/internal/validation"
)

// Param sets *field from the path parameter name, the method's key.
func Param[T any](
	c *echo.Context,
	name string,
	field *T,
) (string, bool) {
	return behaviorserver.Param(c, name, field)
}

// Query sets *field from the query parameter name, when the request has it.
func Query[T any](
	c *echo.Context,
	name string,
	field *T,
) (string, bool) {
	return behaviorserver.Query(c, name, field)
}

// Body decodes the JSON body into in. An empty body leaves in as it is, for
// Valid to report what is missing.
func Body(
	c *echo.Context,
	in any,
) (string, bool) {
	return behaviorserver.Body(c, in)
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
