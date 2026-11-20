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

package commanddsl

import (
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/guide"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each command of the binary, and how many each layer made.
type ListOutput struct {
	_           struct{}  `title:"Commands"`
	Items       []Summary `                 json:"items"`
	Command     int       `                 json:"command"`
	Endpoint    int       `                 json:"endpoint"`
	Behavior    int       `                 json:"behavior"`
	HandWritten int       `                 json:"hand_written"`
}

// Summary is one command, and the layer and group that made it.
type Summary struct {
	Command string `json:"command"`
	Layer   string `json:"layer"`
	Group   string `json:"group,omitempty" table:"-"`
	Doc     string `json:"doc"                       label:"Description"`
}

// GetInput names the group.
type GetInput struct {
	Name string `json:"name" validate:"required" doc:"The command path of the group, such as token or agent"`
}

// GetOutput is a group and each of its commands.
type GetOutput struct {
	_        struct{}  `title:"Group"`
	Name     string    `              json:"name"`
	Doc      string    `              json:"doc"      label:"Description"`
	Commands []Command `              json:"commands"                     table:"sections"`
}

// Command is one command of a group: its flags and refusals, for a local
// group's.
type Command struct {
	Command string   `json:"command"`
	Method  string   `json:"method"`
	Short   string   `json:"short"           label:"Description"`
	Fails   []string `json:"fails,omitempty" label:"Refuses"`
	Flags   []Flag   `json:"flags,omitempty"`
}

// Flag is one flag of a local command, from a field of its input.
type Flag struct {
	Flag     string `json:"flag"`
	Short    string `json:"short,omitempty"`
	Type     string `json:"type"`
	Validate string `json:"validate"`
	Doc      string `json:"doc"             label:"Help"`
}

// CreateInput is the name and the doc of the group to create.
type CreateInput struct {
	Name        string `json:"name"        validate:"required" doc:"The group's name: its command under osapi"`
	Description string `json:"description" validate:"required" doc:"What the group is, in one line"`
}

// CreateOutput is the files that create wrote, and the steps after it.
type CreateOutput struct {
	_     struct{} `title:"Local group"`
	Name  string   `                    json:"name"`
	Files []File   `                    json:"files"`
	Next  []Step   `                    json:"next"`
}

// File is a file that create wrote.
type File struct {
	Path string `json:"path"`
}

// Step is a command to run after create.
type Step struct {
	Command string `json:"command"`
}

// SpecInput is the input of spec, which has none.
type SpecInput struct{}

// SpecOutput is each kind of term of the DSL, with its terms.
type SpecOutput struct {
	Kinds []guide.Section `json:"kinds" table:"sections"`
}

// ValidateInput is the input of validate, which has none.
type ValidateInput struct{}

// ValidateOutput is each problem of each local group, and of the output of
// each group, and the fix of each rule that one breaks.
type ValidateOutput struct {
	_        struct{}          `title:"Validation"`
	Local    int               `                   json:"local"              label:"Local groups"`
	Groups   int               `                   json:"groups"`
	Valid    bool              `                   json:"valid"`
	Problems []command.Problem `                   json:"problems,omitempty"`
	Fixes    map[string]any    `                   json:"fixes,omitempty"`
}
