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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/validation/rule"

	"github.com/getkin/kin-openapi/openapi3"
)

// bearer is the security scheme of every endpoint that is not public.
const bearer = "BearerAuth"

// errorSchema is the name of the shared error schema.
const errorSchema = "ErrorResponse"

// reasonDoc is the description of the reason field, before what each
// reason means.
const reasonDoc = "Why the request failed, as a word that a program can branch on. " +
	"The HTTP status is the class of the problem; the reason is the cause, " +
	"and the error is the message for a person, which can change between releases."

// ErrorSchema is the schema of ErrorResponse, the body of each error
// answer: its message, and its reason, one of the values of types.Reason, with what each means.
func ErrorSchema() *openapi3.Schema {
	text := func(doc string, example string) *openapi3.SchemaRef {
		s := openapi3.NewStringSchema()
		s.Description = doc
		s.Example = example

		return s.NewRef()
	}
	doc := reasonDoc
	reason := openapi3.NewStringSchema()
	for _, r := range types.Reason("").Values() {
		doc += " " + string(r) + ": " + ReasonDocs[r]
		reason.Enum = append(reason.Enum, string(r))
	}
	reason.Description = doc
	reason.Example = string(types.ReasonNotFound)

	s := openapi3.NewObjectSchema()
	s.Properties = openapi3.Schemas{
		"error":   text("A description of the error that occurred.", "Failed to retrieve status."),
		"details": text("Additional details about the error.", "Failed due to network timeout."),
		"reason":  reason.NewRef(),
	}

	return s
}

// MethodExtension is the extension of each operation of the spec that names
// the method in its group, such as get or drain.
const MethodExtension = "x-osapi-method"

// AreaExtension is the extension of each operation and each tag of the spec
// that names its area, one of area.All, such as fleet. The full spec groups
// the tags by it in x-tagGroups.
const AreaExtension = "x-osapi-area"

// AreasExtension is the extension of the full spec that lists the areas,
// area.All, in their order, each with its name and its label. A generator
// that reads the spec, such as the SDK's pages, takes the areas from it.
const AreasExtension = "x-osapi-areas"

// GroupExtension is the extension of each tag of the spec that names its
// group, such as dns or agent: the segment of its paths that the group's own
// operations start at. The SDK names its services from it, as the tag's
// x-displayName is a title for a person, such as Package updates.
const GroupExtension = "x-osapi-group"

// ProgramExtension marks an operation that a program calls, not a person,
// which NoCommand sets: it has no CLI command, and the SDK needs no example
// of it, since an example would run it against a controller.
const ProgramExtension = "x-osapi-program"

// SecretExtension marks a field, or the value of a map, of type
// secret.Value: a reference, secret://<vault>/<key>, or the value.
const SecretExtension = "x-osapi-secret"

// OpenAPI is the group's part of the spec. It has the shared ErrorResponse
// schema, so the spec stands alone.
func OpenAPI(
	g *Group,
) (*openapi3.T, error) {
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info:    &openapi3.Info{Title: g.Title, Version: "1.0.0"},
		Paths:   openapi3.NewPaths(),
		Tags: openapi3.Tags{{
			Name:        g.Tag,
			Description: g.TagDoc,
			Extensions: map[string]any{
				"x-displayName": g.TagDisplay,
				GroupExtension:  g.Name,
				AreaExtension:   string(g.Area),
			},
		}},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				bearer: &openapi3.SecuritySchemeRef{Value: openapi3.NewJWTSecurityScheme()},
			},
			Schemas: openapi3.Schemas{errorSchema: ErrorSchema().NewRef()},
		},
	}

	b := &builder{
		schemas:  doc.Components.Schemas,
		params:   openapi3.ParametersMap{},
		docs:     g.Docs,
		examples: g.Examples,
	}
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
		item.SetOperation(m.HTTPMethod(), op)
	}
	if len(b.params) > 0 {
		doc.Components.Parameters = b.params
	}

	return doc, nil
}

// builder adds each named type it meets to the components, once.
type builder struct {
	schemas  openapi3.Schemas
	params   openapi3.ParametersMap
	docs     map[reflect.Type]string
	examples map[reflect.Type]map[string]string
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
		// The method's name in its group, such as get or drain, from which
		// the SDK's generator names each typed method, and the group's area,
		// which puts the method's SDK page in its folder.
		Extensions: map[string]any{MethodExtension: m.Name, AreaExtension: string(g.Area)},
	}
	if m.noCommand {
		op.Extensions[ProgramExtension] = true
	}
	// The examples of the method, as the input that each one sends. The
	// SDK's generator writes a program that calls the method with each one.
	examples, err := specExamples(m)
	if err != nil {
		return nil, err
	}
	if len(examples) > 0 {
		op.Extensions[ExamplesExtension] = examples
	}
	if perm := g.Permission(m); perm != "" {
		op.Security = &openapi3.SecurityRequirements{{bearer: []string{string(perm)}}}
	} else {
		op.Security = &openapi3.SecurityRequirements{}
	}

	var body []reflect.StructField
	form := false
	for _, f := range fields(m.In) {
		switch in := f.Tag.Get("in"); in {
		case "path", "query":
			s, err := b.ref(f.Type)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}
			p := &openapi3.Parameter{
				Name:        jsonName(f),
				In:          in,
				Required:    in == "path" || validates(f, "required"),
				Description: f.Tag.Get("doc"),
			}
			if p.Schema, err = constrained(s, f, false); err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}
			// oapi-codegen copies this onto the parameter's field, so a Go
			// client made from the spec validates as the server does.
			if v := f.Tag.Get("validate"); v != "" {
				p.Extensions = extension(v)
			}
			// A shared parameter is in the components once, and each
			// operation refers to it.
			if name := f.Tag.Get("param"); name != "" {
				b.params[name] = &openapi3.ParameterRef{Value: p}
				op.Parameters = append(op.Parameters,
					&openapi3.ParameterRef{Ref: "#/components/parameters/" + name})

				continue
			}
			op.Parameters = append(op.Parameters, &openapi3.ParameterRef{Value: p})
		default:
			form = form || in == "form"
			body = append(body, f)
		}
	}
	if len(body) > 0 {
		s, required, err := b.body(body)
		if err != nil {
			return nil, err
		}
		rb := openapi3.NewRequestBody().WithRequired(required)
		// A field tagged in:"form" makes the body a multipart form, as an
		// upload of a file is.
		if form {
			rb = rb.WithFormDataSchemaRef(s)
		} else {
			rb = rb.WithJSONSchemaRef(s)
		}
		op.RequestBody = &openapi3.RequestBodyRef{Value: rb}
	}

	returns := m.returns
	if returns == "" {
		returns = m.summary
	}
	success := openapi3.NewResponse().WithDescription(returns)
	if m.status != http.StatusNoContent {
		out, err := b.ref(m.Out)
		if err != nil {
			return nil, fmt.Errorf("response: %w", err)
		}
		success = success.WithJSONSchemaRef(out)
		if m.location != "" {
			success.Headers = openapi3.Headers{"Location": &openapi3.HeaderRef{
				Value: &openapi3.Header{Parameter: openapi3.Parameter{
					Description: "The URL of the resource that the request created.",
					Schema:      openapi3.NewStringSchema().NewRef(),
				}},
			}}
		}
		// A status that Answers added has the output too.
		for _, a := range m.answers {
			op.Responses.Set(strconv.Itoa(a.Code), &openapi3.ResponseRef{
				Value: openapi3.NewResponse().WithDescription(a.Doc).WithJSONSchemaRef(out),
			})
		}
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
// required: a field tagged in:"body" when it is validate:"required", an
// object when one of its fields is.
func (b *builder) body(
	fs []reflect.StructField,
) (*openapi3.SchemaRef, bool, error) {
	if len(fs) == 1 && fs[0].Tag.Get("in") == "body" {
		s, err := b.ref(fs[0].Type)

		return s, validates(fs[0], "required"), err
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
	if values, ok := rule.Values(t); ok {
		return b.enum(t, values), nil
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
		s.Description = b.docs[t]
		for name, ex := range b.examples[t] {
			if p, ok := s.Properties[name]; ok && p.Value != nil {
				p.Value.Example = example(p.Value, ex)
			}
		}
		b.schemas[t.Name()] = openapi3.NewSchemaRef("", s)
	}

	return &openapi3.SchemaRef{Ref: ref}, nil
}

// enum is a $ref to the schema of t, a closed set, which it adds the first
// time, with the values of its Values.
func (b *builder) enum(
	t reflect.Type,
	values []string,
) *openapi3.SchemaRef {
	if _, ok := b.schemas[t.Name()]; !ok {
		s := openapi3.NewStringSchema()
		s.Description = b.docs[t]
		// A type of pkg/sdk/types that describes itself has one description
		// in each spec that uses it, so two specs merge.
		if d, ok := reflect.Zero(t).Interface().(interface{ Doc() string }); ok &&
			s.Description == "" {
			s.Description = d.Doc()
		}
		for _, v := range values {
			s.Enum = append(s.Enum, v)
		}
		b.schemas[t.Name()] = openapi3.NewSchemaRef("", s)
	}

	return &openapi3.SchemaRef{Ref: "#/components/schemas/" + t.Name()}
}

// inline is the schema of a type that is not a named struct.
func (b *builder) inline(
	t reflect.Type,
) (*openapi3.Schema, error) {
	if t == reflect.TypeFor[time.Time]() {
		return openapi3.NewDateTimeSchema(), nil
	}
	if t == reflect.TypeFor[secret.Value]() {
		s := openapi3.NewStringSchema().WithFormat("password")
		s.WriteOnly = true
		s.Extensions = map[string]any{SecretExtension: true}

		return s, nil
	}
	switch t.Kind() {
	case reflect.String:
		return openapi3.NewStringSchema(), nil
	case reflect.Bool:
		return openapi3.NewBoolSchema(), nil
	case reflect.Int64:
		// An int64 is an int64, and the other integers have no format, as
		// oapi-codegen maps them back.
		return openapi3.NewInt64Schema(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return openapi3.NewIntegerSchema(), nil
	case reflect.Float32:
		return openapi3.NewFloat64Schema(), nil
	case reflect.Float64:
		// A float64 is a double, and a float32 a number with no format, as
		// oapi-codegen maps them back.
		return openapi3.NewFloat64Schema().WithFormat("double"), nil
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
// when its json tag has neither omitempty nor omitzero; in a request, when
// it is tagged validate:"required".
func (b *builder) object(
	fs []reflect.StructField,
	response bool,
) (*openapi3.Schema, error) {
	s := openapi3.NewObjectSchema()
	s.Properties = openapi3.Schemas{}
	for _, f := range fs {
		name := jsonName(f)
		fs, err := b.field(f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name, err)
		}
		if fs.Ref == "" {
			fs.Value.Description = f.Tag.Get("doc")
		}
		if s.Properties[name], err = constrained(fs, f, true); err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name, err)
		}
		required := validates(f, "required")
		if response {
			tag := f.Tag.Get("json")
			required = !strings.Contains(tag, ",omitempty") && !strings.Contains(tag, ",omitzero")
		}
		if required {
			s.Required = append(s.Required, name)
		}
	}

	return s, nil
}

// field is the schema of a field: bytes are a file in a form and base64 in
// JSON, and each other type is its schema.
func (b *builder) field(
	f reflect.StructField,
) (*openapi3.SchemaRef, error) {
	if f.Type != reflect.TypeFor[[]byte]() {
		return b.ref(f.Type)
	}
	bytes := openapi3.NewBytesSchema()
	if f.Tag.Get("in") == "form" {
		bytes.Format = "binary"
	}

	return openapi3.NewSchemaRef("", bytes), nil
}

// constrained is s with the field's example, pattern and default, and the
// constraints of its validate tag. A field's own schema carries the whole tag
// in rule.Extension; a parameter carries it itself.
func constrained(
	s *openapi3.SchemaRef,
	f reflect.StructField,
	property bool,
) (*openapi3.SchemaRef, error) {
	v := f.Tag.Get("validate")
	if s.Value == nil {
		// A $ref, such as a closed set, keeps its constraints in the
		// components; the tag is still checked.
		_, err := rule.Apply(v, f.Type, target{&openapi3.Schema{}})

		return s, err
	}
	if ex, ok := f.Tag.Lookup("example"); ok {
		s.Value.Example = example(s.Value, ex)
	}
	if p := f.Tag.Get("pattern"); p != "" {
		s.Value.Pattern = p
	}
	// A path parameter is always given, so a default means nothing there:
	// the CLI's default, such as --target _all, comes from the tag itself.
	if d, ok := f.Tag.Lookup("default"); ok && f.Tag.Get("in") != "path" {
		s.Value.Default = example(s.Value, d)
	}
	if _, err := rule.Apply(v, f.Type, target{s.Value}); err != nil {
		return nil, err
	}
	if property && v != "" {
		s.Value.Extensions = extension(v)
	}
	// A secret is write-only, so a client never shows it, and a password,
	// so a form hides it. Its description says what it takes. A map of
	// secrets has the same schema for each value.
	if secret.Is(f.Type) {
		s.Value.WriteOnly = true
		if s.Value.Extensions == nil {
			s.Value.Extensions = map[string]any{}
		}
		s.Value.Extensions[SecretExtension] = true
		s.Value.Description = secret.Describe(s.Value.Description)
	}

	return s, nil
}

// extension is the extension that carries a whole validate tag.
func extension(
	validate string,
) map[string]any {
	return map[string]any{rule.Extension: map[string]string{"validate": validate}}
}

// target is a schema as a rule.Target, which the constraints of a validate
// tag are set on.
type target struct{ s *openapi3.Schema }

func (t target) Enum(values []string) {
	for _, v := range values {
		t.s.Enum = append(t.s.Enum, v)
	}
}

func (t target) Pattern(re string) { t.s.Pattern = re }

func (t target) Format(format string) { t.s.Format = format }

// Items is the target of the array's items. A schema with no items has
// nothing to constrain, so the rules go to a schema nobody reads.
func (t target) Items() rule.Target {
	if t.s.Items == nil || t.s.Items.Value == nil {
		return target{&openapi3.Schema{}}
	}

	return target{t.s.Items.Value}
}

// Bound sets min, max or len as the constraint that fits the type: the
// length of a string, the items of an array, the value of a number.
func (t target) Bound(
	r string,
	n float64,
) {
	u := uint64(n)
	isMin, isMax := r != "max", r != "min"
	switch {
	case t.s.Type.Is(openapi3.TypeString):
		if isMin {
			t.s.MinLength = u
		}
		if isMax {
			t.s.MaxLength = &u
		}
	case t.s.Type.Is(openapi3.TypeArray):
		if isMin {
			t.s.MinItems = u
		}
		if isMax {
			t.s.MaxItems = &u
		}
	default:
		if isMin {
			t.s.Min = &n
		}
		if isMax {
			t.s.Max = &n
		}
	}
}

// validates reports whether the field's validate tag has the rule.
func validates(
	f reflect.StructField,
	name string,
) bool {
	return rule.Has(f.Tag.Get("validate"), name)
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

// upper is s with its first letter in upper case, and "" for "".
func upper(
	s string,
) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

// example is the text of an example tag as the value its schema's type
// holds: a number for an integer or a number, a bool for a boolean, and the
// text for anything else, or when it does not parse.
func example(
	s *openapi3.Schema,
	text string,
) any {
	switch {
	case s.Type.Is(openapi3.TypeInteger):
		if n, err := strconv.ParseInt(text, 10, 64); err == nil {
			return n
		}
	case s.Type.Is(openapi3.TypeNumber):
		if n, err := strconv.ParseFloat(text, 64); err == nil {
			return n
		}
	case s.Type.Is(openapi3.TypeBoolean):
		if v, err := strconv.ParseBool(text); err == nil {
			return v
		}
	case s.Type.Is(openapi3.TypeObject) || s.Type.Is(openapi3.TypeArray):
		// An object or a list is its JSON: a string there is not a valid
		// example of the schema.
		var v any
		if err := json.Unmarshal([]byte(text), &v); err == nil {
			return v
		}
	}

	return text
}

// formats are the validate rules that are a format, as the behavior DSL
// maps them.
