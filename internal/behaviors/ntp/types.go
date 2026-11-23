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

package ntp

// GetInput is get's input: get has none.
type GetInput struct{}

// CreateInput is create's input: the servers chrony uses.
type CreateInput struct {
	Servers []string `json:"servers" validate:"required,min=1,max=16,dive,ntp_server" doc:"The NTP servers, hostnames or IP addresses" telemetry:"attr"`
}

// UpdateInput is update's input: the servers that replace the ones osapi set.
type UpdateInput struct {
	Servers []string `json:"servers" validate:"required,min=1,max=16,dive,ntp_server" doc:"The NTP servers, hostnames or IP addresses" telemetry:"attr"`
}

// DeleteInput is delete's input: delete has none.
type DeleteInput struct{}

// Status is get's output: whether chrony is synchronized, and its sources.
type Status struct {
	Synchronized  bool     `json:"synchronized"             doc:"Whether chrony is synchronized to a source"`
	Stratum       int      `json:"stratum,omitempty"        doc:"The NTP stratum of the host"`
	Offset        string   `json:"offset,omitempty"         doc:"The offset of the system clock from NTP time, such as +0.000003422s"`
	CurrentSource string   `json:"current_source,omitempty" doc:"The source that chrony follows now"`
	Servers       []string `json:"servers,omitempty"        doc:"The sources that chrony has"`
}

// Result is the output of each write: whether it changed the host.
type Result struct {
	Changed bool `json:"changed" doc:"Whether the host changed"`
}
