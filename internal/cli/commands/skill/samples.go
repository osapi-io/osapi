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

package skill

import "github.com/osapi-io/osapi/internal/skills"

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	sampleList = ListOutput{Items: []skills.Skill{
		{
			Name:        "build-a-behavior",
			Description: "Build or change an osapi dsl behavior.",
			References:  []string{"references/services.md"},
		},
		{
			Name:        "osapi",
			Description: "Operate a fleet of hosts through osapi.",
			Init:        true,
			References:  []string{"references/behaviors.md", "references/targets.md"},
		},
		{
			Name:        "osapi-operations",
			Description: "Run and configure the components of osapi.",
			Init:        true,
			References:  []string{"references/config.md", "references/tokens.md"},
		},
	}}

	sampleGet = GetOutput{
		Name: "build-a-behavior",
		Content: "---\nname: build-a-behavior\n" +
			"description: Build or change an osapi dsl behavior.\n---\n\n" +
			"# Build a behavior\n\n" +
			"A behavior is the typed definition under internal/behaviors/<name>/.\n",
	}
)

// sampleInstall is the sample output of the install example, which the docs page of the command
// shows as the CLI prints it.
var sampleInstall = InstallOutput{
	Version: "v1.4.0",
	Tools:   []string{"claude"},
	Changed: 5,
	Files: []skills.Result{
		{Path: ".claude/skills/osapi/SKILL.md", Status: skills.Created},
		{Path: ".claude/skills/osapi/references/behaviors.md", Status: skills.Created},
		{Path: ".claude/skills/osapi-operations/SKILL.md", Status: skills.Created},
		{Path: ".claude/skills/osapi-operations/references/config.md", Status: skills.Created},
		{Path: "CLAUDE.md", Status: skills.Updated},
	},
}
