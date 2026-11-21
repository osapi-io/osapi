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

package route

// ListInput names the interface whose routes to read.
type ListInput struct {
	Name string `json:"name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
}

// Route is one route of an interface that osapi writes.
type Route struct {
	Name        string `json:"name"             validate:"required,interface_name"        doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"          telemetry:"attr"`
	Destination string `json:"destination"      validate:"required,cidr"                  doc:"Where the route goes, in CIDR notation, such as 10.0.0.0/8; not a default route" telemetry:"attr"`
	Via         string `json:"via"              validate:"required,ip"                    doc:"The gateway, an address of the destination's family, such as 192.168.1.1"`
	Metric      int    `json:"metric,omitempty" validate:"omitempty,min=0,max=4294967295" doc:"The metric, from 0 to 4294967295; none when it is 0"`
}

// CreateInput is a route that osapi does not have yet.
type CreateInput Route

// UpdateInput is a new gateway or metric of a route that osapi has.
type UpdateInput Route

// DeleteInput names a route that osapi wrote.
type DeleteInput struct {
	Name        string `json:"name"        validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary" telemetry:"attr"`
	Destination string `json:"destination" validate:"required,cidr"           doc:"The destination of the route, in CIDR notation, such as 10.0.0.0/8"     telemetry:"attr"`
}

// Entry is a route of the host's main routing table.
type Entry struct {
	Destination string `json:"destination" doc:"The destination in CIDR notation, or default"`
	Gateway     string `json:"gateway"     doc:"The gateway; empty for a route on the link"`
	Interface   string `json:"interface"   doc:"The interface the route goes out of"`
	Family      string `json:"family"      doc:"inet or inet6"`
	Metric      int    `json:"metric"      doc:"The metric; 0 when the kernel gives none"`
	Proto       string `json:"proto"       doc:"Where the route came from, such as kernel, dhcp or static; empty on macOS"`
	Scope       string `json:"scope"       doc:"The scope, such as link; empty for a global route"`
}

// Result is a route after a write or a removal.
type Result struct {
	Destination string `json:"destination" doc:"The destination of the route"`
	Changed     bool   `json:"changed"     doc:"Whether the host changed"`
}
