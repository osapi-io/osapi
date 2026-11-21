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

package dns

// GetInput names the interface whose resolvers to read.
type GetInput struct {
	InterfaceName string `json:"interface_name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
}

// UpdateInput is the resolvers to set on an interface.
type UpdateInput struct {
	InterfaceName string   `json:"interface_name"           validate:"required,interface_name"        doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
	Servers       []string `json:"servers,omitempty"        validate:"omitempty,max=16,dive,ip"       doc:"The DNS servers, IPv4 or IPv6; the interface's own when empty"`
	SearchDomains []string `json:"search_domains,omitempty" validate:"omitempty,max=16,dive,hostname" doc:"The search domains; the interface's own when empty"`
	OverrideDHCP  bool     `json:"override_dhcp,omitempty"  validate:"omitempty"                      doc:"Use only these servers: drop the ones DHCP gives"`
}

// DeleteInput names the interface whose resolvers osapi set.
type DeleteInput struct {
	InterfaceName string `json:"interface_name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
}

// Entry is the resolvers of an interface.
type Entry struct {
	Servers       []string `json:"servers"        doc:"The DNS servers"`
	SearchDomains []string `json:"search_domains" doc:"The search domains"`
}

// Result is an interface after an update or a removal.
type Result struct {
	InterfaceName string `json:"interface_name" doc:"The interface"`
	Changed       bool   `json:"changed"        doc:"Whether the host changed"`
}
