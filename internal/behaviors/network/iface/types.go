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

package iface

// ListInput takes nothing: list reads every interface.
type ListInput struct{}

// GetInput names the interface to read.
type GetInput struct {
	Name string `json:"name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
}

// Config is the config osapi writes for an interface. dhcp4 and dhcp6 are
// always written; another empty field is left out of the file. Type dummy
// writes the interface under netplan's dummy-devices, and netplan makes it.
type Config struct {
	Name       string   `json:"name"                  validate:"required,interface_name"        doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"                                  telemetry:"attr"`
	Type       string   `json:"type,omitempty"        validate:"omitempty,oneof=ethernet dummy" doc:"ethernet, the default, for an interface the host has, or dummy for a dummy interface that netplan makes"`
	DHCP4      bool     `json:"dhcp4,omitempty"       validate:"omitempty"                      doc:"Take an IPv4 address from DHCP"`
	DHCP6      bool     `json:"dhcp6,omitempty"       validate:"omitempty"                      doc:"Take an IPv6 address from DHCP"`
	Addresses  []string `json:"addresses,omitempty"   validate:"omitempty,max=16,dive,cidr"     doc:"The static addresses, in CIDR notation, such as 10.0.0.5/24"`
	Gateway4   string   `json:"gateway4,omitempty"    validate:"omitempty,ipv4"                 doc:"The IPv4 default gateway"`
	Gateway6   string   `json:"gateway6,omitempty"    validate:"omitempty,ipv6"                 doc:"The IPv6 default gateway"`
	MTU        int      `json:"mtu,omitempty"         validate:"omitempty,min=68,max=9000"      doc:"The MTU, from 68 to 9000"`
	MACAddress string   `json:"mac_address,omitempty" validate:"omitempty,mac_address"          doc:"The MAC address to set"`
	WakeOnLAN  bool     `json:"wakeonlan,omitempty"   validate:"omitempty"                      doc:"Turn on wake-on-LAN"`
}

// CreateInput is the config of an interface osapi has no config of.
type CreateInput Config

// UpdateInput is the config to write over the one osapi has.
type UpdateInput Config

// DeleteInput names the interface whose config osapi wrote.
type DeleteInput struct {
	Name string `json:"name" validate:"required,interface_name" doc:"The interface, such as eth0, or a fact such as @fact.interface.primary"`
}

// Entry is an interface as the host reports it.
type Entry struct {
	Name      string   `json:"name"      doc:"The interface"`
	State     string   `json:"state"     doc:"up or down"`
	MAC       string   `json:"mac"       doc:"Its MAC address; empty for one without, such as lo"`
	MTU       int      `json:"mtu"       doc:"Its MTU"`
	IPv4      string   `json:"ipv4"      doc:"Its first IPv4 address that is not link-local"`
	IPv6      string   `json:"ipv6"      doc:"Its first IPv6 address that is not link-local"`
	Addresses []string `json:"addresses" doc:"Each of its addresses, in CIDR notation"`
	DHCP4     bool     `json:"dhcp4"     doc:"Whether DHCP gave it an IPv4 address: a route of it has proto dhcp"`
	Primary   bool     `json:"primary"   doc:"Whether it has the default route"`
}

// Result is an interface after a write or a removal.
type Result struct {
	Name    string `json:"name"    doc:"The interface"`
	Changed bool   `json:"changed" doc:"Whether the host changed"`
}
