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

package endpoint

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// bearer is the security scheme of every endpoint that is not public.
const bearer = "BearerAuth"

// ErrorRef is the shared ErrorResponse schema, as a group's spec refers to
// it from internal/controller/api/<group>/gen, where the hand-written specs
// are and the generated ones are compared with them.
const ErrorRef = "../../components/gen/api.yaml#/components/schemas/ErrorResponse"

// errorSchema is the name of the shared error schema.
const errorSchema = "ErrorResponse"

// OpenAPI is the group's part of the spec. errorRef is where the shared
// ErrorResponse schema is, relative to the file the spec is written to.
func OpenAPI(
	g *Group,
	errorRef string,
) (*openapi3.T, error) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: g.Title, Version: "1.0.0", Description: g.Doc},
		Paths:   openapi3.NewPaths(),
		Tags: openapi3.Tags{{
			Name:        g.Tag,
			Description: g.TagDoc,
			Extensions:  map[string]any{"x-displayName": g.TagDisplay},
		}},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				bearer: &openapi3.SecuritySchemeRef{Value: openapi3.NewJWTSecurityScheme()},
			},
			Schemas: openapi3.Schemas{errorSchema: &openapi3.SchemaRef{Ref: errorRef}},
		},
	}

	b := &builder{schemas: doc.Components.Schemas}
	for _, m := range g.Methods {
		op, err := b.operation(g, m)
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", g.Name, m.Name, err)
		}
		path := g.Path(m)
		item := doc.Paths.Value(path)
		if item == nil {
			item = &openapi3.PathItem{}
			doc.Paths.Set(path, item)
		}
		item.SetOperation(m.Verb, op)
	}

	return doc, nil
}

// builder adds each named type it meets to the components, once.
type builder struct {
	schemas openapi3.Schemas
}

func (b *builder) operation(
	g *Group,
	m *Method,
) (*openapi3.Operation, error) {
	op := &openapi3.Operation{
		OperationID: m.OperationID(g),
		Summary:     m.summary,
		Description: m.explain,
		Tags:        []string{g.Tag},
		Responses:   openapi3.NewResponses(),
	}
	if perm := g.Permission(m); perm != "" {
		op.Security = &openapi3.SecurityRequirements{{bearer: []string{perm}}}
	} else {
		op.Security = &openapi3.SecurityRequirements{}
	}

	var body []reflect.StructField
	for _, f := range fields(m.In) {
		switch in := f.Tag.Get("in"); in {
		case "path", "query":
			s, err := b.ref(f.Type)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}
			op.Parameters = append(op.Parameters, &openapi3.ParameterRef{Value: &openapi3.Parameter{
				Name:        jsonName(f),
				In:          in,
				Required:    in == "path" || validates(f, "required"),
				Description: f.Tag.Get("doc"),
				Schema:      constrained(s, f),
			}})
		default:
			body = append(body, f)
		}
	}
	if len(body) > 0 {
		s, required, err := b.body(body)
		if err != nil {
			return nil, err
		}
		op.RequestBody = &openapi3.RequestBodyRef{Value: openapi3.NewRequestBody().
			WithRequired(required).
			WithJSONSchemaRef(s)}
	}

	returns := m.returns
	if returns == "" {
		returns = m.summary
	}
	success := openapi3.NewResponse().WithDescription(returns)
	if m.status != http.StatusNoContent {
		s, err := b.ref(m.Out)
		if err != nil {
			return nil, fmt.Errorf("response: %w", err)
		}
		success = success.WithJSONSchemaRef(s)
	}
	op.Responses.Set(strconv.Itoa(m.status), &openapi3.ResponseRef{Value: success})

	if len(fields(m.In)) > 0 {
		b.fail(op, http.StatusBadRequest, "The request is not valid.")
	}
	if g.Permission(m) != "" {
		b.fail(op, http.StatusUnauthorized, "Unauthorized - API key required")
		b.fail(op, http.StatusForbidden, "Forbidden - Insufficient permissions")
	}
	for _, e := range m.errors {
		b.fail(op, e.Code, e.Doc)
	}
	b.fail(op, http.StatusInternalServerError, "An error on the controller.")

	return op, nil
}

// fail adds an error response with the shared error schema.
func (b *builder) fail(
	op *openapi3.Operation,
	code int,
	doc string,
) {
	op.Responses.Set(strconv.Itoa(code), &openapi3.ResponseRef{Value: openapi3.NewResponse().
		WithDescription(doc).
		WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/" + errorSchema})})
}

// body is the request body: the type of a field tagged in:"body", or an
// object of the input's other fields. It reports whether the body is
// required, which it is when one of its fields is.
func (b *builder) body(
	fs []reflect.StructField,
) (*openapi3.SchemaRef, bool, error) {
	if len(fs) == 1 && fs[0].Tag.Get("in") == "body" {
		s, err := b.ref(fs[0].Type)

		return s, true, err
	}
	s, err := b.object(fs, false)
	if err != nil {
		return nil, false, err
	}

	return openapi3.NewSchemaRef("", s), len(s.Required) > 0, nil
}

// ref is the schema of t: a $ref to the components for a named struct,
// which it adds the first time, or the schema in place for any other type.
func (b *builder) ref(
	t reflect.Type,
) (*openapi3.SchemaRef, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct || t.Name() == "" || t == reflect.TypeFor[time.Time]() {
		s, err := b.inline(t)
		if err != nil {
			return nil, err
		}

		return openapi3.NewSchemaRef("", s), nil
	}

	ref := "#/components/schemas/" + t.Name()
	if _, ok := b.schemas[t.Name()]; !ok {
		// The name goes in first, so a type that refers to itself ends.
		b.schemas[t.Name()] = &openapi3.SchemaRef{}
		s, err := b.object(fields(t), true)
		if err != nil {
			delete(b.schemas, t.Name())

			return nil, fmt.Errorf("%s: %w", t.Name(), err)
		}
		b.schemas[t.Name()] = openapi3.NewSchemaRef("", s)
	}

	return &openapi3.SchemaRef{Ref: ref}, nil
}

// inline is the schema of a type that is not a named struct.
func (b *builder) inline(
	t reflect.Type,
) (*openapi3.Schema, error) {
	if t == reflect.TypeFor[time.Time]() {
		return openapi3.NewDateTimeSchema(), nil
	}
	switch t.Kind() {
	case reflect.String:
		return openapi3.NewStringSchema(), nil
	case reflect.Bool:
		return openapi3.NewBoolSchema(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return openapi3.NewIntegerSchema(), nil
	case reflect.Float32, reflect.Float64:
		return openapi3.NewFloat64Schema(), nil
	case reflect.Interface:
		return &openapi3.Schema{}, nil
	case reflect.Slice, reflect.Array:
		items, err := b.ref(t.Elem())
		if err != nil {
			return nil, err
		}
		s := openapi3.NewArraySchema()
		s.Items = items

		return s, nil
	case reflect.Map:
		s := openapi3.NewObjectSchema()
		if t.Elem().Kind() == reflect.Interface {
			free := true
			s.AdditionalProperties = openapi3.AdditionalProperties{Has: &free}

			return s, nil
		}
		values, err := b.ref(t.Elem())
		if err != nil {
			return nil, err
		}
		s.AdditionalProperties = openapi3.AdditionalProperties{Schema: values}

		return s, nil
	case reflect.Struct:
		return b.object(fields(t), true)
	default:
		return nil, fmt.Errorf("type %s: kind %s has no schema", t, t.Kind())
	}
}

// object is an object of the fields. In a response, a field is required
// when its json tag has no omitempty; in a request, when it is tagged
// validate:"required".
func (b *builder) object(
	fs []reflect.StructField,
	response bool,
) (*openapi3.Schema, error) {
	s := openapi3.NewObjectSchema()
	s.Properties = openapi3.Schemas{}
	for _, f := range fs {
		name := jsonName(f)
		fs, err := b.ref(f.Type)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name, err)
		}
		if fs.Ref == "" {
			fs.Value.Description = f.Tag.Get("doc")
		}
		s.Properties[name] = constrained(fs, f)
		required := validates(f, "required")
		if response {
			required = !strings.Contains(f.Tag.Get("json"), ",omitempty")
		}
		if required {
			s.Required = append(s.Required, name)
		}
	}

	return s, nil
}

// constrained is s with the constraints of the field's validate tag: the
// values of oneof, and min and max as the length of a string.
func constrained(
	s *openapi3.SchemaRef,
	f reflect.StructField,
) *openapi3.SchemaRef {
	if s.Value == nil {
		return s
	}
	for _, rule := range strings.Split(f.Tag.Get("validate"), ",") {
		name, param, _ := strings.Cut(rule, "=")
		switch name {
		case "oneof":
			for _, v := range strings.Fields(param) {
				s.Value.Enum = append(s.Value.Enum, v)
			}
		case "min", "max":
			n, err := strconv.ParseUint(param, 10, 64)
			if err != nil || !s.Value.Type.Is(openapi3.TypeString) {
				continue
			}
			if name == "min" {
				s.Value.MinLength = n
			} else {
				s.Value.MaxLength = &n
			}
		}
	}

	return s
}

// validates reports whether the field's validate tag has the rule.
func validates(
	f reflect.StructField,
	rule string,
) bool {
	return slices.Contains(strings.Split(f.Tag.Get("validate"), ","), rule)
}

// fields is the exported fields of a struct type that JSON reads, or none
// for any other type.
func fields(
	t reflect.Type,
) []reflect.StructField {
	if t.Kind() != reflect.Struct {
		return nil
	}
	var out []reflect.StructField
	for f := range t.Fields() {
		if f.IsExported() && jsonName(f) != "-" {
			out = append(out, f)
		}
	}

	return out
}

// jsonName is the field's name in JSON: the json tag's name, or the Go name.
func jsonName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}

	return name
}

func lowerVerb(
	verb string,
) string {
	return strings.ToLower(verb)
}

// upper is s with its first letter in upper case, and "" for "".
func upper(
	s string,
) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
