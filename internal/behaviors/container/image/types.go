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

package image

// ListInput is list's input; list takes none.
type ListInput struct{}

// PullInput is pull's input.
type PullInput struct {
	Image string `json:"image" validate:"required,image_ref" doc:"The image, such as alpine:3.20" telemetry:"attr"`
}

// DeleteInput is delete's input.
type DeleteInput struct {
	Image string `json:"image"           validate:"required,image_ref" doc:"The image, by its reference or its ID"`
	Force bool   `json:"force,omitempty" validate:"omitempty"          doc:"Remove it also when a container uses it"`
}

// Entry is one image, as list answers it.
type Entry struct {
	ID         string `json:"id"         doc:"The image's ID"`
	Repository string `json:"repository" doc:"Its repository, such as alpine"`
	Tag        string `json:"tag"        doc:"Its tag, such as 3.20"`
	Size       string `json:"size"       doc:"Its size, as docker prints it, such as 7.8MB"`
}

// PullResult is pull's output.
type PullResult struct {
	Image   string `json:"image"    doc:"The image"`
	ImageID string `json:"image_id" doc:"The ID of the image that the host has now"`
	Size    int64  `json:"size"     doc:"Its size in bytes"`
	Changed bool   `json:"changed"  doc:"Whether the pull got a new image: false when the host had it already"`
}

// Result is delete's output.
type Result struct {
	Image   string `json:"image"   doc:"The image"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
