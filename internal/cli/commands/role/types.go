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

package role

// ListInput is the input of roles list, which has none.
type ListInput struct{}

// Role is one role, with what it grants.
type Role struct {
	_ struct{} `title:"Role"`
	// Name is the role's name, such as admin.
	Name string `json:"name"`
	// Source is built-in, or config for a custom role of the config.
	Source string `json:"source"`
	// Permissions are what the role grants. roles get shows them; a table
	// of roles leaves them out, because they are many.
	Permissions []string `json:"permissions" table:"-"`
}

// GetInput names the role to read.
type GetInput struct {
	Name string `json:"name" validate:"required" doc:"The role, such as admin, or a custom role of the config."`
}

// ListOutput is each role: the built-in ones, then those of the config.
type ListOutput struct {
	_     struct{} `title:"Roles"`
	Roles []Role   `              json:"roles"`
}
