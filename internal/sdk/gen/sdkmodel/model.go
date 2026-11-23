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

// Package sdkmodel is what the Go SDK is made of, read from the full spec of
// the API: each service, its methods with their parameters, body and result,
// and each type the methods send and receive. internal/sdk/gen writes the
// SDK's typed services from it. A generator of the SDK's reference pages or
// of its examples reads the same model, so the code, the pages and the
// examples cannot name a method differently.
package sdkmodel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// MethodExtension is the extension of each operation of the spec that names
// the method in its group, such as get or drain. The endpoint layer writes
// it.
const MethodExtension = "x-osapi-method"

// ExamplesExtension is the extension of an operation of the spec that has
// the examples of its method, each with its doc and its input. The endpoint
// layer writes it.
const ExamplesExtension = "x-osapi-examples"

// AreaExtension is the extension of each operation of the spec that names
// its area, such as fleet. The endpoint layer writes it.
const AreaExtension = "x-osapi-area"

// AreasExtension is the extension of the spec that lists the areas in their
// order, each with its name and its label. The endpoint layer writes it.
const AreasExtension = "x-osapi-areas"

// Area is one area of the spec: the operations of its services are under
// it, such as in the folder of the SDK's pages.
type Area struct {
	// Name is the name of the area, such as fleet.
	Name string `json:"name"`
	// Label is what a reader sees, such as Fleet.
	Label string `json:"label"`
}

// Model is the SDK that the spec describes.
type Model struct {
	// Areas are the areas of the spec, in its order.
	Areas []*Area
	// Services are the services of the SDK, in the order of their names.
	Services []*Service
	// Types are the types that the methods send and receive, in the order
	// of their names. A collection is not one: it is a Collection of its
	// item type.
	Types []*Type
}

// Service is the methods of one group of the API, such as NetworkDNS.
type Service struct {
	// Name is the Go name of the service, such as NetworkDNS.
	Name string
	// Doc is the description of the group's tag.
	Doc string
	// Methods are the service's methods, in the order of their paths.
	Methods []*Method
	// Node is true for a group of a node, whose paths start with
	// /api/node/, and false for a group of the controller.
	Node bool
	// Area is the name of the area of the service's operations, one of the
	// Areas of the model.
	Area string
}

// Method is one operation of the spec.
type Method struct {
	// Name is the Go name of the method, such as Get.
	Name string
	// OperationID is the operationId of the operation.
	OperationID string
	// HTTPMethod is the method of the request, such as GET.
	HTTPMethod string
	// Path is the path of the operation, such as
	// /api/node/{hostname}/network/dns/{interface_name}.
	Path string
	// Summary is the summary of the operation.
	Summary string
	// Description is the description of the operation.
	Description string
	// Permission is the permission the operation needs, or "".
	Permission string
	// PathParams are the parameters in the path, in the order of the path.
	PathParams []*Param
	// Query are the parameters of the query, in the order of the spec.
	Query []*Param
	// Input is the one struct that a method with a JSON body takes: a field
	// for each parameter of the path but the target, then a field for each
	// property of the body. It is nil when the method has no JSON body.
	Input *Type
	// Multipart is true when the body is a multipart form, such as the
	// upload of a file. The generator writes no method for it: a person
	// writes it.
	Multipart bool
	// Result is the type of a success's body, or nil when a success has
	// none, such as a 204.
	Result *TypeRef
	// Statuses are the status codes that the spec declares, in order.
	Statuses []int
	// Examples are the examples of the method's definition, in their
	// order.
	Examples []*Example
}

// Example is one example of a method: a call with an input.
type Example struct {
	// Doc says what the example shows.
	Doc string `json:"doc"`
	// Input is the value of each parameter and each property of the body
	// that the example sets, by its name in the spec.
	Input map[string]any `json:"input"`
}

// Target is the parameter of the path that names the hosts, hostname, or
// nil when the path has none.
func (m *Method) Target() *Param {
	if len(m.PathParams) > 0 && m.PathParams[0].Name == "hostname" {
		return m.PathParams[0]
	}

	return nil
}

// Keys is the parameters of the path but the target, such as the name of a
// service.
func (m *Method) Keys() []*Param {
	if m.Target() != nil {
		return m.PathParams[1:]
	}

	return m.PathParams
}

// RequiredQuery is the parameters of the query that the request must have.
func (m *Method) RequiredQuery() []*Param {
	var out []*Param
	for _, p := range m.Query {
		if p.Required {
			out = append(out, p)
		}
	}

	return out
}

// OptionalQuery is the parameters of the query that the request can leave
// out.
func (m *Method) OptionalQuery() []*Param {
	var out []*Param
	for _, p := range m.Query {
		if !p.Required {
			out = append(out, p)
		}
	}

	return out
}

// Param is a parameter of an operation.
type Param struct {
	// Name is the name in the spec, such as interface_name.
	Name string
	// GoName is the Go name of a field, such as InterfaceName.
	GoName string
	// ArgName is the Go name of an argument, such as interfaceName.
	ArgName string
	// Doc is the description of the parameter.
	Doc string
	// Type is the type of the parameter: a string, an int or a bool.
	Type *TypeRef
	// Required is true when the request must have it.
	Required bool
}

// Kind is the kind of a TypeRef.
type Kind int

// The kinds of a TypeRef.
const (
	// Scalar is a Go type of the standard library, such as string, int64
	// or time.Time, in Name.
	Scalar Kind = iota
	// Named is a type of the model, in Name.
	Named
	// Slice is a slice of Elem.
	Slice
	// Map is a map of strings to Elem.
	Map
	// Collection is a Collection of Elem: the {job_id, results} answer of a
	// node.
	Collection
	// Any is a value of any JSON type.
	Any
)

// TypeRef is the Go type of a value.
type TypeRef struct {
	Kind Kind
	// Name is the Go name of a Scalar or a Named type, such as string or
	// NetworkDnsEntry.
	Name string
	// Elem is the type of the items of a Slice, a Map or a Collection.
	Elem *TypeRef
}

// String is the Go spelling of the type, such as []string or
// Collection[NetworkDnsEntry].
func (t *TypeRef) String() string {
	switch t.Kind {
	case Slice:
		return "[]" + t.Elem.String()
	case Map:
		return "map[string]" + t.Elem.String()
	case Collection:
		return "Collection[" + t.Elem.String() + "]"
	case Any:
		return "any"
	default:
		return t.Name
	}
}

// Type is a struct of the SDK, from a schema of the spec.
type Type struct {
	// Name is the Go name, the name of the schema, such as NetworkDnsEntry.
	Name string
	// Doc is the description of the schema.
	Doc string
	// Fields are the properties of the schema, in the order of their names.
	Fields []*Field
}

// Field is a property of a schema.
type Field struct {
	// Name is the name in the spec, such as search_domains.
	Name string
	// GoName is the Go name, such as SearchDomains.
	GoName string
	// Doc is the description of the property.
	Doc string
	// Type is the Go type.
	Type *TypeRef
	// Required is true when the schema requires the property.
	Required bool
	// Enum are the values the property can have, or nil.
	Enum []string
	// Path is true for a field of an Input that goes in the path, not in
	// the body.
	Path bool
}

// New is the model of doc.
func New(
	doc *openapi3.T,
) (*Model, error) {
	b := &builder{doc: doc, types: map[string]*Type{}}
	areas, err := specAreas(doc)
	if err != nil {
		return nil, err
	}
	b.areas = areas
	services, err := b.services()
	if err != nil {
		return nil, err
	}
	m := &Model{Areas: areas, Services: services}
	for _, t := range b.types {
		m.Types = append(m.Types, t)
	}
	slices.SortFunc(m.Types, func(a, b *Type) int { return strings.Compare(a.Name, b.Name) })

	return m, nil
}

// specAreas is the list of areas of doc, from its AreasExtension.
func specAreas(
	doc *openapi3.T,
) ([]*Area, error) {
	var out []*Area
	data, err := json.Marshal(doc.Extensions[AreasExtension])
	if err == nil {
		err = json.Unmarshal(data, &out)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", AreasExtension, err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("the spec has no %s", AreasExtension)
	}

	return out, nil
}

// builder makes the model, and each type a method meets, once.
type builder struct {
	doc   *openapi3.T
	types map[string]*Type
	areas []*Area
}

// area is the area of op, from its AreaExtension: one of the areas of the
// spec.
func (b *builder) area(
	op *openapi3.Operation,
) (string, error) {
	name, _ := op.Extensions[AreaExtension].(string)
	for _, a := range b.areas {
		if a.Name == name {
			return name, nil
		}
	}

	return "", fmt.Errorf(
		"the %s %q is not one of the spec's %s",
		AreaExtension,
		name,
		AreasExtension,
	)
}

// verbs is the order of the methods of one path.
var verbs = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

func (b *builder) services() ([]*Service, error) {
	byTag := map[string]*Service{}
	order := make([]*Service, 0, len(b.doc.Tags))
	for _, tag := range b.doc.Tags {
		s := &Service{Doc: tag.Description}
		byTag[tag.Name] = s
		order = append(order, s)
	}
	paths := b.doc.Paths.InMatchingOrder()
	slices.Sort(paths)
	group := map[*Service]string{}
	node := map[*Service]bool{}
	for _, path := range paths {
		item := b.doc.Paths.Value(path)
		for _, verb := range verbs {
			op := item.GetOperation(verb)
			if op == nil {
				continue
			}
			if len(op.Tags) != 1 || byTag[op.Tags[0]] == nil {
				return nil, fmt.Errorf("%s: the operation has no tag of the spec", op.OperationID)
			}
			s := byTag[op.Tags[0]]
			name, err := serviceName(b.doc.Tags.Get(op.Tags[0]), path)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op.OperationID, err)
			}
			if group[s] != "" && group[s] != name {
				return nil, fmt.Errorf(
					"%s: the paths of tag %s name the service %s and %s",
					op.OperationID, op.Tags[0], group[s], name,
				)
			}
			group[s] = name
			node[s] = strings.HasPrefix(path, "/api/node/")
			a, err := b.area(op)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op.OperationID, err)
			}
			if s.Area != "" && s.Area != a {
				return nil, fmt.Errorf(
					"%s: the operations of tag %s are in the areas %s and %s",
					op.OperationID, op.Tags[0], s.Area, a,
				)
			}
			s.Area = a
			m, err := b.method(verb, path, op)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op.OperationID, err)
			}
			s.Methods = append(s.Methods, m)
		}
	}

	var out []*Service
	names := map[string][]*Service{}
	for _, s := range order {
		if len(s.Methods) == 0 {
			continue
		}
		s.Name = group[s]
		s.Node = node[s]
		names[s.Name] = append(names[s.Name], s)
		out = append(out, s)
	}
	// A node's group that has the name of a group of the controller, such
	// as file, is the node's: NodeFile.
	for name, same := range names {
		if len(same) < 2 {
			continue
		}
		for _, s := range same {
			if node[s] {
				s.Name = "Node" + name
			}
		}
	}
	seen := map[string]bool{}
	for _, s := range out {
		for _, m := range s.Methods {
			if m.Input != nil {
				m.Input.Name = s.Name + m.Name + "Input"
			}
		}
		if seen[s.Name] {
			return nil, fmt.Errorf("two services have the name %s", s.Name)
		}
		seen[s.Name] = true
		methods := map[string]bool{}
		for _, m := range s.Methods {
			if methods[m.Name] {
				return nil, fmt.Errorf("%s has two methods with the name %s", s.Name, m.Name)
			}
			methods[m.Name] = true
		}
	}
	slices.SortFunc(out, func(a, b *Service) int { return strings.Compare(a.Name, b.Name) })

	return out, nil
}

// serviceName is the Go name of the service of a tag: the words of the path
// up to the group's own, which the last part of the tag's display name
// names, such as NetworkDNS for /api/node/{hostname}/network/dns.
func serviceName(
	tag *openapi3.Tag,
	path string,
) (string, error) {
	display, _ := tag.Extensions["x-displayName"].(string)
	parts := strings.Split(display, "/")
	group := strings.ToLower(parts[len(parts)-1])
	if group == "" {
		return "", fmt.Errorf("tag %s has no x-displayName", tag.Name)
	}
	var words []string
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		if seg == "api" || seg == "node" || strings.HasPrefix(seg, "{") {
			continue
		}
		words = append(words, seg)
		if seg == group {
			return GoName(strings.Join(words, "_")), nil
		}
	}

	return "", fmt.Errorf("path %s has no segment %s, the group of tag %s", path, group, tag.Name)
}

func (b *builder) method(
	verb string,
	path string,
	op *openapi3.Operation,
) (*Method, error) {
	name, _ := op.Extensions[MethodExtension].(string)
	if name == "" {
		return nil, fmt.Errorf("the operation has no %s", MethodExtension)
	}
	m := &Method{
		Name:        GoName(name),
		OperationID: op.OperationID,
		HTTPMethod:  verb,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
	}
	if raw, ok := op.Extensions[ExamplesExtension]; ok {
		data, err := json.Marshal(raw)
		if err == nil {
			err = json.Unmarshal(data, &m.Examples)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ExamplesExtension, err)
		}
	}
	if op.Security != nil {
		for _, req := range *op.Security {
			for _, scopes := range req {
				if len(scopes) > 0 {
					m.Permission = scopes[0]
				}
			}
		}
	}

	byName := map[string]*Param{}
	for _, ref := range op.Parameters {
		p := ref.Value
		t, err := b.ref(p.Schema, "")
		if err != nil {
			return nil, fmt.Errorf("parameter %s: %w", p.Name, err)
		}
		if t.Kind != Scalar {
			return nil, fmt.Errorf(
				"parameter %s is a %s, not a string, a number or a bool",
				p.Name,
				t,
			)
		}
		param := &Param{
			Name:     p.Name,
			GoName:   GoName(p.Name),
			ArgName:  argName(p.Name),
			Doc:      p.Description,
			Type:     t,
			Required: p.Required,
		}
		switch p.In {
		case openapi3.ParameterInPath:
			byName[p.Name] = param
		case openapi3.ParameterInQuery:
			m.Query = append(m.Query, param)
		default:
			return nil, fmt.Errorf(
				"parameter %s is in %s, not in the path or the query",
				p.Name,
				p.In,
			)
		}
	}
	for _, seg := range strings.Split(path, "/") {
		if !strings.HasPrefix(seg, "{") {
			continue
		}
		p := byName[strings.Trim(seg, "{}")]
		if p == nil {
			return nil, fmt.Errorf("path %s has %s, which no parameter is", path, seg)
		}
		m.PathParams = append(m.PathParams, p)
	}

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		content := op.RequestBody.Value.Content
		switch {
		case content.Get("application/json") != nil:
			if err := b.input(m, content.Get("application/json").Schema); err != nil {
				return nil, fmt.Errorf("body: %w", err)
			}
		case content.Get("multipart/form-data") != nil:
			m.Multipart = true
		default:
			return nil, fmt.Errorf("the body is not JSON and not a multipart form")
		}
	}

	for _, code := range sortedCodes(op.Responses) {
		m.Statuses = append(m.Statuses, code)
		r := op.Responses.Status(code).Value
		var t *TypeRef
		if mt := r.Content.Get("application/json"); mt != nil {
			var err error
			if t, err = b.ref(mt.Schema, ""); err != nil {
				return nil, fmt.Errorf("response %d: %w", code, err)
			}
		}
		if code < 200 || code > 299 {
			continue
		}
		if m.Result != nil && (t == nil || t.String() != m.Result.String()) {
			return nil, fmt.Errorf("its successes have different bodies")
		}
		m.Result = t
	}

	return m, nil
}

// input makes m's Input from the schema of its body. Its name is set when
// the service has its name.
func (b *builder) input(
	m *Method,
	body *openapi3.SchemaRef,
) error {
	if body.Ref == "" || body.Value == nil {
		return fmt.Errorf("the body is not a schema of the components")
	}
	in := &Type{Doc: body.Value.Description}
	seen := map[string]bool{}
	for _, p := range m.Keys() {
		in.Fields = append(in.Fields, &Field{
			Name:     p.Name,
			GoName:   p.GoName,
			Doc:      p.Doc,
			Type:     p.Type,
			Required: true,
			Path:     true,
		})
		seen[p.GoName] = true
	}
	fields, err := b.fields(strings.TrimPrefix(body.Ref, "#/components/schemas/"), body.Value)
	if err != nil {
		return err
	}
	for _, f := range fields {
		if seen[f.GoName] {
			return fmt.Errorf("the body and the path both have %s", f.Name)
		}
		in.Fields = append(in.Fields, f)
	}
	m.Input = in

	return nil
}

// sortedCodes is the status codes of responses, in order.
func sortedCodes(
	responses *openapi3.Responses,
) []int {
	var codes []int
	for key := range responses.Map() {
		if code, err := strconv.Atoi(key); err == nil {
			codes = append(codes, code)
		}
	}
	slices.Sort(codes)

	return codes
}

// ref is the Go type of s. A schema of the components is a Named type, or a
// Collection when it is the {job_id, results} answer of a node. An object
// with properties inside a schema is a Named type too, with the name
// inline.
func (b *builder) ref(
	s *openapi3.SchemaRef,
	inline string,
) (*TypeRef, error) {
	if s == nil || s.Value == nil {
		return &TypeRef{Kind: Any}, nil
	}
	if s.Ref != "" {
		name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		if item := collectionItem(s.Value); item != nil {
			elem, err := b.ref(item, "")
			if err != nil {
				return nil, err
			}

			return &TypeRef{Kind: Collection, Elem: elem}, nil
		}
		if err := b.named(name, s.Value); err != nil {
			return nil, err
		}

		return &TypeRef{Kind: Named, Name: name}, nil
	}

	v := s.Value
	switch {
	case v.Type == nil || len(v.Type.Slice()) == 0:
		return &TypeRef{Kind: Any}, nil
	case v.Type.Is(openapi3.TypeString):
		switch v.Format {
		case "date-time":
			return &TypeRef{Kind: Scalar, Name: "time.Time"}, nil
		default:
			return &TypeRef{Kind: Scalar, Name: "string"}, nil
		}
	case v.Type.Is(openapi3.TypeInteger):
		if v.Format == "int64" {
			return &TypeRef{Kind: Scalar, Name: "int64"}, nil
		}

		return &TypeRef{Kind: Scalar, Name: "int"}, nil
	case v.Type.Is(openapi3.TypeNumber):
		return &TypeRef{Kind: Scalar, Name: "float64"}, nil
	case v.Type.Is(openapi3.TypeBoolean):
		return &TypeRef{Kind: Scalar, Name: "bool"}, nil
	case v.Type.Is(openapi3.TypeArray):
		elem, err := b.ref(v.Items, inline)
		if err != nil {
			return nil, err
		}

		return &TypeRef{Kind: Slice, Elem: elem}, nil
	case v.Type.Is(openapi3.TypeObject):
		if len(v.Properties) > 0 {
			if inline == "" {
				return nil, fmt.Errorf("an object with properties has no name")
			}
			if err := b.named(inline, v); err != nil {
				return nil, err
			}

			return &TypeRef{Kind: Named, Name: inline}, nil
		}
		ap := v.AdditionalProperties
		if ap.Schema != nil {
			elem, err := b.ref(ap.Schema, inline)
			if err != nil {
				return nil, err
			}

			return &TypeRef{Kind: Map, Elem: elem}, nil
		}

		return &TypeRef{Kind: Map, Elem: &TypeRef{Kind: Any}}, nil
	}

	return nil, fmt.Errorf("a schema of type %v", v.Type.Slice())
}

// collectionItem is the schema of the items of s when s is the {job_id,
// results} answer of a node, or nil.
func collectionItem(
	s *openapi3.Schema,
) *openapi3.SchemaRef {
	if len(s.Properties) != 2 || s.Properties["job_id"] == nil {
		return nil
	}
	results := s.Properties["results"]
	if results == nil || results.Value == nil || results.Value.Type == nil ||
		!results.Value.Type.Is(openapi3.TypeArray) || results.Value.Items == nil ||
		results.Value.Items.Ref == "" {
		return nil
	}

	return results.Value.Items
}

// named adds the type name of s, once.
func (b *builder) named(
	name string,
	s *openapi3.Schema,
) error {
	if _, ok := b.types[name]; ok {
		return nil
	}
	t := &Type{Name: name, Doc: s.Description}
	b.types[name] = t
	fields, err := b.fields(name, s)
	if err != nil {
		return err
	}
	t.Fields = fields

	return nil
}

// fields is a field for each property of s, the schema name, in the order
// of their names.
func (b *builder) fields(
	name string,
	s *openapi3.Schema,
) ([]*Field, error) {
	var out []*Field
	props := make([]string, 0, len(s.Properties))
	for p := range s.Properties {
		props = append(props, p)
	}
	slices.Sort(props)
	for _, p := range props {
		goName := GoName(p)
		ft, err := b.ref(s.Properties[p], name+goName)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", name, p, err)
		}
		f := &Field{
			Name:     p,
			GoName:   goName,
			Doc:      s.Properties[p].Value.Description,
			Type:     ft,
			Required: slices.Contains(s.Required, p),
		}
		for _, e := range s.Properties[p].Value.Enum {
			f.Enum = append(f.Enum, fmt.Sprint(e))
		}
		out = append(out, f)
	}

	return out, nil
}
