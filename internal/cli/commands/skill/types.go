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

// Package token is the token group of local commands: osapi controller token generate,
// roles list and validate, which sign and read the bearer tokens of the API
// with the signing key of the config. They run in the CLI, and call no API.
package skill

import "github.com/osapi-io/osapi/internal/skills"

// ListInput is the input of list, which has none.
type ListInput struct{}

// ListOutput is each skill this binary has.
type ListOutput struct {
	_     struct{}       `title:"Skills"`
	Items []skills.Skill `               json:"items"`
}

// GetInput names the skill, and one of its reference files.
type GetInput struct {
	Name      string `json:"name"      validate:"required"  doc:"Name of the skill"`
	Reference string `json:"reference" validate:"omitempty" doc:"A reference file of the skill, relative to its directory"`
}

// GetOutput is a skill's file: its text, which prints as it is.
type GetOutput struct {
	Name      string `json:"name"`
	Reference string `json:"reference,omitempty"`
	Content   string `json:"content"             render:"body"`
}

// InstallInput is where to install the skills, and for which tools.
type InstallInput struct {
	Path  string   `json:"path" validate:"omitempty"                         doc:"The repository to install into; the current directory without it"         default:"."`
	Tools []string `json:"tool" validate:"omitempty,dive,oneof=claude codex" doc:"Agent tool to install for: claude, codex (repeatable); claude without it"`
}

// InstallOutput is each file install wrote, and how many changed.
type InstallOutput struct {
	_       struct{}        `title:"Skills"`
	Version string          `               json:"version"`
	Tools   []string        `               json:"tools"`
	Changed int             `               json:"changed"`
	Files   []skills.Result `               json:"files"`
}
