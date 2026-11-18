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

// Package genutil holds what the endpoint layer's writers share: the fields
// of an input as JSON reads them, and the house layout of a jen group.
// routegen writes routes and decoders, and commandgen writes commands, with
// it.
package genutil

import (
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
)

func StructFields(
	t reflect.Type,
) []reflect.StructField {
	if t.Kind() != reflect.Struct {
		return nil
	}
	var out []reflect.StructField
	for f := range t.Fields() {
		if f.IsExported() && JSONName(f) != "-" {
			out = append(out, f)
		}
	}

	return out
}

func JSONName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}

	return name
}

// multi lays out a group one item to a line, as the house style writes
// parameter lists and literals.
func Multi(
	open string,
	closing string,
	items ...jen.Code,
) *jen.Statement {
	return jen.Custom(
		jen.Options{Open: open, Close: closing, Separator: ",", Multi: true},
		items...)
}

func Params(
	items ...jen.Code,
) *jen.Statement {
	return Multi("(", ")", items...)
}
