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

package endpointdsl

import (
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/guide"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each endpoint group.
type ListOutput struct {
	_     struct{}  `title:"Endpoint groups"`
	Items []Summary `                        json:"items"`
}

// Summary is one endpoint group and the names of its methods.
type Summary struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Doc     string   `json:"doc"     label:"Description"`
}

// GetInput names the group.
type GetInput struct {
	Name string `json:"name" validate:"required" doc:"The endpoint group, such as agent"`
}

// GetOutput is a group and each of its methods.
type GetOutput struct {
	_       struct{} `title:"Endpoint group"`
	Name    string   `                       json:"name"`
	Doc     string   `                       json:"doc"     label:"Description"`
	Methods []Method `                       json:"methods"`
	Types   []Type   `                       json:"types"`
}

// Type is the input and the output type of one method, with their fields:
// a required input field is marked *.
type Type struct {
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Method is one method of a group: its route, permission and operation.
type Method struct {
	Name        string `json:"name"`
	Verb        string `json:"verb"`
	Path        string `json:"path"`
	Permission  string `json:"permission"   none:"none (public)"`
	OperationID string `json:"operation_id"                      label:"Operation ID"`
}

// OpenAPIOutput is a group's part of the OpenAPI spec, as YAML, which
// prints as it is.
type OpenAPIOutput struct {
	Name string `json:"name"`
	YAML string `json:"yaml" render:"body"`
}

// CreateInput is the name and the doc of the group to create.
type CreateInput struct {
	Name        string `json:"name"        validate:"required" doc:"The group's name: its path segment and CLI group"`
	Description string `json:"description" validate:"required" doc:"What the group is, in one line"`
}

// CreateOutput is the files that create wrote, and the steps after it.
type CreateOutput struct {
	_     struct{} `title:"Endpoint group"`
	Name  string   `                       json:"name"`
	Files []File   `                       json:"files"`
	Next  []Step   `                       json:"next"`
}

// File is a file that create wrote.
type File struct {
	Path string `json:"path"`
}

// Step is a command to run after create.
type Step struct {
	Command string `json:"command"`
}

// SpecInput is the input of spec get, which has none.
type SpecInput struct{}

// SpecOutput is each kind of term of the DSL, with its terms.
type SpecOutput struct {
	Kinds []guide.Section `json:"kinds" table:"sections"`
}

// ValidateInput is the input of validate, which has none.
type ValidateInput struct{}

// ValidateOutput is each problem of each endpoint group, and the fix of
// each rule that one breaks.
type ValidateOutput struct {
	_        struct{}           `title:"Validation"`
	Groups   int                `                   json:"groups"`
	Valid    bool               `                   json:"valid"`
	Problems []endpoint.Problem `                   json:"problems,omitempty"`
	Fixes    map[string]any     `                   json:"fixes,omitempty"`
}
