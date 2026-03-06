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

package file

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateContext is the data available to Go templates during rendering.
type TemplateContext struct {
	// Facts contains agent facts (architecture, kernel, etc.).
	Facts map[string]any
	// Vars contains user-supplied template variables.
	Vars map[string]any
	// Hostname is the agent's hostname.
	Hostname string
}

// renderTemplate parses rawTemplate as a Go text/template and executes it
// with the provider's cached facts, the supplied vars, and the hostname.
func (p *FileProvider) renderTemplate(
	rawTemplate []byte,
	vars map[string]any,
) ([]byte, error) {
	tmpl, err := template.New("file").Parse(string(rawTemplate))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	ctx := TemplateContext{
		Facts:    p.cachedFacts,
		Vars:     vars,
		Hostname: p.hostname,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
