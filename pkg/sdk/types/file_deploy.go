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

package types

import "slices"

// ContentType is how a deploy treats a file: as bytes, or as a template.
type ContentType string

// The content types of a file.
const (
	// ContentRaw writes the bytes as they are.
	ContentRaw ContentType = "raw"
	// ContentTemplate renders the file with Go text/template and the agent's
	// facts before it writes it.
	ContentTemplate ContentType = "template"
)

// Values are each content type.
func (ContentType) Values() []ContentType {
	return []ContentType{ContentRaw, ContentTemplate}
}

// Doc is the description of ContentType in the spec of each group that uses
// it.
func (ContentType) Doc() string {
	return "How a deploy treats a file: raw writes its bytes as they are; " +
		"template renders it with Go text/template and the agent's facts."
}

// Valid reports whether c is one of the content types.
func (c ContentType) Valid() bool {
	return slices.Contains(c.Values(), c)
}
