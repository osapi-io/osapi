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

// Package specgen writes the specs as the repository keeps them, for the
// endpoint generator and the behavior generator alike: the spec of each
// group, and the full spec of the API, which joins them.
package specgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// The info of the full spec: osapi's own title and tagline, not those of a
// group.
const (
	title       = "osapi"
	description = `Typed, repeatable operations for Linux hosts. Give your agent a fleet, not
a shell.

osapi lets an AI agent manage Linux hosts through typed operations, not
through a login. Each operation is a behavior: one thing on a host,
defined one time, with a fixed input and output. osapi checks each
request before it queues it. A second run changes nothing. Each change is
recorded.
`
	version    = "1.0.0"
	contactURL = "https://github.com/osapi-io/osapi"
	license    = "MIT License"
	licenseURL = "https://opensource.org/licenses/MIT"
)

// Encode is doc as YAML under the comment header, such as "Code generated
// by ... DO NOT EDIT.". The document goes through JSON first, so each value,
// an extension included, is written as the spec's JSON form has it, and the
// keys are in order.
func Encode(
	doc *openapi3.T,
	header string,
) ([]byte, error) {
	tree, err := toTree(doc)
	if err != nil {
		return nil, err
	}

	return write(tree, header)
}

// Full is the spec of the API, from the spec of each group in docs, under
// the comment header. Each group keeps its part in the order of docs: its
// tag, its paths and its schemas. A tag gets the title of its group in front,
// with _ for each space, such as Agent_Management_API_agent_operations, and
// x-tagGroups puts it under that title. x-osapi-areas lists the areas, with
// the label and in the order of area.All. The UI's client files take their
// names from the tags. The shared ErrorResponse schema is first. Two groups
// that serve one path, or that give one schema or parameter two shapes, are
// an error.
func Full(
	docs []*openapi3.T,
	header string,
) ([]byte, error) {
	shared := endpoint.ErrorSchema()
	errorResponse, err := toTree(shared)
	if err != nil {
		return nil, err
	}
	var (
		tags, groups []any
		paths        = &part{what: "path"}
		schemas      = &part{what: "schema"}
		security     = &part{what: "security scheme"}
		params       = &part{what: "parameter"}
	)
	// The first entry has nothing to conflict with.
	_ = schemas.put("ErrorResponse", errorResponse, ordered(shared), "the shared ErrorResponse")
	for _, doc := range docs {
		tree, err := toTree(doc)
		if err != nil {
			return nil, err
		}
		root := tree.(map[string]any)
		group := doc.Info.Title
		prefix := strings.ReplaceAll(group, " ", "_") + "_"

		renamed := map[string]string{}
		var names []any
		for _, t := range list(root["tags"]) {
			tag := t.(map[string]any)
			name := prefix + tag["name"].(string)
			renamed[tag["name"].(string)] = name
			tag["name"] = name
			tags = append(tags, tag)
			names = append(names, name)
		}
		groups = append(groups, map[string]any{"name": group, "tags": names})

		for _, item := range object(root["paths"]) {
			for _, op := range object(item) {
				if op, ok := op.(map[string]any); ok {
					for i, t := range list(op["tags"]) {
						if n, ok := renamed[t.(string)]; ok {
							op["tags"].([]any)[i] = n
						}
					}
				}
			}
		}

		components := object(root["components"])
		for _, p := range []struct {
			to  *part
			add any
		}{
			{paths, root["paths"]},
			{schemas, components["schemas"]},
			{security, components["securitySchemes"]},
			{params, components["parameters"]},
		} {
			if err := p.to.merge(p.add, group); err != nil {
				return nil, err
			}
		}
	}

	return write(mapping(
		"openapi", "3.0.0",
		"info", mapping(
			"title", title,
			"description", description,
			"version", version,
			"contact", mapping("url", contactURL),
			"license", mapping("name", license, "url", licenseURL),
		),
		"tags", tags,
		"paths", paths.node,
		"components", mapping(
			"schemas", schemas.node,
			"securitySchemes", security.node,
			"parameters", params.node,
		),
		"x-tagGroups", groups,
		endpoint.AreasExtension, areas(),
	), header)
}

// areas is area.All as the spec lists it: each area's name and label, in
// order.
func areas() *yaml.Node {
	out := &yaml.Node{Kind: yaml.SequenceNode}
	for _, a := range area.All {
		out.Content = append(out.Content, mapping("name", string(a.Name), "label", a.Label))
	}

	return out
}

// part is one map of the full spec, such as its paths, with the groups'
// entries in the order they were added.
type part struct {
	what  string
	node  *yaml.Node
	from  map[string]string
	value map[string]any
}

// merge adds each entry of m, a map, in the order that the spec of its
// group has the keys. An entry that another group added with the same value
// is there once; with a different value, it is an error.
func (p *part) merge(
	m any,
	group string,
) error {
	entries := object(m)
	n := node(entries)
	for i := 0; i+1 < len(n.Content); i += 2 {
		k := n.Content[i].Value
		if err := p.add(k, entries[k], group); err != nil {
			return err
		}
	}

	return nil
}

// add adds the entry key, from group.
func (p *part) add(
	key string,
	value any,
	group string,
) error {
	return p.put(key, value, node(value), group)
}

// put adds the entry key, from group, written as n.
func (p *part) put(
	key string,
	value any,
	n *yaml.Node,
	group string,
) error {
	if p.node == nil {
		p.node = &yaml.Node{Kind: yaml.MappingNode}
		p.from = map[string]string{}
		p.value = map[string]any{}
	}
	if from, ok := p.from[key]; ok {
		// A path is served once; a schema or a parameter that two groups
		// share has one shape.
		if p.what == "path" || !reflect.DeepEqual(p.value[key], value) {
			return fmt.Errorf("%s %s: %s and %s both define it", p.what, key, from, group)
		}

		return nil
	}
	p.from[key], p.value[key] = group, value
	p.node.Content = append(p.node.Content, node(key), n)

	return nil
}

// ordered is the ErrorResponse schema s with its properties in the order
// that a person reads them: the message, its details, and the reason.
func ordered(
	s *openapi3.Schema,
) *yaml.Node {
	property := func(name string) *yaml.Node {
		p := s.Properties[name].Value
		kv := []any{"type", "string", "description", p.Description}
		if len(p.Enum) > 0 {
			kv = append(kv, "enum", p.Enum)
		}

		return mapping(append(kv, "example", p.Example)...)
	}

	return mapping(
		"type", "object",
		"properties", mapping(
			"error", property("error"),
			"details", property("details"),
			"reason", property("reason"),
		),
	)
}

// mapping is a mapping node of the keys and values, in their order. A
// part that no group added to is left out.
func mapping(
	kv ...any,
) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode}
	for i := 0; i < len(kv); i += 2 {
		if v, ok := kv[i+1].(*yaml.Node); ok && v == nil {
			continue
		}
		n.Content = append(n.Content, node(kv[i]), node(kv[i+1]))
	}

	return n
}

// node is v as a YAML node: a node as it is, any other value encoded, with
// the keys of a map in order.
func node(
	v any,
) *yaml.Node {
	if n, ok := v.(*yaml.Node); ok {
		return n
	}
	n := &yaml.Node{}
	// A value from JSON, or a string, always encodes.
	_ = n.Encode(v)

	return n
}

// object is v as a map, or an empty one when v is not a map.
func object(
	v any,
) map[string]any {
	m, _ := v.(map[string]any)

	return m
}

// list is v as a list, or nil when v is not a list.
func list(
	v any,
) []any {
	l, _ := v.([]any)

	return l
}

// toTree is v as the tree of maps, lists and values that its JSON is.
func toTree(
	v any,
) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// A number stays as its JSON has it, so 134217728 is not written as
	// 1.34217728e+08.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var tree any
	if err := dec.Decode(&tree); err != nil {
		return nil, err
	}

	return numbers(tree), nil
}

// numbers is v with each json.Number as an int64 when it is a whole
// number, and else as a float64.
func numbers(
	v any,
) any {
	switch v := v.(type) {
	case map[string]any:
		for k, e := range v {
			v[k] = numbers(e)
		}
	case []any:
		for i, e := range v {
			v[i] = numbers(e)
		}
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return n
		}
		f, _ := v.Float64()

		return f
	}

	return v
}

// write is v as YAML under the comment header.
func write(
	v any,
	header string,
) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("# " + header + "\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
