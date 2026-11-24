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

package behaviordsl

import (
	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/guide"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each behavior.
type ListOutput struct {
	_     struct{}  `title:"Behaviors"`
	Items []Summary `                  json:"items"`
}

// Summary is one behavior and the names of its methods.
type Summary struct {
	Name     string    `json:"name"`
	Group    string    `json:"group,omitempty"`
	Of       string    `json:"of,omitempty"`
	Area     area.Name `json:"area,omitempty"`
	Methods  []string  `json:"methods"`
	Doc      string    `json:"doc"                label:"Description"`
	Internal bool      `json:"internal,omitempty"`
}

// GetInput names the behavior.
type GetInput struct {
	Name string `json:"name" validate:"required" doc:"The behavior, from osapi dsl behavior list"`
}

// GetOutput is a behavior: where osapi serves it, where it runs, what it
// is built on, its methods, types and settings, and what its calls record.
type GetOutput struct {
	_          struct{}           `title:"Behavior"`
	Name       string             `                 json:"name"`
	Doc        string             `                 json:"doc"             label:"Description"`
	Placement  behavior.Placement `                 json:"placement"`
	Group      string             `                 json:"group,omitempty"`
	Of         string             `                 json:"of,omitempty"`
	Area       area.Name          `                 json:"area,omitempty"`
	Internal   bool               `                 json:"internal"`
	API        bool               `                 json:"api"             label:"API"`
	CLI        bool               `                 json:"cli"             label:"CLI"`
	Next       bool               `                 json:"next"`
	Platforms  []string           `                 json:"platforms"                           none:"none"`
	Sudo       []string           `                 json:"sudo"                                none:"none"`
	Runs       []string           `                 json:"runs"                                none:"none"`
	Writes     []string           `                 json:"writes"                              none:"none"`
	KeepsState bool               `                 json:"keeps_state"`
	Uses       []string           `                 json:"uses"                                none:"none"`
	UsedBy     []string           `                 json:"used_by"         label:"Used by"     none:"none"`
	Methods    []Method           `                 json:"methods"`
	Types      []Type             `                 json:"types"`
	Settings   []Setting          `                 json:"settings"                                        empty:"Declare them with behavior.Config."`
	Telemetry  Telemetry          `                 json:"telemetry"`
}

// Method is one method of a behavior: where it is served and run.
type Method struct {
	Name       string        `json:"name"`
	Verb       behavior.Verb `json:"verb"`
	HTTPMethod string        `json:"http_method,omitempty" label:"HTTP"`
	Route      string        `json:"route,omitempty"                           none:"-"`
	Permission string        `json:"permission,omitempty"                      none:"-"`
	Command    string        `json:"command,omitempty"                                  table:"-"`
	Doc        string        `json:"doc"                   label:"Description"`
}

// Type is the input and the output type of one method, with their fields:
// CreateInput { key*, value* }, a required input field marked *.
type Type struct {
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Setting is one of a behavior's config keys, and the agent start flag
// that sets it.
type Setting struct {
	Key     string `json:"key"`
	Flag    string `json:"flag"    label:"Agent flag" prefix:"--"`
	Type    string `json:"type"`
	Default string `json:"default"`
}

// Telemetry is what a behavior's calls record: the span and metrics every
// method gets, and the input fields each method records on the span.
type Telemetry struct {
	Span       string              `json:"span"`
	ChildSpans []string            `json:"child_spans" label:"Child spans"     none:"none"`
	Metrics    []string            `json:"metrics"`
	Attributes []string            `json:"attributes"`
	Recorded   string              `json:"recorded"    label:"Recorded inputs" none:"none"`
	Inputs     map[string][]string `json:"inputs"                                          render:"-"`
}

// MethodInput names the behavior and its method.
// OpenAPIOutput is a served behavior's part of the OpenAPI spec, as YAML,
// which prints as it is.
type OpenAPIOutput struct {
	Name string `json:"name"`
	YAML string `json:"yaml" render:"body"`
}

// MethodInput is the behavior and the method to show.
type MethodInput struct {
	Name   string `json:"name"   validate:"required" doc:"The behavior, from osapi dsl behavior list"`
	Method string `json:"method" validate:"required" doc:"The method, from osapi dsl behavior get"`
}

// MethodOutput is one method in full: where it is served, its fields with
// their constraints, its output and its examples. --json has the JSON
// Schema of its input and output.
type MethodOutput struct {
	_          struct{}           `title:"Method"`
	Name       string             `               json:"name"`
	Verb       behavior.Verb      `               json:"verb"`
	Doc        string             `               json:"doc"                   label:"Description"`
	Long       string             `               json:"long,omitempty"        label:"Detail"`
	HTTPMethod string             `               json:"http_method,omitempty" label:"HTTP"`
	Route      string             `               json:"route,omitempty"                           none:"other behaviors only; it is not served"`
	Permission string             `               json:"permission,omitempty"`
	Command    string             `               json:"command,omitempty"`
	InputType  string             `               json:"input_type"            label:"Input type"`
	OutputType string             `               json:"output_type"           label:"Output type"`
	Fields     []Field            `               json:"fields"                label:"Input"                                                     empty:"The method takes no input; its command needs only --target."`
	Results    []Result           `               json:"results"               label:"Output"`
	Examples   []ExampleRow       `               json:"example_rows"          label:"Examples"`
	Input      *behavior.Schema   `               json:"input"                                                                                                                                                       render:"-"`
	Output     *behavior.Schema   `               json:"output"                                                                                                                                                      render:"-"`
	Calls      []behavior.Example `               json:"examples,omitempty"                                                                                                                                          render:"-"`
	Refusals   []behavior.Refusal `               json:"refusals,omitempty"                                                                                                                                          render:"-"`
}

// Field is one input field: its flag and request field.
type Field struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Constraints string `json:"constraints"`
	Description string `json:"description"`
}

// Result is one field of the output.
type Result struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ExampleRow is one example of a method: its input and output, as fields.
type ExampleRow struct {
	Example string `json:"example"`
	Input   string `json:"input"`
	Output  string `json:"output"`
}

// CreateInput is the behavior to create, and where.
type CreateInput struct {
	Name        string   `json:"name"        validate:"required"  doc:"The behavior's name: lowercase letters and digits"`
	Description string   `json:"description" validate:"required"  doc:"What it manages, in one line"`
	Group       string   `json:"group"       validate:"omitempty" doc:"Nest it under a group, such as network: its routes, commands and directory"`
	Package     string   `json:"package"     validate:"omitempty" doc:"The Go package and directory, when it is not the name: package is a Go keyword, so --name package --package packages"`
	Requires    []string `json:"requires"    validate:"omitempty" doc:"Behaviors it is built on, such as managedfile: it writes the Requires, the host.Require and a provider field of each"`
	Root        string   `json:"root"        validate:"omitempty" doc:"The osapi checkout to write into"                                                                                     default:"."`
}

// CreateOutput is the files that create wrote, and the steps after it.
type CreateOutput struct {
	_     struct{} `title:"Behavior"`
	Name  string   `                 json:"name"`
	Files []File   `                 json:"files"`
	Next  []Step   `                 json:"next"`
}

// File is a file that create wrote.
type File struct {
	Path string `json:"path"`
}

// Step is a step after create.
type Step struct {
	Step string `json:"step"`
}

// SpecInput is the input of spec get, which has none.
type SpecInput struct{}

// SpecOutput is each kind of term of the DSL, with its terms.
type SpecOutput struct {
	Kinds []guide.Section `json:"kinds" table:"sections"`
}

// ValidateInput is the input of validate, which has none.
type ValidateInput struct{}

// ValidateOutput is each problem of each behavior, and the fix of each rule
// that one breaks.
type ValidateOutput struct {
	_         struct{}           `title:"Validation"`
	Behaviors int                `                   json:"behaviors"`
	Valid     bool               `                   json:"valid"`
	Problems  []behavior.Problem `                   json:"problems,omitempty"`
	Fixes     map[string]any     `                   json:"fixes,omitempty"`
}
