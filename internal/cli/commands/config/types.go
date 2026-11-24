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

package config

import osapiconfig "github.com/osapi-io/osapi/internal/config"

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each setting, and the file it was read from.
type ListOutput struct {
	_     struct{}  `title:"Settings"`
	File  string    `                 json:"file"`
	Items []Setting `                 json:"items"`
}

// Setting is one key with its value in effect, and where it came from.
type Setting struct {
	Name     string `json:"name"               label:"Key"`
	Value    any    `json:"value"`
	Source   string `json:"source"`
	Env      string `json:"env"`
	Flag     string `json:"flag,omitempty"                         table:"-"`
	Type     string `json:"type"                                   table:"-"`
	Default  any    `json:"default,omitempty"                      table:"-"`
	Validate string `json:"validate,omitempty"                     table:"-"`
	Doc      string `json:"doc,omitempty"      label:"Description" table:"-"`
	Behavior string `json:"behavior,omitempty"                     table:"-"`
	Secret   bool   `json:"secret,omitempty"                                 render:"-"`
}

// GetInput names the setting to show.
type GetInput struct {
	Key string `json:"key" validate:"required" doc:"The setting to show, such as controller.api.port"`
}

// GetOutput is one setting in full, and the file it was read from.
type GetOutput struct {
	_       struct{} `title:"Config"`
	File    string   `               json:"file"`
	Setting Setting  `               json:"setting"`
}

// ValidateInput is the input of validate, which has none.
type ValidateInput struct{}

// ValidateOutput is whether the config is valid, each problem, and the fix
// of each rule that a problem breaks.
type ValidateOutput struct {
	_        struct{}              `title:"Config"`
	File     string                `               json:"file"`
	Valid    bool                  `               json:"valid"`
	Problems []osapiconfig.Problem `               json:"problems"`
	// Fixes are the fix of each rule a problem breaks, by rule.
	Fixes map[string]any `json:"fixes,omitempty"`
}
