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
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/validation/rule"
)

// Schema is a JSON Schema for one type, as much of the draft as osapi's
// inputs and outputs use.
type Schema struct {
	Type        string             `json:"type,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	// Order is the order the properties were declared in, for rendering.
	Order                []string `json:"-"`
	Required             []string `json:"required,omitempty"`
	Items                *Schema  `json:"items,omitempty"`
	AdditionalProperties *Schema  `json:"additionalProperties,omitempty"`
	Enum                 []string `json:"enum,omitempty"`
	Pattern              string   `json:"pattern,omitempty"`
	Format               string   `json:"format,omitempty"`
	MinLength            *int     `json:"minLength,omitempty"`
	MaxLength            *int     `json:"maxLength,omitempty"`
	Minimum              *float64 `json:"minimum,omitempty"`
	Maximum              *float64 `json:"maximum,omitempty"`
	MinItems             *int     `json:"minItems,omitempty"`
	MaxItems             *int     `json:"maxItems,omitempty"`
	// WriteOnly marks a value a client sends and is never shown back, such as
	// a field tagged secret.
	WriteOnly bool `json:"writeOnly,omitempty"`
	// Validate is the field's whole validate tag, which the spec carries in
	// rule.Extension.
	Validate string `json:"-"`
}

// SchemaFor builds the JSON Schema for t from its struct tags: json names
// the field, doc describes it, and validate gives its constraints. A struct
// that holds itself, such as a step whose steps are steps, is expanded one
// time: inside itself it is an object that names its type.
func SchemaFor(
	t reflect.Type,
) (*Schema, error) {
	return schemaFor(t, map[reflect.Type]bool{})
}

func schemaFor(
	t reflect.Type,
	open map[reflect.Type]bool,
) (*Schema, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}, nil
	case reflect.Bool:
		return &Schema{Type: "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}, nil
	case reflect.Slice, reflect.Array:
		items, err := schemaFor(t.Elem(), open)
		if err != nil {
			return nil, err
		}

		return &Schema{Type: "array", Items: items}, nil
	case reflect.Map:
		values, err := schemaFor(t.Elem(), open)
		if err != nil {
			return nil, err
		}

		return &Schema{Type: "object", AdditionalProperties: values}, nil
	case reflect.Struct:
		if open[t] {
			return &Schema{Type: "object", Description: "a " + t.Name() + ", as above"}, nil
		}
		open[t] = true
		defer delete(open, t)

		return structSchema(t, open)
	default:
		return nil, fmt.Errorf("type %s: kind %s has no JSON Schema", t, t.Kind())
	}
}

func structSchema(
	t reflect.Type,
	open map[reflect.Type]bool,
) (*Schema, error) {
	s := &Schema{Type: "object", Properties: map[string]*Schema{}}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := JSONName(f)
		if name == "-" {
			continue
		}

		fs, err := schemaFor(f.Type, open)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", t.Name(), f.Name, err)
		}
		if doc := f.Tag.Get("doc"); doc != "" {
			fs.Description = doc
		}
		if Secret(f) {
			fs.Format = "password"
			fs.WriteOnly = true
		}

		fs.Validate = f.Tag.Get("validate")
		required, err := rule.Apply(fs.Validate, target{fs})
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", t.Name(), f.Name, err)
		}
		if required {
			s.Required = append(s.Required, name)
		}

		s.Properties[name] = fs
		s.Order = append(s.Order, name)
	}

	return s, nil
}

// JSONName is the field's name in JSON: the json tag's name, or the Go name.
func JSONName(
	f reflect.StructField,
) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return f.Name
	}

	return name
}

// target is a Schema as a rule.Target, which the constraints of a validate
// tag are set on.
type target struct{ s *Schema }

func (t target) Enum(values []string) { t.s.Enum = values }

func (t target) Pattern(re string) { t.s.Pattern = re }

func (t target) Format(format string) { t.s.Format = format }

// Items is the target of the array's items. A schema with no items has
// nothing to constrain, so the rules go to a schema nobody reads.
func (t target) Items() rule.Target {
	if t.s.Items == nil {
		return target{&Schema{}}
	}

	return target{t.s.Items}
}

// Bound sets min, max or len as the constraint that fits the type: length
// for a string, items for an array, value for a number.
func (t target) Bound(
	r string,
	n float64,
) {
	i := int(n)
	switch t.s.Type {
	case "string":
		if r != "max" {
			t.s.MinLength = &i
		}
		if r != "min" {
			t.s.MaxLength = &i
		}
	case "array":
		if r != "max" {
			t.s.MinItems = &i
		}
		if r != "min" {
			t.s.MaxItems = &i
		}
	default:
		if r != "max" {
			t.s.Minimum = &n
		}
		if r != "min" {
			t.s.Maximum = &n
		}
	}
}

// Secret reports whether f carries a secret, tagged secret:"true". A secret
// is redacted from the audit log, marked write-only in the spec, and never
// recorded on a span.
func Secret(
	f reflect.StructField,
) bool {
	return f.Tag.Get("secret") == "true"
}

// TypeName is the name of t as a type list shows it: []Entry for a list.
func TypeName(
	t reflect.Type,
) string {
	if t.Kind() == reflect.Slice {
		return "[]" + t.Elem().Name()
	}

	return t.Name()
}

// Shape is t with the fields of its schema s, as dsl behavior get and
// dsl endpoint get show a type: CreateInput { key*, value* }, with
// each required field of an input marked *.
func Shape(
	t reflect.Type,
	s *Schema,
	input bool,
) string {
	name := TypeName(t)
	if s.Items != nil {
		s = s.Items
	}
	if len(s.Order) == 0 {
		return name
	}
	parts := make([]string, 0, len(s.Order))
	for _, f := range s.Order {
		if input && slices.Contains(s.Required, f) {
			f += "*"
		}
		parts = append(parts, f)
	}

	return name + " { " + strings.Join(parts, ", ") + " }"
}

// FieldShape is t with its top-level JSON fields, as Shape shows a type,
// from its struct tags alone: a required input field, validate:"required",
// is marked *. It does not walk into the fields, so a type that holds itself,
// such as a span tree, has a shape too.
func FieldShape(
	t reflect.Type,
	input bool,
) string {
	name := TypeName(t)
	for t.Kind() == reflect.Slice || t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return name
	}
	var parts []string
	for i := range t.NumField() {
		f := t.Field(i)
		json := strings.Split(f.Tag.Get("json"), ",")[0]
		if !f.IsExported() || json == "-" || json == "" {
			continue
		}
		if input && slices.Contains(strings.Split(f.Tag.Get("validate"), ","), "required") {
			json += "*"
		}
		parts = append(parts, json)
	}
	if len(parts) == 0 {
		return name
	}

	return name + " { " + strings.Join(parts, ", ") + " }"
}
