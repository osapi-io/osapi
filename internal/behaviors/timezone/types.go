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

package timezone

// GetInput takes nothing: the timezone is of the whole host.
type GetInput struct{}

// Entry is the timezone of the host.
type Entry struct {
	Timezone  string `json:"timezone"   doc:"The IANA name of the timezone, such as America/New_York"`
	UTCOffset string `json:"utc_offset" doc:"The offset from UTC now, such as -05:00"`
}

// UpdateInput is the timezone to set.
type UpdateInput struct {
	Timezone string `json:"timezone" validate:"required,max=64,timezone_name" doc:"The IANA name of the timezone, such as UTC or America/New_York" telemetry:"attr"`
}

// Result is the timezone after the update.
type Result struct {
	Timezone string `json:"timezone" doc:"The timezone of the host"`
	Changed  bool   `json:"changed"  doc:"Whether the host changed"`
}
