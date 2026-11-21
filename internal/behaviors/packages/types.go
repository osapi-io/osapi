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

package packages

// ListInput takes nothing: it lists every installed package.
type ListInput struct{}

// GetInput names one installed package.
type GetInput struct {
	Name string `json:"name" validate:"required,max=255,package_name" doc:"The package, such as curl"`
}

// CreateInput is the package to install.
type CreateInput struct {
	Name    string `json:"name"    validate:"required,max=255,package_name"     doc:"The package to install, such as curl"                                   telemetry:"attr"`
	Version string `json:"version,omitempty" validate:"omitempty,max=255,package_version" doc:"The version to install, such as 8.5.0-2ubuntu10; the newest when empty"`
}

// DeleteInput is the package to remove.
type DeleteInput struct {
	Name string `json:"name" validate:"required,max=255,package_name" doc:"The package to remove, such as curl"`
}

// Installed is one installed package.
type Installed struct {
	Name         string `json:"name"                   doc:"The package"`
	Version      string `json:"version"                doc:"The installed version"`
	Architecture string `json:"architecture,omitempty" doc:"The architecture it is built for, such as amd64"`
	Manager      string `json:"manager"                doc:"The package database that has it: dpkg, rpm or brew"`
	Description  string `json:"description,omitempty"  doc:"Its one-line summary"`
	Size         int64  `json:"size,omitempty"         doc:"Its installed size in bytes"                         render:"bytes"`
}

// Result is the package after an install or a removal.
type Result struct {
	Name    string `json:"name"    doc:"The package"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
