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

package updatepackages

// ListInput takes nothing: it lists every update the index has.
type ListInput struct{}

// UpdateInput takes nothing: it refreshes the whole index.
type UpdateInput struct{}

// Available is one package that the index has a newer version of.
type Available struct {
	Name           string `json:"name"            doc:"The package"`
	CurrentVersion string `json:"current_version" doc:"The installed version"`
	NewVersion     string `json:"new_version"     doc:"The version the index has"`
}

// Result is the index after a refresh.
type Result struct {
	Changed bool `json:"changed" doc:"Whether the refresh got new index data"`
}
