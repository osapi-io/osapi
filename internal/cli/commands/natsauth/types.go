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

package natsauth

// CreateInput is the input of a create.
type CreateInput struct {
	Out   string `json:"out"   validate:"required"  doc:"The directory that the files go to. It is made, mode 0700, when it is not there"`
	Print bool   `json:"print" validate:"omitempty" doc:"Also print the server's part: its authorization block, or its resolver file"`
}

// Written is a file that a create wrote.
type Written struct {
	Role string `json:"role"`
	Path string `json:"path"`
}

// CreateOutput is what a create wrote. It never holds a secret, except
// Server with --print, which the operator asked for.
type CreateOutput struct {
	_      struct{}  `title:"NATS auth"`
	Files  []Written `                  json:"files"`
	Server string    `                  json:"server,omitempty" label:"Server config"`
}
