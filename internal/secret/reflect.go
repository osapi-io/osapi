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

package secret

import (
	"reflect"
	"slices"
	"strings"
	"sync"
)

// valueType is the type of a secret.
var valueType = reflect.TypeFor[Value]()

// Is reports whether a field of type t holds a secret: a Value, or a
// pointer, a slice, an array or a map of them, such as
// map[string]secret.Value.
func Is(
	t reflect.Type,
) bool {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		return Is(t.Elem())
	default:
		return t == valueType
	}
}

// Fields are the JSON names of the fields of t, a struct or a pointer to
// one, that hold a secret, at each depth: the names that the audit redacts.
func Fields(
	t reflect.Type,
) []string {
	var names []string
	fields(t, map[reflect.Type]bool{}, &names)
	slices.Sort(names)

	return slices.Compact(names)
}

func fields(
	t reflect.Type,
	seen map[reflect.Type]bool,
	names *[]string,
) {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice ||
		t.Kind() == reflect.Array || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct || seen[t] {
		return
	}
	seen[t] = true
	for f := range t.Fields() {
		name := jsonName(f)
		// JSON reads the fields of an embedded struct, also of one whose
		// type is not exported.
		if (!f.IsExported() && !f.Anonymous) || name == "-" {
			continue
		}
		if Is(f.Type) {
			*names = append(*names, name)

			continue
		}
		fields(f.Type, seen, names)
	}
}

// jsonName is the name of f in JSON: its json tag's, or the field's.
func jsonName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}

	return name
}

// References are the references in x, at each depth, in the order of its
// fields, with the keys of a map in order.
func References(
	x any,
) []string {
	var refs []string
	walk(reflect.ValueOf(x), func(v Value) {
		if v.IsReference() {
			refs = append(refs, string(v))
		}
	})

	return refs
}

// Resolve replaces each reference in *x, at each depth, with the value
// that fn gives for it. x is a pointer. The first error of fn stops it, and
// is its error.
func Resolve(
	x any,
	fn func(ref string) (string, error),
) error {
	var err error
	set(reflect.ValueOf(x), func(v Value) Value {
		if err != nil || !v.IsReference() {
			return v
		}
		value, e := fn(string(v))
		if e != nil {
			err = e

			return v
		}

		return Value(value)
	})

	return err
}

// set replaces each Value in v, at each depth, with what fn gives for it.
// v is settable, or holds a pointer, a slice or a map, whose elements are.
func set(
	v reflect.Value,
	fn func(Value) Value,
) {
	if !v.IsValid() {
		return
	}
	if v.Type() == valueType {
		if v.CanSet() {
			v.SetString(string(fn(Value(v.String()))))
		}

		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		set(v.Elem(), fn)
	case reflect.Struct:
		for i := range v.NumField() {
			if f := v.Type().Field(i); f.IsExported() || f.Anonymous {
				set(v.Field(i), fn)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			set(v.Index(i), fn)
		}
	case reflect.Map:
		// The element of a map is not settable, so a copy changes and
		// goes back in.
		iter := v.MapRange()
		for iter.Next() {
			e := reflect.New(v.Type().Elem()).Elem()
			e.Set(iter.Value())
			set(e, fn)
			v.SetMapIndex(iter.Key(), e)
		}
	}
}

// walk calls fn with each Value in v, at each depth.
func walk(
	v reflect.Value,
	fn func(Value),
) {
	if !v.IsValid() {
		return
	}
	if v.Type() == valueType {
		fn(Value(v.String()))

		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		walk(v.Elem(), fn)
	case reflect.Struct:
		for i := range v.NumField() {
			if f := v.Type().Field(i); f.IsExported() || f.Anonymous {
				walk(v.Field(i), fn)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			walk(v.Index(i), fn)
		}
	case reflect.Map:
		keys := v.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int {
			return strings.Compare(a.String(), b.String())
		})
		for _, k := range keys {
			walk(v.MapIndex(k), fn)
		}
	}
}

// Plain is a copy of x in which each Value is a string, so it marshals its
// value: the CLI sends the request with it, and the dispatch of a behavior
// queues the job with it. Each other marshal of x gives the reference, or
// Redacted. A struct that holds a secret becomes a struct of the same
// fields and tags, so JSON reads it as it reads x. x is as it is when it
// holds no secret.
func Plain(
	x any,
) any {
	if x == nil {
		return nil
	}
	v := reflect.ValueOf(x)
	t := plainType(v.Type())
	if t == v.Type() {
		return x
	}

	return convert(v, t).Interface()
}

// plainTypes are the plain types that plainType made, by the type that
// each is made from.
var plainTypes sync.Map

// plainType is t with each Value a string. A type that holds no secret is
// t.
func plainType(
	t reflect.Type,
) reflect.Type {
	if p, ok := plainTypes.Load(t); ok {
		return p.(reflect.Type)
	}
	p := makePlainType(t)
	plainTypes.Store(t, p)

	return p
}

func makePlainType(
	t reflect.Type,
) reflect.Type {
	if t == valueType {
		return reflect.TypeFor[string]()
	}
	switch t.Kind() {
	case reflect.Pointer:
		if e := plainType(t.Elem()); e != t.Elem() {
			return reflect.PointerTo(e)
		}
	case reflect.Slice:
		if e := plainType(t.Elem()); e != t.Elem() {
			return reflect.SliceOf(e)
		}
	case reflect.Array:
		if e := plainType(t.Elem()); e != t.Elem() {
			return reflect.ArrayOf(t.Len(), e)
		}
	case reflect.Map:
		if e := plainType(t.Elem()); e != t.Elem() {
			return reflect.MapOf(t.Key(), e)
		}
	case reflect.Struct:
		return plainStruct(t)
	}

	return t
}

// plainStruct is t with its exported fields, each with its plain type and
// its tags, when one of them holds a secret, or t. JSON reads only the
// exported fields, so the copy has no other. The fields of an embedded
// struct whose type is not exported are fields of the copy, as JSON reads
// them.
func plainStruct(
	t reflect.Type,
) reflect.Type {
	out, changed := plainFields(t)
	if !changed {
		return t
	}

	return reflect.StructOf(out)
}

// plainFields are the fields of plainStruct, and whether one of them holds
// a secret.
func plainFields(
	t reflect.Type,
) ([]reflect.StructField, bool) {
	var out []reflect.StructField
	changed := false
	for f := range t.Fields() {
		if !f.IsExported() {
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				inner, c := plainFields(f.Type)
				out = append(out, inner...)
				changed = changed || c
			}

			continue
		}
		p := plainType(f.Type)
		changed = changed || p != f.Type
		out = append(out, reflect.StructField{
			Name: f.Name, Type: p, Tag: f.Tag, Anonymous: f.Anonymous,
		})
	}

	return out, changed
}

// convert is v as a value of t, its plain type.
func convert(
	v reflect.Value,
	t reflect.Type,
) reflect.Value {
	if v.Type() == t {
		return v
	}
	out := reflect.New(t).Elem()
	switch v.Kind() {
	case reflect.String:
		out.SetString(v.String())
	case reflect.Pointer:
		if !v.IsNil() {
			out.Set(reflect.New(t.Elem()))
			out.Elem().Set(convert(v.Elem(), t.Elem()))
		}
	case reflect.Slice:
		if !v.IsNil() {
			out.Set(reflect.MakeSlice(t, v.Len(), v.Len()))
			for i := range v.Len() {
				out.Index(i).Set(convert(v.Index(i), t.Elem()))
			}
		}
	case reflect.Array:
		for i := range v.Len() {
			out.Index(i).Set(convert(v.Index(i), t.Elem()))
		}
	case reflect.Map:
		if !v.IsNil() {
			out.Set(reflect.MakeMapWithSize(t, v.Len()))
			iter := v.MapRange()
			for iter.Next() {
				out.SetMapIndex(iter.Key(), convert(iter.Value(), t.Elem()))
			}
		}
	case reflect.Struct:
		for i := range t.NumField() {
			f := t.Field(i)
			out.Field(i).Set(convert(v.FieldByName(f.Name), f.Type))
		}
	}

	return out
}
