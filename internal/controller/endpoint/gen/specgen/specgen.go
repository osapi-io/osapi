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

// Package specgen writes the spec of an endpoint group as the repository
// keeps it, for the endpoint generator and the behavior generator alike: the
// same encoding, header and reference to the shared ErrorResponse.
package specgen

import (
	"bytes"
	"encoding/json"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// componentsSpec is the spec, from the repository's root, that each API
// document shares ErrorResponse from.
const componentsSpec = "internal/controller/api/components/gen/api.yaml"

// ErrorRef is the reference to the shared ErrorResponse from a spec in dir,
// a directory from the repository's root.
func ErrorRef(
	dir string,
) string {
	rel, _ := filepath.Rel(dir, componentsSpec)

	return filepath.ToSlash(rel) + "#/components/schemas/ErrorResponse"
}

// Encode is doc as YAML under the comment header, such as "Code generated
// by ... DO NOT EDIT.". The document goes through JSON first, so each value,
// an extension included, is written as the spec's JSON form has it, and the
// keys are in order.
func Encode(
	doc *openapi3.T,
	header string,
) ([]byte, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString("# " + header + "\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(tree); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
