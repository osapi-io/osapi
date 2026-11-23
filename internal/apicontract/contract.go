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

// Package apicontract reduces an OpenAPI document to its contract: one line
// for each tag, operation and schema property, sorted. Two documents with the
// same lines accept the same requests and give the same responses, whatever
// their prose, their order or where a parameter is declared. The tests that
// hold a generated spec to a hand-written one, and the full spec to its
// golden copy, compare these lines.
package apicontract

import (
	"fmt"
	"sort"
	"strings"
)

// Lines is the contract of doc, an OpenAPI document read into maps, such as
// by yaml.Unmarshal.
func Lines(
	doc map[string]any,
) []string {
	var out []string
	for _, t := range list(doc["tags"]) {
		out = append(out, "tag "+str(obj(t)["name"]))
	}
	for path, item := range obj(doc["paths"]) {
		for verb, raw := range obj(item) {
			op := obj(raw)
			if op["responses"] == nil {
				continue
			}
			out = append(out, operation(doc, strings.ToUpper(verb), path, op))
		}
	}
	for name, raw := range obj(obj(doc["components"])["schemas"]) {
		// A schema that is only a $ref to a schema of its own name, in
		// another file, declares nothing: a spec puts it there so its
		// responses can say #/components/schemas/<name>, and the file it
		// refers to defines it.
		if refName(str(obj(raw)["$ref"])) == name {
			continue
		}
		out = append(out, schema(name, obj(raw))...)
	}
	sort.Strings(out)

	return out
}

// operation is one line: the verb, the path, the operationId, the tags, the
// auth, each parameter, the request body and each response.
func operation(
	doc map[string]any,
	verb string,
	path string,
	op map[string]any,
) string {
	tags := make([]string, 0, len(list(op["tags"])))
	for _, t := range list(op["tags"]) {
		tags = append(tags, str(t))
	}
	sort.Strings(tags)

	params := make([]string, 0, len(list(op["parameters"])))
	for _, raw := range list(op["parameters"]) {
		params = append(params, parameter(doc, obj(raw)))
	}
	sort.Strings(params)

	body := ""
	if rb := obj(op["requestBody"]); rb != nil {
		body = content(obj(rb["content"]))
		if rb["required"] == true {
			body += "!"
		}
	}

	responses := make([]string, 0, len(obj(op["responses"])))
	for code, raw := range obj(op["responses"]) {
		responses = append(responses, code+"="+content(obj(obj(raw)["content"])))
	}
	sort.Strings(responses)

	return fmt.Sprintf("op %s %s id=%s tags=%v auth=%s params=%v body=%s responses=%v",
		verb, path, str(op["operationId"]), tags, auth(op), params, body, responses)
}

// auth is the scopes an operation needs, or public when it has none.
func auth(
	op map[string]any,
) string {
	var scopes []string
	for _, req := range list(op["security"]) {
		for scheme, raw := range obj(req) {
			for _, sc := range list(raw) {
				scopes = append(scopes, scheme+":"+str(sc))
			}
		}
	}
	if len(scopes) == 0 {
		return "public"
	}
	sort.Strings(scopes)

	return strings.Join(scopes, ",")
}

// parameter is a parameter's name, place, whether it is required and its
// schema, with a $ref to components/parameters resolved, so a parameter
// declared in place and one declared in components read the same.
func parameter(
	doc map[string]any,
	p map[string]any,
) string {
	if r := str(p["$ref"]); r != "" {
		resolved := obj(obj(obj(doc["components"])["parameters"])[refName(r)])
		if resolved == nil {
			return "ref:" + refName(r)
		}
		p = resolved
	}
	required := ""
	if p["required"] == true {
		required = "!"
	}

	return fmt.Sprintf("%s:%s%s:%s", str(p["in"]), str(p["name"]), required, sig(obj(p["schema"])))
}

// content is each media type of a body with its schema.
func content(
	c map[string]any,
) string {
	if len(c) == 0 {
		return "-"
	}
	var parts []string
	for media, raw := range c {
		parts = append(parts, media+":"+sig(obj(obj(raw)["schema"])))
	}
	sort.Strings(parts)

	return strings.Join(parts, "|")
}

// schema is one line for each property of a component schema, or one line
// for a schema that is a $ref or not an object.
func schema(
	name string,
	s map[string]any,
) []string {
	props := obj(s["properties"])
	if str(s["$ref"]) != "" || len(props) == 0 {
		return []string{"schema " + name + " = " + sig(s)}
	}
	required := map[string]bool{}
	for _, r := range list(s["required"]) {
		required[str(r)] = true
	}
	out := make([]string, 0, len(props))
	for prop, raw := range props {
		mark := ""
		if required[prop] {
			mark = "!"
		}
		out = append(out, fmt.Sprintf("schema %s.%s%s = %s", name, prop, mark, sig(obj(raw))))
	}

	return out
}

// sig is a schema in short form: a $ref by its name; an array as [items]; a
// map as map[values], or map[any] when its values are free; an object in place as {name:sig, ...}, with ! on what
// is required; anything else as its type, its format after a slash, and its
// enum in parentheses.
func sig(
	s map[string]any,
) string {
	if r := str(s["$ref"]); r != "" {
		return refName(r)
	}
	switch str(s["type"]) {
	case "array":
		return "[" + sig(obj(s["items"])) + "]"
	case "object":
		if ap := obj(s["additionalProperties"]); ap != nil {
			return "map[" + sig(ap) + "]"
		}
		if s["additionalProperties"] == true {
			return "map[any]"
		}
		required := map[string]bool{}
		for _, r := range list(s["required"]) {
			required[str(r)] = true
		}
		var fields []string
		for prop, raw := range obj(s["properties"]) {
			mark := ""
			if required[prop] {
				mark = "!"
			}
			fields = append(fields, prop+mark+":"+sig(obj(raw)))
		}
		sort.Strings(fields)

		return "{" + strings.Join(fields, ",") + "}"
	}
	out := str(s["type"])
	if f := str(s["format"]); f != "" {
		out += "/" + f
	}
	if enum := list(s["enum"]); len(enum) > 0 {
		values := make([]string, 0, len(enum))
		for _, v := range enum {
			values = append(values, fmt.Sprint(v))
		}
		out += "(" + strings.Join(values, "|") + ")"
	}

	return out
}

// refName is the last part of a $ref, the name it points at, whichever file
// it is in.
func refName(
	ref string,
) string {
	return ref[strings.LastIndex(ref, "/")+1:]
}

func obj(
	v any,
) map[string]any {
	m, _ := v.(map[string]any)

	return m
}

func list(
	v any,
) []any {
	l, _ := v.([]any)

	return l
}

func str(
	v any,
) string {
	s, _ := v.(string)

	return s
}
