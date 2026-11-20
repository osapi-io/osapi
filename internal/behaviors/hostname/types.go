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

package hostname

// GetInput is get's input, which has none.
type GetInput struct{}

// Entry is the host's name.
type Entry struct {
	Name string `json:"name" doc:"The hostname of the host"`
}

// UpdateInput is the name to set.
type UpdateInput struct {
	Name string `json:"name" validate:"required,max=253,hostname" doc:"The hostname to set, such as web-01" telemetry:"attr"`
}

// Result is the name after an update, and whether the host changed.
type Result struct {
	Name    string `json:"name"    doc:"The hostname of the host"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
