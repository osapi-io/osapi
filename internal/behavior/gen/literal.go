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

	"github.com/dave/jennifer/jen"
)

// literal is v as Go source: an example's input as the composite literal a
// person would write, with zero fields left out.
func literal(
	v reflect.Value,
) (jen.Code, error) {
	t := v.Type()

	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return scalar(v), nil
	case reflect.Struct:
		fields := jen.Dict{}
		for i := range t.NumField() {
			f := t.Field(i)
			if !f.IsExported() || v.Field(i).IsZero() {
				continue
			}
			c, err := literal(v.Field(i))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}
			fields[jen.Id(f.Name)] = c
		}

		return typeOf(t).Values(fields), nil
	case reflect.Slice:
		if v.IsNil() {
			return jen.Nil(), nil
		}
		elems := make([]jen.Code, 0, v.Len())
		for i := range v.Len() {
			c, err := literal(v.Index(i))
			if err != nil {
				return nil, err
			}
			elems = append(elems, c)
		}

		return typeOf(t).Values(elems...), nil
	case reflect.Map:
		if v.IsNil() {
			return jen.Nil(), nil
		}
		entries := jen.Dict{}
		for _, k := range v.MapKeys() {
			kc, err := literal(k)
			if err != nil {
				return nil, err
			}
			vc, err := literal(v.MapIndex(k))
			if err != nil {
				return nil, err
			}
			entries[kc] = vc
		}

		return typeOf(t).Values(entries), nil
	default:
		return nil, fmt.Errorf("a %s has no Go literal here", t)
	}
}

// scalar is a string, bool or number, converted to its named type when it
// has one.
func scalar(
	v reflect.Value,
) *jen.Statement {
	var lit *jen.Statement
	switch v.Kind() {
	case reflect.String:
		lit = jen.Lit(v.String())
	case reflect.Bool:
		lit = jen.Lit(v.Bool())
	case reflect.Float32, reflect.Float64:
		lit = jen.Lit(v.Float())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		lit = jen.Lit(int(v.Uint()))
	default:
		lit = jen.Lit(int(v.Int()))
	}

	t := v.Type()
	// An untyped literal already is a string, bool, int or float64.
	if t.PkgPath() == "" {
		switch t.Kind() {
		case reflect.String, reflect.Bool, reflect.Int, reflect.Float64:
			return lit
		}
	}

	return typeOf(t).Call(lit)
}

// typeOf is t's name in Go source.
func typeOf(
	t reflect.Type,
) *jen.Statement {
	if t.Name() != "" {
		if t.PkgPath() == "" {
			return jen.Id(t.Name())
		}

		return jen.Qual(t.PkgPath(), t.Name())
	}

	switch t.Kind() {
	case reflect.Slice:
		return jen.Index().Add(typeOf(t.Elem()))
	case reflect.Map:
		return jen.Map(typeOf(t.Key())).Add(typeOf(t.Elem()))
	default:
		return jen.Op("*").Add(typeOf(t.Elem()))
	}
}
