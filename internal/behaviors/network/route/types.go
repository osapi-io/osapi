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

// ListInput takes nothing: list reads every route.
type ListInput struct{}

// GetInput names the interface whose routes to read.
type GetInput struct {
	InterfaceName string `json:"interface_name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
}

// Config is the routes osapi writes for an interface.
type Config struct {
	InterfaceName string   `json:"interface_name" validate:"required,interface_name"          doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"                                                                                   telemetry:"attr"`
	Routes        []string `json:"routes"         validate:"required,min=1,max=64,dive,route" doc:"Each route as ip route writes it: <destination> via <gateway>, with metric <n> after it when you give one, such as 10.0.0.0/8 via 192.168.1.1 metric 100"`
}

// CreateInput is the routes of an interface osapi has no routes of.
type CreateInput Config

// UpdateInput is the routes to write over the ones osapi has.
type UpdateInput Config

// DeleteInput names the interface whose routes osapi wrote.
type DeleteInput struct {
	InterfaceName string `json:"interface_name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
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

// Result is an interface's routes after a write or a removal.
type Result struct {
	InterfaceName string `json:"interface_name" doc:"The interface"`
	Changed       bool   `json:"changed"        doc:"Whether the host changed"`
}
